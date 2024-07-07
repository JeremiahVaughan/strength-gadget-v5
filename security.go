package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/argon2"
)

const (

	// MockVerificationCode for local development we mock the verification code to avoid sending emails with real email servers
	MockVerificationCode = "ABCDEF"
)

// const (
// 	PasswordResetAttemptType = "14cb4661-74e5-49e8-8532-ebe93d1e806a"
// 	LoginAttemptType         = "288e1dae-5865-4707-b242-ce818ee8145f"
// 	VerificationAttemptType  = "ca33f4f1-e2ba-49e1-8222-be982a57c231"
// )

type AccessAttemptType int

const (
	PasswordResetAttemptType AccessAttemptType = iota
	LoginAttemptType
	VerificationAttemptType
)

// todo validate the length of input so it can't crash your stuff
type VerificationRequest struct {
	Email string
	Code  string
}

type VerificationCode struct {
	Id      string
	Code    string
	Expires uint64
}

type UserSession struct {
	UserId        int64
	Authenticated bool

	WorkoutSessionExists bool
	WorkoutSession       WorkoutSession
}

type WorkoutSession struct {
	// CurrentWorkoutSeed is the concatenation of the current year, julian day, and UserId this ensures a unique workout per user, per day
	// We don't want all users to have the same workout because then they would bottleneck in the gym on the same equipment.
	// We don't want the users getting bored either, so we are giving them a different workout everyday.
	CurrentWorkoutSeed    int64                     `json:"currentWorkoutSeed"`
	CurrentWorkoutRoutine RoutineType               `json:"currentWorkoutRoutine"`
	RandomizedIndices     DailyWorkoutRandomIndices `json:"randomizedIndices"`
	ProgressIndex         int                       `json:"progressIndex"`

	// CurrentOffsets are determined during exercise selection
	CurrentOffsets DailyWorkoutOffsets `json:"currentOffsets"`
}

func (w *WorkoutSession) saveToRedis(ctx context.Context, userId int64) error {
	bytes, err := json.Marshal(w)
	if err != nil {
		return fmt.Errorf("error, when marshalling the new workout session to json. Error: %v", err)
	}

	key := strconv.FormatInt(userId, 10)
	err = RedisConnectionPool.Set(ctx, key, string(bytes), WorkoutSessionExpiration).Err()
	if err != nil {
		return fmt.Errorf("error, when attempting to persist the users workout session in redis. Error: %v", err)
	}
	return nil
}

type User struct {
	Id            int64
	Email         string
	EmailVerified bool
	PasswordHash  string
}

type UserService struct {
}

// userErr Structuring errors to better ensure the user does not get exposed to information he/she shouldn't have.
type userErr string

func (us *UserService) FetchFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(AuthSessionKey).(*User)
	if !ok {
		return nil, fmt.Errorf("error, could not locate the user in session context")
	}
	return user, nil
}

type ContextKey string

// AuthSessionKey use for both locating the user session ID from the http cookie and locating the user struct in context
const AuthSessionKey ContextKey = "auth_session_key"

// WorkoutSessionKey (also userId) workout data is seperated from auth session data so that the user may logout without losing their workout progress
const WorkoutSessionKey ContextKey = "workout_session_key"

type ForgotPassword struct {
	Email       string `json:"email"`
	ResetCode   string `json:"resetCode"`
	NewPassword string `json:"newPassword"`
}

func emailIsValid(email string) userErr {
	if email == "" {
		return ErrorMissingEmailAddress
	}

	if !EmailIsValidFormat(email) {
		return ErrorInvalidEmailAddress
	}

	return ""
}

func emailAlreadyExists(ctx context.Context, email string) (*bool, error) {
	var emailExists bool
	row := ConnectionPool.QueryRow(ctx, "SELECT email FROM athlete WHERE email = $1", email)
	queryErr := row.Scan(&email)
	if queryErr != nil {
		if queryErr.Error() == pgx.ErrNoRows.Error() {
			emailExists = false
		} else {
			return nil, fmt.Errorf("an unexpected error occurred when attempting to check if the email: %s exists. Error: %v", email, queryErr)
		}
	} else {
		emailExists = true
	}

	return &emailExists, nil
}

func validateLogoutRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(string(AuthSessionKey))
	if err != nil {
		return "", fmt.Errorf("error, no session_key provided in request when attempting to logout. Error: %v", err)
	}
	return cookie.Value, nil
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HandleUnexpectedError(w, fmt.Errorf("error, only POST method is supported"))
		return
	}

	sessionKey, err := validateLogoutRequest(r)
	if err != nil {
		HandleUnexpectedError(w, err)
		return
	}

	err = logout(r.Context(), w, sessionKey)
	if err != nil {
		HandleUnexpectedError(w, err)
		return
	}

	w.Header().Set("HX-Redirect", EndpointLogin)
}

func logout(ctx context.Context, w http.ResponseWriter, sessionKey string) error {
	err := RedisConnectionPool.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("unable to delete session. Error: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   string(AuthSessionKey),
		Value:  "",
		MaxAge: -1, // This deletes the cookie
	})

	return nil
}

// todo add if this account exists, you will get an email

func startNewSession(ctx context.Context, userId int64) (*http.Cookie, *http.Cookie, error) {
	// The extra twelve hours is to ensure the user isn't being prompted to, login will in the middle of using the app.
	// For example without the extra twelve hours, if they were to log in to use the app at 0930 am on sunday, then
	// started using the app immediately. The following week they would be logged off at 0930. This means the user is likely
	// to be using the app at this time as it may be part of their schedule. The offset of twelve hours ensures less disruption.
	hoursInOneWeekPlusTwelve := 180 * time.Hour
	authSessionLength := hoursInOneWeekPlusTwelve
	authSessionKey := GenerateSessionKey()
	// purposely using a string as the storage type for sessions because expiring the hash is done in a separate command. If the expiry command were to fail, then this would mean an immortal session was just created.
	err := RedisConnectionPool.Set(ctx, authSessionKey, userId, authSessionLength).Err()
	if err != nil {
		return nil, nil, fmt.Errorf("error, when persisting the session when startNewSession(). Error: %v", err)
	}
	expirationTime := time.Now().Add(authSessionLength)
	var domain string
	if Environment == EnvironmentLocal {
		domain = "localhost"
	} else {
		domain = ".strengthgadget.com"
	}
	authCookie := &http.Cookie{
		Name:     string(AuthSessionKey),
		Value:    authSessionKey,
		HttpOnly: true,
		Secure:   true,
		Expires:  expirationTime,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Domain:   domain,
	}

	uid := strconv.FormatInt(userId, 10)
	workoutCookie := &http.Cookie{
		Name:  string(WorkoutSessionKey),
		Value: uid,
	}

	return authCookie, workoutCookie, nil
}

func verifyValidUserProvidedPassword(hashedInput string, userHash string) userErr {
	if subtle.ConstantTimeCompare([]byte(hashedInput), []byte(userHash)) == 1 {
		return ""
	} else {
		return ErrorUserFeedbackWrongPasswordOrUsername
	}
}

func getSalt(hash string) (string, error) {
	fields := strings.Split(hash, "$")
	if len(fields) != 6 {
		return "", errors.New("user hash contains corrupted formatting")
	}

	return fields[4], nil
}

type ForgotPasswordFields struct {
	NewPassword   TextInput
	PasswordMatch TextInput
	Email         TextInput
	ResetCode     TextInput
	Submit        Button
	ValidForm     bool
}

func GetEmailAndPassword(requestBody io.Reader) (*Credentials, error) {
	bytes, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to read request body: %v", err)
	}

	var credentials Credentials
	err = json.Unmarshal(bytes, &credentials)
	if err != nil {
		return nil, fmt.Errorf("unable to decode credentials due to: %v", err)
	}
	return &credentials, nil
}

