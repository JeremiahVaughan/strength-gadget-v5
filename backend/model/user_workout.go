package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"strconv"
	"strengthgadget.com/m/v2/constants"
	"strings"
	"time"
)

// ExerciseUserDataMap is a type that represents a mapping between exercise ids and user exercise data.
type ExerciseUserDataMap map[string]ExerciseUserData

type UserWorkout struct {
	Weekday                  time.Weekday         `json:"weekday"`
	ProgressIndex            WorkoutProgressIndex `json:"progressIndex,omitempty"`
	WorkoutRoutine           RoutineType          `json:"workoutRoutine"`
	SlottedWarmupExercises   []uint16             `json:"-"`
	SlottedMainExercises     []uint16             `json:"-"`
	SlottedCoolDownExercises []uint16             `json:"-"`
	// UserExerciseDataMap also is used to tell if an exercise has already been selected or not
	UserExerciseDataMap ExerciseUserDataMap `json:"-"`
	Exists              bool                `json:"-"`
}

const (
	userWorkoutKey              = "userWorkoutKey"
	WorkoutProgressIndexKey     = "workoutProgressIndexKey"
	slottedWarmupExercisesKey   = "slottedWarmupExercises"
	slottedMainExercisesKey     = "slottedMainExercises"
	slottedCoolDownExercisesKey = "slottedCoolDownExercises"
	UserExerciseUserDataKey     = "userExerciseUserData"
)

func (use *UserWorkout) ToRedis(ctx context.Context, userId string, client *redis.Client, exp time.Duration) (err error) {
	// initialize pipeline
	pipe := client.Pipeline()

	// marshal CurrentStepPointer and WorkoutRoutine into JSON and store as a Redis string
	// storing the progress index in a separate redis key, so it can be updated individually
	temp := use.ProgressIndex
	use.ProgressIndex = nil
	userWorkout, err := json.Marshal(use)
	if err != nil {
		return fmt.Errorf("error, when marshalling user workout for redis. Error: %v", err)
	}
	use.ProgressIndex = temp
	pipe.Set(ctx, GetUserKey(userId, userWorkoutKey), userWorkout, exp)

	serializedProgressIndex, err := json.Marshal(use.ProgressIndex)
	if err != nil {
		return fmt.Errorf("error, when marshalling user workout progress index for redis. Error: %v", err)
	}
	pipe.Set(ctx, GetUserKey(userId, WorkoutProgressIndexKey), serializedProgressIndex, exp)

	// store SlottedWarmupExercises in a sorted set
	key := GetUserKey(userId, slottedWarmupExercisesKey)
	for i, exerciseIndex := range use.SlottedWarmupExercises {
		member := serializeUniqueMember(i, exerciseIndex)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(i), Member: member})
	}
	pipe.Expire(ctx, key, exp)

	// store SlottedMainExercises in a sorted set
	key = GetUserKey(userId, slottedMainExercisesKey)
	for i, exerciseIndex := range use.SlottedMainExercises {
		member := serializeUniqueMember(i, exerciseIndex)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(i), Member: member})
	}
	pipe.Expire(ctx, key, exp)

	// store SlottedCoolDownExercises in a sorted set
	key = GetUserKey(userId, slottedCoolDownExercisesKey)
	for i, exerciseIndex := range use.SlottedCoolDownExercises {
		member := serializeUniqueMember(i, exerciseIndex)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(i), Member: member})
	}
	pipe.Expire(ctx, key, exp)

	// UserExerciseDataMap and Daily Workout Slot Index stored as a Redis Hash
	key = GetUserKey(userId, UserExerciseUserDataKey)
	for hKey, hVal := range use.UserExerciseDataMap {
		var bytes []byte
		bytes, err = json.Marshal(hVal)
		if err != nil {
			return fmt.Errorf("error, when marshalling user exercise data for redis. Error: %v", err)
		}
		pipe.HSet(ctx, key, hKey, bytes)
	}
	pipe.Expire(ctx, key, exp)

	// execute the commands in the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error, when executing exec for ToRedis(). Error: %v", err)
	}

	return nil
}

