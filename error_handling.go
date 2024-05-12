package main

import (
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
)

func HandleUnexpectedError(w http.ResponseWriter, err error) {
	sentry.CaptureException(err)
	log.Printf("ERROR: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
}
