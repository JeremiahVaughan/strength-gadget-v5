package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type NewWorkoutRequest struct {
	CurrentRoutine RoutineType
}

func validateNewWorkoutRequest(req *NewWorkoutRequest) error {
	var errorFeedback []error
	if req.CurrentRoutine != LOWER && req.CurrentRoutine != CORE && req.CurrentRoutine != UPPER {
		errorFeedback = append(errorFeedback, errors.New("invalid routine type"))
	}
	if len(errorFeedback) > 0 {
		return fmt.Errorf("errors, when validating request: %v", errorFeedback)
	}
	return nil
}

func HandleWorkoutComplete(w http.ResponseWriter, r *http.Request) {
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

	err = r.ParseForm()
	if err != nil {
		err = fmt.Errorf("error, when parsing form for HandleWorkoutComplete(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	current := r.FormValue("currentWorkoutRoutine")
	currentWorkoutRoutine, err := strconv.Atoi(current)
	if err != nil {
		err = fmt.Errorf("error, when attempting to convert currentWorkoutRoutine from a string to a number. Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	req := NewWorkoutRequest{
		CurrentRoutine: RoutineType(currentWorkoutRoutine),
	}
	err = validateNewWorkoutRequest(&req)
	if err != nil {
		err = fmt.Errorf("error, when validateNewWorkoutRequest() for HandleWorkoutComplete(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		type Completed struct {
			NewWorkout            Button
			CurrentWorkoutRoutine RoutineType
		}
		completed := Completed{
			NewWorkout: Button{
				Id:    "new_workout",
				Label: "New Workout",
				Color: PrimaryButtonColor,
				Type:  ButtonTypeSubmit,
			},
			CurrentWorkoutRoutine: req.CurrentRoutine,
		}
		err = templateMap["workout-completed-page.html"].ExecuteTemplate(w, "base", completed)
		if err != nil {
			err = fmt.Errorf("error, when executing workout completed template. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	case http.MethodPost:
		nextRoutine := req.CurrentRoutine.GetNextRoutine()
		err = persistWorkoutRoutine(r.Context(), userSession.UserId, nextRoutine)
		if err != nil {
			err = fmt.Errorf("error, when persistWorkoutRoutine() for HandleWorkoutComplete(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		userSession.WorkoutSession, err = createNewWorkout(r.Context(), userSession.UserId, nextRoutine)
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
