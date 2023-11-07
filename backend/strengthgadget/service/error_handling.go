package service

import (
	"log"
	"net/http"
	"strengthgadget.com/m/v2/model"
)

func GenerateResponse(w http.ResponseWriter, e *model.Error) {
	log.Printf("ERROR: %v", e.InternalError)
	http.Error(w, e.UserFeedbackError.Message, e.UserFeedbackError.ResponseCode)
}
