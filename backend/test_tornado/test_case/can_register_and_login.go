package test_case

import (
	"fmt"
	"log"
	"net/http"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	testConstants "strengthgadget.com/m/v2/test_tornado/constants"
	testModel "strengthgadget.com/m/v2/test_tornado/model"
	"strengthgadget.com/m/v2/test_tornado/service"
	"time"
)

func CanRegisterAndLogin() error {
	response, err := checkIfLoggedIn(nil)
	if response.ResponseCode != http.StatusUnauthorized {
		return fmt.Errorf("error, expected logged out when first executing test script, but got response code: %d. Error: %v", response.ResponseCode, err)
	}

	var registeredEmail string
	registeredEmail, err = registerNewUser()
	if err != nil {
		return fmt.Errorf("error, when registerNewUser() for CanRegisterAndLogin(). Error: %v", err)
	}

	// todo implement a way to test verification code expiration (e.g., let it expire and ensure it can't be used still)
	response, err = attemptLoginWithoutVerification(registeredEmail)
	if response.ResponseCode != http.StatusForbidden {
		return fmt.Errorf("error, when attempting to login without verification. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	response, err = checkIfLoggedIn(nil)
	if response.ResponseCode != http.StatusUnauthorized {
		return fmt.Errorf("error, expected logged out after attempted login without verification, but got response code: %d. Error: %v", response.ResponseCode, err)
	}

	requestOkCount := 0
	tooManyRecentResendVerificationRequests := 0
	for i := 0; i < 10; i++ {
		response, err = attemptResendVerificationCode(registeredEmail)
		if response.ResponseCode == http.StatusOK {
			requestOkCount++
		} else if response.ResponseCode == http.StatusBadRequest {
			tooManyRecentResendVerificationRequests++
		} else {
			return fmt.Errorf("error, unexpected response when attemptResendVerificationCode() for CanRegisterAndLogin(). Error: %v, Response code: %d", err, response.ResponseCode)
		}
	}
	if requestOkCount == 0 || tooManyRecentResendVerificationRequests == 0 {
		return fmt.Errorf("error, expected at least one 200 or 400 but got: 200: %d, 400: %d", requestOkCount, tooManyRecentResendVerificationRequests)
	}
	waitForTooManyRecentResendVerificationCountToReset()

	invalidRequestCount := 0
	tooManyRecentVerificationRequestsCount := 0
	for i := 0; i < 10; i++ {
		response, err = attemptVerificationWithInvalidCode(registeredEmail)
		if response.ResponseCode == http.StatusBadRequest {
			invalidRequestCount++
		} else if response.ResponseCode == http.StatusForbidden {
			tooManyRecentVerificationRequestsCount++
		} else {
			return fmt.Errorf("error, unexpected response when attemptVerificationWithInvalidCode() for CanRegisterAndLogin(). Error: %v, Response code: %d", err, response.ResponseCode)
		}
	}
	if invalidRequestCount == 0 || tooManyRecentVerificationRequestsCount == 0 {
		return fmt.Errorf("error, expected at least one of either 400 or 403 errors but got: 400: %d, 403: %d", invalidRequestCount, tooManyRecentVerificationRequestsCount)
	}
	waitForTooManyRecentVerificationCountToReset()

	response, err = attemptVerificationWithValidCode(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting verification with valid verification code. Error: %v. Response code: %d", err, response.ResponseCode)
	}
	sessionCookie := response.Cookie

	response, err = checkIfLoggedIn(sessionCookie)
	if err != nil || response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, expected logged in but got response code: %+v. Error: %v", response, err)
	}

	response, err = attemptLogout(sessionCookie)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to logout, but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, sessionCookie, err)
	}

	response, err = checkIfLoggedIn(sessionCookie)
	if response.ResponseCode != http.StatusUnauthorized {
		return fmt.Errorf("error, expected logged in but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, sessionCookie, err)
	}

	err = attemptLoginWithUserThatDoesNotExist()
	if err != nil {
		return fmt.Errorf("error, when attemptLoginWithUserThatDoesNotExist() for CanRegisterAndLogin(). Error: %v", err)
	}

	password := testConstants.ValidPassword
	response, err = attemptLogin(registeredEmail, password)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attemptLogin() for CanRegisterAndLogin() when attempting login after registration. Error: %v", err)
	}

	response, err = attemptLogout(response.Cookie)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to logout after post registration login, but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, response.Cookie, err)
	}

	err = attemptResetPassword(registeredEmail)
	if err != nil {
		return fmt.Errorf("error, when attempting password reset. Error: %v", err)
	}

	response, err = attemptLoginWithOldPassword(registeredEmail, password)
	if response.ResponseCode != http.StatusBadRequest {
		return fmt.Errorf("error, attempted login with old password and expected a 400 but got %d. Error: %v. Cookie: %+v", response.ResponseCode, err, response.Cookie)
	}

	password = testConstants.ValidNewPassword
	response, err = attemptLogin(registeredEmail, password)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting login after changing password. Error: %v", err)
	}
	sessionCookie = response.Cookie

	response, err = checkIfLoggedIn(sessionCookie)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, expected logged in after post reset password login but got response code: %d. Cookie: %+v. Error: %v", response.ResponseCode, sessionCookie, err)
	}

	err = service.SendNotification(fmt.Sprintf("##################\n.\n.\n.\n.\n%s\n%s\n.", registeredEmail, testConstants.ValidNewPassword))
	if err != nil {
		return fmt.Errorf("error, then attempting to public test email and test password: %v", err)
	}
	log.Printf("integration test EMAIL: %s", registeredEmail)
	log.Printf("integration test PASSWORD: %s", testConstants.ValidNewPassword)
	log.Printf("integration test Active Session Cookie: %+v", sessionCookie)

	return nil
}

