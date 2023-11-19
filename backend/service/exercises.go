package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"strconv"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strings"
	"time"
)

func FinishCurrentAndFetchNextExercise(ctx context.Context, measurement string) (*model.ExerciseResponse, error) {
	parsedMeasurement, err := strconv.Atoi(measurement)
	if err != nil {
		return nil, fmt.Errorf("error, when converting measurement from string to int. Error: %v", err)
	}

	currentSuperset, err := FetchCurrentSuperset(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, when FetchCurrentSuperset() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
	}

	currentSuperset = updateSuperSetWithCurrentMeasurement(currentSuperset, parsedMeasurement)

	user, err := fetchUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, could not FinishCurrentAndFetchNextExercise() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
	}

	var currentExerciseId string
	if currentSuperset.Exercises != nil {
		currentExerciseId = currentSuperset.Exercises[currentSuperset.CurrentExercisePointer].Id
	}

	var muscleGroupsNotInRecovery []model.MuscleGroup
	muscleGroupsNotInRecovery, err = fetchAllMuscleGroupsNotInRecovery(ctx, currentSuperset, false)
	if err != nil {
		return nil, fmt.Errorf("error, when fetchAllMuscleGroupsNotInRecovery() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
	}
	numberOfAvailableMuscleGroups := len(muscleGroupsNotInRecovery)
	numberOfExerciseInSuperset := config.NumberOfExerciseInSuperset
	currentSuperset = markPreviousExerciseAsCompleted(currentSuperset, numberOfAvailableMuscleGroups, numberOfExerciseInSuperset)

	var randomExercise *model.Exercise
	numberOfActiveExercises := len(currentSuperset.Exercises)
	if !isSuperSetFull(numberOfActiveExercises, numberOfAvailableMuscleGroups) {
		randomExercise, err = fetchRandomExercise(
			ctx,
			currentSuperset,
			currentExerciseId,
			muscleGroupsNotInRecovery,
			user.Id,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when fetchRandomExercise() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
		}
		currentSuperset.Exercises = append(currentSuperset.Exercises, model.Exercise{
			Id:                       randomExercise.Id,
			LastCompletedMeasurement: randomExercise.LastCompletedMeasurement,
		})
	} else if IsSupersetComplete(currentSuperset) {
		// todo address the awkwardness of having the very last set contain a single exercise, maybe try to keep all the super sets the same length by just adding more exercises from the same muscle group over and over again
		// todo could also just change the reps and weight to indicate that they completed the previous set but then there is the question of if they are supposed to rest or not.
		err = markCompletionOfCurrentSuperset(ctx, user, currentSuperset)
		if err != nil {
			return nil, fmt.Errorf("error, when markCompletionOfCurrentSuperset() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
		}

		var muscleGroupsCompletedCount int
		muscleGroupsCompletedCount, err = getCurrentWorkoutMuscleGroupsWorkedCount(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("error, when getCurrentWorkoutMuscleGroupsWorkedCount() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
		}

		muscleGroupsCount, err := getTotalMuscleGroupsCount(ctx)
		if err != nil {
			return nil, fmt.Errorf("error, when getTotalMuscleGroupsCount() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
		}

		var result *model.ExerciseResponse
		if hasMuscleGroupWorkedSessionLimitBeenReached(muscleGroupsCount, muscleGroupsCompletedCount) {
			result, err = getCompletedWorkoutResponse(ctx, user, currentSuperset)
			if err != nil {
				return nil, fmt.Errorf("error, when getCompletedWorkoutResponse() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
			}
		} else {
			result, err = ShuffleExercise(ctx)
			if err != nil {
				return nil, fmt.Errorf("error, when ShuffleExercise() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
			}
		}
		return result, nil
	}

	err = setCurrentSupersetForUser(ctx, user, currentSuperset)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to set next exercise for user: %s. Error: %v", user.Email, err)
	}

	response, err := SuperSetToExerciseResponse(ctx, currentSuperset)
	if err != nil {
		return nil, fmt.Errorf("error, when SuperSetToExerciseResponse() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
	}
	return response, nil
}

func isSuperSetFull(numberOfActiveExercises int, numberOfAvailableMuscleGroups int) bool {
	return numberOfActiveExercises == config.NumberOfExerciseInSuperset || numberOfAvailableMuscleGroups == 0
}

func IsSupersetComplete(currentSuperset *model.SuperSet) bool {
	return currentSuperset.SetCompletionCount == config.NumberOfSetsInSuperSet
}

func getCompletedWorkoutResponse(ctx context.Context, user *model.User, currentSuperset *model.SuperSet) (*model.ExerciseResponse, error) {
	currentSuperset.WorkoutComplete = true
	err := setCurrentSupersetForUser(ctx, user, currentSuperset)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to set current superset for user user: %s. Error: %v", user.Email, err)
	}

	response, err := SuperSetToExerciseResponse(ctx, currentSuperset)
	if err != nil {
		return nil, fmt.Errorf("error, when SuperSetToExerciseResponse() for getCompletedWorkoutResponse(). Error: %v", err)
	}
	return response, nil
}

func getCurrentWorkoutMuscleGroupsWorkedCount(ctx context.Context, user *model.User) (int, error) {
	var result int
	redisResult, err := config.RedisConnectionPool.HLen(ctx, getCompletedMuscleGroupsInSessionCountKey(user.Id)).Result()
	exists := !errors.Is(err, redis.Nil)
	if err != nil && exists {
		return 0, fmt.Errorf("error, when attempting to fetch current completed workout count for user %s . Error: %v", user.Email, err)
	}
	if exists {
		result = int(redisResult)
	}
	return result, nil
}

func getTotalMuscleGroupsCount(ctx context.Context) (int, error) {
	var result int
	redisResult, err := config.RedisConnectionPool.Get(ctx, constants.TotalMuscleGroupCountKey).Result()
	exists := !errors.Is(err, redis.Nil)
	if err != nil && exists {
		return 0, fmt.Errorf("error, when attempting to fetch total muscle group count from redis. Error: %v", err)
	}

	if !exists {
		err = config.ConnectionPool.QueryRow(
			ctx,
			"SELECT count(1) FROM muscle_group",
		).Scan(
			&result,
		)
		if err != nil {
			return 0, fmt.Errorf("error, when attempting to execute sql statement: %v", err)
		}
		err = config.RedisConnectionPool.Set(ctx, constants.TotalMuscleGroupCountKey, result, time.Hour).Err()
		if err != nil {
			return 0, fmt.Errorf("error, when attempting to cache the total muscle group count. Error: %v", err)
		}
	} else {
		result, err = strconv.Atoi(redisResult)
		if err != nil {
			return 0, fmt.Errorf("error, when attempting to convert redis result into result: %v", err)
		}
	}
	return result, nil
}

func hasMuscleGroupWorkedSessionLimitBeenReached(totalMuscleGroupsCount int, count int) bool {
	// Adding one before division if totalMuscleGroupsCount be odd to handle ceiling
	halfMuscleGroups := totalMuscleGroupsCount / 2
	if totalMuscleGroupsCount%2 != 0 {
		halfMuscleGroups++
	}

	return halfMuscleGroups <= count
}

func updateSuperSetWithCurrentMeasurement(currentSuperset *model.SuperSet, measurement int) *model.SuperSet {
	currentExercise := currentSuperset.Exercises[currentSuperset.CurrentExercisePointer]
	currentExercise.LastCompletedMeasurement = measurement
	currentSuperset.Exercises[currentSuperset.CurrentExercisePointer] = currentExercise
	return currentSuperset
}

func ShuffleExercise(ctx context.Context) (*model.ExerciseResponse, error) {
	user, err := fetchUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, could not FinishCurrentAndFetchNextExercise() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
	}

	currentSuperset, err := FetchCurrentSuperset(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, when FetchCurrentSuperset() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
	}

	if currentSuperset == nil {
		currentSuperset = &model.SuperSet{}
	}

	var currentExerciseId string
	if currentSuperset.Exercises != nil {
		currentExerciseId = currentSuperset.Exercises[currentSuperset.CurrentExercisePointer].Id
	}

	var muscleGroupsNotInRecovery []model.MuscleGroup
	muscleGroupsNotInRecovery, err = fetchAllMuscleGroupsNotInRecovery(ctx, currentSuperset, true)
	if err != nil {
		return nil, fmt.Errorf("error, when fetchAllMuscleGroupsNotInRecovery() for ShuffleExercise(). Error: %v", err)
	}

	var randomExercise *model.Exercise
	randomExercise, err = fetchRandomExercise(ctx, currentSuperset, currentExerciseId, muscleGroupsNotInRecovery, user.Id)
	if err != nil {
		return nil, fmt.Errorf("error, when fetchRandomExercise() for FinishCurrentAndFetchNextExercise(). Error: %v", err)
	}

	if randomExercise != nil {
		currentSuperset.Exercises[currentSuperset.CurrentExercisePointer] = model.Exercise{
			Id:                       randomExercise.Id,
			LastCompletedMeasurement: randomExercise.LastCompletedMeasurement,
		}
	} else {
		currentSuperset.WorkoutComplete = true
	}

	err = setCurrentSupersetForUser(ctx, user, currentSuperset)
	if err != nil {
		return nil, fmt.Errorf("error, when set current super set after shuffling current exercise for user: %s. Error: %v", user.Email, err)
	}

	response, err := SuperSetToExerciseResponse(ctx, currentSuperset)
	if err != nil {
		return nil, fmt.Errorf("error, when SuperSetToExerciseResponse() for ShuffleExercise(). Exercise: %+v. Error: %v", randomExercise, err)
	}
	return response, nil
}

func fetchRandomExercise(
	ctx context.Context,
	currentSuperset *model.SuperSet,
	currentExerciseId string,
	availableMuscleGroups []model.MuscleGroup,
	userId string,
) (*model.Exercise, error) {
	selectedMuscleGroup := selectRandomMuscleGroup(availableMuscleGroups)
	if selectedMuscleGroup == nil {
		return nil, nil
	}

	exercises, err := fetchAllExercisesForMuscleGroup(ctx, *selectedMuscleGroup)
	if err != nil {
		return nil, fmt.Errorf("error, when fetchAllExercisesForMuscleGroup() for ReadyForNextExercise(). Error: %v", err)
	}

	var selectedExercise *model.Exercise
	uniqueExerciseAttemptLimit := 4
	for i := 0; i < uniqueExerciseAttemptLimit; i++ {
		selectedExercise = selectRandomExercise(exercises)
		if selectedExercise == nil {
			return nil, fmt.Errorf("no exercises available for the %s muscle groups", selectedMuscleGroup.Name)
		}
		if currentSuperset.Exercises == nil {
			currentSuperset.Exercises = make([]model.Exercise, 1)
			break
		}
		if selectedExercise.Id != currentExerciseId {
			break
		}
	}

	measurement, err := fetchLastCompletedExerciseMeasurements(ctx, userId, selectedExercise.Id)
	if err != nil {
		return nil, fmt.Errorf("error, when fetchLastCompletedExerciseMeasurements() for fetchLastCompletedExerciseMeasurements(). Error: %v", err)
	}
	selectedExercise.LastCompletedMeasurement = measurement
	return selectedExercise, nil
}

func fetchLastCompletedExerciseMeasurements(ctx context.Context, userId string, exerciseId string) (int, error) {
	var lastMeasurement int
	err := config.ConnectionPool.QueryRow(
		ctx,
		"SELECT lcm.measurement\nFROM last_completed_measurement lcm\nWHERE user_id = $1\n    AND exercise_id = $2",
		userId,
		exerciseId,
	).Scan(
		&lastMeasurement,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// this just means the user has never completed this exercise before
			return 0, nil
		} else {
			return 0, fmt.Errorf("error, when attempting to execute sql statement: %v", err)
		}
	}
	return lastMeasurement, nil
}

func markPreviousExerciseAsCompleted(currentSuperset *model.SuperSet, numberOfAvailableMuscleGroups int, numberOfExerciseInSuperset int) *model.SuperSet {
	numberOfActiveExercises := len(currentSuperset.Exercises)
	currentExerciseNumber := currentSuperset.CurrentExercisePointer + 1
	if currentExerciseNumber == numberOfExerciseInSuperset || (numberOfAvailableMuscleGroups == 0 && numberOfActiveExercises == currentExerciseNumber) {
		currentSuperset.CurrentExercisePointer = 0
		currentSuperset.SetCompletionCount++
	} else {
		currentSuperset.CurrentExercisePointer++
	}
	return currentSuperset
}

func markCompletionOfCurrentSuperset(ctx context.Context, user *model.User, currentSuperset *model.SuperSet) error {
	err := updateCurrentExerciseMeasurementsForSuperset(ctx, user.Id, currentSuperset)
	if err != nil {
		return fmt.Errorf("error, when attempting to update all super set exercise measurements after superset completion. Error: %v", err)
	}

	muscleGroups, err := fetchAllMuscleGroupsForExercises(ctx, currentSuperset.Exercises)
	if err != nil {
		return fmt.Errorf("error, when fetchAllMuscleGroupsForExercises() for markCompletionOfCurrentSuperset(). Error: %v", err)
	}
	pipeline := config.RedisConnectionPool.Pipeline()
	countKey := getCompletedMuscleGroupsInSessionCountKey(user.Id)
	for _, group := range muscleGroups {
		pipeline.HSet(ctx, countKey, group.Id, 1)
		pipeline.Expire(ctx, countKey, time.Duration(config.CurrentWorkoutExpirationTimeInHours)*time.Hour)

		key := getUserMuscleGroupInRecoveryKey(user.Id, group.Id)
		pipeline.Set(ctx, key, group.Id, config.MuscleGroupRecoveryWindowInHours*time.Hour)
	}
	_, err = pipeline.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error, when attempting mark muscle groups as in recovery mode for user %s. Error: %v", user.Email, err)
	}

	key := getCurrentSupersetForUserKey(user.Id)
	err = config.RedisConnectionPool.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error, when attempting to delect current superset after completion of current superset. Error: %v", err)
	}
	return nil
}

