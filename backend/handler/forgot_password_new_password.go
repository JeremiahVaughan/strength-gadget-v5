package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strengthgadget.com/m/v2/auth"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func validateForgotPasswordNewPasswordRequest(ctx context.Context, req *model.ForgotPassword, user *model.User) *model.Error {
	err := service.IsVerificationCodeValid(ctx, user, &model.VerificationRequest{
		Email: req.Email,
		Code:  req.ResetCode,
	}, constants.PasswordResetAttemptType)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when service.IsVerificationCodeValid() for validateForgotPasswordNewPasswordRequest() for user email: %s. Error: %v", req.Email, err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}
	return nil
}

func HandleForgotPasswordNewPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "error, only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	body, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
		return
	}

	var req *model.ForgotPassword
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

	err := forgotPasswordNewPassword(r.Context(), req)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func forgotPasswordNewPassword(ctx context.Context, req *model.ForgotPassword) (err *model.Error) {
	err = emailIsValid(req.Email)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to validate email for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	user, err := auth.GetUser(ctx, req.Email)
	if err != nil {
		var feedback model.UserFeedbackError
		if err.UserFeedbackError == constants.ErrorUserFeedbackWrongPasswordOrUsername {
			feedback = constants.ErrorEmailDoesNotExists
		} else {
			feedback = constants.ErrorUnexpectedTryAgain
		}
		return &model.Error{
			InternalError:     fmt.Errorf("error, when auth.GetUser() for forgotPasswordNewPassword() for user: %s. Error: %v", req.Email, err),
			UserFeedbackError: feedback,
		}
	}

	err = validateForgotPasswordNewPasswordRequest(ctx, req, user)
	successfulAttempt := err == nil
	deferErr := service.RecordAccessAttempt(ctx, user, successfulAttempt, constants.PasswordResetAttemptType)
	if deferErr != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to record password reset attempt for user %s. Defer error: %v. Original error: %v", user.Email, deferErr, err),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		}
	}
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to validate request for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	err = service.PasswordIsAcceptable(req.NewPassword)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to validate password for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	hashedPassword, err := service.ObtainHashFromPassword(req.NewPassword)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to hash password for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	e := service.UpdateUserPassword(ctx, user, hashedPassword)
	if e != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to update password for forgotPasswordNewPassword(). Error: %v", e),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		}
	}
	return nil
}
