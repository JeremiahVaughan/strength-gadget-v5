package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HealthResponse struct {
	Status     string `json:"status"`
	AppVersion string `json:"appVersion"`
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	// todo make health check also ping each of its dependencies to ensure connectivity between the dependency and it
	status := HealthResponse{
		Status:     "ok",
		AppVersion: Version,
	}

	result, err := json.Marshal(status)
	if err != nil {
		errMessage := fmt.Sprintf("error, when attempting to marshal health response: %v", err)
		log.Print(errMessage)
		http.Error(w, errMessage, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(result)
	if err != nil {
		errMessage := fmt.Sprintf("error, when attempting to write health response result: %v", err)
		log.Print(errMessage)
		http.Error(w, errMessage, http.StatusInternalServerError)
		return
	}
}
