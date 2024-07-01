package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	Environment string

	SentryEndpoint string

	RegistrationEmailFrom         string
	RegistrationEmailFromPassword string

	VerificationCodeLength = 6
	//LockoutDurationInHours                           = 24
	AllowedVerificationAttemptsWithTheExcessiveRetryLockoutWindow int
	VerificationCodeValidityWindowInMin                           int
	EmailRootCa                                                   string
	TrustedUiOrigins                                              []string

	AllowedVerificationResendCodeAttemptsWithinOneHour int

	AllowedLoginAttemptsBeforeTriggeringLockout int

	VerificationExcessiveRetryAttemptLockoutDurationInSeconds                 int
	WindowLengthInSecondsForTheNumberOfAllowedVerificationEmailsBeforeLockout int
	WindowLengthInSecondsForTheNumberOfAllowedLoginAttemptsBeforeLockout      int

	// WorkoutSessionExpiration should the workout expire before they complete it, then
	// they will need to complete the same workout routine again. This 48 hours ensures
	// they have had enough rest to do so.
	WorkoutSessionExpiration = time.Duration(time.Hour * 48)

	NumberOfExerciseInSuperset = 3

	NumberOfSetsInSuperSet = 4

	// CurrentSupersetExpirationTimeInHours this addresses the edge case where the user doesn't finish the superset within a reasonable amount of time.
	// the superset is assumed to be aborted regardless of the progress made in that superset.
	CurrentSupersetExpirationTimeInHours = 6

	Version             string
	ConnectionPool      *pgxpool.Pool
	RedisConnectionPool *redis.Client

	HttpServer *http.Server

	AllowedIpRanges []*net.IPNet

	lowerWorkout AvailableWorkoutExercises
	coreWorkout  AvailableWorkoutExercises
	upperWorkout AvailableWorkoutExercises
)

