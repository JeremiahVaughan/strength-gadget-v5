package main

import (
	"context"
    "database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nalgeon/redka"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var (
	DefaultExerciseTimeOptions   = generateDefaultTimeOptions()
	DefaultExerciseWeightOptions = generateDefaultWeightOptions()

	Environment string

	SentryEndpoint string

	RegistrationEmailFrom         string
	RegistrationEmailFromPassword string

	VerificationCodeLength = 6
	//LockoutDurationInHours                           = 24
	AllowedVerificationAttemptsWithTheExcessiveRetryLockoutWindow int
	VerificationCodeValidityWindowInMin                           int
	EmailRootCa                                                   string

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

	DomainName string

	Version             string
	ConnectionPool      *sql.DB
	RedisConnectionPool *redka.DB

	HttpServer *http.Server

	AllowedIpRanges []*net.IPNet

	lowerWorkout AvailableWorkoutExercises
	coreWorkout  AvailableWorkoutExercises
	upperWorkout AvailableWorkoutExercises

	DebugMode string
)

type Config struct {
    Database Database `json:"database"`
}

type Database struct {
    DataDirectory string `json:"dataDirectory"`
    MigrationDirectory string `json:"migrationDirectory"`
}

func NewConfig(ctx context.Context) (Config, error) {
    bytes, err := fetchConfigFromS3(ctx, "strengthgadget")
    if err != nil {
        return Config{}, fmt.Errorf("error, when fetching config file. Error: %v", err)
    }

    var c Config
    err = json.Unmarshal(bytes, &c)
    if err != nil {
        return Config{}, fmt.Errorf("error, when decoding config file. Error: %v", err)
    }

    return c, nil
}

func generateDefaultTimeOptions() MeasurementOptions {
	timeSelectionCap := 1200
	timeInterval := 15
	return generateTimeOptions(timeInterval, timeSelectionCap)
}

func generateDefaultWeightOptions() MeasurementOptions {
	weightSelectionCap := 600
	weightInterval := 5
	return generateWeightOptions(weightInterval, weightSelectionCap)
}

func init() {
	var err error
	exerciseMap := generateExerciseMap()
	muscleGroupMap := generateMuscleGroupMap()
	lowerWorkout, err = generateWorkoutExercises(exerciseMap, muscleGroupMap, LOWER)
	if err != nil {
		log.Fatalf("error, when generateWorkoutExercises() for lower body. Error: %v", err)
	}
	coreWorkout, err = generateWorkoutExercises(exerciseMap, muscleGroupMap, CORE)
	if err != nil {
		log.Fatalf("error, when generateWorkoutExercises() for core body. Error: %v", err)
	}
	upperWorkout, err = generateWorkoutExercises(exerciseMap, muscleGroupMap, UPPER)
	if err != nil {
		log.Fatalf("error, when generateWorkoutExercises() for upper body. Error: %v", err)
	}
}

// generateMuscleGroupMap return value key is muscle group Id
func generateMuscleGroupMap() map[int]MuscleGroup {
	result := make(map[int]MuscleGroup, len(AllMuscleGroups))
	for _, mg := range AllMuscleGroups {
		result[mg.Id] = mg
	}
	return result
}




func InitConfig(ctx context.Context) (Config, error) {
	var err error

    config, err := NewConfig(ctx)
    if err != nil {
        return Config{}, fmt.Errorf("error, when creating new config. Error: %v", err)
    }

	AllowedIpRanges, err = initAllowedIpRanges()
	if err != nil {
        return Config{}, fmt.Errorf("error, when initAllowedIpRanges() for InitConfig(). Error: %v", err)
	}

	RedisConnectionPool, err = connectToRedisDatabase()
	if err != nil {
		return Config{}, fmt.Errorf("error, when connectToRedisDatabase() for InitConfig(). Error: %v", err)
	}

	return config, nil
}

func initAllowedIpRanges() ([]*net.IPNet, error) {
	var blocksSlice []string
	var err error
	if Environment == EnvironmentLocal {
		blocksSlice = []string{
			"127.0.0.1/32", // ipv4 loop-back
			"::1/128",      // ipv6 loop-back
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


func connectToRedisDatabase() (*redka.DB, error) {
	return redka.Open("/session_data/redis.db", nil)
}


func GetSuperSetExpiration() time.Duration {
	return time.Duration(CurrentSupersetExpirationTimeInHours) * time.Hour
}
