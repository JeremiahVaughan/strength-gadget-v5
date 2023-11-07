package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strengthgadget.com/m/v2/auth"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func validateForgotPasswordResetCodeRequest(req *model.ForgotPassword) *model.Error {
	var errorFeedback []error

	err := emailIsValid(req.Email)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to emailIsValid() for validateForgotPasswordResetCodeRequest(). Email: %s, error: %v", req.Email, err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	if req.ResetCode == "" {
		errorFeedback = append(errorFeedback, errors.New("reset code is required"))
	}

	if len(errorFeedback) > 0 {
		return &model.Error{
			InternalError:     fmt.Errorf("errors, when validating request: %v", errorFeedback),
			UserFeedbackError: constants.ErrorPasswordResetCodeIsInvalid,
		}
	}
	return nil
}

func HandleForgotPasswordResetCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "error, only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	body, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
		return
	}

	var req model.ForgotPassword
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

	err := validateForgotPasswordResetCodeRequest(&req)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	user, err := auth.GetUser(r.Context(), req.Email)
	if err != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to get user for password reset code: %v", err),
			UserFeedbackError: constants.ErrorUserFeedbackAccessDenied,
		})
		return
	}

	err = service.IsVerificationCodeValid(r.Context(), user, &model.VerificationRequest{
		Email: req.Email,
		Code:  req.ResetCode,
	}, constants.PasswordResetAttemptType)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}