func attemptLoginWithOldPassword(registeredEmail, oldPassword string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.Login,
		model.Credentials{
			Email:    registeredEmail,
			Password: oldPassword,
		},
		nil,
		nil,
	)
}

func attemptResetPassword(registeredEmail string) error {
	var response *testModel.TestRequestResponse
	var err error
	requestSuccessfulCount := 0
	tooManyRecentResendResetPasswordCodeCount := 0
	for i := 0; i < 10; i++ {
		response, err = attemptSendResendResetCode(registeredEmail)
		if response.ResponseCode == http.StatusOK {
			requestSuccessfulCount++
		} else if response.ResponseCode == http.StatusBadRequest {
			tooManyRecentResendResetPasswordCodeCount++
		} else {
			return fmt.Errorf("error, when attempting request a resend for password reset code for forgot reset password mechanism. Response Code: %d. Error: %v", response.ResponseCode, err)
		}
	}
	if requestSuccessfulCount == 0 || tooManyRecentResendResetPasswordCodeCount == 0 {
		return fmt.Errorf("error, expected at least one of either 200 or 400 errors for resend reset password code execessive request test but got: 200: %d, 400: %d", requestSuccessfulCount, tooManyRecentResendResetPasswordCodeCount)
	}

	waitForTooManyRecentResendPasswordResetRequestsCountToReset()
	response, err = attemptSendResendResetCode(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to Resend password reset code after waiting for exausting reset attempts. Response code: %d. Error: %v", response.ResponseCode, err)
	}

	requestSuccessfulCount = 0
	tooManyRecentResetPasswordCodeCount := 0
	for i := 0; i < 10; i++ {
		response, err = attemptSendInvalidForgotPasswordResetCode(registeredEmail)
		if response.ResponseCode == http.StatusBadRequest {
			requestSuccessfulCount++
		} else if response.ResponseCode == http.StatusForbidden {
			tooManyRecentResetPasswordCodeCount++
		} else {
			return fmt.Errorf("error, when attempting to attemptSendInvalidForgotPasswordResetCode() for attemptResetPassword(). Response Code: %d. Error: %v", response.ResponseCode, err)
		}
	}
	if requestSuccessfulCount == 0 || tooManyRecentResendResetPasswordCodeCount == 0 {
		return fmt.Errorf("error, expected at least one of either 200 or 403 errors for resend reset password code execessive request test but got: 200: %d, 403: %d", requestSuccessfulCount, tooManyRecentResendResetPasswordCodeCount)
	}

	waitForTooManyRecentPasswordResetAttemptsCountToReset()
	response, err = attemptSendValidForgotPasswordResetCode(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attemptSendValidForgotPasswordResetCode() for attemptResetPassword(). Error: %v", err)
	}

	response, err = attemptToSendInvalidNewPassword(registeredEmail)
	if response.ResponseCode != http.StatusBadRequest {
		return fmt.Errorf("error, attempting to send an invalid new password and expected to recieve a 400 response but instead got: %d. Error: %v", response.ResponseCode, err)
	}

	response, err = attemptToSendValidNewPassword(registeredEmail)
	if response.ResponseCode != http.StatusOK {
		return fmt.Errorf("error, when attempting to send a valid new password. Expected response code 200 but got: %d. Error: %v", response.ResponsePayload, err)
	}

	return nil
}

