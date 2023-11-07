package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	RegistrationEmailFrom         string
	RegistrationEmailFromPassword string

	VerificationCodeLength = 6
	//LockoutDurationInHours                           = 24
	AllowedVerificationAttemptsWithTheExcessiveRetryLockoutWindow int
	VerificationCodeValidityWindowInMin                           int
	LocalDevelopment                                              string
	EmailRootCa                                                   string
	TrustedUiOrigins                                              []string

	// todo make this configurable
	AllowedVerificationResendCodeAttemptsWithinOneHour = 3

	AllowedLoginAttemptsBeforeTriggeringLockout int

	VerificationExcessiveRetryAttemptLockoutDurationInSeconds                 int
	WindowLengthInSecondsForTheNumberOfAllowedVerificationEmailsBeforeLockout int
	WindowLengthInSecondsForTheNumberOfAllowedLoginAttemptsBeforeLockout      int

	// MuscleGroupRecoveryWindowInHours setting this for 72 hours minus 4 hours for variance in workout start time. I was subtracting 2 hours but that turned out to not be enough as my exercises were finally showing up half-way through my workout.
	MuscleGroupRecoveryWindowInHours = time.Duration(72 - 4)

	NumberOfExerciseInSuperset = 3

	NumberOfSetsInSuperSet = 4

	// CurrentSupersetExpirationTimeInHours this addresses the edge case where the user doesn't finish the superset within a reasonable amount of time.
	// the superset is assumed to be aborted regardless of the progress made in that superset.
	CurrentSupersetExpirationTimeInHours = 2

	// CurrentWorkoutExpirationTimeInHours this is used to set the expiration time for the muscle groups worked counter.
	// This counter is used to keep the number of exercises spread evenly as possible throughout the week. The aim here
	// is to ensure workout times stay consistent, rather than 2 hours one day and 30 min the next day. The current rule
	// is that if the workout counter is equal or more than half the total amount of muscle groups then we will stop the workout.
	CurrentWorkoutExpirationTimeInHours = 4

	Version             string
	ConnectionPool      *pgxpool.Pool
	RedisConnectionPool *redis.Client
)

func InitConfig(ctx context.Context) error {
	var errorMsgs []string
	RegistrationEmailFrom = os.Getenv("REGISTRATION_EMAIL_FROM")
	if RegistrationEmailFrom == "" {
		errorMsgs = append(errorMsgs, "REGISTRATION_EMAIL_FROM")
	}
	RegistrationEmailFromPassword = os.Getenv("REGISTRATION_EMAIL_FROM_PASSWORD")
	if RegistrationEmailFromPassword == "" {
		errorMsgs = append(errorMsgs, "REGISTRATION_EMAIL_FROM_PASSWORD")
	}
	databaseConnectionString := os.Getenv("DATABASE_CONNECTION_STRING")
	if databaseConnectionString == "" {
		errorMsgs = append(errorMsgs, "DATABASE_CONNECTION_STRING")
	}
	Version = os.Getenv("VERSION")
	if Version == "" {
		errorMsgs = append(errorMsgs, "VERSION")
	}
	databaseRootCa := os.Getenv("DATABASE_ROOT_CA")
	if databaseRootCa == "" {
		errorMsgs = append(errorMsgs, "DATABASE_ROOT_CA")
	}
	trustedUiOriginsString := os.Getenv("TRUSTED_UI_ORIGIN")
	if trustedUiOriginsString == "" {
		errorMsgs = append(errorMsgs, "TRUSTED_UI_ORIGIN")
	} else {
		TrustedUiOrigins = strings.Split(trustedUiOriginsString, ",")
	}
	EmailRootCa = os.Getenv("EMAIL_ROOT_CA")
	if EmailRootCa == "" {
		errorMsgs = append(errorMsgs, "EMAIL_ROOT_CA")
	}
	redisConnectionString := os.Getenv("REDIS_CONNECTION_STRING")
	if redisConnectionString == "" {
		errorMsgs = append(errorMsgs, "REDIS_CONNECTION_STRING")
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		errorMsgs = append(errorMsgs, "REDIS_PASSWORD")
	}

	toParse := os.Getenv("VERIFICATION_EXCESSIVE_RETRY_ATTEMPT_LOCKOUT_DURATION_IN_SECONDS")
	var err error
	VerificationExcessiveRetryAttemptLockoutDurationInSeconds, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "VERIFICATION_EXCESSIVE_RETRY_ATTEMPT_LOCKOUT_DURATION_IN_SECONDS")
	}

	toParse = os.Getenv("ALLOWED_VERIFICATION_ATTEMPTS_WITH_THE_EXCESSIVE_RETRY_LOCKOUT_WINDOW")
	AllowedVerificationAttemptsWithTheExcessiveRetryLockoutWindow, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "ALLOWED_VERIFICATION_ATTEMPTS_WITH_THE_EXCESSIVE_RETRY_LOCKOUT_WINDOW")
	}

	toParse = os.Getenv("WINDOW_LENGTH_IN_SECONDS_FOR_THE_NUMBER_OF_ALLOWED_VERIFICATION_EMAILS_BEFORE_LOCKOUT")
	WindowLengthInSecondsForTheNumberOfAllowedVerificationEmailsBeforeLockout, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "WINDOW_LENGTH_IN_SECONDS_FOR_THE_NUMBER_OF_ALLOWED_VERIFICATION_EMAILS_BEFORE_LOCKOUT")
	}

	toParse = os.Getenv("WINDOW_LENGTH_IN_SECONDS_FOR_THE_NUMBER_OF_ALLOWED_LOGIN_ATTEMPTS_BEFORE_LOCKOUT")
	WindowLengthInSecondsForTheNumberOfAllowedLoginAttemptsBeforeLockout, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "WINDOW_LENGTH_IN_SECONDS_FOR_THE_NUMBER_OF_ALLOWED_LOGIN_ATTEMPTS_BEFORE_LOCKOUT")
	}

	toParse = os.Getenv("ALLOWED_LOGIN_ATTEMPTS_BEFORE_TRIGGERING_LOCKOUT")
	AllowedLoginAttemptsBeforeTriggeringLockout, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "ALLOWED_LOGIN_ATTEMPTS_BEFORE_TRIGGERING_LOCKOUT")
	}

	toParse = os.Getenv("VERIFICATION_CODE_VALIDITY_WINDOW_IN_MIN")
	VerificationCodeValidityWindowInMin, err = strconv.Atoi(toParse)
	if toParse == "" || err != nil {
		errorMsgs = append(errorMsgs, "VERIFICATION_CODE_VALIDITY_WINDOW_IN_MIN")
	}

	LocalDevelopment = os.Getenv("LOCAL_DEVELOPMENT")
	if len(errorMsgs) != 0 {
		return fmt.Errorf("missing required environmental variables: %v", errorMsgs)
	}

	RedisConnectionPool, err = connectToRedisDatabase(redisConnectionString, redisPassword)
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

	return nil
}

