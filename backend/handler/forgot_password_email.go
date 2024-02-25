package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strengthgadget.com/m/v2/auth"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func HandleForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "error, only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	body, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
		return
	}

	var req model.User
	e = json.Unmarshal(body, &req)
	if e != nil {
		http.Error(w, "error, failed to parse JSON", http.StatusBadRequest)
		return
	}
	e = r.Body.Close()
	if e != nil {
		http.Error(w, fmt.Sprintf("error, when attempting to close request body: %v", e), http.StatusInternalServerError)
		return
	}

	email := req.Email
	if !service.EmailIsValidFormat(email) {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("the user provided an email address with an invalid format when submitting to reset password: %s", email),
			UserFeedbackError: model.ErrorInvalidEmailAddress,
		})
		return
	}

	user, err := auth.GetUser(r.Context(), email)
	if err != nil {
		if err.UserFeedbackError == model.ErrorUserFeedbackWrongPasswordOrUsername {
			log.Printf("error, user entered an email that does not exist when reseting password. Email: %s", email)
			return
		} else {
			service.GenerateResponse(w, err)
			return
		}
	}

	verificationCodeRateLimitReached, e := service.HasVerificationCodeRateLimitBeenReached(r.Context(), user.Email)
	if e != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to check if verification code limit has been reached for user for password reset: %s. Error: %v", email, e),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		})
		return
	}

	if verificationCodeRateLimitReached {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, user reached the verification code rate limit for password reset. User: %s", user.Email),
			UserFeedbackError: model.ErrorPasswordResetCodeRateLimitReached,
		})
		return
	}

	e = service.SendForgotPasswordEmail(r.Context(), user)
	if e != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, failed to send forgot password email: %v", e),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
