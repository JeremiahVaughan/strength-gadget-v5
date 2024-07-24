package main

import (
	"fmt"
)

type swapExerciseRequest struct {
	ExerciseId string `json:"exerciseId"`
	WorkoutId  string `json:"workoutId"`
}

func validateSwapExerciseRequest(req *swapExerciseRequest) error {
	var errorFeedback []error
	if req.ExerciseId == "" {
		errorFeedback = append(errorFeedback, fmt.Errorf("must provide exercise Id"))
	}
	if req.WorkoutId == "" {
		errorFeedback = append(errorFeedback, fmt.Errorf("must provide workout Id"))
	}
	if len(errorFeedback) > 0 {
		return fmt.Errorf("errors, when validating request: %v", errorFeedback)
	}
	return nil
}

