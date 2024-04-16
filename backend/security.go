package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/argon2"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (

	// MockVerificationCode for local development we mock the verification code to avoid sending emails with real email servers
	MockVerificationCode = "ABCDEF"
)

const (
	PasswordResetAttemptType = "14cb4661-74e5-49e8-8532-ebe93d1e806a"
	LoginAttemptType         = "288e1dae-5865-4707-b242-ce818ee8145f"
	VerificationAttemptType  = "ca33f4f1-e2ba-49e1-8222-be982a57c231"
)

// todo validate the length of input so it can't crash your stuff
type VerificationRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerificationCode struct {
	Id      string
	Code    string
	Expires uint64
	UserId  string
}

type UserSession struct {
	UserId        string
	SessionKey    string
	Authenticated bool
}

type User struct {
	Id            string
	Email         string
	EmailVerified bool
	PasswordHash  string
}

type UserService struct {
}

// Error Structuring errors to better ensure the user does not get exposed to information he/she shouldn't have.
type Error struct {
	InternalError     error
	UserFeedbackError UserFeedbackError
}

type UserFeedbackError struct {
	Message      string
	ResponseCode int
}

func (us *UserService) FetchFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(SessionKey).(*User)
	if !ok {
		return nil, fmt.Errorf("error, could not locate the user in session context")
	}
	return user, nil
}

// SessionKey use for both locating the user session ID from the http cookie and locating the user struct in context
const SessionKey = "session_key"

type ForgotPassword struct {
	Email       string `json:"email"`
	ResetCode   string `json:"resetCode"`
	NewPassword string `json:"newPassword"`
}

func HandleVerification(w http.ResponseWriter, r *http.Request) {
	verificationRequest, err := getVerificationRequestBody(r.Body)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	err = emailIsValid(verificationRequest.Email)
	if err != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to validate email for HandleVerification(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	// choosing to let empty codes not count against verification attempts since this doesn't give an attacker any advantage
	if verificationRequest.Code == "" {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, user %s failed to provide a verification code in the verification request", verificationRequest.Email),
			UserFeedbackError: ErrorUserFeedbackAccessDenied,
		})
		return
	}

	user, err := GetUser(r.Context(), verificationRequest.Email)
	if err != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to retrieve user for email verification. User email: %s. Error: %v", verificationRequest.Email, err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	if user.EmailVerified {
		// user email has already been verified
		return
	}

	err = IsVerificationCodeValid(r.Context(), user, verificationRequest, VerificationAttemptType)
	successfulAttempt := err == nil
	deferErr := RecordAccessAttempt(r.Context(), user, successfulAttempt, VerificationAttemptType)
	if deferErr != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to record verification attempt for user %s. Defer error: %v. Original error: %v", user.Email, deferErr, err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	cookie, e := startNewSession(r.Context(), user.Id)
	if e != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when persisting session key upon verification completion: %v", e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}
	http.SetCookie(w, cookie)
}

func getVerificationRequestBody(requestBody io.Reader) (*VerificationRequest, *Error) {
	bytes, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, unable to parse request body: %v", err),
			UserFeedbackError: ErrorUserFeedbackAccessDenied,
		}
	}

	var verificationRequest VerificationRequest
	err = json.Unmarshal(bytes, &verificationRequest)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("unable to decode verification request due to: %v", err),
			UserFeedbackError: ErrorUserFeedbackAccessDenied,
		}
	}
	return &verificationRequest, nil
}

func HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	verificationRequest, err := getVerificationRequestBody(r.Body)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	if verificationRequest.Email == "" {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, user failed to provide an email in the resend verification request"),
			UserFeedbackError: ErrorUserFeedbackAccessDenied,
		})
		return
	}

	user, err := GetUser(r.Context(), verificationRequest.Email)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	if user.EmailVerified {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, user attempting to request another email verification code when there account was already verified"),
			UserFeedbackError: ErrorVerificationNoLongerRequired,
		})
		return
	}

	verificationCodeRateLimitReached, e := HasVerificationCodeRateLimitBeenReached(r.Context(), user.Email)
	if e != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to check if verification code limit has been reached for user for resend verification. Email: %s. Error: %v", user.Email, e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}

	if verificationCodeRateLimitReached {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, user reached the verification code rate limit for resend verification. User: %s", user.Email),
			UserFeedbackError: ErrorEmailVerificationRateLimitReached,
		})
		return
	}

	e = resendVerification(r.Context(), user)
	if e != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to resendVerification() for HandleResendVerification(). Error: %v", e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}
}

