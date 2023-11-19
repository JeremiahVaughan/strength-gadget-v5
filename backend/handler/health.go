package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/model"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	// todo make health check also ping each of its dependencies to ensure connectivity between the dependency and it
	status := model.HealthResponse{
		Status:     "ok",
		AppVersion: config.Version,
	}

	result, err := json.Marshal(status)
	if err != nil {
		errMessage := fmt.Sprintf("error, when attempting to marshal health response: %v", err)
		log.Printf(errMessage)
		http.Error(w, errMessage, http.StatusInternalServerError)
		return
	}
	_, err = w.Write(result)
	if err != nil {
		errMessage := fmt.Sprintf("error, when attempting to write health response result: %v", err)
		log.Printf(errMessage)
		http.Error(w, errMessage, http.StatusInternalServerError)
		return
	}
}
