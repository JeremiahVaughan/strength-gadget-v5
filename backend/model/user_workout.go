package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type UserWorkout struct {
	Weekday                  time.Weekday      `json:"weekday"`
	ProgressIndex            [][]int           `json:"progressIndex"`
	WorkoutRoutine           RoutineType       `json:"workoutRoutine"`
	SlottedWarmupExercises   []uint16          `json:"-"`
	SlottedMainExercises     []uint16          `json:"-"`
	SlottedCoolDownExercises []uint16          `json:"-"`
	ExerciseMeasurements     map[string]uint16 `json:"-"`
	Exists                   bool              `json:"-"`
}

const (
	userWorkoutKey              = "userWorkoutKey"
	slottedWarmupExercisesKey   = "slottedWarmupExercises"
	slottedMainExercisesKey     = "slottedMainExercises"
	slottedCoolDownExercisesKey = "slottedCoolDownExercises"
	userExerciseMeasurementsKey = "userExerciseMeasurements"
)

func (use *UserWorkout) ToRedis(ctx context.Context, userId string, client *redis.Client, exp time.Duration) (err error) {
	// initialize pipeline
	pipe := client.Pipeline()

	// marshal CurrentStepPointer and WorkoutRoutine into JSON and store as a Redis string
	userWorkout, err := json.Marshal(use)
	if err != nil {
		return fmt.Errorf("error, when marshalling user workout for redis. Error: %v", err)
	}
	pipe.Set(ctx, getUserKey(userId, userWorkoutKey), userWorkout, exp)

	// store SlottedWarmupExercises in a sorted set
	key := getUserKey(userId, slottedWarmupExercisesKey)
	for i, exerciseId := range use.SlottedWarmupExercises {
		member := serializeUniqueMember(i, exerciseId)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(i), Member: member})
	}
	pipe.Expire(ctx, key, exp)

	// store SlottedMainExercises in a sorted set
	key = getUserKey(userId, slottedMainExercisesKey)
	for i, exerciseId := range use.SlottedMainExercises {
		member := serializeUniqueMember(i, exerciseId)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(i), Member: member})
	}
	pipe.Expire(ctx, key, exp)

	// store SlottedCoolDownExercises in a sorted set
	key = getUserKey(userId, slottedCoolDownExercisesKey)
	for i, exerciseId := range use.SlottedCoolDownExercises {
		member := serializeUniqueMember(i, exerciseId)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(i), Member: member})
	}
	pipe.Expire(ctx, key, exp)

	// ExerciseMeasurements stored as a Redis Hash
	key = getUserKey(userId, userExerciseMeasurementsKey)
	for hKey, hVal := range use.ExerciseMeasurements {
		pipe.HSet(ctx, key, hKey, hVal)
	}
	pipe.Expire(ctx, key, exp)

	// execute the commands in the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error, when executing exec for ToRedis(). Error: %v", err)
	}

	return nil
}

func serializeUniqueMember(score int, exerciseIndex uint16) string {
	return strconv.Itoa(score) + ":" + strconv.Itoa(int(exerciseIndex))
}

