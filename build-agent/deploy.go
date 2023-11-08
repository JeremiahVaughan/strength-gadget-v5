package main

import (
	"context"
	"deploy/config"
	"deploy/constants"
	"deploy/model"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Deployment struct {
	Environment    string
	DesiredVersion string
	EnvVars        map[string]string
}

type Deploy struct {
	WantDeployedEnvironments    map[string]string `yaml:"environments"`
	AlreadyDeployedEnvironments map[string]string `yaml:"alreadyDeployedEnvironments"`
}

const currentlyDeployedParamKey = "currently_deployed"
const paramNotFoundError = "ParameterNotFound: "

func main() {
	// controlling config init through the main method so unit tests can still run without having to provide env vars.
	config.InitConfig()
	ctx := context.Background()

	log.Printf("Starting deployment jobs...")
	ssmClient, err := getSsmClientForTerraformState()
	if err != nil {
		log.Fatalf("error, when attempting to create an SSM client for Terraform State. Error: %v", err)
	}

	var masterRepoRoot string
	masterRepoRoot, err = os.Getwd()
	if err != nil {
		log.Fatalf("an unexpected error has occurred when attempting to retrieve the current working directory: %v", err)
	}

	err = os.Setenv(constants.TfVarAppName, constants.AppName)
	if err != nil {
		log.Fatalf("error, when attempting to set TF app name env var: %v", err)
	}
	err = setupEcr(masterRepoRoot)
	if err != nil {
		log.Fatalf("error, when attempting to setup ECR: %v", err)
	}

	var ecrUrl string
	ecrUrl, err = getEcrUrl()
	if err != nil {
		log.Fatalf("error, when attempting to fetch ECR url: %v", err)
	}

	var environmentsToDeploy map[string]string
	environmentsToDeploy, err = getEnvironmentsInNeedOfDeployment(ssmClient)
	if err != nil {
		log.Fatalf("error when attempting to fetch environments to deploy to. Error: %v", err)
	}

	// deploying one environment at a time todo see if you can make this happen in parallel. Not sure what to do about handling clones or checkouts
	for environment, deploymentVersion := range environmentsToDeploy {
		var deployment *Deployment
		deployment, err = getEnvironmentDeploymentDetails(ctx, environment, deploymentVersion, ssmClient)
		if err != nil {
			log.Fatalf("error, when attempting to fetch environmental deployment. Environment: %s. Error: %v", environment, err)
		}

		err = deployToEnvironment(*deployment, masterRepoRoot, ecrUrl)
		if err != nil {
			log.Fatalf("error when attempting to deploy to environment: %s. Error: %v", environment, err)
		}
	}

	err = recordSuccessfulDeployments(environmentsToDeploy, ssmClient)
	if err != nil {
		log.Fatalf("error has occured when attempting to record if deployments were successful: %v", err)
	}
}

func setupEcr(masterRepoRoot string) error {
	artifactInfraPath := fmt.Sprintf("%s/infrastructure/terraform/live/artifacts", masterRepoRoot)
	err := terraform(artifactInfraPath, "apply")
	if err != nil {
		return fmt.Errorf("error, when applying the terraform artifacts layer: %v", err)
	}
	return err
}

func getSsmClientForTerraformState() (*ssm.Client, error) {
	return getSsmClient(config.AppConfig.TerraformStateBucketKey, config.AppConfig.TerraformStateBucketSecret)
}

func getSsmClient(awsKey, awsSecret string) (*ssm.Client, error) {
	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				awsKey,
				awsSecret,
				"",
			),
		),
		awsConfig.WithRegion(config.AppConfig.TerraformStateBucketRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("error, when creating AWS config: %v", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)
	return ssmClient, err
}

