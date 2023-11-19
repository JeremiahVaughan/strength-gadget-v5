package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
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
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, failed to perform exercises handler action: %v", err),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		})
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