func deserializeUniqueMember(member string) (uint16, error) {
	parts := strings.Split(member, ":")
	if len(parts) != 2 {
		return 00, fmt.Errorf("invalid member format")
	}

	exerciseIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("error parsing exerciseIndex: %v", err)
	}

	return uint16(exerciseIndex), nil
}
func (use *UserWorkout) FromRedis(ctx context.Context, userId string, client *redis.Client) error {
	// Initialize pipeline
	pipe := client.Pipeline()

	// Get userWorkout from Redis
	getWorkout := pipe.Get(ctx, getUserKey(userId, userWorkoutKey))

	// Get sorted set of slottedWarmupExercises
	getWarmupExercises := pipe.ZRange(ctx, getUserKey(userId, slottedWarmupExercisesKey), 0, -1)

	// Get sorted set of slottedMainExercises
	getMainExercises := pipe.ZRange(ctx, getUserKey(userId, slottedMainExercisesKey), 0, -1)

	// Get sorted set of slottedCoolDownExercises
	getCoolDownExercises := pipe.ZRange(ctx, getUserKey(userId, slottedCoolDownExercisesKey), 0, -1)

	// Get Hash of user exercise measurements
	getMeasurements := pipe.HGetAll(ctx, getUserKey(userId, userExerciseMeasurementsKey))

	// Execute the pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// nothing to update as there is no UserWorkout currently stored for this user
			use.Exists = false
			return nil
		}
		return fmt.Errorf("error executing pipeline for FromRedis(). Error: %v", err)
	}

	// Unmarshal userWorkout
	userWorkoutResult, err := getWorkout.Result()
	if err != nil {
		return fmt.Errorf("error, when getting user workout from redis. Error: %v", err)
	}
	use.Exists = true
	err = json.Unmarshal([]byte(userWorkoutResult), use)
	if err != nil {
		return fmt.Errorf("error unmarshalling user workout. Error: %v", err)
	}

	// Get and convert SlottedWarmupExercises from []string to []uint16
	slottedWarmupExercises, err := getWarmupExercises.Result()
	if err != nil {
		return fmt.Errorf("error, when retrieving slotted warmup exercises from redis. Error: %v", err)
	}
	for _, se := range slottedWarmupExercises {
		var exercisePosition uint16
		exercisePosition, err = deserializeUniqueMember(se)
		if err != nil {
			return fmt.Errorf("error, when deserializing unique member for warmup exercises. Error: %v", err)
		}
		use.SlottedWarmupExercises = append(use.SlottedWarmupExercises, exercisePosition)
	}

	// Get and convert SlottedMainExercises from []string to []uint16
	slottedMainExercises, err := getMainExercises.Result()
	if err != nil {
		return fmt.Errorf("error, when retrieving slotted main exercises from redis. Error: %v", err)
	}
	for _, se := range slottedMainExercises {
		var exercisePosition uint16
		exercisePosition, err = deserializeUniqueMember(se)
		if err != nil {
			return fmt.Errorf("error, when deserializing unique member for main exercises. Error: %v", err)
		}
		use.SlottedMainExercises = append(use.SlottedMainExercises, exercisePosition)
	}

	// Get and convert SlottedCoolDownExercises from []string to []uint16
	slottedCoolDownExercises, err := getCoolDownExercises.Result()
	if err != nil {
		return fmt.Errorf("error, when retrieving slotted cool down exercises from redis. Error: %v", err)
	}
	for _, se := range slottedCoolDownExercises {
		var exercisePosition uint16
		exercisePosition, err = deserializeUniqueMember(se)
		if err != nil {
			return fmt.Errorf("error, when deserializing unique member for cooldown exercises. Error: %v", err)
		}
		use.SlottedCoolDownExercises = append(use.SlottedCoolDownExercises, exercisePosition)
	}

	// Get and convert ExerciseMeasurements from map[string]string to map[string]uint16
	userExerciseMeasurements, err := getMeasurements.Result()
	if err != nil {
		return fmt.Errorf("error, when retrieving updated user exercise measurements from redis. Error: %v", err)
	}

	use.ExerciseMeasurements = make(map[string]uint16)
	for k, v := range userExerciseMeasurements {
		var exerciseMeasurement uint64
		exerciseMeasurement, err = strconv.ParseUint(v, 10, 16)
		if err != nil {
			return fmt.Errorf("error, when converting string to uint16. Error: %v", err)
		}
		use.ExerciseMeasurements[k] = uint16(exerciseMeasurement)
	}

	return nil
}