func recordSuccessfulDeployments(deployed map[string]string, ssmClient *ssm.Client) error {
	deployedMarshalled, err := json.Marshal(deployed)
	if err != nil {
		return fmt.Errorf("error, when attempting to marshall what was deployed: %v", err)
	}

	currentlyDeployed := currentlyDeployedParamKey
	result := string(deployedMarshalled)
	overwrite := true
	input := &ssm.PutParameterInput{
		Name:      &currentlyDeployed,
		Type:      types.ParameterTypeString,
		Value:     &result,
		Overwrite: &overwrite,
	}
	_, err = ssmClient.PutParameter(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error happend when attempting to record hashes to aws param store: %v", err)
	}

	return nil
}

func deployToEnvironment(deployment Deployment, masterRepoRoot string, ecrUrl string) error {
	repoRoot := fmt.Sprintf("%s/%s", constants.RepoRoot, deployment.Environment)
	for key, value := range deployment.EnvVars {
		err := os.Setenv(key, value)
		if err != nil {
			return fmt.Errorf("error, when attempting to set environmental variable: %s. Error: %v", key, err)
		}
	}

	if deployment.DesiredVersion != "latest" {
		cmd := exec.Command("mkdir", "-p", repoRoot)
		var output []byte
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error, when executing command to create directory for cloning the repository to be deployed. Output: %s. Error: %v", string(output), err)
		}
		repoReadToken := os.Getenv("TF_VAR_git_clone_token")
		if repoReadToken == "" {
			return fmt.Errorf("error, missing TF_VAR_git_clone_token env var")
		}
		gitUserName := os.Getenv("TF_VAR_git_user_name")
		if gitUserName == "" {
			return fmt.Errorf("error, missing TF_VAR_git_user_name env var")
		}
		cmd = exec.Command(
			"git",
			"clone",
			"--branch",
			deployment.DesiredVersion,
			"--depth",
			"1",
			fmt.Sprintf("https://%s:%s@git.jetbrains.space/strengthgadget/strengthgadget/strength-gadget-v4.git", gitUserName, repoReadToken),
		)
		cmd.Dir = repoRoot
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error executing command. Output: %s. Error: %v", string(output), err)
		}
	} else {
		repoRoot = masterRepoRoot
	}

	log.Printf("Deploying any changed APIs...")
	ecsPath := fmt.Sprintf("%s/infrastructure/terraform/live/%s/ecs", repoRoot, deployment.Environment)
	err := generateApiArtifact(ecrUrl, repoRoot)
	if err != nil {
		return fmt.Errorf("error, when generating api artifact: %v", err)
	}
	err = os.Setenv(constants.TfVarBuildNumber, config.AppConfig.BuildNumber)
	if err != nil {
		return fmt.Errorf("error, when attempting to set aws build number env variable: %v", err)
	}
	err = terraform(ecsPath, "apply")
	if err != nil {
		return fmt.Errorf("error, when running terraform to apply for deploying the backend API. Error: %v", err)
	}

	cmd := exec.Command("npm", "i")
	uiDir := fmt.Sprintf("%s/ui", repoRoot)
	cmd.Dir = uiDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error, when attempting to install UI dependencies: %v", err)
	}
	log.Printf(string(output))

	err = updateAppVersionNumber(uiDir)
	if err != nil {
		return fmt.Errorf("error, when attempting to updateAppVersionNumber() for deployToEnvironment(). Error: %v", err)
	}
	cmd = exec.Command(
		"npx",
		"nx",
		"build",
		"--configuration=production",
	)
	cmd.Dir = uiDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error, when attempting to build ui files: %v", err)
	}
	log.Printf(string(output))

	cloudFrontPath := fmt.Sprintf("%s/infrastructure/terraform/live/%s/cloudfront", repoRoot, deployment.Environment)
	err = terraform(cloudFrontPath, "apply")
	if err != nil {
		return fmt.Errorf("error, when running terraform to apply for deploying the backend API. Error: %v", err)
	}

	backOffInterval := 3 * time.Second
	for i := 0; i < 120; i++ {
		err = ensureBackendDeployed()
		if err == nil {
			break
		}
		time.Sleep(backOffInterval)
	}
	if err != nil {
		return fmt.Errorf("error, when checking if backend deployed: %v", err)
	}

	log.Printf("Deployment to %s has completed successfully", deployment.Environment)
	return nil
}

