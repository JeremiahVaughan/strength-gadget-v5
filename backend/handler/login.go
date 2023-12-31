package handler

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strengthgadget.com/m/v2/auth"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
	"strings"
	"time"
)

// todo add if this account exists, you will get an email

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	cred, err := auth.GetEmailAndPassword(r.Body)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	ctx := r.Context()
	var loginAttemptLockout bool
	loginAttemptLockout, err = service.HasLoginAttemptRateLimitBeenReached(ctx, cred.Email)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	if loginAttemptLockout {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, user %s hit the login attempt lockout for too many recent failed login attempts", cred.Email),
			UserFeedbackError: constants.ErrorLoginAttemptRateLimitReached,
		})
		return
	}

	err = service.IncrementLoginAttemptCount(r.Context(), cred.Email)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	user, err := auth.GetUser(ctx, cred.Email)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	salt, err := getSalt(user.PasswordHash)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	hashedInput, err := service.GenerateSecureHash(cred.Password, salt)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	// todo compare legacy versions of the hash should new versions of Argon2 come out or if the parameters change.
	// todo also, update the hash if a legacy version of the hash is detected. -- is this even possible? -- Maybe I can verify the hash with the legacy version of Argon2 somehow then if it matches. Create with the new version
	err = verifyValidUserProvidedPassword(*hashedInput, user.PasswordHash)
	successfulAttempt := err == nil
	deferErr := service.RecordAccessAttempt(ctx, user, successfulAttempt, constants.LoginAttemptType)
	if deferErr != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to record login attempt for user %s. Defer error: %v. Original error: %v", user.Email, deferErr, err),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		})
		return
	}
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	if !user.EmailVerified {
		err = &model.Error{
			InternalError:     errors.New("user attempted to login without verifying email first"),
			UserFeedbackError: constants.ErrorUnverifiedEmailAddress,
		}
		service.GenerateResponse(w, err)
		return
	}

	cookie, e := startNewSession(ctx, user.Id)
	if e != nil {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, when persisting session key upon login: %v", e),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		})
		return
	}
	http.SetCookie(w, cookie)
}

func startNewSession(ctx context.Context, userId string) (*http.Cookie, error) {
	// The extra twelve hours is to ensure the user isn't being prompted to, login will in the middle of using the app.
	// For example without the extra twelve hours, if they were to log in to use the app at 0930 am on sunday, then
	// started using the app immediately. The following week they would be logged off at 0930. This means the user is likely
	// to be using the app at this time as it may be part of their schedule. The offset of twelve hours ensures less disruption.
	hoursInOneWeekPlusTwelve := 180 * time.Hour
	sessionLength := hoursInOneWeekPlusTwelve
	sessionKey := service.GenerateSessionKey()
	// purposely using a string as the storage type for sessions because expiring the hash is done in a separate command. If the expiry command were to fail, then this would mean an immortal session was just created.
	err := config.RedisConnectionPool.Set(ctx, sessionKey, userId, sessionLength).Err()
	if err != nil {
		return nil, fmt.Errorf("error, when persisting the session when startNewSession(). Error: %v", err)
	}
	expirationTime := time.Now().Add(sessionLength)
	var domain string
	if config.Environment == constants.EnvironmentLocal {
		domain = "localhost"
	} else {
		domain = ".strengthgadget.com"
	}
	cookie := &http.Cookie{
		Name:     model.SessionKey,
		Value:    sessionKey,
		HttpOnly: true,
		Secure:   true,
		Expires:  expirationTime,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   domain,
	}
	return cookie, nil
}

func verifyValidUserProvidedPassword(hashedInput string, userHash string) *model.Error {
	if subtle.ConstantTimeCompare([]byte(hashedInput), []byte(userHash)) == 1 {
		return nil
	} else {
		return &model.Error{
			InternalError:     errors.New("user password did not match their password hash"),
			UserFeedbackError: constants.ErrorUserFeedbackWrongPasswordOrUsername,
		}
	}
}

func getSalt(hash string) (string, *model.Error) {
	fields := strings.Split(hash, "$")
	if len(fields) != 6 {
		return "", &model.Error{
			InternalError:     errors.New("user hash contains corrupted formatting"),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		}
	}
	return fields[4], nil
}