func GetUser(ctx context.Context, email string) (*User, userErr, error) {
	var user User
	row := ConnectionPool.QueryRow(
		ctx,
		`SELECT id,
                password_hash,
                email,
                (SELECT EXISTS(SELECT 1
                FROM access_attempt
                WHERE user_id = athlete.id
                    AND type = $1
                    AND access_granted = true)) as email_verified
        FROM athlete
        WHERE email = $2`,
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrorUserFeedbackWrongPasswordOrUsername, nil
		} else {
			return nil, "", fmt.Errorf("an unexpected error occurred when attempting to retrieve user: %s. Error: %v", email, err)
		}
	}
	return &user, "", nil
}

// GenerateSecureHash creating a hash according to this guide: https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go
func GenerateSecureHash(password, salt string) (*string, error) {
	decodeString, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred when attempting to base64 decode the salt: %v", err)
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

func PasswordIsAcceptable(password string) (userErr, error) {
	if len(password) == 0 {
		return ErrorPasswordWasNotProvided, nil
	}

	if len(password) < 12 {
		return ErrorPasswordMustBeAtLeastTwelveCharsLong, nil
	}

	nonNumberMatch, err := regexp.MatchString("\\D", password)
	if err != nil {
		return "", fmt.Errorf("an unexpected error occurred when checking password for non-numeric charactures: %v", err)
	}
	if !nonNumberMatch {
		return ErrorPasswordCannotContainAllNumbers, nil
	}

	return "", nil
}

func ObtainHashFromPassword(password string) (*string, error) {
	salt, err := generateSalt()
	if err != nil {
		return nil, fmt.Errorf("error, when generating a salt for ObtainHashFromPassword(). Error: %v", err)
	}
	hashedPassword, err := GenerateSecureHash(password, *salt)
	if err != nil {
		return nil, fmt.Errorf("error, when generating a hash from password for ObtainHashFromPassword(). Error: %v", err)
	}
	return hashedPassword, nil
}

func generateSalt() (*string, error) {
	saltSize := 128 // 128 bits is the salt size used by Bcrypt
	b := make([]byte, saltSize)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred when generating the salt: %v", err)
	}
	result := base64.StdEncoding.EncodeToString(b)
	return &result, nil
}

func IncrementLoginAttemptCount(ctx context.Context, email string) error {
	key := LoginAttemptPrefix + email
	err := IncrementRateLimitingCount(ctx, key, WindowLengthInSecondsForTheNumberOfAllowedLoginAttemptsBeforeLockout)
	if err != nil {
		return fmt.Errorf("error, when attempting to IncrementRateLimitingCount() for IncrementLoginAttemptCount(). Error: %v", err)
	}
	return nil
}

func HasLoginAttemptRateLimitBeenReached(ctx context.Context, email string) (bool, error) {
	key := LoginAttemptPrefix + email
	result, err := HasRateLimitBeenReached(ctx, key, AllowedLoginAttemptsBeforeTriggeringLockout)
	if err != nil {
		return false, fmt.Errorf("error, when HasRateLimitBeenReached() for HasLoginAttemptRateLimitBeenReached(). Error: %v", err)
	}
	return result, nil
}

func PersistNewUser(ctx context.Context, email string, hashedPassword *string) error {
	tx, e := ConnectionPool.Begin(ctx)
	if e != nil {
		return fmt.Errorf("an error has occurred when attempting to create transaction: %v", e)
	}

	e = func(tx pgx.Tx) error {
		var id int64
		err := tx.QueryRow(
			ctx,
			`INSERT INTO athlete (email, password_hash) 
            VALUES ($1, $2)
            RETURNING id`,
			email,
			hashedPassword,
		).Scan(
			&id,
		)
		if err != nil {
			return fmt.Errorf("error, when attempting to execute sql statement: %v", err)
		}

		e = GenerateNewVerificationCode(ctx, tx, id, email, false)
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
	_, err := ConnectionPool.Exec(ctx, `UPDATE athlete SET password_hash = $1 WHERE id = $2`, passwordHash, user.Id)
	if err != nil {
		return fmt.Errorf("an error has occurred when attempting to update user password: %v", err)
	}
	return nil
}
