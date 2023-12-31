package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	model2 "strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/test_tornado/config"
	"strengthgadget.com/m/v2/test_tornado/model"
)

func TestRequest(
	httpMethod string,
	endpoint string,
	body any,
	cookie *http.Cookie,
	responsePayloadType any,
) (*model.TestRequestResponse, error) {
	var request *http.Request
	var err error
	url := fmt.Sprintf("%s%s", config.AppUrl, endpoint)
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
			return &model.TestRequestResponse{
				ResponseCode: response.StatusCode,
			}, fmt.Errorf("error, when reading error response body: %v", err)
		}
		return &model.TestRequestResponse{
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
		if c.Name == model2.SessionKey {
			authCookie = c
		}
	}
	if responsePayloadType != nil && httpMethod == http.MethodGet {
		var rb []byte
		rb, err = io.ReadAll(response.Body)
		if err != nil {
			return &model.TestRequestResponse{
				ResponseCode: response.StatusCode,
			}, fmt.Errorf("error, when reading response body: %v", err)
		}
		err = json.Unmarshal(rb, responsePayloadType)
		if err != nil {
			return nil, fmt.Errorf("error, when unmarshalling response body: %v", err)
		}
	}
	return &model.TestRequestResponse{
		ResponseCode:    response.StatusCode,
		Cookie:          authCookie,
		ResponsePayload: responsePayloadType,
	}, nil
}

func SendNotification(result string) error {
	message := model2.Notification{Content: result}
	requestBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error, when attempting to marshall payload to json: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, config.NotificationEndpoint, bytes.NewBuffer(requestBody))
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