func resendVerification(ctx context.Context, user *User) error {
	tx, e := ConnectionPool.Begin(ctx)
	if e != nil {
		return fmt.Errorf("an error has occurred when attempting to create transaction for resendVerification()): %v", e)
	}
	e = func(tx pgx.Tx) error {
		e = GenerateNewVerificationCode(ctx, tx, user.Id, user.Email, false)
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

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, e := io.ReadAll(r.Body)
	if e != nil {
		http.Error(w, "error, failed to read request body", http.StatusBadRequest)
		return
	}

	var request Credentials
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
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to validate email for HandleRegister(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	// todo find a way for emails to become available again if verification takes too long. This prevents an edge case where the wrong email is accidentally used and prevents the person who actually owns that email from creating an account.
	// todo let user know in the UI right after submitting registration that they must verify within blank hours otherwise they will have to reregister.
	emailExists, err := emailAlreadyExists(r.Context(), request.Email)
	if err != nil {
		GenerateResponse(w, err)
		return
	}
	if *emailExists {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("a user attempted to register with an email that already exists: %s", request.Email),
			UserFeedbackError: ErrorEmailAlreadyExists,
		})
		return
	}
	err = PasswordIsAcceptable(request.Password)
	if err != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to validate password for HandleRegister(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	hashedPassword, err := ObtainHashFromPassword(request.Password)
	if err != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to hash password for HandleRegister(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		})
		return
	}

	e = PersistNewUser(r.Context(), request.Email, hashedPassword)
	if e != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to persist new user: %v", e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
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

func emailIsValid(email string) *Error {
	if email == "" {
		return &Error{
			InternalError:     fmt.Errorf("the user did not provide an email address with their credentials: %s", email),
			UserFeedbackError: ErrorMissingEmailAddress,
		}
	}

	if !EmailIsValidFormat(email) {
		return &Error{
			InternalError:     fmt.Errorf("the user provided an email address with an invalid format: %s", email),
			UserFeedbackError: ErrorInvalidEmailAddress,
		}
	}
	return nil
}

func emailAlreadyExists(ctx context.Context, email string) (*bool, *Error) {
	var emailExists bool
	row := ConnectionPool.QueryRow(ctx, "SELECT email FROM \"user\" WHERE email = $1", email)
	queryErr := row.Scan(&email)
	if queryErr != nil {
		if queryErr.Error() == pgx.ErrNoRows.Error() {
			emailExists = false
		} else {
			return nil, &Error{
				InternalError:     fmt.Errorf("an unexpected error occurred when attempting to check if the email: %s exists. Error: %v", email, queryErr),
				UserFeedbackError: ErrorUnexpectedTryAgain,
			}
		}
	} else {
		emailExists = true
	}
	return &emailExists, nil
}

func validateLogoutRequest(r *http.Request) (string, *Error) {
	cookie, err := r.Cookie(SessionKey)
	if err != nil {
		return "", &Error{
			InternalError:     fmt.Errorf("error, no session_key provided in request when attempting to logout. Error: %v", err),
			UserFeedbackError: ErrorUserNotLoggedIn,
		}
	}
	return cookie.Value, nil
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, only POST method is supported"),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}

	sessionKey, err := validateLogoutRequest(r)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	err = logout(r.Context(), w, sessionKey)
	if err != nil {
		GenerateResponse(w, err)
		return
	}
}

func logout(ctx context.Context, w http.ResponseWriter, sessionKey string) *Error {
	err := RedisConnectionPool.Del(ctx, sessionKey).Err()
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("unable to delete session. Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:   SessionKey,
		Value:  "",
		MaxAge: -1, // This deletes the cookie
	})
	return nil
}

