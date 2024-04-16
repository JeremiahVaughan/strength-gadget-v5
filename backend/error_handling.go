package main

import (
	"github.com/getsentry/sentry-go"
	"log"
	"net/http"
)

func GenerateResponse(w http.ResponseWriter, e *Error) {
	sentry.CaptureException(e.InternalError)
	log.Printf("ERROR: %v", e.InternalError)
	http.Error(w, e.UserFeedbackError.Message, e.UserFeedbackError.ResponseCode)
}