func attemptToSendValidNewPassword(registeredEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.ForgotPasswordPrefix+constants.NewPassword,
		model.ForgotPassword{
			Email:       registeredEmail,
			ResetCode:   constants.MockVerificationCode,
			NewPassword: testConstants.ValidNewPassword,
		},
		nil,
		nil,
	)
}

func attemptToSendInvalidNewPassword(registeredEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.ForgotPasswordPrefix+constants.NewPassword,
		model.ForgotPassword{
			Email:       registeredEmail,
			ResetCode:   constants.MockVerificationCode,
			NewPassword: testConstants.InValidPassword,
		},
		nil,
		nil,
	)
}

func attemptSendValidForgotPasswordResetCode(registeredEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.ForgotPasswordPrefix+constants.ResetCode,
		model.ForgotPassword{
			Email:     registeredEmail,
			ResetCode: constants.MockVerificationCode,
		},
		nil,
		nil,
	)
}

func attemptSendResendResetCode(registeredEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.ForgotPasswordPrefix+constants.Email,
		model.ForgotPassword{
			Email: registeredEmail,
		},
		nil,
		nil,
	)
}

func attemptSendInvalidForgotPasswordResetCode(registeredEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.ForgotPasswordPrefix+constants.ResetCode,
		model.ForgotPassword{
			Email:     registeredEmail,
			ResetCode: testConstants.InvalidVerificationCode,
		},
		nil,
		nil,
	)
}

func attemptLogin(registeredEmail string, validPassword string) (*testModel.TestRequestResponse, error) {
	tooManyRecentLoginAttemptsCount := 0
	loginAttemptRejectedDueToInvalidUsernameOrPasswordCount := 0
	for i := 0; i < 10; i++ {
		response, err := attemptLoginWithInvalidCredentials(registeredEmail)
		if response.ResponseCode == http.StatusBadRequest {
			loginAttemptRejectedDueToInvalidUsernameOrPasswordCount++
		} else if response.ResponseCode == http.StatusUnauthorized {
			tooManyRecentLoginAttemptsCount++
		} else {
			return nil, fmt.Errorf("error, when attempting to attemptLoginWithInvalidCredentials() for attemptLogin(). Response Code: %d. Nil Cookie: %+v, Error: %v", response.ResponseCode, response.Cookie, err)
		}
	}
	if loginAttemptRejectedDueToInvalidUsernameOrPasswordCount == 0 || tooManyRecentLoginAttemptsCount == 0 {
		return nil, fmt.Errorf("error, expected at least one of either 400 or 401 errors for excessive login requests test but got: 400: %d, 401: %d", loginAttemptRejectedDueToInvalidUsernameOrPasswordCount, tooManyRecentLoginAttemptsCount)
	}

	waitForTooManyRecentPasswordLoginAttemptsCountToReset()
	return attemptLoginWithValidCredentials(registeredEmail, validPassword)
}

func attemptLoginWithUserThatDoesNotExist() error {
	tooManyRecentLoginAttemptsCount := 0
	loginAttemptRejectedDueToInvalidUsernameOrPasswordCount := 0
	for i := 0; i < 10; i++ {
		response, err := attemptLoginWithInvalidCredentials("oeueounaoethusneohauneotu@gmail.com")
		if response.ResponseCode == http.StatusBadRequest {
			loginAttemptRejectedDueToInvalidUsernameOrPasswordCount++
		} else if response.ResponseCode == http.StatusUnauthorized {
			tooManyRecentLoginAttemptsCount++
		} else {
			return fmt.Errorf("error, when attempting to attemptLoginWithInvalidCredentials() for attemptLoginWithUserThatDoesNotExist(). Response Code: %d. Nil Cookie: %+v, Error: %v", response.ResponseCode, response.Cookie, err)
		}
	}
	if loginAttemptRejectedDueToInvalidUsernameOrPasswordCount == 0 || tooManyRecentLoginAttemptsCount == 0 {
		return fmt.Errorf("error, expected at least one of either 400 or 401 errors for excessive login requests with a user that does not existing test but got: 400: %d, 401: %d", loginAttemptRejectedDueToInvalidUsernameOrPasswordCount, tooManyRecentLoginAttemptsCount)
	}
	return nil
}