func updateCurrentExerciseMeasurementsForSuperset(ctx context.Context, userId string, superset *model.SuperSet) error {
	var valueStrings []string
	var valueArgs []interface{}

	for i, exercise := range superset.Exercises {
		// Create the placeholder for this row
		ph := fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)

		// Append the placeholder string to our slice
		valueStrings = append(valueStrings, ph)

		// Append the actual values to our slice
		valueArgs = append(valueArgs, exercise.LastCompletedMeasurement, userId, exercise.Id)
	}

	// Create the base query string with placeholders
	query := fmt.Sprintf(
		"INSERT INTO last_completed_measurement (measurement, user_id, exercise_id) VALUES %s ON CONFLICT (user_id, exercise_id) DO UPDATE SET measurement = EXCLUDED.measurement;",
		strings.Join(valueStrings, ","),
	)

	// Execute the query
	_, err := config.ConnectionPool.Exec(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("error, when executing query to create init table: %v", err)
	}

	//tx, err := config.ConnectionPool.Begin(ctx)
	//if err != nil {
	//	return fmt.Errorf("error, when attempting to start a transaction: %v", err)
	//}
	//
	//err = func() error {
	//	for _, exercise := range superset.Exercises {
	//		_, err = tx.Exec(
	//			ctx,
	//			"INSERT INTO last_completed_measurement (user_id, exercise_id, measurement)\nVALUES ($1, $2, $3)\nON CONFLICT (user_id, exercise_id)\nDO UPDATE SET measurement = EXCLUDED.measurement;\n",
	//			exercise.LastCompletedMeasurement,
	//			userId,
	//			exercise.Id,
	//		)
	//		if err != nil {
	//			return fmt.Errorf("error, when executing query to create init table: %v", err)
	//		}
	//	}
	//	return nil
	//}()
	//if err != nil {
	//	rollBackErr := tx.Rollback(ctx)
	//	if rollBackErr != nil {
	//		return fmt.Errorf("error, when attempting to roll back commit: Rollback Error: %v, Original Error: %v", rollBackErr, err)
	//	}
	//	return fmt.Errorf("error, when attempting to perform database transaction: %v", err)
	//}
	//err = tx.Commit(ctx)
	//if err != nil {
	//	return fmt.Errorf("error, when attempting to commit the transaction to the database: %v", err)
	//}
	return nil
}

