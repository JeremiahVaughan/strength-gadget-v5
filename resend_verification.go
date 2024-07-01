package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4"
)

func HandleResendVerification(w http.ResponseWriter, r *http.Request, fields *VerificationFields) (*VerificationFields, error) {
	if fields.Email.Value == "" {
		return nil, fmt.Errorf("error, user failed to provide an email in the resend verification request")
	}

	user, e, err := GetUser(r.Context(), fields.Email.Value)
	if err != nil {
		return nil, fmt.Errorf("error, unable to query database when GetUser for HandleResendVerification. Error: %v", err)
	}
	if e != "" {
		return nil, fmt.Errorf("error, expected to retrieve user from database but couldn't. Error: %v", err)
	}

	if user.EmailVerified {
		fields.ConfirmationCode.ErrorMsg = ErrorVerificationCodeAlreadyVerified
		return fields, nil
	}

	verificationCodeResendRateLimitReached, err := HasVerificationCodeRateLimitBeenReached(r.Context(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to check if verification code limit has been reached for user for resend verification. Email: %s. Error: %v", user.Email, e)
	}

	if verificationCodeResendRateLimitReached {
		fields.ConfirmationCode.ErrorMsg = ErrorTooManyRecentResendVerificationCodeAttempts
		return fields, nil
	}

	err = resendVerification(r.Context(), user)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to resendVerification() for HandleResendVerification(). Error: %v", e)
	}

	fields.ValidForm = true
	return fields, nil
}

func resendVerification(ctx context.Context, user *User) error {
	tx, err := ConnectionPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("an error has occurred when attempting to create transaction for resendVerification()): %v", err)
	}
	err = func(tx pgx.Tx) error {
		err = GenerateNewVerificationCode(ctx, tx, user.Id, user.Email, false)
		if err != nil {
			return fmt.Errorf("error, when generateNewVerificationCode() for resendVerification(): %v", err)
		}
		return nil
	}(tx)
	if err != nil {
		rollBackErr := tx.Rollback(ctx)
		if rollBackErr != nil {
			return fmt.Errorf("error happened when attempting to roll back transaction after query failed. ERROR: %v. RollBack Error: %v", err, rollBackErr)
		}
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("an error occurred when attempting to commit new user and verification code to DB: %v", err)
	}
	return nil
}
