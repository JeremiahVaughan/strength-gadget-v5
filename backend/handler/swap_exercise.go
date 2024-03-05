package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
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

func HandleSwapExercise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "error, only PUT method is supported", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
		return
	}

	var req swapExerciseRequest
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

	err = validateSwapExerciseRequest(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, userError := model.SwapExercise(
		r.Context(),
		config.RedisConnectionPool,
		config.ConnectionPool,
		req.ExerciseId,
		req.WorkoutId,
		config.NumberOfSetsInSuperSet,
		config.NumberOfExerciseInSuperset,
		config.CurrentSupersetExpirationTimeInHours,
	)
	if userError != nil {
		service.GenerateResponse(w, userError)
		return
	}
	responseData, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "error, failed to create JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseData)
	if err != nil {
		http.Error(w, "error, failed to write response", http.StatusInternalServerError)
		return
	}
}