func fetchAllMuscleGroupsForExercises(ctx context.Context, exercises []model.Exercise) (map[string]model.MuscleGroup, error) {
	muscleGroups := make(map[string]model.MuscleGroup)
	if len(exercises) == 0 {
		return muscleGroups, nil
	}

	var exerciseIds []string
	for _, e := range exercises {
		exerciseIds = append(exerciseIds, e.Id)
	}

	// todo cache this in redis to save on hitting the database
	exercisesArgs, insertValues := getExerciseArgsAndInsertValues(exerciseIds)
	rows, err := config.ConnectionPool.Query(
		ctx,
		fmt.Sprintf("SELECT id, name\nFROM muscle_group\nJOIN exercise_muscle_group ON muscle_group.id = exercise_muscle_group.muscle_group_id\nWHERE exercise_muscle_group.exercise_id IN (%s)", exercisesArgs),
		insertValues...,
	)
	defer rows.Close()

	for rows.Next() {
		var muscleGroup model.MuscleGroup
		err = rows.Scan(
			&muscleGroup.Id,
			&muscleGroup.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when scanning database rows: %v", err)
		}
		muscleGroups[muscleGroup.Id] = muscleGroup
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error, when iterating through database rows: %v", err)
	}
	return muscleGroups, nil
}

func getExerciseArgsAndInsertValues(exerciseIds []string) (string, []any) {
	var exercisesArgsSlice []string
	var insertValues []any
	for i, exerciseId := range exerciseIds {
		exercisesArgsSlice = append(exercisesArgsSlice, fmt.Sprintf("$%d", i+1))
		insertValues = append(insertValues, exerciseId)
	}
	return strings.Join(exercisesArgsSlice, ", "), insertValues
}

