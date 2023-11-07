package test_case

import (
	"fmt"
	"log"
	"net/http"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/test_tornado/service"
)

func CanDoWorkout() error {
	email, sessionCookie, err := newUser()
	if err != nil {
		return fmt.Errorf("error, when newUser() for CanDoWorkout(). Error: %v", err)
	}
	log.Printf("user created for CanDoWorkout. Email: %s", email)

	exercise, err := attemptToGetCurrentExercise(sessionCookie)
	if err != nil {
		return fmt.Errorf("error, when attempting to fetch exercise right after logging in for the first time. Cookie: %+v. Exercise: %+v, Error: %v", sessionCookie, exercise, err)
	}

	exercise, err = attemptToGetCurrentExercise(sessionCookie)
	if err != nil {
		return fmt.Errorf("error, when attempting to fetch exercise right after logging in to prove this call is idempotent. Cookie: %+v. Exercise: %+v, Error: %v", sessionCookie, exercise, err)
	}

	exerciseFirst, err := attemptToShuffleCurrentExercise(sessionCookie)
	if err != nil || exercise == nil {
		return fmt.Errorf("error, when attempting shuffle the first exercise. Cookie: %+v. Payload: %+v, Error: %v", sessionCookie, exerciseFirst, err)
	}

	exerciseSecond, err := attemptToShuffleCurrentExercise(sessionCookie)
	if err != nil || exerciseSecond == nil || exerciseFirst.Id == exerciseSecond.Id {
		return fmt.Errorf("error, when attempting shuffle the second exercise to prove this call is idempotent. Cookie: %+v. First exercise: %s, Second Exercise: %s. Error: %v", sessionCookie, exerciseFirst.Name, exerciseSecond.Name, err)
	}

	var nextExerciseResponse *model.ExerciseResponse
	for i := 0; i < 60; i++ {
		nextExerciseResponse, err = attemptToGetNextExercise(sessionCookie)
		if err != nil {
			return fmt.Errorf("error, when attempting to fetch the next %d exercise. Cookie: %+v. Error: %v", i+1, sessionCookie, err)
		}
	}
	if !nextExerciseResponse.WorkoutComplete {
		return fmt.Errorf("error, expected all muscle groups to be in recovery")
	}

	return nil
}

func newUser() (string, *http.Cookie, error) {
	validEmail, responseCode, err := attemptWithValidCredentials(constants.Register)
	if responseCode != http.StatusOK {
		return "", nil, fmt.Errorf("error, when attempt to register for newUser(). Response Code: %d. Error: %s", responseCode, err)
	}
	response, err := attemptVerificationWithValidCode(validEmail)
	if response.ResponseCode != http.StatusOK {
		return "", nil, fmt.Errorf("error, when attempting verification with valid verification code for newUser(). Error: %v. Response code: %d", err, response.ResponseCode)
	}
	return validEmail, response.Cookie, nil
}

func attemptToGetCurrentExercise(cookie *http.Cookie) (*model.Exercise, error) {
	response, err := service.TestRequest(
		http.MethodGet,
		constants.CurrentExercise,
		nil,
		cookie,
		&model.Exercise{},
	)
	if err != nil {
		return nil, fmt.Errorf("error, when service.TestRequest() for attemptToGetCurrentExercise(). Error: %v", err)
	}
	result := response.ResponsePayload.(*model.Exercise)
	return result, err
}

func attemptToShuffleCurrentExercise(cookie *http.Cookie) (*model.Exercise, error) {
	response, err := service.TestRequest(
		http.MethodGet,
		constants.ShuffleExercise,
		nil,
		cookie,
		&model.Exercise{},
	)
	if err != nil {
		return nil, fmt.Errorf("error, when service.TestRequest() for attemptToShuffleCurrentExercise(). Error: %v", err)
	}
	result := response.ResponsePayload.(*model.Exercise)
	return result, err
}

func attemptToGetNextExercise(cookie *http.Cookie) (*model.ExerciseResponse, error) {
	response, err := service.TestRequest(
		http.MethodGet,
		constants.ReadyForNextExercise,
		nil,
		cookie,
		&model.ExerciseResponse{},
	)
	if err != nil {
		return nil, fmt.Errorf("error, when service.TestRequest() for attemptToGetNextExercise(). Error: %v", err)
	}
	result := response.ResponsePayload.(*model.ExerciseResponse)
	return result, err
}
