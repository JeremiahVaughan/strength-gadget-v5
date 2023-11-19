package handler

import (
	"fmt"
	"net/http"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func HandleIsLoggedIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error, only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	loggedIn, err := isLoggedIn(r)
	if err != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to check if user is logged in. Error: %v", err),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		})
	}

	if loggedIn {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func isLoggedIn(r *http.Request) (bool, error) {
	return service.IsAuthenticated(r)
}