func setCurrentSupersetForUser(ctx context.Context, user *model.User, superSet *model.SuperSet) error {
	bytes, err := json.Marshal(superSet)
	if err != nil {
		return fmt.Errorf("error, when attempting to marshal the current superset into json. Error: %v", err)
	}

	key := getCurrentSupersetForUserKey(user.Id)
	err = config.RedisConnectionPool.Set(ctx, key, string(bytes), time.Duration(config.CurrentSupersetExpirationTimeInHours)*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error, when attempting to set redis key: %s. Error: %v", key, err)
	}
	return nil
}

func selectRandomMuscleGroup(availableMuscleGroups []model.MuscleGroup) *model.MuscleGroup {
	muscleGroupCount := len(availableMuscleGroups)
	if muscleGroupCount == 0 {
		return nil
	}
	result := availableMuscleGroups[rand.Intn(muscleGroupCount)]
	return &result
}

func selectRandomExercise(availableExercises []model.Exercise) *model.Exercise {
	exerciseCount := len(availableExercises)
	if exerciseCount == 0 {
		return nil
	}
	result := availableExercises[rand.Intn(exerciseCount)]
	return &result
}

func fetchAllMuscleGroups(ctx context.Context) ([]model.MuscleGroup, error) {
	var muscleGroups []model.MuscleGroup
	muscleGroupsFromRedis, err := config.RedisConnectionPool.Get(ctx, constants.CachedMuscleGroupsKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var rows pgx.Rows
			rows, err = config.ConnectionPool.Query(
				ctx,
				"SELECT id, name FROM muscle_group",
			)
			defer rows.Close()

			for rows.Next() {
				var muscleGroup model.MuscleGroup
				err = rows.Scan(
					&muscleGroup.Id,
					&muscleGroup.Name,
				)
				if err != nil {
					return nil, fmt.Errorf("error, when scanning database rows: %v", err)
				}
				muscleGroups = append(muscleGroups, muscleGroup)
			}
			err = rows.Err()
			if err != nil {
				return nil, fmt.Errorf("error, when iterating through database rows: %v", err)
			}
			var bytes []byte
			bytes, err = json.Marshal(&muscleGroups)
			if err != nil {
				return nil, fmt.Errorf("error, when marshalling musclegroups for caching in redis. Error: %v", err)
			}
			err = config.RedisConnectionPool.Set(ctx, constants.CachedMuscleGroupsKey, bytes, time.Hour).Err()
			if err != nil {
				return nil, fmt.Errorf("error, when attempting to cache muscle groups in redis. Error: %v", err)
			}
			return muscleGroups, nil
		} else {
			return nil, fmt.Errorf("error, when attempting to fetch muscle groups from redis. Error: %v", err)
		}
	}
	err = json.Unmarshal([]byte(muscleGroupsFromRedis), &muscleGroups)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to unmarshal muscle groups from redis. Error: %v", err)
	}
	return muscleGroups, nil
}

