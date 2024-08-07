package main

import (
	"context"
	"fmt"
	"net/http"
)

func HandleWorkoutComplete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userSession, err := FetchUserSession(r)
	if err != nil {
		err = fmt.Errorf("error, when FetchUserSession() for HandleWorkoutComplete(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !userSession.Authenticated {
		// HX-Redirect only works if the page has already been loaded so we have to use full redirect instead
		http.Redirect(w, r, EndpointLogin, http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		type Completed struct {
			NewWorkout            Button
		}
		completed := Completed{
			NewWorkout: Button{
				Id:    "new_workout",
				Label: "New Workout",
				Color: PrimaryButtonColor,
				Type:  ButtonTypeSubmit,
			},
		}
		err = templateMap["workout-completed-page.html"].ExecuteTemplate(w, "base", completed)
		if err != nil {
			err = fmt.Errorf("error, when executing workout completed template. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	case http.MethodPost:
		var currentWorkoutRoutine RoutineType
		currentWorkoutRoutine, err = fetchCurrentWorkoutRoutine(ctx, ConnectionPool, userSession.UserId)
		if err != nil {
			err = fmt.Errorf("error, when attempting to fetchCurrentWorkoutRoutine() for HandleWorkoutComplete(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		userSession.WorkoutSession, err = createNewWorkout(ctx, userSession.UserId, currentWorkoutRoutine)
		if err != nil {
			err = fmt.Errorf("error, when createNewWorkout() for HandleWorkoutComplete(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		userSession.WorkoutSessionExists = true
		redirectExercisePage(w, r, userSession)
	}
}

func persistWorkoutRoutine(ctx context.Context, userId int64, routineType RoutineType) error {
	_, err := ConnectionPool.Exec(
		ctx,
		`UPDATE athlete 
		SET current_routine = $1
		WHERE id = $2`,
		routineType,
		userId,
	)
	if err != nil {
		return fmt.Errorf("error, when attempting to persist a request for a stress test. Error: %v", err)
	}
	return nil
}