func (use *UserWorkout) ToRedisUpdateExerciseSwap(
	ctx context.Context,
	userId string,
	client *redis.Client,
	exerciseSlotIndex int,
	oldExerciseIndex uint16,
	newExerciseIndex uint16,
	oldExerciseId string,
	newExerciseId string,
	exerciseUserData ExerciseUserData,
	slottedExerciseKey string,
) error {
	// initialize pipeline
	pipe := client.Pipeline()

	// update selected exercise
	oldMember := serializeUniqueMember(exerciseSlotIndex, oldExerciseIndex)
	pipe.ZRem(ctx, slottedExerciseKey, oldMember)
	newMember := serializeUniqueMember(exerciseSlotIndex, newExerciseIndex)
	pipe.ZAdd(ctx, slottedExerciseKey, redis.Z{Score: float64(exerciseSlotIndex), Member: newMember})

	// update exercise user data
	exerciseMeasurementKey := GetUserKey(userId, UserExerciseUserDataKey)
	pipe.HDel(ctx, exerciseMeasurementKey, oldExerciseId)
	bytes, err := json.Marshal(exerciseUserData)
	if err != nil {
		return fmt.Errorf("error marshaling exerciseUserData: %v", err)
	}
	pipe.HSet(ctx, exerciseMeasurementKey, newExerciseId, bytes)

	// execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error, when executing redis query to ToRedisUpdateExerciseSwap(). Error: %v", err)
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
	getWorkout := pipe.Get(ctx, GetUserKey(userId, userWorkoutKey))

	getWorkoutProgressIndex := pipe.Get(ctx, GetUserKey(userId, WorkoutProgressIndexKey))

	// Get sorted set of slottedWarmupExercises
	getWarmupExercises := pipe.ZRange(ctx, GetUserKey(userId, slottedWarmupExercisesKey), 0, -1)

	// Get sorted set of slottedMainExercises
	getMainExercises := pipe.ZRange(ctx, GetUserKey(userId, slottedMainExercisesKey), 0, -1)

	// Get sorted set of slottedCoolDownExercises
	getCoolDownExercises := pipe.ZRange(ctx, GetUserKey(userId, slottedCoolDownExercisesKey), 0, -1)

	// Get Hash of user exercise measurements
	getMeasurements := pipe.HGetAll(ctx, GetUserKey(userId, UserExerciseUserDataKey))

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

	// Unmarshal workout progress index
	workoutProgressIndex, err := getWorkoutProgressIndex.Result()
	if err != nil {
		return fmt.Errorf("error, when getting user workout progress index from redis. Error: %v", err)
	}
	err = json.Unmarshal([]byte(workoutProgressIndex), &use.ProgressIndex)
	if err != nil {
		return fmt.Errorf("error unmarshalling user workout progress index. Error: %v", err)
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

	// Get and convert map[string]string to map[string]ExerciseUserData
	userExerciseMeasurements, err := getMeasurements.Result()
	if err != nil {
		return fmt.Errorf("error, when retrieving updated user exercise user data from redis. Error: %v", err)
	}

	use.UserExerciseDataMap = make(map[string]ExerciseUserData)
	for k, v := range userExerciseMeasurements {
		eud := ExerciseUserData{}
		err = json.Unmarshal([]byte(v), &eud)
		if err != nil {
			return fmt.Errorf("error, when unmarshalling exercise user data from redis. Error: %v", err)
		}
		use.UserExerciseDataMap[k] = eud
	}

	return nil
}

func (use *UserWorkout) InitSlottedExercises(exercisesPerSuperSet int, dailyWorkout DailyWorkout) ([]string, error) {
	numberOfCardioExercisesPerWorkout := 1

	use.UserExerciseDataMap = make(ExerciseUserDataMap)

	for i := 0; i < numberOfCardioExercisesPerWorkout; i++ {
		startingExercise := uint16(rand.Intn(len(dailyWorkout.CardioExercises)))
		nextExercise, err := getNextAvailableExercise(
			startingExercise,
			dailyWorkout.CardioExercises,
			use.UserExerciseDataMap,
			len(use.SlottedWarmupExercises),
			constants.DailyWorkoutSlotPhaseWarmup,
		)
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
		nextExercise, err := getNextAvailableExercise(
			startingExercise,
			exercises,
			use.UserExerciseDataMap,
			len(use.SlottedMainExercises),
			constants.DailyWorkoutSlotPhaseMainFocused,
		)
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
		nextExercise, err := getNextAvailableExercise(
			startingExercise,
			exercises,
			use.UserExerciseDataMap,
			len(use.SlottedMainExercises),
			constants.DailyWorkoutSlotPhaseMainFiller,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when getNextAvailableExercise() for main filler exercises. Error: %v", err)
		}
		use.SlottedMainExercises = append(use.SlottedMainExercises, nextExercise)
	}

	numberOfCoolDownExercises := len(dailyWorkout.CoolDownExercises)
	for i := 0; i < numberOfCoolDownExercises; i++ {
		exercises := dailyWorkout.CoolDownExercises[i]
		startingExercise := uint16(rand.Intn(len(exercises)))
		nextExercise, err := getNextAvailableExercise(
			startingExercise,
			exercises,
			use.UserExerciseDataMap,
			len(use.SlottedCoolDownExercises),
			constants.DailyWorkoutSlotPhaseCoolDown,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when getNextAvailableExercise() for cool down exercises. Error: %v", err)
		}
		use.SlottedCoolDownExercises = append(use.SlottedCoolDownExercises, nextExercise)
	}

	var exerciseIds []string
	exerciseIds = []string{}
	for k := range use.UserExerciseDataMap {
		exerciseIds = append(exerciseIds, k)
	}
	return exerciseIds, nil
}

// getNextAvailableExercise finds the next available exercise from the exercise pool based on the starting exercise index and the already slotted exercises.
// It returns the index of the next available exercise in the exercise pool.
// If the exercise pool is empty, it returns the starting exercise index.
func getNextAvailableExercise(
	currentExerciseIndex uint16,
	exercisePool []Exercise,
	alreadySlottedExercises ExerciseUserDataMap,
	dailyWorkoutSlotIndex int,
	dailyWorkoutSlotPhase constants.DailyWorkoutSlotPhase,
) (uint16, error) {
	exercisePoolSize := len(exercisePool)
	if exercisePoolSize == 0 {
		return currentExerciseIndex, fmt.Errorf("error, cannot have an empty exercise pool")
	}

	result := currentExerciseIndex
	counter := int(currentExerciseIndex)

	for range exercisePool {
		selectedIndex := counter % exercisePoolSize
		selectedExercise := exercisePool[selectedIndex]
		if isNewExercise(selectedExercise.Id, alreadySlottedExercises) {
			result = uint16(selectedIndex)
			alreadySlottedExercises[selectedExercise.Id] = ExerciseUserData{
				Measurement:           0, // init to zero because exercise measurements are updated later
				DailyWorkoutSlotIndex: dailyWorkoutSlotIndex,
				DailyWorkoutSlotPhase: dailyWorkoutSlotPhase,
			}
			break
		}
		counter++
	}
	return result, nil
}

func isNewExercise(selectedExerciseId string, alreadySlottedExercises ExerciseUserDataMap) bool {
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

func GetUserKey(userId, key string) string {
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
	var us UserService
	user, err := us.FetchFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, could not UserService.FetchFromContext() for fetchAllMuscleGroupsNotInRecovery(). Error: %v", err)
	}

	// todo need to address the edge case where a user workout expires due to taking longer than 6 hours to complete
	userWorkout := UserWorkout{}
	err = userWorkout.FromRedis(ctx, user.Id, redisDb)
	if err != nil {
		return nil, fmt.Errorf("error, when fetching user workout from redis. Error: %v", err)
	}

	var dailyWorkout DailyWorkout
	weekday := time.Now().Weekday()
	if !userWorkout.Exists {
		userWorkout.ProgressIndex = []int{
			0,
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
		var exerciseMeasurements map[string]int
		exerciseMeasurements, err = fetchExerciseMeasurements(
			ctx,
			db,
			user.Id,
			slottedExercises,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when fetchExerciseMeasurements() for GetCurrentWorkout(). Error: %v", err)
		}

		for k, v := range exerciseMeasurements {
			d, ok := userWorkout.UserExerciseDataMap[k]
			if !ok {
				return nil, fmt.Errorf("error, expected exercise to exist in exercise data but it did not")
			}
			d.Measurement = v
			userWorkout.UserExerciseDataMap[k] = d
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
) (map[string]int, error) {
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

	result := make(map[string]int)
	for rows.Next() {
		var exerciseId string
		var measurement int
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

func fetchExerciseMeasurement(
	ctx context.Context,
	db *pgxpool.Pool,
	userId string,
	exerciseId string,
) (int, error) {
	var exerciseMeasurement int
	err := db.QueryRow(
		ctx,
		`SELECT measurement 
		 FROM last_completed_measurement
		 WHERE user_id = $1
		   AND exercise_id = $2`,
		userId,
		exerciseId,
	).Scan(
		&exerciseMeasurement,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// means the user hasn't done this exercise before for a measurement to be
			// recorded
			return 0, nil
		} else {
			return 0, fmt.Errorf("error, when attempting to execute sql statement: %v", err)
		}
	}
	return exerciseMeasurement, nil
}

func SwapExercise(
	ctx context.Context,
	redisDb *redis.Client,
	db *pgxpool.Pool,
	exerciseId string,
	numberOfSetsInSuperSet int,
	numberOfExerciseInSuperset int,
) (*UserWorkoutDto, *Error) {
	var us UserService
	user, err := us.FetchFromContext(ctx)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, could not UserService.FetchFromContext() for SwapExercise(). Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}

	userWorkout := UserWorkout{}
	err = userWorkout.FromRedis(ctx, user.Id, redisDb)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, when fetching user workout from redis for swapping exercise. Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}

	if !userWorkout.Exists {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, user %s attempted to fetch an user workout that expired", user.Id),
			UserFeedbackError: ErrorCouldNotLocateUserWorkout,
		}
	}

	dailyWorkout := DailyWorkout{}
	err = dailyWorkout.FromRedis(
		ctx,
		redisDb,
		getDailyWorkoutHashKey(userWorkout.WorkoutRoutine),
		userWorkout.Weekday,
	)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, when fetching the daily workout from redis for swapping an exercise. Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}

	// we are using the progress index passed directly from the client to avoid a race condition where the server side
	// might not have been updated yet
	exerciseUserData, ok := userWorkout.UserExerciseDataMap[exerciseId]
	if !ok {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, expected exercise data to exist for exercise id %s but it did not", exerciseId),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	workoutPhase := exerciseUserData.DailyWorkoutSlotPhase
	dailyWorkoutSlotIndex := exerciseUserData.DailyWorkoutSlotIndex
	var oldExercise Exercise
	var newExercise Exercise
	var currentExerciseIndex uint16
	var nextExerciseIndex uint16
	var exercisePool []Exercise
	var key string
	switch workoutPhase {
	case constants.DailyWorkoutSlotPhaseWarmup:
		exercisePool = dailyWorkout.CardioExercises
		currentExerciseIndex = userWorkout.SlottedWarmupExercises[dailyWorkoutSlotIndex]
		key = GetUserKey(user.Id, slottedWarmupExercisesKey)
	case constants.DailyWorkoutSlotPhaseMainFocused:
		exercisePool = dailyWorkout.MuscleCoverageMainExercises[dailyWorkoutSlotIndex]
		currentExerciseIndex = userWorkout.SlottedMainExercises[dailyWorkoutSlotIndex]
		key = GetUserKey(user.Id, slottedMainExercisesKey)
	case constants.DailyWorkoutSlotPhaseMainFiller:
		exercisePool = dailyWorkout.AllMainExercises
		currentExerciseIndex = userWorkout.SlottedMainExercises[dailyWorkoutSlotIndex]
		key = GetUserKey(user.Id, slottedMainExercisesKey)
	case constants.DailyWorkoutSlotPhaseCoolDown:
		exercisePool = dailyWorkout.CoolDownExercises[dailyWorkoutSlotIndex]
		currentExerciseIndex = userWorkout.SlottedCoolDownExercises[dailyWorkoutSlotIndex]
		key = GetUserKey(user.Id, slottedCoolDownExercisesKey)
	default:
		return nil, &Error{
			InternalError:     fmt.Errorf("error, unexpected daily workout slot phase provided: %d", workoutPhase),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	nextExerciseIndex, err = getNextAvailableExercise(
		currentExerciseIndex,
		exercisePool,
		userWorkout.UserExerciseDataMap,
		dailyWorkoutSlotIndex,
		workoutPhase,
	)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, when getNextAvailableExercise() for SwapExercise(). Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}

	oldExercise = exercisePool[currentExerciseIndex]
	newExercise = exercisePool[nextExerciseIndex]

	if newExercise.Id != oldExercise.Id { // edge case that can happen if we don't have enough exercises in a particular pool
		var newMeasurement int
		newMeasurement, err = fetchExerciseMeasurement(ctx, db, user.Id, newExercise.Id)
		if err != nil {
			return nil, &Error{
				InternalError:     fmt.Errorf("error, when fetching the next exercise newMeasurement for SwapExercise(). Error: %v", err),
				UserFeedbackError: ErrorUnexpectedTryAgain,
			}
		}
		exerciseUserData.Measurement = newMeasurement
		err = userWorkout.ToRedisUpdateExerciseSwap(
			ctx,
			user.Id,
			redisDb,
			dailyWorkoutSlotIndex,
			currentExerciseIndex,
			nextExerciseIndex,
			oldExercise.Id,
			newExercise.Id,
			exerciseUserData,
			key,
		)
		if err != nil {
			return nil, &Error{
				InternalError:     fmt.Errorf("error, when attempting to save swapped exercise to redis. Error: %v", err),
				UserFeedbackError: ErrorUnexpectedTryAgain,
			}
		}
	}

	var updatedUserWorkout UserWorkout
	err = updatedUserWorkout.FromRedis(ctx, user.Id, redisDb)
	if err != nil {
		return nil, &Error{
			InternalError: fmt.Errorf("error, when fetching updatedUserWorkout from redis for swapping exercise "+
				"for returning results to UI. Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}

	userWorkoutDto := UserWorkoutDto{}
	userWorkoutDto.Fill(
		updatedUserWorkout,
		dailyWorkout,
		numberOfSetsInSuperSet,
		numberOfExerciseInSuperset,
	)

	return &userWorkoutDto, nil
}
