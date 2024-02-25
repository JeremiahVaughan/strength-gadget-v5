package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/argon2"
	"net/http"
	"regexp"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/model"
)

// GenerateSecureHash creating a hash according to this guide: https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go
func GenerateSecureHash(password, salt string) (*string, *model.Error) {
	decodeString, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("an error has occurred when attempting to base64 decode the salt: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}
	// using Argon2id as it provides the most protection according to: https://security.stackexchange.com/a/197550
	memory := 32768
	time := 3
	threads := 1
	keyLen := 32 // 32 MB
	hashedPassword := argon2.IDKey(
		[]byte(password),
		decodeString,
		// recommended values for logging in according to: https://pkg.go.dev/golang.org/x/crypto/argon2
		uint32(time),
		uint32(memory),
		uint8(threads),
		uint32(keyLen),
	)
	encodedHashedPassword := base64.StdEncoding.EncodeToString(hashedPassword)
	// making the hash format align with: https://github.com/P-H-C/phc-winner-argon2#command-line-utility
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d,l=%d$%s$%s", argon2.Version, memory, time, threads, keyLen, salt, encodedHashedPassword)
	return &encodedHash, nil
}

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := checkForValidActiveSession(r)
		if err != nil {
			GenerateResponse(w, err)
			return
		}

		if !session.Authenticated {
			GenerateResponse(w, &model.Error{
				InternalError:     fmt.Errorf("user attempted to access a protected resource without being authenticated"),
				UserFeedbackError: model.ErrorUserNotLoggedIn,
			})
			return
		}

		user := &model.User{Id: session.UserId}

		// attach the user to the context
		ctx := context.WithValue(r.Context(), model.SessionKey, user)

		// and call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func IsAuthenticated(r *http.Request) (bool, error) {
	session, err := checkForValidActiveSession(r)
	if err != nil {
		return false, fmt.Errorf("error when checkForValidActiveSession() for IsAuthenticated(). Error: %v", err)
	}
	return session.Authenticated, nil
}

func PasswordIsAcceptable(password string) *model.Error {
	if len(password) == 0 {
		return &model.Error{
			InternalError:     errors.New("user did not provide a password"),
			UserFeedbackError: model.ErrorPasswordWasNotProvided,
		}
	}

	if len(password) < 12 {
		return &model.Error{
			InternalError:     errors.New("user chose a password that was too short"),
			UserFeedbackError: model.ErrorPasswordMustBeAtLeastTwelveCharsLong,
		}
	}

	nonNumberMatch, err := regexp.MatchString("\\D", password)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("an unexpected error occurred when checking password for non-numeric charactures: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}
	if !nonNumberMatch {
		return &model.Error{
			InternalError:     errors.New("user attempted to create a password with all numbers"),
			UserFeedbackError: model.ErrorPasswordCannotContainAllNumbers,
		}
	}
	return nil
}

func ObtainHashFromPassword(password string) (*string, *model.Error) {
	salt, err := generateSalt()
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("error, when generating a salt for ObtainHashFromPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}
	hashedPassword, err := GenerateSecureHash(password, *salt)
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("error, when generating a hash from password for ObtainHashFromPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}
	return hashedPassword, nil
}

func generateSalt() (*string, *model.Error) {
	saltSize := 128 // 128 bits is the salt size used by Bcrypt
	b := make([]byte, saltSize)
	_, err := rand.Read(b)
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("an error has occurred when generating the salt: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}
	result := base64.StdEncoding.EncodeToString(b)
	return &result, nil
}

func IncrementLoginAttemptCount(ctx context.Context, email string) *model.Error {
	key := model.LoginAttemptPrefix + email
	err := IncrementRateLimitingCount(ctx, key, config.WindowLengthInSecondsForTheNumberOfAllowedLoginAttemptsBeforeLockout)
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to IncrementRateLimitingCount() for IncrementLoginAttemptCount(). Error: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}
	return nil
}

func HasLoginAttemptRateLimitBeenReached(ctx context.Context, email string) (bool, *model.Error) {
	key := model.LoginAttemptPrefix + email
	result, err := HasRateLimitBeenReached(ctx, key, config.AllowedLoginAttemptsBeforeTriggeringLockout)
	if err != nil {
		return false, &model.Error{
			InternalError:     fmt.Errorf("error, when HasRateLimitBeenReached() for HasLoginAttemptRateLimitBeenReached(). Error: %v", err),
			UserFeedbackError: model.ErrorUnexpectedTryAgain,
		}
	}
	return result, nil
}

func PersistNewUser(ctx context.Context, email string, hashedPassword *string) error {
	userId := uuid.New().String()
	tx, e := config.ConnectionPool.Begin(ctx)
	if e != nil {
		return fmt.Errorf("an error has occurred when attempting to create transaction: %v", e)
	}

	e = func(tx pgx.Tx) error {
		_, e = tx.Exec(
			ctx,
			"INSERT INTO \"user\" (id, email, password_hash) VALUES ($1, $2, $3)",
			userId,
			email,
			hashedPassword,
		)
		if e != nil {
			return fmt.Errorf("an error has occurred when attempting to insert a new user: %s. Error: %v", email, e)
		}

		e = GenerateNewVerificationCode(ctx, tx, userId, email, false)
		if e != nil {
			return fmt.Errorf("error, when generateNewVerificationCode() for registration handler: %v", e)
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

func UpdateUserPassword(ctx context.Context, user *model.User, passwordHash *string) error {
	_, err := config.ConnectionPool.Exec(ctx, "UPDATE \"user\" SET password_hash = $1 WHERE id = $2", passwordHash, user.Id)
	if err != nil {
		return fmt.Errorf("an error has occurred when attempting to update user password: %v", err)
	}
	return nil
}
