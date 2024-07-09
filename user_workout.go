package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

type UserWorkoutDto struct {
	ProgressIndex     WorkoutProgressIndex `json:"progressIndex"`
	Weekday           time.Weekday         `json:"weekday"`
	WorkoutId         int                  `json:"workoutId"`
	WarmupExercises   []Exercise           `json:"warmupExercises"`
	MainExercises     []Exercise           `json:"mainExercises"`
	CoolDownExercises []Exercise           `json:"coolDownExercises"`
}


// ChoosenExercisesMap is a type that represents which exercises have already been selected, and the value holds the current mesarument
type ChoosenExercisesMap map[int]int

// ExerciseMeasurementsMap key is exerciseId, value is current measurement value
type ExerciseMeasurementsMap map[int]int

type UserWorkout struct {
	Weekday                  time.Weekday         `json:"weekday"`
	ProgressIndex            WorkoutProgressIndex `json:"progressIndex,omitempty"`
	WorkoutRoutine           RoutineType          `json:"workoutRoutine"`
	WorkoutId                int                  `json:"workoutId"`
	SlottedWarmupExercises   []uint16             `json:"-"`
	SlottedMainExercises     []uint16             `json:"-"`
	SlottedCoolDownExercises []uint16             `json:"-"`
	// UserExerciseDataMap also is used to tell if an exercise has already been selected or not
	UserExerciseDataMap ChoosenExercisesMap `json:"-"`
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

func (use *UserWorkout) ToRedis(ctx context.Context, userId int64, client *redis.Client, exp time.Duration) (err error) {
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


// getNextAvailableExercise finds the next available exercise from the exercise pool based on the starting exercise index and the already slotted exercises.
// It returns the index of the next available exercise in the exercise pool.
// If the exercise pool doesn't contain any available exercises, it returns the starting exercise.
func getNextAvailableExercise(
	currentOffset int,
	randomPool []int,
	exercisePool []Exercise,
	alreadySlottedExercises ChoosenExercisesMap,
) (nextExercise Exercise, nextOffset int, err error) {
	exercisePoolSize := len(exercisePool)
	if exercisePoolSize == 0 {
		return Exercise{}, 0, fmt.Errorf("error, cannot have an empty exercise pool")
	}

	counter := currentOffset
	var selectedExercise Exercise
	for i := 0; i <= len(exercisePool); i++ {
		selectedExercise, err = getSelectedExercise(counter, randomPool, exercisePool)
		if err != nil {
			return Exercise{}, 0, fmt.Errorf("error, when getSelectedExercise() for getNextAvailableExercise(). Error: %v", err)
		}
		if isNewExercise(selectedExercise.Id, alreadySlottedExercises) {
			alreadySlottedExercises[selectedExercise.Id] = 0 // init to zero because exercise measurements are updated later
			break
		}
		counter++
	}
	return selectedExercise, counter, nil
}

func getSelectedExercise(
	currentOffset int,
	randomPool []int,
	exercisePool []Exercise,
) (Exercise, error) {
	exercisePoolLength := len(exercisePool)
	if exercisePoolLength == 0 {
		return Exercise{}, errors.New("error, exercisePool cannot be empty")
	}
	randomPoolLength := len(randomPool)
	if randomPoolLength == 0 {
		return Exercise{}, errors.New("error, randomPool cannot be empty")
	}
	if randomPoolLength != exercisePoolLength {
		return Exercise{}, fmt.Errorf("error, randomPoolLength %d does not equal exercisePoolLength %d", randomPoolLength, exercisePoolLength)
	}
	selectedIndex := currentOffset % exercisePoolLength
	if selectedIndex >= randomPoolLength {
		return Exercise{}, fmt.Errorf(
			`error, selectedIndex is out of bounds with the randomPool. 
			selectedIndex: %d,
			currentOffset: %d,
			randomPoolLength: %d,
			randomPool: %+v,
			exercisePoolLength: %d,
			exercisePool: %+v,`,
			selectedIndex,
			currentOffset,
			randomPoolLength,
			randomPool,
			exercisePoolLength,
			exercisePool,
		)
	}
	actualIndex := randomPool[selectedIndex]
	if actualIndex >= exercisePoolLength {
		return Exercise{}, fmt.Errorf(
			`error, actualIndex is out of bounds for given exercise pool. 
			actualIndex: %d,
			currentOffset: %d,
			randomPoolLength: %d,
			randomPool: %+v,
			exercisePoolLength: %d,
			exercisePool: %+v,`,
			actualIndex,
			currentOffset,
			randomPoolLength,
			randomPool,
			exercisePoolLength,
			exercisePool,
		)
	}
	return exercisePool[actualIndex], nil
}

func isNewExercise(selectedExerciseId int, alreadySlottedExercises ChoosenExercisesMap) bool {
	_, alreadyExists := alreadySlottedExercises[selectedExerciseId]
	return !alreadyExists
}

func calculateNumberOfSets(workout AvailableWorkoutExercises, exercisesPerSuperSet int) int {
	length := len(workout.MainExercises)
	result := length / exercisesPerSuperSet

	if length%exercisesPerSuperSet != 0 {
		result += 1
	}
	return result
}

func GetUserKey(userId int64, key string) string {
	uid := strconv.FormatInt(userId, 10)
	return uid + ":" + key
}


func fetchCurrentWorkoutRoutine(ctx context.Context, db *pgxpool.Pool, userId int64) (RoutineType, error) {
	var result RoutineType
	err := db.QueryRow(
		ctx,
		`SELECT current_routine
        FROM public.athlete
        WHERE id = $1`,
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
	userId int64,
	choosenExercises ChoosenExercisesMap,
) (currentMeasurements ChoosenExercisesMap, err error) {
	placeholders := make([]string, len(choosenExercises))
	var args []interface{}
	args = append(args, userId) // user id will be our first argument

	exerciseIds := make([]any, len(choosenExercises))
	i := 0
	for exerciseId := range choosenExercises {
		exerciseIds[i] = exerciseId
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		i++
	}

	args = append(args, exerciseIds...)

	query := fmt.Sprintf(
		`SELECT exercise_id, measurement 
        FROM last_completed_measurement 
        WHERE user_id = $1 
            AND exercise_id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to retrieve records. Error: %v", err)
	}

	for rows.Next() {
		var exerciseId int
		var measurement int
		err = rows.Scan(
			&exerciseId,
			&measurement,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when scanning database rows: %v", err)
		}
		choosenExercises[exerciseId] = measurement
	}

	// add placeholders for measurements that haven't been persisted yet
	for exerciseId := range choosenExercises {
		_, ok := choosenExercises[exerciseId] 
		if !ok {
			choosenExercises[exerciseId] = 0
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error, when iterating through database rows: %v", err)
	}
	return choosenExercises, nil
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

// func SwapExercise(
// 	ctx context.Context,
// 	redisDb *redis.Client,
// 	db *pgxpool.Pool,
// 	exerciseId string,
// 	workoutId string,
// 	numberOfSetsInSuperSet int,
// 	numberOfExerciseInSuperset int,
// 	exp int,
// ) (*UserWorkoutDto, error) {
// 	var us UserService
// 	user, err := us.FetchFromContext(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, could not userservice.fetchfromcontext() for swapexercise(). error: %v", err)
// 	}

// 	userWorkout := UserWorkout{}
// 	err = userWorkout.FromRedis(ctx, user.Id, redisDb)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, when fetching user workout from redis for swapping exercise. Error: %v", err)
// 	}

// 	if !userWorkout.Exists || userWorkout.WorkoutId != workoutId {
// 		return nil, fmt.Errorf("error, user %s attempted to fetch an user workout that expired for SwapExercise()", user.Id)
// 	}

// 	dailyWorkout := AvailableWorkoutExercises{}
// 	err = dailyWorkout.FromRedis(
// 		ctx,
// 		redisDb,
// 		getDailyWorkoutHashKey(userWorkout.WorkoutRoutine),
// 		userWorkout.Weekday,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, when fetching the daily workout from redis for swapping an exercise. Error: %v", err)
// 	}

// 	// we are using the progress index passed directly from the client to avoid a race condition where the server side
// 	// might not have been updated yet
// 	exerciseUserData, ok := userWorkout.UserExerciseDataMap[exerciseId]
// 	if !ok {
// 		return nil, fmt.Errorf("error, expected exercise data to exist for exercise id %s but it did not", exerciseId)
// 	}
// 	workoutPhase := exerciseUserData.DailyWorkoutSlotPhase
// 	dailyWorkoutSlotIndex := exerciseUserData.DailyWorkoutSlotIndex
// 	var oldExercise Exercise
// 	var newExercise Exercise
// 	var currentExerciseIndex uint16
// 	var nextExerciseIndex uint16
// 	var exercisePool []Exercise
// 	var key string
// 	switch workoutPhase {
// 	case DailyWorkoutSlotPhaseWarmup:
// 		exercisePool = dailyWorkout.CardioExercises
// 		currentExerciseIndex = userWorkout.SlottedWarmupExercises[dailyWorkoutSlotIndex]
// 		key = GetUserKey(user.Id, slottedWarmupExercisesKey)
// 	case DailyWorkoutSlotPhaseMainFocused:
// 		exercisePool = dailyWorkout.MainExercises[dailyWorkoutSlotIndex]
// 		currentExerciseIndex = userWorkout.SlottedMainExercises[dailyWorkoutSlotIndex]
// 		key = GetUserKey(user.Id, slottedMainExercisesKey)
// 	case DailyWorkoutSlotPhaseMainFiller:
// 		exercisePool = dailyWorkout.AllMainExercises
// 		currentExerciseIndex = userWorkout.SlottedMainExercises[dailyWorkoutSlotIndex]
// 		key = GetUserKey(user.Id, slottedMainExercisesKey)
// 	case DailyWorkoutSlotPhaseCoolDown:
// 		exercisePool = dailyWorkout.CoolDownExercises[dailyWorkoutSlotIndex]
// 		currentExerciseIndex = userWorkout.SlottedCoolDownExercises[dailyWorkoutSlotIndex]
// 		key = GetUserKey(user.Id, slottedCoolDownExercisesKey)
// 	default:
// 		return nil, fmt.Errorf("error, unexpected daily workout slot phase provided: %d", workoutPhase)
// 	}
// 	nextExerciseIndex, err = getNextAvailableExercise(
// 		currentExerciseIndex,
// 		exercisePool,
// 		userWorkout.UserExerciseDataMap,
// 		dailyWorkoutSlotIndex,
// 		workoutPhase,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, when getNextAvailableExercise() for SwapExercise(). Error: %v", err)
// 	}

// 	oldExercise = exercisePool[currentExerciseIndex]
// 	newExercise = exercisePool[nextExerciseIndex]

// 	if newExercise.Id != oldExercise.Id { // edge case that can happen if we don't have enough exercises in a particular pool
// 		var newMeasurement int
// 		newMeasurement, err = fetchExerciseMeasurement(ctx, db, user.Id, newExercise.Id)
// 		if err != nil {
// 			return nil, fmt.Errorf("error, when fetching the next exercise newMeasurement for SwapExercise(). Error: %v", err)
// 		}
// 		exerciseUserData.Measurement = newMeasurement
// 		err = userWorkout.ToRedisUpdateExerciseSwap(
// 			ctx,
// 			user.Id,
// 			redisDb,
// 			dailyWorkoutSlotIndex,
// 			currentExerciseIndex,
// 			nextExerciseIndex,
// 			oldExercise.Id,
// 			newExercise.Id,
// 			exerciseUserData,
// 			key,
// 			exp,
// 		)
// 		if err != nil {
// 			return nil, fmt.Errorf("error, when attempting to save swapped exercise to redis. Error: %v", err)
// 		}
// 	}

// 	var updatedUserWorkout UserWorkout
// 	err = updatedUserWorkout.FromRedis(ctx, user.Id, redisDb)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, when fetching updatedUserWorkout from redis for swapping exercise for returning results to UI. Error: %v", err)
// 	}

// 	userWorkoutDto := UserWorkoutDto{}
// 	userWorkoutDto.Fill(
// 		updatedUserWorkout,
// 		dailyWorkout,
// 		numberOfSetsInSuperSet,
// 		numberOfExerciseInSuperset,
// 	)

// 	return &userWorkoutDto, nil
// }