func getUserMuscleGroupInRecoveryKey(userId string, muscleGroupId string) string {
	return fmt.Sprintf("%s:%s", userId, muscleGroupId)
}

func getCompletedMuscleGroupsInSessionCountKey(userId string) string {
	return fmt.Sprintf("%s:%s", constants.MuscleGroupsCompletedInSessionKey, userId)
}

func getCurrentSupersetForUserKey(userId string) string {
	return fmt.Sprintf("%s%s", constants.CurrentSupersetPrefix, userId)
}

func fetchAllMuscleGroupsNotInRecovery(ctx context.Context, currentSuperset *model.SuperSet, shuffle bool) ([]model.MuscleGroup, error) {
	// todo find a way to get these muscle groups from redis or some other cache to save on calls to the backend
	muscleGroupsAlreadyActiveInCurrentSuperset, err := fetchAllMuscleGroupsForExercises(ctx, currentSuperset.Exercises)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting fetchAllMuscleGroupsForExercises() for fetchAllMuscleGroupsNotInRecovery(). Exercise Ids: %+v. Error: %v", currentSuperset.Exercises, err)
	}

	allMuscleGroups, err := fetchAllMuscleGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, when fetchAllMuscleGroups() for fetchMuscleGroupsCurrentlyInRecovery(). Error: %v", err)
	}

	user, err := fetchUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, could not fetchUserFromContext() for fetchAllMuscleGroupsNotInRecovery(). Error: %v", err)
	}

	pipe := config.RedisConnectionPool.Pipeline()
	redisResults := make([]*redis.IntCmd, 0, len(allMuscleGroups))
	for _, group := range allMuscleGroups {
		redisResults = append(redisResults, pipe.Exists(ctx, getUserMuscleGroupInRecoveryKey(user.Id, group.Id)))
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, when executing pipeline command to redis. Error: %v", err)
	}

	var results []model.MuscleGroup
	for i, rr := range redisResults {
		exists, _ := rr.Result()
		// if no muscle group recovery entry exists in redis then it is considered available for the current workout
		if exists == 0 {
			muscleGroup := allMuscleGroups[i]
			// muscle groups that are already in the super set should not be selected again unless it is a shuffle operation
			_, ok := muscleGroupsAlreadyActiveInCurrentSuperset[muscleGroup.Id]
			if !ok || shuffle {
				results = append(results, muscleGroup)
			}
		}
	}
	return results, nil
}