// todo add if this account exists, you will get an email

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	cred, err := GetEmailAndPassword(r.Body)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	ctx := r.Context()
	var loginAttemptLockout bool
	loginAttemptLockout, err = HasLoginAttemptRateLimitBeenReached(ctx, cred.Email)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	if loginAttemptLockout {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, user %s hit the login attempt lockout for too many recent failed login attempts", cred.Email),
			UserFeedbackError: ErrorLoginAttemptRateLimitReached,
		})
		return
	}

	err = IncrementLoginAttemptCount(r.Context(), cred.Email)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	user, err := GetUser(ctx, cred.Email)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	salt, err := getSalt(user.PasswordHash)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	hashedInput, err := GenerateSecureHash(cred.Password, salt)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	// todo compare legacy versions of the hash should new versions of Argon2 come out or if the parameters change.
	// todo also, update the hash if a legacy version of the hash is detected. -- is this even possible? -- Maybe I can verify the hash with the legacy version of Argon2 somehow then if it matches. Create with the new version
	err = verifyValidUserProvidedPassword(*hashedInput, user.PasswordHash)
	successfulAttempt := err == nil
	deferErr := RecordAccessAttempt(ctx, user, successfulAttempt, LoginAttemptType)
	if deferErr != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to record login attempt for user %s. Defer error: %v. Original error: %v", user.Email, deferErr, err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	if !user.EmailVerified {
		err = &Error{
			InternalError:     errors.New("user attempted to login without verifying email first"),
			UserFeedbackError: ErrorUnverifiedEmailAddress,
		}
		GenerateResponse(w, err)
		return
	}

	cookie, e := startNewSession(ctx, user.Id)
	if e != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when persisting session key upon login: %v", e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
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
	sessionKey := GenerateSessionKey()
	// purposely using a string as the storage type for sessions because expiring the hash is done in a separate command. If the expiry command were to fail, then this would mean an immortal session was just created.
	err := RedisConnectionPool.Set(ctx, sessionKey, userId, sessionLength).Err()
	if err != nil {
		return nil, fmt.Errorf("error, when persisting the session when startNewSession(). Error: %v", err)
	}
	expirationTime := time.Now().Add(sessionLength)
	var domain string
	if Environment == EnvironmentLocal {
		domain = "localhost"
	} else {
		domain = ".strengthgadget.com"
	}
	cookie := &http.Cookie{
		Name:     SessionKey,
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

func verifyValidUserProvidedPassword(hashedInput string, userHash string) *Error {
	if subtle.ConstantTimeCompare([]byte(hashedInput), []byte(userHash)) == 1 {
		return nil
	} else {
		return &Error{
			InternalError:     errors.New("user password did not match their password hash"),
			UserFeedbackError: ErrorUserFeedbackWrongPasswordOrUsername,
		}
	}
}

func getSalt(hash string) (string, *Error) {
	fields := strings.Split(hash, "$")
	if len(fields) != 6 {
		return "", &Error{
			InternalError:     errors.New("user hash contains corrupted formatting"),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	return fields[4], nil
}

func HandleIsLoggedIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error, only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	loggedIn, err := isLoggedIn(r)
	if err != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to check if user is logged in. Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
	}

	if loggedIn {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func isLoggedIn(r *http.Request) (bool, error) {
	return IsAuthenticated(r)
}

func validateForgotPasswordNewPasswordRequest(ctx context.Context, req *ForgotPassword, user *User) *Error {
	err := IsVerificationCodeValid(ctx, user, &VerificationRequest{
		Email: req.Email,
		Code:  req.ResetCode,
	}, PasswordResetAttemptType)
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when IsVerificationCodeValid() for validateForgotPasswordNewPasswordRequest() for user email: %s. Error: %v", req.Email, err),
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

	var req *ForgotPassword
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
		GenerateResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func validateForgotPasswordResetCodeRequest(req *ForgotPassword) *Error {
	var errorFeedback []error

	err := emailIsValid(req.Email)
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to emailIsValid() for validateForgotPasswordResetCodeRequest(). Email: %s, error: %v", req.Email, err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	if req.ResetCode == "" {
		errorFeedback = append(errorFeedback, errors.New("reset code is required"))
	}

	if len(errorFeedback) > 0 {
		return &Error{
			InternalError:     fmt.Errorf("errors, when validating request: %v", errorFeedback),
			UserFeedbackError: ErrorPasswordResetCodeIsInvalid,
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

	var req ForgotPassword
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
		GenerateResponse(w, err)
		return
	}

	user, err := GetUser(r.Context(), req.Email)
	if err != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to get user for password reset code: %v", err),
			UserFeedbackError: ErrorUserFeedbackAccessDenied,
		})
		return
	}

	err = IsVerificationCodeValid(r.Context(), user, &VerificationRequest{
		Email: req.Email,
		Code:  req.ResetCode,
	}, PasswordResetAttemptType)
	if err != nil {
		GenerateResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func forgotPasswordNewPassword(ctx context.Context, req *ForgotPassword) (err *Error) {
	err = emailIsValid(req.Email)
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to validate email for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	user, err := GetUser(ctx, req.Email)
	if err != nil {
		var feedback UserFeedbackError
		if err.UserFeedbackError == ErrorUserFeedbackWrongPasswordOrUsername {
			feedback = ErrorEmailDoesNotExists
		} else {
			feedback = ErrorUnexpectedTryAgain
		}
		return &Error{
			InternalError:     fmt.Errorf("error, when GetUser() for forgotPasswordNewPassword() for user: %s. Error: %v", req.Email, err),
			UserFeedbackError: feedback,
		}
	}

	err = validateForgotPasswordNewPasswordRequest(ctx, req, user)
	successfulAttempt := err == nil
	deferErr := RecordAccessAttempt(ctx, user, successfulAttempt, PasswordResetAttemptType)
	if deferErr != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to record password reset attempt for user %s. Defer error: %v. Original error: %v", user.Email, deferErr, err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to validate request for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	err = PasswordIsAcceptable(req.NewPassword)
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to validate password for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	hashedPassword, err := ObtainHashFromPassword(req.NewPassword)
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to hash password for forgotPasswordNewPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}

	e := UpdateUserPassword(ctx, user, hashedPassword)
	if e != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to update password for forgotPasswordNewPassword(). Error: %v", e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	return nil
}

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

	var req User
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
	if !EmailIsValidFormat(email) {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("the user provided an email address with an invalid format when submitting to reset password: %s", email),
			UserFeedbackError: ErrorInvalidEmailAddress,
		})
		return
	}

	user, err := GetUser(r.Context(), email)
	if err != nil {
		if err.UserFeedbackError == ErrorUserFeedbackWrongPasswordOrUsername {
			log.Printf("error, user entered an email that does not exist when reseting password. Email: %s", email)
			return
		} else {
			GenerateResponse(w, err)
			return
		}
	}

	verificationCodeRateLimitReached, e := HasVerificationCodeRateLimitBeenReached(r.Context(), user.Email)
	if e != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, when attempting to check if verification code limit has been reached for user for password reset: %s. Error: %v", email, e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}

	if verificationCodeRateLimitReached {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, user reached the verification code rate limit for password reset. User: %s", user.Email),
			UserFeedbackError: ErrorPasswordResetCodeRateLimitReached,
		})
		return
	}

	e = SendForgotPasswordEmail(r.Context(), user)
	if e != nil {
		GenerateResponse(w, &Error{
			InternalError:     fmt.Errorf("error, failed to send forgot password email: %v", e),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func GetEmailAndPassword(requestBody io.Reader) (*Credentials, *Error) {
	bytes, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, when attempting to read request body: %v", err),
			UserFeedbackError: ErrorUserFeedbackAccessDenied,
		}
	}

	var credentials Credentials
	err = json.Unmarshal(bytes, &credentials)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("unable to decode credentials due to: %v", err),
			UserFeedbackError: ErrorUserFeedbackAccessDenied,
		}
	}
	return &credentials, nil
}

