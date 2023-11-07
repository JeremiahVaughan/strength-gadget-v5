package config

import (
	"deploy/constants"
	"deploy/model"
	"log"
	"os"
)

var AppConfig model.AppConfig

func InitConfig() {
	// Get environment variable values
	AppConfig.TerraformStateBucketKey = os.Getenv("TF_VAR_terraform_state_bucket_key")
	AppConfig.TerraformStateBucketSecret = os.Getenv("TF_VAR_terraform_state_bucket_secret")
	AppConfig.TerraformStateBucketRegion = os.Getenv("TF_VAR_terraform_state_bucket_region")
	AppConfig.BuildNumber = os.Getenv(constants.BuildNumber)

	// Check that required environment variables are set
	if AppConfig.TerraformStateBucketKey == "" {
		log.Fatalf("TF_VAR_terraform_state_bucket_key environment variable is not set")
	}
	if AppConfig.TerraformStateBucketSecret == "" {
		log.Fatalf("TF_VAR_terraform_state_bucket_secret environment variable is not set")
	}
	if AppConfig.TerraformStateBucketRegion == "" {
		log.Fatalf("TF_VAR_terraform_state_bucket_region environment variable is not set")
	}
	if AppConfig.BuildNumber == "" {
		log.Fatalf("BUILD_NUMBER environment variable is not set")
	}
}
