package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	MockValidPassword    = "cbj4vbz9dhn@XWE*zxm"
	MockValidNewPassword = "G9AegHkXkuG47Y2kqBZs"
	MockInValidPassword  = "123"
	MockValidEmail       = "local_test_user@gmail.com"
	MockInValidEmail     = "local_test_user"

	MockInvalidVerificationCode = "ABCDEFGHIJKLMNOPQRSTUVWXZ"
)

var tests []TestCase

var AppUrl = "http://strengthgadget:8080"

var NotificationEndpoint string

type TestRequestResponse struct {
	ResponseCode    int
	Cookie          *http.Cookie
	ResponsePayload any
}

type TestCase func() error

type IntegrationTestClient struct {
}

func (c *IntegrationTestClient) GetValidEmail() string {
	validEmailParts := strings.Split(MockValidEmail, "@")
	return fmt.Sprintf("%s%d@%s", validEmailParts[0], rand.Intn(4294967294), validEmailParts[1])
}

func (c *IntegrationTestClient) TestRequest(
	httpMethod string,
	endpoint string,
	body any,
	cookie *http.Cookie,
	responsePayloadType any,
) (*TestRequestResponse, error) {

	var request *http.Request
	var err error
	url := fmt.Sprintf("%s%s", AppUrl, endpoint)
	if body != nil {
		var requestBody []byte
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error, when attempting to marshall payload to json: %v", err)
		}
		request, err = http.NewRequest(httpMethod, url, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, fmt.Errorf("error, when creating post request. ERROR: %v", err)
		}
	} else {
		request, err = http.NewRequest(httpMethod, url, nil)
		if err != nil {
			return nil, fmt.Errorf("error, when creating get request. ERROR: %v", err)
		}
	}

	if cookie != nil {
		request.AddCookie(cookie)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if response != nil {
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				log.Printf("error, when attempting to close response body: %v", err)
			}
		}(response.Body)
	}
	if response != nil && response.StatusCode != http.StatusOK {
		var rb []byte
		rb, err = io.ReadAll(response.Body)
		if err != nil {
			return &TestRequestResponse{
				ResponseCode: response.StatusCode,
			}, fmt.Errorf("error, when reading error response body: %v", err)
		}
		return &TestRequestResponse{
			ResponseCode: response.StatusCode,
		}, fmt.Errorf("error, when performing get request. ERROR: %v. RESPONSE CODE: %d. RESPONSE MESSAGE: %s", err, response.StatusCode, string(rb))
	}
	if err != nil {
		if response != nil {
			err = fmt.Errorf("error: %v. RESPONSE CODE: %d", err, response.StatusCode)
		}
		return nil, fmt.Errorf("error, when performing post request. ERROR: %v", err)
	}
	var authCookie *http.Cookie
	for _, c := range response.Cookies() {
		if c.Name == string(AuthSessionKey) {
			authCookie = c
		}
	}
	if responsePayloadType != nil && httpMethod == http.MethodGet {
		var rb []byte
		rb, err = io.ReadAll(response.Body)
		if err != nil {
			return &TestRequestResponse{
				ResponseCode: response.StatusCode,
			}, fmt.Errorf("error, when reading response body: %v", err)
		}
		err = json.Unmarshal(rb, responsePayloadType)
		if err != nil {
			return nil, fmt.Errorf("error, when unmarshalling response body: %v", err)
		}
	}
	return &TestRequestResponse{
		ResponseCode:    response.StatusCode,
		Cookie:          authCookie,
		ResponsePayload: responsePayloadType,
	}, nil
}

