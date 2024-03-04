package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/model"
	"strings"
)

func RecordIncrementedWorkoutStep(ctx context.Context, req model.RecordIncrementedWorkoutStepRequest) *model.Error {
	// todo if the server progress index is ahead of the client progress index, then we need to return current workout because this indicates the user switched to an older client we may just want to implement a sync mechanism that is triggered on gain of focus to handle these types of out of sync problems in bulk, so I don't have to address every use-case when they pop up
	var us model.UserService
	user, err := us.FetchFromContext(ctx)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when UserService.FetchFromContext() for RecordIncrementedWorkoutStep(). Error: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}

	userWorkout := model.UserWorkout{}
	// todo consider just fetching the workout routine to make this more efficient
	err = userWorkout.FromRedis(ctx, user.Id, config.RedisConnectionPool)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, unable to fetch user workout from redis. Error: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}

	if !userWorkout.Exists || userWorkout.WorkoutId != req.WorkoutId {
		return &model.Error{
			InternalError:     fmt.Errorf("error, user %s attempted to fetch an user workout that expired for RecordIncrementedWorkoutStep()", user.Id),
			UserFeedbackError: model.ErrorClientOutOfSync,
		}
	}

	if haveMainExercisesJustBeenCompleted(req.IncrementedProgressIndex) {
		// We are incrementing based on the value in redis in case this command gets sent more than once
		// using redis as the source of truth makes incrementing the workout routine idempotent.
		nextWorkoutRoutine := userWorkout.WorkoutRoutine.GetNextRoutine()

		_, err = config.ConnectionPool.Exec(ctx,
			`UPDATE "user"
			SET current_routine = $1
			WHERE id = $2;`,
			nextWorkoutRoutine,
			user.Id,
		)
		if err != nil {
			return &model.Error{
				InternalError:     fmt.Errorf("error, updating current routine in database for user: %s. Error: %v", user.Id, err),
				UserFeedbackError: model.ErrorUnexpectedTryAgain,
			}
		}
	} else if hasWorkoutBeenCompleted(req.IncrementedProgressIndex) {
		err = updateUserExerciseData(ctx, user)
		if err != nil {
			return &model.Error{
				InternalError:     fmt.Errorf("error, updating exercise data for user due to workout completeion: %s. Error: %v", user.Id, err),
				UserFeedbackError: model.ErrorUnexpectedTryAgain,
			}
		}
	}

	if req.LastCompletedMeasurement > 0 {
		err = updateExerciseUserData(ctx, req, user)
		if err != nil {
			return &model.Error{
				InternalError:     fmt.Errorf("error, updating exercise user data for user: %s. Error: %v", user.Id, err),
				UserFeedbackError: model.ErrorUnexpectedTryAgain,
			}
		}
	}

	serializedProgressIndex, err := json.Marshal(req.IncrementedProgressIndex)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when marshalling user workout progress index for redis. Error: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}
	err = config.RedisConnectionPool.Set(
		ctx,
		model.GetUserKey(user.Id, model.WorkoutProgressIndexKey),
		serializedProgressIndex,
		config.GetSuperSetExpiration(),
	).Err()
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting a mutation in redis for RecordIncrementedWorkoutStep(). Error: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}
	return nil
}

func updateUserExerciseData(ctx context.Context, user *model.User) error {
	redisMap, err := config.RedisConnectionPool.HGetAll(
		ctx,
		model.GetUserKey(user.Id, model.UserExerciseUserDataKey),
	).Result()
	if err != nil {
		return fmt.Errorf("error, when retrieving updated user exercise user data from redis. Error: %v", err)
	}

	// key is exercise id
	exerciseUserData := make(map[string]model.ExerciseUserData)
	for k, v := range redisMap {
		eud := model.ExerciseUserData{}
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
	_, err = config.ConnectionPool.Exec(
		ctx,
		sqlStatement,
		args...,
	)
	if err != nil {
		return fmt.Errorf("error, updating exercise user data in database for user: %s. Error: %v", user.Id, err)
	}
	return nil
}

func generateQueryForExerciseUserData(exerciseUserData map[string]model.ExerciseUserData, user *model.User) (sqlStatement string, args []any) {
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

func hasWorkoutBeenCompleted(index model.WorkoutProgressIndex) bool {
	return len(index)-1 == model.WorkoutPhaseCompleted
}

func updateExerciseUserData(ctx context.Context, req model.RecordIncrementedWorkoutStepRequest, user *model.User) error {
	key := model.GetUserKey(user.Id, model.UserExerciseUserDataKey)
	redisString, err := config.RedisConnectionPool.HGet(
		ctx,
		key,
		req.ExerciseId,
	).Result()
	if err != nil {
		return fmt.Errorf("error, unable to get exercise user data from redis for exercise: %s. Error: %v", req.ExerciseId, err)
	}
	var value model.ExerciseUserData
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
	err = config.RedisConnectionPool.HSet(
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

func haveMainExercisesJustBeenCompleted(incrementedProgressIndex model.WorkoutProgressIndex) bool {
	return len(incrementedProgressIndex)-1 == model.WorkoutPhaseCoolDown &&
		incrementedProgressIndex[model.WorkoutPhaseCoolDown] == 0
}
