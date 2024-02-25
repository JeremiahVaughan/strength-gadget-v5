package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func validateRecordIncrementedWorkoutStepRequest(req *model.RecordIncrementedWorkoutStepRequest) error {
	var errorFeedback []error
	if !req.IncrementedProgressIndex.IsValid() {
		errorFeedback = append(errorFeedback, errors.New("must provide between 1 inclusive and 4 inclusive for the workout phase"))
	}
	if req.ExerciseId == "" {
		errorFeedback = append(errorFeedback, errors.New("exerciseId is required"))
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

	var req model.RecordIncrementedWorkoutStepRequest
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

	err = service.RecordIncrementedWorkoutStep(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("error, failed to record incremented workout step handler action: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