func updateAppVersionNumber(uiDir string) error {
	// Open a new file for writing only
	file, err := os.OpenFile(
		fmt.Sprintf("%s/.env", uiDir),
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		return fmt.Errorf("error, failed creating file for updateAppVersionNumber(). Error: %v", err)
	}

	// data to write
	data := []byte(fmt.Sprintf("VITE_APP_VERSION=%s\n", config.AppConfig.BuildNumber))

	// write data to file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error, failed writing to file for updateAppVersionNumber(). Error: %v", err)
	}

	// Close the file
	err = file.Close()
	if err != nil {
		return fmt.Errorf("error, failed closing file for updateAppVersionNumber(). Error: %v", err)
	}
	return nil
}

func ensureBackendDeployed() (err error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.%s/api/health", os.Getenv("TF_VAR_domain_name")), nil)
	if err != nil {
		return fmt.Errorf("error, when generating get request: %v", err)
	}

	c := &http.Client{}
	response, err := c.Do(request)
	if err != nil || response.StatusCode != http.StatusOK {
		return fmt.Errorf("error, when performing get request. ERROR: %v. RESPONSE CODE: %d", err, response.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		errClose := Body.Close()
		if errClose != nil {
			err = fmt.Errorf("error, when closing response body: %v", errClose)
		}
	}(response.Body)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error, when reading response body: %v", err)
	}

	var healthResponse model.HealthResponse
	err = json.Unmarshal(responseBody, &healthResponse)
	if err != nil {
		return fmt.Errorf("error, when unmarshalling response body: %v", err)
	}

	deployedVersionNumber := getBuildNumber(healthResponse)
	if deployedVersionNumber != config.AppConfig.BuildNumber {
		return fmt.Errorf("error, build number does not match expected. EXPECTED: %s. GOT: %s", config.AppConfig.BuildNumber, deployedVersionNumber)
	}
	return nil
}

func getBuildNumber(healthResponse model.HealthResponse) string {
	return strings.Split(healthResponse.AppVersion, ".")[2]
}

func unmarshalIntoSyncMap(data []byte) (*sync.Map, error) {
	var result sync.Map
	var tmpMap map[string]interface{}
	err := json.Unmarshal(data, &tmpMap)
	if err != nil {
		return nil, fmt.Errorf("an unexpected error has occurred when attempting to unmarshall into temporay map: %v", err)
	}

	for k, v := range tmpMap {
		result.Store(k, v)
	}
	return &result, nil
}

func getEnvironmentsInNeedOfDeployment(ssmClient *ssm.Client) (map[string]string, error) {
	file, err := os.ReadFile("deployed.yml")
	if err != nil {
		return nil, fmt.Errorf("error has happend when attempting to open the deployed.yml file: %v", err)
	}
	var deployed Deploy
	err = yaml.Unmarshal(file, &deployed)
	if err != nil {
		return nil, fmt.Errorf("error has occurred when unmarshalling the deployed.yml file: %v", err)
	}

	param, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(fmt.Sprintf(currentlyDeployedParamKey)),
		WithDecryption: aws.Bool(true),
	})
	skipUnmarshal := false
	if err != nil {
		if err.Error() == paramNotFoundError {
			skipUnmarshal = true
		} else {
			return nil, fmt.Errorf("an unexpected error has occurred when attempting to fetch an aws param: %v", err)
		}
	}

	// todo see if you can move this into terraform
	if !skipUnmarshal {
		err = json.Unmarshal([]byte(*param.Parameter.Value), &deployed.AlreadyDeployedEnvironments)
		if err != nil {
			return nil, fmt.Errorf("error, when attempting to unmarshall what is currently deployed. Error: %v", err)
		}
	}

	needsDeployment := determineWhatNeedsDeploying(deployed.AlreadyDeployedEnvironments, deployed.WantDeployedEnvironments)

	return needsDeployment, nil
}