func (use *UserWorkout) InitSlottedExercises(exercisesPerSuperSet int, dailyWorkout DailyWorkout) ([]string, error) {
	numberOfCardioExercisesPerWorkout := 1

	// alreadySlottedExercises key is exercise id
	alreadySlottedExercises := make(map[string]bool)

	for i := 0; i < numberOfCardioExercisesPerWorkout; i++ {
		startingExercise := uint16(rand.Intn(len(dailyWorkout.CardioExercises)))
		nextExercise, err := getNextAvailableExercise(startingExercise, dailyWorkout.CardioExercises, alreadySlottedExercises)
		if err != nil {
			return nil, fmt.Errorf("error, when getNextAvailableExercise() for cardio exercises. Error: %v", err)
		}
		use.SlottedWarmupExercises = append(use.SlottedWarmupExercises, nextExercise)
	}

	numberOfMainExercises := len(dailyWorkout.MuscleCoverageMainExercises)
	minimumMainExercisesForWorkout := numberOfMainExercises
	for i := 0; i < minimumMainExercisesForWorkout; i++ {
		exercises := dailyWorkout.MuscleCoverageMainExercises[i]
		startingExercise := uint16(rand.Intn(len(exercises)))
		nextExercise, err := getNextAvailableExercise(startingExercise, exercises, alreadySlottedExercises)
		if err != nil {
			return nil, fmt.Errorf("error, when getNextAvailableExercise() for main exercises. Error: %v", err)
		}
		use.SlottedMainExercises = append(use.SlottedMainExercises, nextExercise)
	}

	// The point of filler exercises is to make all sets even otherwise the last set may end up being a single exercise
	numberOfSets := calculateNumberOfSets(dailyWorkout, exercisesPerSuperSet)
	totalExercises := numberOfSets * exercisesPerSuperSet
	requiredFillerExercises := totalExercises - numberOfMainExercises
	for i := 0; i < requiredFillerExercises; i++ {
		exercises := dailyWorkout.AllMainExercises
		startingExercise := uint16(rand.Intn(len(exercises)))
		nextExercise, err := getNextAvailableExercise(startingExercise, exercises, alreadySlottedExercises)
		if err != nil {
			return nil, fmt.Errorf("error, when getNextAvailableExercise() for main filler exercises. Error: %v", err)
		}
		use.SlottedMainExercises = append(use.SlottedMainExercises, nextExercise)
	}

	numberOfCoolDownExercises := len(dailyWorkout.CoolDownExercises)
	for i := 0; i < numberOfCoolDownExercises; i++ {
		exercises := dailyWorkout.CoolDownExercises[i]
		startingExercise := uint16(rand.Intn(len(exercises)))
		nextExercise, err := getNextAvailableExercise(startingExercise, exercises, alreadySlottedExercises)
		if err != nil {
			return nil, fmt.Errorf("error, when getNextAvailableExercise() for cool down exercises. Error: %v", err)
		}
		use.SlottedCoolDownExercises = append(use.SlottedCoolDownExercises, nextExercise)
	}

	var exerciseIds []string
	exerciseIds = []string{}
	for k := range alreadySlottedExercises {
		exerciseIds = append(exerciseIds, k)
	}
	return exerciseIds, nil
}

// getNextAvailableExercise finds the next available exercise from the exercise pool based on the starting exercise index and the already slotted exercises.
// It returns the index of the next available exercise in the exercise pool.
// If the exercise pool is empty, it returns the starting exercise index.
func getNextAvailableExercise(startingExercise uint16, exercisePool []Exercise, alreadySlottedExercises map[string]bool) (uint16, error) {
	exercisePoolSize := len(exercisePool)
	if exercisePoolSize == 0 {
		return startingExercise, fmt.Errorf("error, cannot have an empty exercise pool")
	}

	result := startingExercise
	counter := int(startingExercise)

	for range exercisePool {
		selectedIndex := counter % exercisePoolSize
		selectedExercise := exercisePool[selectedIndex]
		if isNewExercise(selectedExercise.Id, alreadySlottedExercises) {
			result = uint16(selectedIndex)
			alreadySlottedExercises[selectedExercise.Id] = true
			break
		}
		counter++
	}
	return result, nil
}

func isNewExercise(selectedExerciseId string, alreadySlottedExercises map[string]bool) bool {
	_, alreadyExists := alreadySlottedExercises[selectedExerciseId]
	return !alreadyExists
}

func calculateNumberOfSets(workout DailyWorkout, exercisesPerSuperSet int) int {
	length := len(workout.MuscleCoverageMainExercises)
	result := length / exercisesPerSuperSet

	if length%exercisesPerSuperSet != 0 {
		result += 1
	}
	return result
}

func getUserKey(userId, key string) string {
	return userId + ":" + key
}

