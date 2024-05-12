package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func CanRegisterAndLogin() error {
	c := IntegrationTestClient{}

	response, err := c.checkIfLoggedIn(nil)
	if response.ResponseCode != http.StatusUnauthorized {
		return fmt.Errorf("error, expected logged out when first executing test script, but got response code: %d. Error: %v", response.ResponseCode, err)
	}

	var registeredEmail string
	registeredEmail, err = c.registerNewUser()
	if err != nil {
		return fmt.Errorf("error, when registerNewUser() for CanRegisterAndLogin(). Error: %v", err)
	}

	// todo implement a way to test verification code expiration (e.g., let it expire and ensure it can't be used still)
	response, err = c.attemptLoginWithoutVerification(registeredEmail)
	if response.ResponseCode != http.StatusForbidden {
		return fmt.Errorf("error, when attempting to login without verification. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	response, err = c.checkIfLoggedIn(nil)
	if response.ResponseCode != http.StatusUnauthorized {
		return fmt.Errorf("error, expected logged out after attempted login without verification, but got response code: %d. Error: %v", response.ResponseCode, err)
	}

	requestOkCount := 0
	tooManyRecentResendVerificationRequests := 0
	for i := 0; i < 10; i++ {
		response, err = c.attemptResendVerificationCode(registeredEmail)
		switch response.ResponseCode {
		case http.StatusOK:
			requestOkCount++
		case http.StatusBadRequest:
			tooManyRecentResendVerificationRequests++
		default:
			return fmt.Errorf("error, unexpected response when attemptResendVerificationCode() for CanRegisterAndLogin(). Error: %v, Response code: %d", err, response.ResponseCode)
		}
	}
	if requestOkCount == 0 || tooManyRecentResendVerificationRequests == 0 {
		return fmt.Errorf("error, expected at least one 200 or 400 but got: 200: %d, 400: %d", requestOkCount, tooManyRecentResendVerificationRequests)
	}
	c.waitForTooManyRecentResendVerificationCountToReset()

	invalidRequestCount := 0
	tooManyRecentVerificationRequestsCount := 0
	for i := 0; i < 10; i++ {
		response, err = c.attemptVerificationWithInvalidCode(registeredEmail)
		switch response.ResponseCode {
		case http.StatusBadRequest:
			invalidRequestCount++
		case http.StatusForbidden:
			tooManyRecentVerificationRequestsCount++
		default:
			return fmt.Errorf("error, unexpected response when attemptVerificationWithInvalidCode() for CanRegisterAndLogin(). Error: %v, Response code: %d", err, response.ResponseCode)
		}
	}
	if invalidRequestCount == 0 || tooManyRecentVerificationRequestsCount == 0 {
		return fmt.Errorf("error, expected at least one of either 400 or 403 errors but got: 400: %d, 403: %d", invalidRequestCount, tooManyRecentVerificationRequestsCount)
	}
	c.waitForTooManyRecentVerificationCountToReset()

	response, err = c.attemptVerificationWithValidCode(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting verification with valid verification code. Error: %v. Response code: %d", err, response.ResponseCode)
	}
	sessionCookie := response.Cookie

	response, err = c.checkIfLoggedIn(sessionCookie)
	if err != nil || response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, expected logged in but got response code: %+v. Error: %v", response, err)
	}

	response, err = c.attemptLogout(sessionCookie)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to logout, but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, sessionCookie, err)
	}

	response, err = c.checkIfLoggedIn(sessionCookie)
	if response.ResponseCode != http.StatusUnauthorized {
		return fmt.Errorf("error, expected logged in but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, sessionCookie, err)
	}

	err = c.attemptLoginWithUserThatDoesNotExist()
	if err != nil {
		return fmt.Errorf("error, when attemptLoginWithUserThatDoesNotExist() for CanRegisterAndLogin(). Error: %v", err)
	}

	password := MockValidPassword
	response, err = c.attemptLogin(registeredEmail, password)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attemptLogin() for CanRegisterAndLogin() when attempting login after registration. Error: %v", err)
	}

	response, err = c.attemptLogout(response.Cookie)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to logout after post registration login, but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, response.Cookie, err)
	}

	err = c.attemptResetPassword(registeredEmail)
	if err != nil {
		return fmt.Errorf("error, when attempting password reset. Error: %v", err)
	}

	response, err = c.attemptLoginWithOldPassword(registeredEmail, password)
	if response.ResponseCode != http.StatusBadRequest {
		return fmt.Errorf("error, attempted login with old password and expected a 400 but got %d. Error: %v. Cookie: %+v", response.ResponseCode, err, response.Cookie)
	}

	password = MockValidNewPassword
	response, err = c.attemptLogin(registeredEmail, password)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting login after changing password. Error: %v", err)
	}
	sessionCookie = response.Cookie

	response, err = c.checkIfLoggedIn(sessionCookie)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, expected logged in after post reset password login but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, sessionCookie, err)
	}

	err = c.SendNotification(fmt.Sprintf("##################\n.\n.\n.\n.\n%s\n%s\n.", registeredEmail, MockValidNewPassword))
	if err != nil {
		return fmt.Errorf("error, then attempting to public test email and test password: %v", err)
	}
	log.Printf("integration test EMAIL: %s", registeredEmail)
	log.Printf("integration test PASSWORD: %s", MockValidNewPassword)
	log.Printf("integration test Active Session Cookie: %+v", sessionCookie)

	return nil
}

