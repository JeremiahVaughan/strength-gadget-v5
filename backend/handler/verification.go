package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strengthgadget.com/m/v2/auth"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func HandleVerification(w http.ResponseWriter, r *http.Request) {
	verificationRequest, err := getVerificationRequestBody(r.Body)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	err = emailIsValid(verificationRequest.Email)
	if err != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to validate email for HandleVerification(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	// choosing to let empty codes not count against verification attempts since this doesn't give an attacker any advantage
	if verificationRequest.Code == "" {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, user %s failed to provide a verification code in the verification request", verificationRequest.Email),
			UserFeedbackError: model.ErrorUserFeedbackAccessDenied,
		})
		return
	}

	user, err := auth.GetUser(r.Context(), verificationRequest.Email)
	if err != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to retrieve user for email verification. User email: %s. Error: %v", verificationRequest.Email, err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	if user.EmailVerified {
		// user email has already been verified
		return
	}

	err = service.IsVerificationCodeValid(r.Context(), user, verificationRequest, constants.VerificationAttemptType)
	successfulAttempt := err == nil
	deferErr := service.RecordAccessAttempt(r.Context(), user, successfulAttempt, constants.VerificationAttemptType)
	if deferErr != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to record verification attempt for user %s. Defer error: %v. Original error: %v", user.Email, deferErr, err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		})
		return
	}
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	cookie, e := startNewSession(r.Context(), user.Id)
	if e != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when persisting session key upon verification completion: %v", e),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		})
		return
	}
	http.SetCookie(w, cookie)
}

func getVerificationRequestBody(requestBody io.Reader) (*model.VerificationRequest, *model.Error) {
	bytes, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("error, unable to parse request body: %v", err),
			UserFeedbackError: model.ErrorUserFeedbackAccessDenied,
		}
	}

	var verificationRequest model.VerificationRequest
	err = json.Unmarshal(bytes, &verificationRequest)
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("unable to decode verification request due to: %v", err),
			UserFeedbackError: model.ErrorUserFeedbackAccessDenied,
		}
	}
	return &verificationRequest, nil
}