func GetUser(ctx context.Context, email string) (*User, *Error) {
	var user User
	row := ConnectionPool.QueryRow(
		ctx,
		"SELECT id,\n       password_hash,\n       email,\n    (SELECT EXISTS(SELECT 1\n        FROM access_attempt\n        WHERE user_id = \"user\".id\n          AND type = $1\n          AND access_granted = true)) as email_verified\nFROM \"user\"\nWHERE email = $2",
		VerificationAttemptType,
		email,
	)
	err := row.Scan(
		&user.Id,
		&user.PasswordHash,
		&user.Email,
		&user.EmailVerified,
	)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return nil, &Error{
				InternalError:     fmt.Errorf("no user found for email: %v", err),
				UserFeedbackError: ErrorUserFeedbackWrongPasswordOrUsername,
			}
		} else {
			return nil, &Error{
				InternalError:     fmt.Errorf("an unexpected error occurred when attempting to retrieve user: %s. Error: %v", email, err),
				UserFeedbackError: ErrorUnexpectedTryAgain,
			}
		}
	}
	return &user, nil
}

// GenerateSecureHash creating a hash according to this guide: https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go
func GenerateSecureHash(password, salt string) (*string, *Error) {
	decodeString, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("an error has occurred when attempting to base64 decode the salt: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
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
			GenerateResponse(w, &Error{
				InternalError:     fmt.Errorf("user attempted to access a protected resource without being authenticated"),
				UserFeedbackError: ErrorUserNotLoggedIn,
			})
			return
		}

		user := &User{Id: session.UserId}

		// attach the user to the context
		ctx := context.WithValue(r.Context(), SessionKey, user)

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

func PasswordIsAcceptable(password string) *Error {
	if len(password) == 0 {
		return &Error{
			InternalError:     errors.New("user did not provide a password"),
			UserFeedbackError: ErrorPasswordWasNotProvided,
		}
	}

	if len(password) < 12 {
		return &Error{
			InternalError:     errors.New("user chose a password that was too short"),
			UserFeedbackError: ErrorPasswordMustBeAtLeastTwelveCharsLong,
		}
	}

	nonNumberMatch, err := regexp.MatchString("\\D", password)
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("an unexpected error occurred when checking password for non-numeric charactures: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	if !nonNumberMatch {
		return &Error{
			InternalError:     errors.New("user attempted to create a password with all numbers"),
			UserFeedbackError: ErrorPasswordCannotContainAllNumbers,
		}
	}
	return nil
}

func ObtainHashFromPassword(password string) (*string, *Error) {
	salt, err := generateSalt()
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, when generating a salt for ObtainHashFromPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}
	hashedPassword, err := GenerateSecureHash(password, *salt)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, when generating a hash from password for ObtainHashFromPassword(). Error: %v", err),
			UserFeedbackError: err.UserFeedbackError,
		}
	}
	return hashedPassword, nil
}