func (c *IntegrationTestClient) attemptLoginWithOldPassword(registeredEmail, oldPassword string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointLogin,
		Credentials{
			Email:    registeredEmail,
			Password: oldPassword,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptResetPassword(registeredEmail string) error {
	var response *TestRequestResponse
	var err error
	requestSuccessfulCount := 0
	tooManyRecentResendResetPasswordCodeCount := 0
	for i := 0; i < 10; i++ {
		response, err = c.attemptSendResendResetCode(registeredEmail)
		switch response.ResponseCode {
		case http.StatusOK:
			requestSuccessfulCount++
		case http.StatusBadRequest:
			tooManyRecentResendResetPasswordCodeCount++
		default:
			return fmt.Errorf("error, when attempting request a resend for password reset code for forgot reset password mechanism. Response Code: %d. Error: %v", response.ResponseCode, err)
		}
	}
	if requestSuccessfulCount == 0 || tooManyRecentResendResetPasswordCodeCount == 0 {
		return fmt.Errorf("error, expected at least one of either 200 or 400 errors for resend reset password code execessive request test but got: 200: %d, 400: %d", requestSuccessfulCount, tooManyRecentResendResetPasswordCodeCount)
	}

	c.waitForTooManyRecentResendPasswordResetRequestsCountToReset()
	response, err = c.attemptSendResendResetCode(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to Resend password reset code after waiting for exausting reset attempts. Response code: %d. Error: %v", response.ResponseCode, err)
	}

	requestSuccessfulCount = 0
	tooManyRecentResetPasswordCodeCount := 0
	for i := 0; i < 10; i++ {
		response, err = c.attemptSendInvalidForgotPasswordResetCode(registeredEmail)
		switch response.ResponseCode {
		case http.StatusBadRequest:
			requestSuccessfulCount++
		case http.StatusForbidden:
			tooManyRecentResetPasswordCodeCount++
		default:
			return fmt.Errorf("error, when attempting to attemptSendInvalidForgotPasswordResetCode() for attemptResetPassword(). Response Code: %d. Error: %v", response.ResponseCode, err)
		}
	}
	if requestSuccessfulCount == 0 || tooManyRecentResendResetPasswordCodeCount == 0 {
		return fmt.Errorf("error, expected at least one of either 200 or 403 errors for resend reset password code execessive request test but got: 200: %d, 403: %d", requestSuccessfulCount, tooManyRecentResendResetPasswordCodeCount)
	}

	c.waitForTooManyRecentPasswordResetAttemptsCountToReset()
	response, err = c.attemptSendValidForgotPasswordResetCode(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attemptSendValidForgotPasswordResetCode() for attemptResetPassword(). Error: %v", err)
	}

	response, err = c.attemptToSendInvalidNewPassword(registeredEmail)
	if response.ResponseCode != http.StatusBadRequest {
		return fmt.Errorf("error, attempting to send an invalid new password and expected to recieve a 400 response but instead got: %d. Error: %v", response.ResponseCode, err)
	}

	response, err = c.attemptToSendValidNewPassword(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to send a valid new password. Expected response code 200 but got: %d. Error: %v", response.ResponsePayload, err)
	}

	return nil
}

func (c *IntegrationTestClient) attemptToSendValidNewPassword(registeredEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointForgotPasswordPrefix+EndpointNewPassword,
		ForgotPassword{
			Email:       registeredEmail,
			ResetCode:   MockVerificationCode,
			NewPassword: MockValidNewPassword,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptToSendInvalidNewPassword(registeredEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointForgotPasswordPrefix+EndpointNewPassword,
		ForgotPassword{
			Email:       registeredEmail,
			ResetCode:   MockVerificationCode,
			NewPassword: MockInValidPassword,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptSendValidForgotPasswordResetCode(registeredEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointForgotPasswordPrefix+EndpointResetCode,
		ForgotPassword{
			Email:     registeredEmail,
			ResetCode: MockVerificationCode,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptSendResendResetCode(registeredEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointForgotPasswordPrefix+EndpointEmail,
		ForgotPassword{
			Email: registeredEmail,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptSendInvalidForgotPasswordResetCode(registeredEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointForgotPasswordPrefix+EndpointResetCode,
		ForgotPassword{
			Email:     registeredEmail,
			ResetCode: MockInvalidVerificationCode,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptLogin(registeredEmail string, validPassword string) (*TestRequestResponse, error) {
	tooManyRecentLoginAttemptsCount := 0
	loginAttemptRejectedDueToInvalidUsernameOrPasswordCount := 0
	for i := 0; i < 10; i++ {
		response, err := c.attemptLoginWithInvalidCredentials(registeredEmail)
		switch response.ResponseCode {
		case http.StatusBadRequest:
			loginAttemptRejectedDueToInvalidUsernameOrPasswordCount++
		case http.StatusUnauthorized:
			tooManyRecentLoginAttemptsCount++
		default:
			return nil, fmt.Errorf("error, when attempting to attemptLoginWithInvalidCredentials() for attemptLogin(). Response Code: %d. Nil Cookie: %+v, Error: %v", response.ResponseCode, response.Cookie, err)
		}
	}
	if loginAttemptRejectedDueToInvalidUsernameOrPasswordCount == 0 || tooManyRecentLoginAttemptsCount == 0 {
		return nil, fmt.Errorf("error, expected at least one of either 400 or 401 errors for excessive login requests test but got: 400: %d, 401: %d", loginAttemptRejectedDueToInvalidUsernameOrPasswordCount, tooManyRecentLoginAttemptsCount)
	}

	c.waitForTooManyRecentPasswordLoginAttemptsCountToReset()
	return c.attemptLoginWithValidCredentials(registeredEmail, validPassword)
}

func (c *IntegrationTestClient) attemptLoginWithUserThatDoesNotExist() error {
	tooManyRecentLoginAttemptsCount := 0
	loginAttemptRejectedDueToInvalidUsernameOrPasswordCount := 0
	for i := 0; i < 10; i++ {
		response, err := c.attemptLoginWithInvalidCredentials("oeueounaoethusneohauneotu@gmail.com")
		switch response.ResponseCode {
		case http.StatusBadRequest:
			loginAttemptRejectedDueToInvalidUsernameOrPasswordCount++
		case http.StatusUnauthorized:
			tooManyRecentLoginAttemptsCount++
		default:
			return fmt.Errorf("error, when attempting to attemptLoginWithInvalidCredentials() for attemptLoginWithUserThatDoesNotExist(). Response Code: %d. Nil Cookie: %+v, Error: %v", response.ResponseCode, response.Cookie, err)
		}
	}
	if loginAttemptRejectedDueToInvalidUsernameOrPasswordCount == 0 || tooManyRecentLoginAttemptsCount == 0 {
		return fmt.Errorf("error, expected at least one of either 400 or 401 errors for excessive login requests with a user that does not existing test but got: 400: %d, 401: %d", loginAttemptRejectedDueToInvalidUsernameOrPasswordCount, tooManyRecentLoginAttemptsCount)
	}
	return nil
}

func (c *IntegrationTestClient) attemptLoginWithInvalidCredentials(registeredEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointLogin,
		Credentials{
			Email:    registeredEmail,
			Password: MockInValidPassword,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptLoginWithValidCredentials(registeredEmail string, validPassword string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointLogin,
		Credentials{
			Email:    registeredEmail,
			Password: validPassword,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) waitForTooManyRecentResendVerificationCountToReset() {
	time.Sleep(5 * time.Second)
}

func (c *IntegrationTestClient) waitForTooManyRecentPasswordResetAttemptsCountToReset() {
	time.Sleep(5 * time.Second)
}

func (c *IntegrationTestClient) waitForTooManyRecentResendPasswordResetRequestsCountToReset() {
	time.Sleep(5 * time.Second)
}

func (c *IntegrationTestClient) waitForTooManyRecentPasswordLoginAttemptsCountToReset() {
	time.Sleep(5 * time.Second)
}

func (c *IntegrationTestClient) attemptResendVerificationCode(email string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointResendVerification,
		VerificationRequest{
			Email: email,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptLogout(cookie *http.Cookie) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointLogout,
		nil,
		cookie,
		nil,
	)
}

func (c *IntegrationTestClient) checkIfLoggedIn(cookie *http.Cookie) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodGet,
		EndpointIsLoggedIn,
		nil,
		cookie,
		nil,
	)
}

func (c *IntegrationTestClient) attemptVerificationWithValidCode(validEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointVerification,
		VerificationRequest{
			Email: validEmail,
			Code:  MockVerificationCode,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) waitForTooManyRecentVerificationCountToReset() {
	time.Sleep(5 * time.Second)
}

func (c *IntegrationTestClient) attemptVerificationWithInvalidCode(validEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointVerification,
		VerificationRequest{
			Email: validEmail,
			Code:  MockInvalidVerificationCode,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptLoginWithoutVerification(validEmail string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		EndpointLogin,
		Credentials{
			Email:    validEmail,
			Password: MockValidPassword,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptWithValidCredentials(register string) (string, int, error) {
	validEmail := c.GetValidEmail()
	response, err := c.TestRequest(
		http.MethodPost,
		register,
		Credentials{
			Email:    validEmail,
			Password: MockValidPassword,
		},
		nil,
		nil,
	)
	return validEmail, response.ResponseCode, err
}

func (c *IntegrationTestClient) attemptWithInvalidPassword(endpoint string) (*TestRequestResponse, error) {
	return c.TestRequest(http.MethodPost, endpoint, Credentials{
		Email:    c.GetValidEmail(),
		Password: MockInValidPassword,
	}, nil, nil)
}

func (c *IntegrationTestClient) attemptWithInvalidEmail(endpoint string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		endpoint, Credentials{
			Email:    MockInValidEmail,
			Password: MockValidPassword,
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptWithEmptyPassword(endpoint string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		endpoint,
		Credentials{
			Email:    c.GetValidEmail(),
			Password: "",
		},
		nil,
		nil,
	)
}

func (c *IntegrationTestClient) attemptWithEmptyUsername(endpoint string) (*TestRequestResponse, error) {
	return c.TestRequest(
		http.MethodPost,
		endpoint,
		Credentials{
			Email:    "",
			Password: MockValidPassword,
		},
		nil,
		nil,
	)
}
