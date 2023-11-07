package handler

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"net/http"
	"strengthgadget.com/m/v2/auth"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	verificationRequest, err := getVerificationRequestBody(r.Body)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	if verificationRequest.Email == "" {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, user failed to provide an email in the resend verification request"),
			UserFeedbackError: constants.ErrorUserFeedbackAccessDenied,
		})
		return
	}

	user, err := auth.GetUser(r.Context(), verificationRequest.Email)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	if user.EmailVerified {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, user attempting to request another email verification code when there account was already verified"),
			UserFeedbackError: constants.ErrorVerificationNoLongerRequired,
		})
		return
	}

	verificationCodeRateLimitReached, e := service.HasVerificationCodeRateLimitBeenReached(r.Context(), user.Email)
	if e != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to check if verification code limit has been reached for user for resend verification. Email: %s. Error: %v", user.Email, e),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		})
		return
	}

	if verificationCodeRateLimitReached {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, user reached the verification code rate limit for resend verification. User: %s", user.Email),
			UserFeedbackError: constants.ErrorEmailVerificationRateLimitReached,
		})
		return
	}

	e = resendVerification(r.Context(), user)
	if e != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to resendVerification() for HandleResendVerification(). Error: %v", e),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		})
		return
	}
}

func resendVerification(ctx context.Context, user *model.User) error {
	tx, e := config.ConnectionPool.Begin(ctx)
	if e != nil {
		return fmt.Errorf("an error has occurred when attempting to create transaction for resendVerification()): %v", e)
	}
	e = func(tx pgx.Tx) error {
		e = service.GenerateNewVerificationCode(ctx, tx, user.Id, user.Email, false)
		if e != nil {
			return fmt.Errorf("error, when generateNewVerificationCode() for resendVerification(): %v", e)
		}
		return nil
	}(tx)
	if e != nil {
		rollBackErr := tx.Rollback(ctx)
		if rollBackErr != nil {
			return fmt.Errorf("error happened when attempting to roll back transaction after query failed. ERROR: %v. RollBack Error: %v", e, rollBackErr)
		}
		return e
	}
	e = tx.Commit(ctx)
	if e != nil {
		return fmt.Errorf("an error occurred when attempting to commit new user and verification code to DB: %v", e)
	}
	return nil
}