func determineWhatNeedsDeploying(alreadyDeployed, wantDeployed map[string]string) map[string]string {
	difference := make(map[string]string)

	for k, v := range wantDeployed {
		if wantDeployed[k] == "latest" {
			difference[k] = v
		} else if v2, ok := alreadyDeployed[k]; !ok || v2 != v {
			difference[k] = v
		}
	}

	return difference
}

func getEnvironmentDeploymentDetails(ctx context.Context, environmentName string, deploymentVersion string, ssmClient *ssm.Client) (*Deployment, error) {
	awsParamPath := fmt.Sprintf("/strengthgadget/%s/env_vars", environmentName)
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(awsParamPath),
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(true),
	}
	paginator := ssm.NewGetParametersByPathPaginator(ssmClient, input)
	envVars := make(map[string]string)
	for paginator.HasMorePages() {
		// Retrieve the page
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error, failed to get a page when fetching env vars from path: %s. Error: %v", awsParamPath, err)
		}

		// Loop through the parameters on the current page
		for _, p := range page.Parameters {
			_, envVarName := filepath.Split(*p.Name)
			envVars[envVarName] = *p.Value
		}
	}

	result := Deployment{
		Environment:    environmentName,
		DesiredVersion: deploymentVersion,
		EnvVars:        envVars,
	}
	return &result, nil
}

func getLowestDirectory(path string) string {
	dir := filepath.Dir(path)
	return filepath.Base(dir)
}

func confirmUniqueNameOfDeploymentDirectories(paths []string) error {
	seen := make(map[string]bool)
	for _, p := range paths {
		dir := getLowestDirectory(p)
		if seen[dir] {
			return errors.New("found duplicate directory: " + dir)
		}
		seen[dir] = true
	}
	return nil
}

