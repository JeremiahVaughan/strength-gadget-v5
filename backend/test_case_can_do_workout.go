package main

import (
	"fmt"
	"net/http"
)

func CanDoWorkout() error {
	// c := IntegrationTestClient{}

	// email, sessionCookie, err := c.newUser()
	// if err != nil {
	// 	return fmt.Errorf("error, when newUser() for CanDoWorkout(). Error: %v", err)
	// }
	// log.Printf("user created for CanDoWorkout. Email: %s", email)

	// todo add the updated pattern for going through a workout
	// .... missing logic
	return nil
}

func (c *IntegrationTestClient) newUser() (string, *http.Cookie, error) {
	validEmail, responseCode, err := c.attemptWithValidCredentials(EndpointRegister)
	if responseCode != http.StatusOK {
		return "", nil, fmt.Errorf("error, when attempt to register for newUser(). Response Code: %d. Error: %s", responseCode, err)
	}
	response, err := c.attemptVerificationWithValidCode(validEmail)
	if response.ResponseCode != http.StatusOK {
		return "", nil, fmt.Errorf("error, when attempting verification with valid verification code for newUser(). Error: %v. Response code: %d", err, response.ResponseCode)
	}
	return validEmail, response.Cookie, nil
}
