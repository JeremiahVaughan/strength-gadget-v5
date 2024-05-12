package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func validateRecordIncrementedWorkoutStepRequest(req *RecordIncrementedWorkoutStepRequest) error {
	var errorFeedback []error
	if !req.IncrementedProgressIndex.IsValid() {
		errorFeedback = append(errorFeedback, errors.New("must provide between 1 inclusive and 4 inclusive for the workout phase"))
	}
	if req.ExerciseId == "" {
		errorFeedback = append(errorFeedback, errors.New("exerciseId is required"))
	}

	if req.WorkoutId == "" {
		errorFeedback = append(errorFeedback, errors.New("workoutId is required"))
	}

	if !IsValidUUID(req.WorkoutId) {
		errorFeedback = append(errorFeedback, errors.New("workoutId is not a valid UUID"))
	}

	if len(errorFeedback) > 0 {
		return fmt.Errorf("errors, when validating request: %v", errorFeedback)
	}
	return nil
}

func HandleRecordIncrementedWorkoutStep(w http.ResponseWriter, r *http.Request) {
	// todo chi is already handling the method check so this is redundant
	if r.Method != http.MethodPut {
		http.Error(w, "error, only PUT method is supported", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
		return
	}

	var req RecordIncrementedWorkoutStepRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "error, failed to parse JSON", http.StatusBadRequest)
		return
	}
	err = r.Body.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("error, when attempting to close request body: %v", err), http.StatusInternalServerError)
		return
	}

	err = validateRecordIncrementedWorkoutStepRequest(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = RecordIncrementedWorkoutStep(r.Context(), req)
	if err != nil {
		HandleUnexpectedError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func RecordIncrementedWorkoutStep(ctx context.Context, req RecordIncrementedWorkoutStepRequest) error {
	// todo if the server progress index is ahead of the client progress index, then we need to return current workout because this indicates the user switched to an older client we may just want to implement a sync mechanism that is triggered on gain of focus to handle these types of out of sync problems in bulk, so I don't have to address every use-case when they pop up
	var us UserService
	user, err := us.FetchFromContext(ctx)
	if err != nil {
		return fmt.Errorf("error, when UserService.FetchFromContext() for RecordIncrementedWorkoutStep(). Error: %v", err)
	}

	userWorkout := UserWorkout{}
	// todo consider just fetching the workout routine to make this more efficient
	err = userWorkout.FromRedis(ctx, user.Id, RedisConnectionPool)
	if err != nil {
		return fmt.Errorf("error, unable to fetch user workout from redis. Error: %v", err)
	}

	if !userWorkout.Exists || userWorkout.WorkoutId != req.WorkoutId {
		return fmt.Errorf("error, user %s attempted to fetch an user workout that expired for RecordIncrementedWorkoutStep()", user.Id)
	}

	if haveMainExercisesJustBeenCompleted(req.IncrementedProgressIndex) {
		// We are incrementing based on the value in redis in case this command gets sent more than once
		// using redis as the source of truth makes incrementing the workout routine idempotent.
		nextWorkoutRoutine := userWorkout.WorkoutRoutine.GetNextRoutine()

		_, err = ConnectionPool.Exec(ctx,
			`UPDATE "user"
			SET current_routine = $1
			WHERE id = $2;`,
			nextWorkoutRoutine,
			user.Id,
		)
		if err != nil {
			return fmt.Errorf("error, updating current routine in database for user: %s. Error: %v", user.Id, err)
		}
	} else if hasWorkoutBeenCompleted(req.IncrementedProgressIndex) {
		err = updateUserExerciseData(ctx, user)
		if err != nil {
			return fmt.Errorf("error, updating exercise data for user due to workout completeion: %s. Error: %v", user.Id, err)
		}
	}

	if req.LastCompletedMeasurement > 0 {
		err = updateExerciseUserData(ctx, req, user)
		if err != nil {
			return fmt.Errorf("error, updating exercise user data for user: %s. Error: %v", user.Id, err)
		}
	}

	serializedProgressIndex, err := json.Marshal(req.IncrementedProgressIndex)
	if err != nil {
		return fmt.Errorf("error, when marshalling user workout progress index for redis. Error: %v", err)
	}
	err = RedisConnectionPool.Set(
		ctx,
		GetUserKey(user.Id, WorkoutProgressIndexKey),
		serializedProgressIndex,
		GetSuperSetExpiration(),
	).Err()
	if err != nil {
		return fmt.Errorf("error, when attempting a mutation in redis for RecordIncrementedWorkoutStep(). Error: %v", err)
	}

	return nil
}

func updateUserExerciseData(ctx context.Context, user *User) error {
	redisMap, err := RedisConnectionPool.HGetAll(
		ctx,
		GetUserKey(user.Id, UserExerciseUserDataKey),
	).Result()
	if err != nil {
		return fmt.Errorf("error, when retrieving updated user exercise user data from redis. Error: %v", err)
	}

	// key is exercise id
	exerciseUserData := make(map[string]ExerciseUserData)
	for k, v := range redisMap {
		eud := ExerciseUserData{}
		err = json.Unmarshal([]byte(v), &eud)
		if err != nil {
			return fmt.Errorf("error, when unmarshalling exercise user data from redis. Error: %v", err)
		}
		exerciseUserData[k] = eud
	}

	if len(exerciseUserData) == 0 {
		return fmt.Errorf("error, empty exerciseUserData was not expected. User: %s", user.Id)
	}

	sqlStatement, args := generateQueryForExerciseUserData(exerciseUserData, user)
	_, err = ConnectionPool.Exec(
		ctx,
		sqlStatement,
		args...,
	)
	if err != nil {
		return fmt.Errorf("error, updating exercise user data in database for user: %s. Error: %v", user.Id, err)
	}
	return nil
}

func generateQueryForExerciseUserData(exerciseUserData map[string]ExerciseUserData, user *User) (sqlStatement string, args []any) {
	var placeHolderCounter int
	var placeHolders []string
	for exerciseId, exerciseData := range exerciseUserData {
		var ph []string
		ph = []string{}
		for i := 0; i < 3; i++ {
			placeHolderCounter++
			ph = append(ph, fmt.Sprintf("$%d", placeHolderCounter))
		}

		queryPart := strings.Join(ph, ", ")
		placeHolders = append(placeHolders, fmt.Sprintf("(%s)", queryPart))

		args = append(args, user.Id, exerciseId, exerciseData.Measurement)
	}
	sqlStatement = fmt.Sprintf(`
		INSERT INTO last_completed_measurement (user_id, exercise_id, measurement)
		VALUES 
			%s
		ON CONFLICT (user_id, exercise_id) DO UPDATE 
		SET 
			measurement = excluded.measurement`,
		strings.Join(placeHolders, ",\n"),
	)
	return sqlStatement, args
}

func hasWorkoutBeenCompleted(index WorkoutProgressIndex) bool {
	return len(index)-1 == WorkoutPhaseCompleted
}

func updateExerciseUserData(ctx context.Context, req RecordIncrementedWorkoutStepRequest, user *User) error {
	key := GetUserKey(user.Id, UserExerciseUserDataKey)
	redisString, err := RedisConnectionPool.HGet(
		ctx,
		key,
		req.ExerciseId,
	).Result()
	if err != nil {
		return fmt.Errorf("error, unable to get exercise user data from redis for exercise: %s. Error: %v", req.ExerciseId, err)
	}
	var value ExerciseUserData
	err = json.Unmarshal([]byte(redisString), &value)
	if err != nil {
		return fmt.Errorf("error, unable to unmarshal exercise user data from redis. Error: %v", err)
	}
	value.Measurement = req.LastCompletedMeasurement
	var redisBytes []byte
	redisBytes, err = json.Marshal(&value)
	if err != nil {
		return fmt.Errorf("error, unable to marshal exercise user data for redis. Error: %v", err)
	}
	err = RedisConnectionPool.HSet(
		ctx,
		key,
		req.ExerciseId,
		redisBytes,
	).Err()
	if err != nil {
		return fmt.Errorf("error, unable to set exercise user data in redis. Error: %v", err)
	}
	return nil
}

func haveMainExercisesJustBeenCompleted(incrementedProgressIndex WorkoutProgressIndex) bool {
	return len(incrementedProgressIndex)-1 == WorkoutPhaseCoolDown &&
		incrementedProgressIndex[WorkoutPhaseCoolDown] == 0
}
