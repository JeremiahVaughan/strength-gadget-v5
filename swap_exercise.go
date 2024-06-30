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

// func HandleSwapExercise(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPut {
// 		http.Error(w, "error, only PUT method is supported", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
// 		return
// 	}

// 	var req swapExerciseRequest
// 	err = json.Unmarshal(body, &req)
// 	if err != nil {
// 		http.Error(w, "error, failed to parse JSON", http.StatusBadRequest)
// 		return
// 	}
// 	err = r.Body.Close()
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("error, when attempting to close request body: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	err = validateSwapExerciseRequest(&req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	result, err := SwapExercise(
// 		r.Context(),
// 		RedisConnectionPool,
// 		ConnectionPool,
// 		req.ExerciseId,
// 		req.WorkoutId,
// 		NumberOfSetsInSuperSet,
// 		NumberOfExerciseInSuperset,
// 		CurrentSupersetExpirationTimeInHours,
// 	)
// 	if err != nil {
// 		HandleUnexpectedError(w, fmt.Errorf("error, when SwapExercise() for HandleSwapExercise(). Error: %v", err))
// 		// todo return user feedback
// 	}

// 	log.Print(result) // todo remove
// 	// todo return template
// }