func GetCurrentWorkout(
	ctx context.Context,
	redisDb *redis.Client,
	db *pgxpool.Pool,
	numberOfSetsInSuperSet,
	numberOfExerciseInSuperset int,
	superSetExpiration time.Duration,
) (*UserWorkoutDto, error) {
	user, err := FetchUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, could not FetchCurrentSuperset() for fetchAllMuscleGroupsNotInRecovery(). Error: %v", err)
	}

	userWorkout := UserWorkout{}
	err = userWorkout.FromRedis(ctx, user.Id, redisDb)
	if err != nil {
		return nil, fmt.Errorf("error, when fetching user workout from redis. Error: %v", err)
	}

	var dailyWorkout DailyWorkout
	weekday := time.Now().Weekday()
	if !userWorkout.Exists {
		userWorkout.ProgressIndex = [][]int{
			{0},
		}
		userWorkout.Weekday = time.Now().Weekday()
		userWorkout.WorkoutRoutine, err = fetchCurrentWorkoutRoutine(ctx, db, user.Id)
		if err != nil {
			return nil, fmt.Errorf("error, when fetchCurrentWorkoutRoutine() for GetCurrentWorkout(). Error: %v", err)
		}

		err = dailyWorkout.FromRedis(
			ctx,
			redisDb,
			getDailyWorkoutHashKey(userWorkout.WorkoutRoutine),
			weekday,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when fetching the daily workout from redis for new workout. Error: %v", err)
		}

		var slottedExercises []string
		slottedExercises, err = userWorkout.InitSlottedExercises(numberOfExerciseInSuperset, dailyWorkout)
		if err != nil {
			return nil, fmt.Errorf("error, when InitSlottedExercises(). Error: %v", err)
		}
		userWorkout.ExerciseMeasurements, err = fetchExerciseMeasurements(
			ctx,
			db,
			user.Id,
			slottedExercises,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when fetchExerciseMeasurements() for GetCurrentWorkout(). Error: %v", err)
		}

		err = userWorkout.ToRedis(
			ctx,
			user.Id,
			redisDb,
			superSetExpiration,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when userWorkout.ToRedis() for GetCurrentWorkout(). Error: %v", err)
		}
	} else {
		err = dailyWorkout.FromRedis(
			ctx,
			redisDb,
			getDailyWorkoutHashKey(userWorkout.WorkoutRoutine),
			weekday,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when fetching the daily workout from redis for existing workout. Error: %v", err)
		}
	}

	result := UserWorkoutDto{}
	result.Fill(
		userWorkout,
		dailyWorkout,
		numberOfSetsInSuperSet,
		numberOfExerciseInSuperset,
	)
	return &result, nil
}

func fetchCurrentWorkoutRoutine(ctx context.Context, db *pgxpool.Pool, userId string) (RoutineType, error) {
	var result RoutineType
	err := db.QueryRow(
		ctx,
		"SELECT current_routine\nFROM public.\"user\"\nWHERE id = $1",
		userId,
	).Scan(
		&result,
	)
	if err != nil {
		return 0, fmt.Errorf("error, when attempting to execute sql statement: %v", err)
	}
	return result, nil
}

func fetchExerciseMeasurements(
	ctx context.Context,
	db *pgxpool.Pool,
	userId string,
	exerciseIds []string,
) (map[string]uint16, error) {
	var placeholders strings.Builder
	var args []interface{}
	args = append(args, userId) // user id will be our first argument

	for i, exerciseId := range exerciseIds {
		if i != 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString(fmt.Sprintf("$%d", i+2))

		args = append(args, exerciseId)
	}

	query := fmt.Sprintf(
		"SELECT exercise_id, measurement FROM last_completed_measurement WHERE user_id = $1 AND exercise_id IN (%s)",
		placeholders.String(),
	)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to retrieve records. Error: %v", err)
	}

	if err != nil {
		return nil, fmt.Errorf("error, when attempting to retrieve records. Error: %v", err)
	}

	result := make(map[string]uint16)
	for rows.Next() {
		var exerciseId string
		var measurement uint16
		err = rows.Scan(
			&exerciseId,
			&measurement,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when scanning database rows: %v", err)
		}
		result[exerciseId] = measurement
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error, when iterating through database rows: %v", err)
	}
	return result, nil
}
