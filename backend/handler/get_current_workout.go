package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/model"
)

func HandleGetCurrentWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error, only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	result, err := model.GetCurrentWorkout(
		r.Context(),
		config.RedisConnectionPool,
		config.ConnectionPool,
		config.NumberOfSetsInSuperSet,
		config.NumberOfExerciseInSuperset,
		config.GetSuperSetExpiration(),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("error, failed to perform  handler action: %v", err), http.StatusInternalServerError)
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