func InitConfig(ctx context.Context) error {

	exerciseMap := generateExerciseMap()

	lowerWorkout = generateWorkoutExercises(exerciseMap, LOWER)
	coreWorkout = generateWorkoutExercises(exerciseMap, CORE)
	upperWorkout = generateWorkoutExercises(exerciseMap, UPPER)

	var err error
	var errorMsgs []string
	Environment = os.Getenv("TF_VAR_environment")
	if Environment == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_environment")
	}

	s := os.Getenv("TF_VAR_allowed_verification_resend_code_attempts_within_one_hour")
	AllowedVerificationResendCodeAttemptsWithinOneHour, err = strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("error, ensure the env var TF_VAR_allowed_verification_resend_code_attempts_within_one_hour has a value and is a number")
	}

	SentryEndpoint = os.Getenv("TF_VAR_sentry_end_point")
	if SentryEndpoint == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_sentry_end_point")
	}

	RegistrationEmailFrom = os.Getenv("TF_VAR_registration_email_from")
	if RegistrationEmailFrom == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_registration_email_from")
	}
	RegistrationEmailFromPassword = os.Getenv("TF_VAR_registration_email_from_password")
	if RegistrationEmailFromPassword == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_registration_email_from_password")
	}
	databaseConnectionString := os.Getenv("TF_VAR_database_connection_string")
	if databaseConnectionString == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_database_connection_string")
	}
	Version = os.Getenv("TF_VAR_version")
	if Version == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_version")
	}
	databaseRootCa := os.Getenv("TF_VAR_database_root_ca")
	if databaseRootCa == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_database_root_ca")
	}
	trustedUiOriginsString := os.Getenv("TF_VAR_trusted_ui_origin")
	if trustedUiOriginsString == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_trusted_ui_origin")
	} else {
		TrustedUiOrigins = strings.Split(trustedUiOriginsString, ",")
	}
	EmailRootCa = os.Getenv("TF_VAR_email_root_ca")
	if EmailRootCa == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_email_root_ca")
	}
	webServerCertKey := os.Getenv("TF_VAR_cloudflare_origin_cert_key")
	if webServerCertKey == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_cloudflare_origin_cert_key")
	}
	webServerCert := os.Getenv("TF_VAR_cloudflare_origin_cert")
	if webServerCert == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_cloudflare_origin_cert")
	}
	redisConnectionString := os.Getenv("TF_VAR_redis_connection_string")
	if redisConnectionString == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_redis_connection_string")
	}

	redisPort := os.Getenv("TF_VAR_redis_port")
	if redisPort == "" {
	}

	redisPassword := os.Getenv("TF_VAR_redis_password")
	if redisPassword == "" {
		errorMsgs = append(errorMsgs, "TF_VAR_redis_password")
	}

	toParse := os.Getenv("TF_VAR_verification_excessive_retry_attempt_lockout_duration_in_seconds")
	VerificationExcessiveRetryAttemptLockoutDurationInSeconds, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "TF_VAR_verification_excessive_retry_attempt_lockout_duration_in_seconds")
	}

	toParse = os.Getenv("TF_VAR_allowed_verification_attempts_with_the_excessive_retry_lockout_window")
	AllowedVerificationAttemptsWithTheExcessiveRetryLockoutWindow, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "TF_VAR_allowed_verification_attempts_with_the_excessive_retry_lockout_window")
	}

	toParse = os.Getenv("TF_VAR_window_length_in_seconds_for_the_number_of_allowed_verification_emails_before_lockout")
	WindowLengthInSecondsForTheNumberOfAllowedVerificationEmailsBeforeLockout, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "TF_VAR_window_length_in_seconds_for_the_number_of_allowed_verification_emails_before_lockout")
	}

	toParse = os.Getenv("TF_VAR_window_length_in_seconds_for_the_number_of_allowed_login_attempts_before_lockout")
	WindowLengthInSecondsForTheNumberOfAllowedLoginAttemptsBeforeLockout, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "TF_VAR_window_length_in_seconds_for_the_number_of_allowed_login_attempts_before_lockout")
	}

	toParse = os.Getenv("TF_VAR_allowed_login_attempts_before_triggering_lockout")
	AllowedLoginAttemptsBeforeTriggeringLockout, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "TF_VAR_allowed_login_attempts_before_triggering_lockout")
	}

	toParse = os.Getenv("TF_VAR_verification_code_validity_window_in_min")
	VerificationCodeValidityWindowInMin, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "TF_VAR_verification_code_validity_window_in_min")
	}

	if len(errorMsgs) != 0 {
		return fmt.Errorf("error, missing env vars. Vars: %s", strings.Join(errorMsgs, ", "))
	}

	AllowedIpRanges, err = initAllowedIpRanges()
	if err != nil {
		return fmt.Errorf("error, when initAllowedIpRanges() for InitConfig(). Error: %v", err)
	}

	RedisConnectionPool, err = connectToRedisDatabase(redisPort)
	if err != nil {
		return fmt.Errorf("error, when connectToRedisDatabase() for InitConfig(). Error: %v", err)
	}

	_, err = RedisConnectionPool.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("error, when attempting to ping the redis database after establishing the initial pooling connection: %v", err)
	}

	ConnectionPool, err = connectToDatabase(ctx, databaseConnectionString, databaseRootCa)
	if err != nil {
		return fmt.Errorf("error, when attempting to establish a connection pool with the database: %v", err)
	}

	err = ConnectionPool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("error, when attempting to ping databse after establishing the initial pooling connection: %v", err)
	}

	HttpServer, err = initHttpServer()
	if err != nil {
		return fmt.Errorf("error, when attempting to setup http server for configuration init. Error: %v", err)
	}
	return nil
}

