package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/test_tornado/config"
	"strengthgadget.com/m/v2/test_tornado/service"
	"time"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("error, when attempting to initialize the application configurations. Error: %v", err)
	}

	for i := 0; i < 60; i++ {
		err = healthCheck()
		if err != nil {
			log.Printf("health check failed retrying in one second. Error: %v", err)
		} else {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatalf("health check failed after exhaustive attempts. Error: %v", err)
	}

	initTestCases()
	var notificationMessage string
	notificationMessage, err = runTestCases()
	if err != nil {
		err = fmt.Errorf("test case(s) failed. Error: %v", err)
		notificationMessage = fmt.Sprintf("%s: %v", notificationMessage, err)
	}
	e := service.SendNotification(notificationMessage)
	if e != nil {
		e = fmt.Errorf("error, when sending notification. Notification Error: %v", e)
	}
	if err != nil || e != nil {
		log.Fatalf("Error: %v. Notification Error: %v", err, e)
	}
}

func healthCheck() error {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", config.AppUrl, constants.Health), nil)
	if err != nil {
		return fmt.Errorf("error, when generating get request: %v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if response != nil && response.StatusCode != http.StatusOK {
		return fmt.Errorf("error, when performing get request. ERROR: %v. RESPONSE CODE: %d", err, response.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("error, when performing get request. ERROR: %v", err)
	}

	defer func(Body io.ReadCloser) {
		errClose := Body.Close()
		if errClose != nil {
			log.Printf("error, when closing response body: %v", errClose)
		}
	}(response.Body)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error, when reading response body: %v", err)
	}

	var hr model.HealthResponse
	err = json.Unmarshal(responseBody, &hr)
	if err != nil {
		return fmt.Errorf("error, when unmarshalling response body: %v", err)
	}
	log.Printf("health check successful: %v", hr)
	return nil
}