func generateApiArtifact(ecrUrl string, repoRoot string) (err error) {
	// Initialize Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error, when creating new docker client: %v", err)
	}

	// Build Docker image
	version := fmt.Sprintf("0.0.%s", config.AppConfig.BuildNumber)
	versionTag := fmt.Sprintf("%s:%s", ecrUrl, version)
	backendCodeDir := fmt.Sprintf("%s/backend/strengthgadget", repoRoot)

	err = os.Setenv("AWS_ACCESS_KEY_ID", os.Getenv("TF_VAR_aws_access_key_id"))
	if err != nil {
		return fmt.Errorf("error, when setting AWS_ACCESS_KEY_ID to login to ECR with. Error: %v", err)
	}

	err = os.Setenv("AWS_SECRET_ACCESS_KEY", os.Getenv("TF_VAR_aws_secret_access_key"))
	if err != nil {
		return fmt.Errorf("error, when setting AWS_SECRET_ACCESS_KEY to login to ECR with. Error: %v", err)
	}

	// Get the login password from AWS ECR
	getLoginCmd := exec.Command("aws", "ecr", "get-login-password", "--region", os.Getenv("TF_VAR_infra_aws_region"))
	password, err := getLoginCmd.Output()
	if err != nil {
		return fmt.Errorf("Arrr! Failed to get login password: %v\n", err)
	}

	// Login to ECR
	log.Printf("logging into ecr...")
	loginCmd := exec.Command("docker", "login", "--username", "AWS", "--password-stdin", ecrUrl)
	loginCmd.Stdin = strings.NewReader(string(password))
	if err = loginCmd.Run(); err != nil {
		return fmt.Errorf("Arrr! Failed to login to ECR: %v\n", err)
	}

	log.Printf("setting docker driver...")
	setDockerDriverCmd := exec.Command("docker", "buildx", "create", "--use")
	setDockerDriverCmd.Dir = backendCodeDir
	output, err := setDockerDriverCmd.Output()
	if err != nil {
		return fmt.Errorf("error, when attempting to set the docker driver. Output: %s. Error: %v", output, err)
	}

	log.Printf("building and pushing image...")
	buildAndPushCmd := exec.Command(
		"docker",
		"buildx",
		"build",
		"--platform",
		"linux/arm64",
		"--push",
		"-t",
		versionTag,
		".",
	)
	buildAndPushCmd.Dir = backendCodeDir

	// Getting the stdout and stderr
	stdout, _ := buildAndPushCmd.StdoutPipe()
	stderr, _ := buildAndPushCmd.StderrPipe()
	if err = buildAndPushCmd.Start(); err != nil {
		return fmt.Errorf("error, failed to start cmd: %v", err)
	}

	// Reading the stdout and stderr
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	if err = buildAndPushCmd.Wait(); err != nil {
		return fmt.Errorf("error, command finished with error: %v", err)
	}

	var cfg aws.Config
	cfg, err = awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("TF_VAR_infra_aws_key_id"),
				os.Getenv("TF_VAR_infra_aws_secret"),
				"",
			),
		),
		awsConfig.WithRegion(os.Getenv("TF_VAR_infra_aws_region")),
	)
	svc := ecr.NewFromConfig(cfg)
	input := &ecr.GetAuthorizationTokenInput{}
	result, err := svc.GetAuthorizationToken(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error, when attempting to fetch an auth token: %v", err)
	}

	authData := result.AuthorizationData[0]
	decodedAuthToken, err := base64.StdEncoding.DecodeString(*authData.AuthorizationToken)
	if err != nil {
		return fmt.Errorf("error, when attempting to decode the authorization token used for ECR: %v", err)
	}

	decodedAuthParts := strings.SplitN(string(decodedAuthToken), ":", 2)
	if len(decodedAuthParts) != 2 {
		return fmt.Errorf("error, failed to parse decoded authorization token")
	}

	ecrAuthConfig := dockerTypes.AuthConfig{
		Username:      decodedAuthParts[0],
		Password:      decodedAuthParts[1],
		ServerAddress: *authData.ProxyEndpoint,
	}

	token, err := encodeECRAuthorizationToken(ecrAuthConfig)
	if err != nil {
		return fmt.Errorf("error, when attempting to encode ECR auth token: %v", err)
	}
	pushOptions := dockerTypes.ImagePushOptions{
		RegistryAuth: token,
	}

	// Push the Docker images to AWS ECR
	pushResponse, err := cli.ImagePush(ctx, versionTag, pushOptions)
	if err != nil {
		return fmt.Errorf("error, when attempting to push an image: %v", err)
	}
	defer func(pushResponse io.ReadCloser) {
		closeErr := pushResponse.Close()
		if closeErr != nil {
			err = fmt.Errorf("error, when attempting to close the push response: %v. original Error: %v", closeErr, err)
		}
	}(pushResponse)

	_, err = io.Copy(os.Stdout, pushResponse)
	if err != nil {
		return fmt.Errorf("error, when attempting to print the push response to standard out: %v", err)
	}
	return
}

func getEcrUrl() (string, error) {
	ssmClient, err := getSsmClient(os.Getenv("TF_VAR_infra_aws_key_id"), os.Getenv("TF_VAR_infra_aws_secret"))
	if err != nil {
		log.Fatalf("error, when attempting to create an SSM client for Terraform State. Error: %v", err)
	}
	awsParamPath := fmt.Sprintf("ecr-repo-url")
	param, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           aws.String(awsParamPath),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("an unexpected error has occurred when attempting to fetch aws params at Path: %s. Error: %v", awsParamPath, err)
	}
	return *param.Parameter.Value, nil
}

func encodeECRAuthorizationToken(authConfig dockerTypes.AuthConfig) (string, error) {
	authBytes, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("error, failed to marshal auth config: %v", err)
	}
	return base64.StdEncoding.EncodeToString(authBytes), nil
}