func initAllowedIpRanges() ([]*net.IPNet, error) {
	var blocksSlice []string
	var err error
	if Environment == EnvironmentLocal {
		blocksSlice = []string{
			"127.0.0.1/32", // ipv4 loop-back
			"::1/128",      // ipv6 loop-back
			"10.0.0.8/32",  // Jeremiah's Iphone
			"10.0.0.24/32", // Jeremiah's Macbook
		}
	} else {
		blocksSlice, err = fetchAllowedIpRanges()
		if err != nil {
			return nil, fmt.Errorf("error, could not fetchAllowedIpRanges() for initAllowedIpRanges(). Error: %v", err)
		}
	}

	var result []*net.IPNet
	result = []*net.IPNet{}
	for _, cidr := range blocksSlice {
		var block *net.IPNet
		_, block, err = net.ParseCIDR(strings.TrimSpace(cidr))
		if err != nil {
			return nil, fmt.Errorf("error, could parse cidr block. Block: %s. Error: %v", cidr, err)
		}
		result = append(result, block)
	}
	return result, nil
}

func fetchAllowedIpRanges() ([]string, error) {
	request, err := http.NewRequest(http.MethodGet, "https://api.cloudflare.com/client/v4/ips", nil)
	if err != nil {
		return nil, fmt.Errorf("error, when generating get request: %v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if response != nil {
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				log.Printf("error, when attempting to close response body: %v", err)
			}
		}(response.Body)
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode > 299) {
		if response.StatusCode == http.StatusNotFound {
			log.Printf("recieved a 404 when attempting url: %s", request.URL)
		}
		var rb []byte
		rb, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("error, when reading error response body: %v", err)
		}
		return nil, fmt.Errorf("error, when performing get request. ERROR: %v. RESPONSE CODE: %d. RESPONSE MESSAGE: %s", err, response.StatusCode, string(rb))
	}
	if err != nil {
		if response != nil {
			err = fmt.Errorf("error: %v. RESPONSE CODE: %d", err, response.StatusCode)
		}
		return nil, fmt.Errorf("error, when performing post request. ERROR: %v", err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error, when reading response body: %v", err)
	}

	var result CloudflareIpsResponse
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return nil, fmt.Errorf("error, when unmarshalling response body: %v", err)
	}
	return result.Result.Ipv4Cidrs, nil
}

func initHttpServer() (*http.Server, error) {
	certPem := os.Getenv("TF_VAR_cloudflare_origin_cert")
	certPemBytes, err := base64.StdEncoding.DecodeString(certPem)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to decode the webserver cert: %v", err)
	}

	keyPem := os.Getenv("TF_VAR_cloudflare_origin_cert_key")
	keyPemBytes, err := base64.StdEncoding.DecodeString(keyPem)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to decode the webserver cert key: %v", err)
	}

	// Load the certificate and key
	cert, err := tls.X509KeyPair(certPemBytes, keyPemBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %v", err)
	}

	// Set up a TLS Config with the certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	serverPort := os.Getenv("TF_VAR_server_port")
	if serverPort == "" {
		return nil, errors.New("error, TF_VAR_server_port env var is required, but was not provided")
	}

	// Create a custom server with TLSConfig
	server := &http.Server{
		Addr:      ":" + serverPort,
		TLSConfig: tlsConfig,
	}
	return server, nil
}

func connectToRedisDatabase(redisPort string) (*redis.Client, error) {
	options := redis.Options{
		Addr: "keydb:6379",
		DB:   0, // use default DB
		// Password: password,
	}
	// Load client cert
	// clientCert, clientKey, err := getClientCertAndKey()
	// if err != nil {
	// 	return nil, fmt.Errorf("error, when getClientCertAndKey() for connectToRedisDatabase(). Error: %v", err)
	// }
	// cert, err := tls.X509KeyPair(clientCert, clientKey)
	// if err != nil {
	// 	return nil, fmt.Errorf("error, when attempting set LoadX509KeyPair() for connectToRedisDatabase(). Error: %v", err) // Ye don't want to sail without a map!
	// }

	// caCert, err := getCaCert()
	// if err != nil {
	// 	return nil, fmt.Errorf("error, when getCaCert() for connectToRedisDatabase(). Error: %v", err)
	// }

	// caCertPool := x509.NewCertPool()
	// caCertPool.AppendCertsFromPEM(caCert)

	// // Create TLS configuration
	// tlsConfig := &tls.Config{
	// 	Certificates:       []tls.Certificate{cert},
	// 	RootCAs:            caCertPool,
	// 	InsecureSkipVerify: false,
	// 	// Remember, settin' InsecureSkipVerify to true is like sailin' without a lookout!
	// 	// InsecureSkipVerify: true, // Only for development or testing!
	// }

	// options.TLSConfig = tlsConfig
	return redis.NewClient(&options), nil
}

