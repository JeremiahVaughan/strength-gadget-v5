package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"io"
	"net/http"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
		return
	}

	var request model.Credentials
	e = json.Unmarshal(body, &request)
	if e != nil {
		http.Error(w, "error, failed to parse JSON", http.StatusBadRequest)
		return
	}
	e = r.Body.Close()
	if e != nil {
		http.Error(w, fmt.Sprintf("error, when attempting to close request body: %v", e), http.StatusInternalServerError)
		return
	}

	// todo add rate limiter
	err := emailIsValid(request.Email)
	if err != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to validate email for HandleRegister(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	// todo find a way for emails to become available again if verification takes too long. This prevents an edge case where the wrong email is accidentally used and prevents the person who actually owns that email from creating an account.
	// todo let user know in the UI right after submitting registration that they must verify within blank hours otherwise they will have to reregister.
	emailExists, err := emailAlreadyExists(r.Context(), request.Email)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}
	if *emailExists {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("a user attempted to register with an email that already exists: %s", request.Email),
			UserFeedbackError: model.ErrorEmailAlreadyExists,
		})
		return
	}
	err = service.PasswordIsAcceptable(request.Password)
	if err != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to validate password for HandleRegister(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	hashedPassword, err := service.ObtainHashFromPassword(request.Password)
	if err != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to hash password for HandleRegister(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	e = service.PersistNewUser(r.Context(), request.Email, hashedPassword)
	if e != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to persist new user: %v", e),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		})
		return
	}
	// todo provide the auth token back to the user for the session.

	//// todo also add a way for the user to resend email verification in case they didn't receive the email or lost it
	//// todo add a password reset mechanism
	//// 		todo its safe to use the same technique for verification code because the email address has already been verified at this point
	////		todo throttle the password reset mechanism to 5 attempts per 24 hours --- seams reasonable
	//// todo add a way for users to change their email
	//// todo add terms of service and privacy policy checkbox
	//// todo change email so it comes from no-reply@strengthgadget.com
	//// todo for email verification go with verification code
	//// 		todo use all caps for ease of readability but accept caps or lowercase for the sake of user experience.
	//// 		todo use code verification because of the silly belief that clicking links in email is evil
	////				todo the silly believe makes people hesitate to click emails
	////  			todo the silly belief makes emails end up in spam folders
	//
}

func emailIsValid(email string) *model.Error {
	if email == "" {
		return &model.Error{
			InternalError:     fmt.Errorf("the user did not provide an email address with their credentials: %s", email),
			UserFeedbackError: model.ErrorMissingEmailAddress,
		}
	}

	if !service.EmailIsValidFormat(email) {
		return &model.Error{
			InternalError:     fmt.Errorf("the user provided an email address with an invalid format: %s", email),
			UserFeedbackError: model.ErrorInvalidEmailAddress,
		}
	}
	return nil
}

func emailAlreadyExists(ctx context.Context, email string) (*bool, *model.Error) {
	var emailExists bool
	row := config.ConnectionPool.QueryRow(ctx, "SELECT email FROM \"user\" WHERE email = $1", email)
	queryErr := row.Scan(&email)
	if queryErr != nil {
		if queryErr.Error() == pgx.ErrNoRows.Error() {
			emailExists = false
		} else {
			return nil, &model.Error{
				InternalError:     fmt.Errorf("an unexpected error occurred when attempting to check if the email: %s exists. Error: %v", email, queryErr),
				UserFeedbackError: model.ErrorUnexpectedTryAgain,
			}
		}
	} else {
		emailExists = true
	}
	return &emailExists, nil
}