func generateSalt() (*string, *Error) {
	saltSize := 128 // 128 bits is the salt size used by Bcrypt
	b := make([]byte, saltSize)
	_, err := rand.Read(b)
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("an error has occurred when generating the salt: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	result := base64.StdEncoding.EncodeToString(b)
	return &result, nil
}

func IncrementLoginAttemptCount(ctx context.Context, email string) *Error {
	key := LoginAttemptPrefix + email
	err := IncrementRateLimitingCount(ctx, key, WindowLengthInSecondsForTheNumberOfAllowedLoginAttemptsBeforeLockout)
	if err != nil {
		return &Error{
			InternalError:     fmt.Errorf("error, when attempting to IncrementRateLimitingCount() for IncrementLoginAttemptCount(). Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	return nil
}

func HasLoginAttemptRateLimitBeenReached(ctx context.Context, email string) (bool, *Error) {
	key := LoginAttemptPrefix + email
	result, err := HasRateLimitBeenReached(ctx, key, AllowedLoginAttemptsBeforeTriggeringLockout)
	if err != nil {
		return false, &Error{
			InternalError:     fmt.Errorf("error, when HasRateLimitBeenReached() for HasLoginAttemptRateLimitBeenReached(). Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}
	return result, nil
}

func PersistNewUser(ctx context.Context, email string, hashedPassword *string) error {
	userId := uuid.New().String()
	tx, e := ConnectionPool.Begin(ctx)
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

func UpdateUserPassword(ctx context.Context, user *User, passwordHash *string) error {
	_, err := ConnectionPool.Exec(ctx, "UPDATE \"user\" SET password_hash = $1 WHERE id = $2", passwordHash, user.Id)
	if err != nil {
		return fmt.Errorf("an error has occurred when attempting to update user password: %v", err)
	}
	return nil
}