func connectToRedisDatabase(connectionString string, password string) (*redis.Client, error) {
	options := redis.Options{
		Addr:     connectionString,
		DB:       0, // use default DB
		Password: password,
	}
	if LocalDevelopment != "true" {
		// Load client cert
		clientCert, clientKey, err := getClientCertAndKey()
		if err != nil {
			return nil, fmt.Errorf("error, when getClientCertAndKey() for connectToRedisDatabase(). Error: %v", err)
		}
		cert, err := tls.X509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, fmt.Errorf("error, when attempting set LoadX509KeyPair() for connectToRedisDatabase(). Error: %v", err) // Ye don't want to sail without a map!
		}

		caCert, err := getCaCert()
		if err != nil {
			return nil, fmt.Errorf("error, when getCaCert() for connectToRedisDatabase(). Error: %v", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Create TLS configuration
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
			// Remember, settin' InsecureSkipVerify to true is like sailin' without a lookout!
			// InsecureSkipVerify: true, // Only for development or testing!
		}

		options.TLSConfig = tlsConfig
	}
	return redis.NewClient(&options), nil
}

func getClientCertAndKey() ([]byte, []byte, error) {
	clientCert, err := base64.StdEncoding.DecodeString(os.Getenv("REDIS_USER_CRT"))
	if err != nil {
		return nil, nil, fmt.Errorf("error, when decoding clientCert. Error: %v", err)
	}

	clientKey, err := base64.StdEncoding.DecodeString(os.Getenv("REDIS_USER_PRIVATE_KEY"))
	if err != nil {
		return nil, nil, fmt.Errorf("error, when decoding clientKey. Error: %v", err)
	}

	return clientCert, clientKey, nil
}

func getCaCert() ([]byte, error) {
	partOne, err := base64.StdEncoding.DecodeString(os.Getenv("REDIS_CA_PEM_PART_ONE"))
	if err != nil {
		return nil, fmt.Errorf("error, when decoding CA cert part one. Error: %v", err)
	}

	partTwo, err := base64.StdEncoding.DecodeString(os.Getenv("REDIS_CA_PEM_PART_TWO"))
	if err != nil {
		return nil, fmt.Errorf("error, when decoding CA cert part two. Error: %v", err)
	}

	partThree, err := base64.StdEncoding.DecodeString(os.Getenv("REDIS_CA_PEM_PART_THREE"))
	if err != nil {
		return nil, fmt.Errorf("error, when decoding CA cert part three. Error: %v", err)
	}

	partFour, err := base64.StdEncoding.DecodeString(os.Getenv("REDIS_CA_PEM_PART_FOUR"))
	if err != nil {
		return nil, fmt.Errorf("error, when decoding CA cert part four. Error: %v", err)
	}

	var result []byte
	result = append(result, partOne...)
	result = append(result, partTwo...)
	result = append(result, partThree...)
	result = append(result, partFour...)
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
	// todo figure out a way to keep the code from having to worry about infrastructure
	if LocalDevelopment != "true" {
		decodedCert, err = base64.StdEncoding.DecodeString(databaseRootCa)
		if err != nil {
			return nil, fmt.Errorf("error, when base64 decoding database CA cert: %v", err)
		}
	} else {
		decodedCert = []byte(databaseRootCa)
	}

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(decodedCert); !ok {
		return nil, fmt.Errorf("error, failed to append CA certificate to the certificate pool")
	}
	return rootCAs, nil
}
