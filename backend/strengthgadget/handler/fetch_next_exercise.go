package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strengthgadget.com/m/v2/service"
)

func HandleReadyForNextExercise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error, only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	result, err := service.FinishCurrentAndFetchNextExercise(
		r.Context(),
		r.URL.Query().Get("measurement"),
	)
	if err != nil {
		errorMsg := fmt.Sprintf("error, failed to perform exercises handler action: %v", err)
		log.Printf(errorMsg)
		http.Error(w, errorMsg, http.StatusInternalServerError)
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