func getClientCertAndKey() ([]byte, []byte, error) {
	clientCert, err := base64.StdEncoding.DecodeString(os.Getenv("TF_VAR_redis_user_crt"))
	if err != nil {
		return nil, nil, fmt.Errorf("error, when decoding clientCert. Error: %v", err)
	}

	clientKey, err := base64.StdEncoding.DecodeString(os.Getenv("TF_VAR_redis_user_private_key"))
	if err != nil {
		return nil, nil, fmt.Errorf("error, when decoding clientKey. Error: %v", err)
	}

	return clientCert, clientKey, nil
}

func getCaCert() ([]byte, error) {
	result, err := base64.StdEncoding.DecodeString(os.Getenv("TF_VAR_redis_ca"))
	if err != nil {
		return nil, fmt.Errorf("error, when decoding CA cert. Error: %v", err)
	}
	return result, nil
}

func connectToDatabase(ctx context.Context, connectionString, databaseRootCa string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("error, when parsing connection string: %v", err)
	}

	var tlsConfig *tls.Config
	tlsConfig, err = generateTlsConfig(databaseRootCa, config)
	if err != nil {
		return nil, fmt.Errorf("error, when generating tls config: %v", err)
	}
	// Add the TLS configuration to the connection config
	config.ConnConfig.TLSConfig = tlsConfig

	// Customize connection pool settings (if desired)
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Minute * 30
	config.MaxConnIdleTime = time.Minute * 5
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("error, unable to create connection pool: %v", err)
	}

	err = attemptToPingDatabaseUntilSuccessful(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("error, exausted attempts to ping the database: %v", err)
	}

	return pool, nil
}

func attemptToPingDatabaseUntilSuccessful(ctx context.Context, pool *pgxpool.Pool) error {
	timeOutInSeconds := 45
	retryInterval := 3
	var err error
	for i := 0; i < (timeOutInSeconds / retryInterval); i++ {
		err = pool.Ping(ctx)
		if err != nil {
			log.Printf("Database ping failed, will attempt again in: %d seconds...", retryInterval)
			time.Sleep(time.Duration(retryInterval) * time.Second)
		} else {
			break
		}
	}
	return err
}

func generateTlsConfig(databaseRootCa string, config *pgxpool.Config) (*tls.Config, error) {
	rootCAs, err := loadRootCA(databaseRootCa)
	if err != nil {
		return nil, fmt.Errorf("error, failed to load root CA for generateTlsConfig(): %v", err)
	}
	tlsConfig := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: config.ConnConfig.Host,
	}
	return tlsConfig, nil
}

func loadRootCA(databaseRootCa string) (*x509.CertPool, error) {
	var err error
	var decodedCert []byte

	// The cert is encoded when deployed because I need to pass it around with terraform
	decodedCert, err = base64.StdEncoding.DecodeString(databaseRootCa)
	if err != nil {
		return nil, fmt.Errorf("error, when base64 decoding database CA cert: %v", err)
	}

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(decodedCert); !ok {
		return nil, fmt.Errorf("error, failed to append CA certificate to the certificate pool")
	}
	return rootCAs, nil
}

func GetSuperSetExpiration() time.Duration {
	return time.Duration(CurrentSupersetExpirationTimeInHours) * time.Hour
}