func fetchAllExercisesForMuscleGroup(ctx context.Context, muscleGroup model.MuscleGroup) ([]model.Exercise, error) {
	// todo implement redis caching to save on queries to the database
	rows, err := config.ConnectionPool.Query(
		ctx,
		"SELECT id, name, demonstration_giphy_id\nFROM exercise\nJOIN exercise_muscle_group emg on exercise.id = emg.exercise_id\nWHERE muscle_group_id = $1",
		muscleGroup.Id,
	)
	defer rows.Close()

	if err != nil {
		return nil, fmt.Errorf("error, when attempting to retrieve records. Error: %v", err)
	}

	var exercises []model.Exercise
	for rows.Next() {
		var exercise model.Exercise
		err = rows.Scan(
			&exercise.Id,
			&exercise.Name,
			&exercise.DemonstrationGiphyId,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when scanning database rows: %v", err)
		}
		exercises = append(exercises, exercise)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error, when iterating through database rows: %v", err)
	}
	return exercises, nil
}

func FetchCurrentExercise(ctx context.Context) (*model.ExerciseResponse, error) {
	var result model.ExerciseResponse
	currentSuperset, err := FetchCurrentSuperset(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to FetchCurrentSuperset() for FetchCurrentExercise(). Error: %v", err)
	}
	if currentSuperset == nil {
		return nil, nil
	}

	// todo confirm if this optimization step is even worth the extra complexity
	if currentSuperset.WorkoutComplete == true {
		result.WorkoutComplete = true
		return &result, nil
	}

	response, err := SuperSetToExerciseResponse(ctx, currentSuperset)
	if err != nil {
		return nil, fmt.Errorf("error, when SuperSetToExerciseResponse() for FetchCurrentExercise(). Error: %v", err)
	}
	return response, nil
}