func (c *IntegrationTestClient) runIntegrationTests() error {
	err := c.InitConfig()
	if err != nil {
		return fmt.Errorf("error, when attempting to initialize the application configurations. Error: %v", err)
	}

	for i := 0; i < 60; i++ {
		err = c.healthCheck()
		if err != nil {
			log.Printf("health check failed retrying in one second. Error: %v", err)
		} else {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return fmt.Errorf("health check failed after exhaustive attempts. Error: %v", err)
	}

	c.initTestCases()
	var notificationMessage string
	notificationMessage, err = c.runTestCases()
	if err != nil {
		err = fmt.Errorf("test case(s) failed. Error: %v", err)
		log.Print(err.Error())
		notificationMessage = fmt.Sprintf("%s: %v", notificationMessage, err)
	}
	notificationErr := c.SendNotification(notificationMessage)
	if notificationErr != nil {
		notificationErr = fmt.Errorf("error, when sending notification. Notification Error: %v", notificationErr)
	}
	if notificationErr != nil {
		return fmt.Errorf("error, notification Error: %v", notificationErr)
	}

	return nil
}

func (c *IntegrationTestClient) registerNewUser() (string, error) {
	response, err := c.attemptWithEmptyUsername(EndpontSignUp)
	if response == nil {
		return "", fmt.Errorf("error, when attempt to register with empty username. Error: %s", err)
	}

	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with empty username. Response Code: %d", response.ResponseCode)
	}

	response, err = c.attemptWithEmptyPassword(EndpontSignUp)
	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with empty password. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	response, err = c.attemptWithInvalidEmail(EndpontSignUp)
	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with invalid email. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	response, err = c.attemptWithInvalidPassword(EndpontSignUp)
	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with invalid password. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	validEmail, responseCode, err := c.attemptWithValidCredentials(EndpontSignUp)
	if responseCode != http.StatusOK {
		return "", fmt.Errorf("error, when attempt to register with valid credentials. Response Code: %d. Error: %s", responseCode, err)
	}
	return validEmail, nil
}

func (c *IntegrationTestClient) runTestCases() (string, error) {
	errChan := make(chan error, len(tests))
	var wg sync.WaitGroup

	for _, test := range tests {
		wg.Add(1)
		go func(t TestCase) {
			defer wg.Done()
			if err := t(); err != nil {
				errChan <- err
			}
		}(test)
	}

	wg.Wait()
	close(errChan)

	var responseErr error
	for err := range errChan {
		if err != nil {
			// Handle yer errors, ye scallywag!
			log.Printf("TEST FAILURE. Error: %v", err)
			responseErr = err
		}
	}
	var resultMessage string
	if responseErr == nil {
		resultMessage = "TEST SUITE SUCCESSFUL!"
	} else {
		resultMessage = "TEST SUITE FAILED"
	}
	log.Print(resultMessage)
	return resultMessage, responseErr
}

func (c *IntegrationTestClient) initTestCases() {
	tests = append(
		tests,
		CanRegisterAndLogin,
		CanDoWorkout,
	)
}

func (c *IntegrationTestClient) SendNotification(result string) error {
	message := Notification{Content: result}
	requestBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error, when attempting to marshall payload to json: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, NotificationEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error, when creating post request. ERROR: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if response != nil {
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				log.Printf("error, when attempting to close response body: %v", err)
			}
		}(response.Body)
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode > 299) {
		var rb []byte
		rb, err = io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error, when reading error response body: %v", err)
		}
		return fmt.Errorf("error, when performing get request. ERROR: %v. RESPONSE CODE: %d. RESPONSE MESSAGE: %s", err, response.StatusCode, string(rb))
	}
	if err != nil {
		if response != nil {
			err = fmt.Errorf("error: %v. RESPONSE CODE: %d", err, response.StatusCode)
		}
		return fmt.Errorf("error, when performing post request. ERROR: %v", err)
	}

	return nil
}

func (c *IntegrationTestClient) InitConfig() error {
	NotificationEndpoint = os.Getenv("NOTIFICATION_ENDPOINT")
	if NotificationEndpoint == "" {
		return fmt.Errorf("error, must provide notification endpoint")
	}
	return nil
}

func (c *IntegrationTestClient) healthCheck() error {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", AppUrl, EndpointHealth), nil)
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

	var hr HealthResponse
	err = json.Unmarshal(responseBody, &hr)
	if err != nil {
		return fmt.Errorf("error, when unmarshalling response body: %v", err)
	}
	log.Printf("health check successful: %v", hr)
	return nil
}
