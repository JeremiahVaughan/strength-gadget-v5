package service

import (
	"context"
	"fmt"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/model"
)

func SendForgotPasswordEmail(ctx context.Context, user *model.User) error {
	tx, err := config.ConnectionPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error, when attempting to start a transaction: %v", err)
	}
	err = func() error {
		err = GenerateNewVerificationCode(ctx, tx, user.Id, user.Email, true)
		if err != nil {
			return fmt.Errorf("error, when generateNewVerificationCode() forSendForgotPasswordEmail(): %v", err)
		}
		return nil
	}()
	if err != nil {
		rollBackErr := tx.Rollback(ctx)
		if rollBackErr != nil {
			return fmt.Errorf("error, when attempting to roll back commit: Rollback Error: %v, Original Error: %v", rollBackErr, err)
		}
		return fmt.Errorf("error, when attempting to perform database transaction: %v", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error, when attempting to commit the transaction to the database: %v", err)
	}
	return nil
}