func FetchCurrentSuperset(ctx context.Context) (*model.SuperSet, error) {
	user, err := fetchUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error, could not FetchCurrentSuperset() for fetchAllMuscleGroupsNotInRecovery(). Error: %v", err)
	}

	key := getCurrentSupersetForUserKey(user.Id)

	result, err := config.RedisConnectionPool.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("error, when attempting to fetch the current exercise for user: %s", user.Email)
	}

	parsedResult := &model.SuperSet{}
	err = json.Unmarshal([]byte(result), parsedResult)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to unmarshall the current super set from json to a struct. Error: %v", err)
	}

	return parsedResult, nil
}

func FetchExercise(ctx context.Context, exerciseId string) (*model.Exercise, error) {
	var exercise model.Exercise
	key := fmt.Sprintf("%s%s", constants.CachedExercisePrefix, exerciseId)
	result, err := config.RedisConnectionPool.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = config.ConnectionPool.QueryRow(
				ctx,
				"SELECT id, name, demonstration_giphy_id, measurement_type_id FROM exercise WHERE id = $1",
				exerciseId,
			).Scan(
				&exercise.Id,
				&exercise.Name,
				&exercise.DemonstrationGiphyId,
				&exercise.MeasurementType,
			)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					// todo implement what happens when an exercise is removed but the user had it selected
					return nil, fmt.Errorf("error, not yet implmented: %v", err)
				} else {
					return nil, fmt.Errorf("error, when attempting to execute sql statement: %v", err)
				}
			}
			var exerciseJson []byte
			exerciseJson, err = json.Marshal(&exercise)
			if err != nil {
				return nil, err
			}
			err = config.RedisConnectionPool.Set(ctx, key, exerciseJson, time.Hour).Err()
			if err != nil {
				return nil, fmt.Errorf("error, when attempting to cache exercise in redis. Error: %v", err)
			}
		} else {
			return nil, fmt.Errorf("error, when attempting to fetch cached exercise from redis. Error: %v", err)
		}
	} else {
		err = json.Unmarshal([]byte(result), &exercise)
		if err != nil {
			return nil, fmt.Errorf("error, when unmarshalling exercise result from redis. Error: %v", err)
		}
	}

	return &exercise, nil
}

func SuperSetToExerciseResponse(ctx context.Context, set *model.SuperSet) (*model.ExerciseResponse, error) {
	var exercise *model.Exercise
	var err error
	if len(set.Exercises) != 0 {
		currentExercise := set.Exercises[set.CurrentExercisePointer]
		exercise, err = FetchExercise(ctx, currentExercise.Id)
		if err != nil {
			return nil, fmt.Errorf("error, when service.FetchExercise() for SuperSetToExerciseResponse(). Exercise Ids: %+v. Current Exercise Pointer: %d. Error: %v", set.Exercises, set.CurrentExercisePointer, err)
		}
		exercise.LastCompletedMeasurement = currentExercise.LastCompletedMeasurement
	}

	return &model.ExerciseResponse{
		Exercise:         exercise,
		SuperSetProgress: set.SuperSetProgress,
	}, nil
}