func attemptLoginWithInvalidCredentials(registeredEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.Login,
		model.Credentials{
			Email:    registeredEmail,
			Password: testConstants.InValidPassword,
		},
		nil,
		nil,
	)
}

func attemptLoginWithValidCredentials(registeredEmail string, validPassword string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.Login,
		model.Credentials{
			Email:    registeredEmail,
			Password: validPassword,
		},
		nil,
		nil,
	)
}

func waitForTooManyRecentResendVerificationCountToReset() {
	time.Sleep(5 * time.Second)
}

func waitForTooManyRecentPasswordResetAttemptsCountToReset() {
	time.Sleep(5 * time.Second)
}

func waitForTooManyRecentResendPasswordResetRequestsCountToReset() {
	time.Sleep(5 * time.Second)
}
func waitForTooManyRecentPasswordLoginAttemptsCountToReset() {
	time.Sleep(5 * time.Second)
}

func attemptResendVerificationCode(email string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(http.MethodPost, constants.ResendVerification, model.VerificationRequest{
		Email: email,
	}, nil, nil)
}

func attemptLogout(cookie *http.Cookie) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(http.MethodPost, constants.Logout, nil, cookie, nil)
}

func checkIfLoggedIn(cookie *http.Cookie) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(http.MethodGet, constants.IsLoggedIn, nil, cookie, nil)
}

func attemptVerificationWithValidCode(validEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.Verification, model.VerificationRequest{
			Email: validEmail,
			Code:  constants.MockVerificationCode,
		},
		nil,
		nil,
	)
}

func waitForTooManyRecentVerificationCountToReset() {
	time.Sleep(5 * time.Second)
}

func attemptVerificationWithInvalidCode(validEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.Verification, model.VerificationRequest{
			Email: validEmail,
			Code:  testConstants.InvalidVerificationCode,
		},
		nil,
		nil,
	)
}

func registerNewUser() (string, error) {
	response, err := attemptWithEmptyUsername(constants.Register)
	if response == nil {
		return "", fmt.Errorf("error, when attempt to register with empty username. Error: %s", err)
	}
	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with empty username. Response Code: %d", response.ResponseCode)
	}

	response, err = attemptWithEmptyPassword(constants.Register)
	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with empty password. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	response, err = attemptWithInvalidEmail(constants.Register)
	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with invalid email. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	response, err = attemptWithInvalidPassword(constants.Register)
	if response.ResponseCode != http.StatusBadRequest {
		return "", fmt.Errorf("error, when attempt to register with invalid password. Response Code: %d. Error: %s", response.ResponseCode, err)
	}

	validEmail, responseCode, err := attemptWithValidCredentials(constants.Register)
	if responseCode != http.StatusOK {
		return "", fmt.Errorf("error, when attempt to register with valid credentials. Response Code: %d. Error: %s", responseCode, err)
	}
	return validEmail, nil
}

func attemptLoginWithoutVerification(validEmail string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(
		http.MethodPost,
		constants.Login,
		model.Credentials{
			Email:    validEmail,
			Password: testConstants.ValidPassword,
		},
		nil,
		nil,
	)
}

func attemptWithValidCredentials(register string) (string, int, error) {
	validEmail := service.GetValidEmail()
	response, err := service.TestRequest(
		http.MethodPost,
		register,
		model.Credentials{
			Email:    validEmail,
			Password: testConstants.ValidPassword,
		},
		nil,
		nil,
	)
	return validEmail, response.ResponseCode, err
}

func attemptWithInvalidPassword(endpoint string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(http.MethodPost, endpoint, model.Credentials{
		Email:    service.GetValidEmail(),
		Password: testConstants.InValidPassword,
	}, nil, nil)
}

func attemptWithInvalidEmail(endpoint string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(http.MethodPost, endpoint, model.Credentials{
		Email:    testConstants.InValidEmail,
		Password: testConstants.ValidPassword,
	}, nil, nil)
}

func attemptWithEmptyPassword(endpoint string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(http.MethodPost, endpoint, model.Credentials{
		Email:    service.GetValidEmail(),
		Password: "",
	}, nil, nil)
}

func attemptWithEmptyUsername(endpoint string) (*testModel.TestRequestResponse, error) {
	return service.TestRequest(http.MethodPost, endpoint, model.Credentials{
		Email:    "",
		Password: testConstants.ValidPassword,
	}, nil, nil)
}
