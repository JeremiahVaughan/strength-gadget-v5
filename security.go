package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

    "database/sql"
)


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

	// Sharing the same data structure as ChoosenExercisesMap, but unlike
	// choosen exercises where the source of truth is CurrentOffsets, the
	// value of this map representing measurements are persisted in the session.
	WorkoutMeasurements ChoosenExercisesMap `json:"workoutMeasurements"`
}

func (w *WorkoutSession) saveToRedis(userId int64) error {
	bytes, err := json.Marshal(w)
	if err != nil {
		return fmt.Errorf("error, when marshalling the new workout session to json. Error: %v", err)
	}

	key := strconv.FormatInt(userId, 10)
	err = RedisConnectionPool.Str().SetExpires(key, string(bytes), WorkoutSessionExpiration)
	if err != nil {
		return fmt.Errorf("error, when attempting to persist the users workout session in redis. Error: %v", err)
	}
	return nil
}

type User struct {
	Id            int64
	Email         string
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
	_, err := RedisConnectionPool.Key().Delete(sessionKey)
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

func startNewSession(userId int64) (*http.Cookie, *http.Cookie, error) {
	// The extra twelve hours is to ensure the user isn't being prompted to, login will in the middle of using the app.
	// For example without the extra twelve hours, if they were to log in to use the app at 0930 am on sunday, then
	// started using the app immediately. The following week they would be logged off at 0930. This means the user is likely
	// to be using the app at this time as it may be part of their schedule. The offset of twelve hours ensures less disruption.
	hoursInOneWeekPlusTwelve := 180 * time.Hour
	authSessionLength := hoursInOneWeekPlusTwelve
	authSessionKey := GenerateSessionKey()
	// purposely using a string as the storage type for sessions because expiring the hash is done in a separate command. If the expiry command were to fail, then this would mean an immortal session was just created.
	err := RedisConnectionPool.Str().SetExpires(authSessionKey, strconv.FormatInt(userId, 10), authSessionLength)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when persisting the session when startNewSession(). Error: %v", err)
	}
	expirationTime := time.Now().Add(authSessionLength)
	var domain string
	if Environment == EnvironmentLocal {
		domain = "localhost"
	} else {
		domain = "strengthgadget.com"
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


func GetUser(ctx context.Context, email string) (*User, userErr, error) {
	var user User
	row := ConnectionPool.QueryRow(
		`SELECT id,
                email,
        FROM athlete
        WHERE email = ?`,
		email,
	)
	err := row.Scan(
		&user.Id,
		&user.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorUserFeedbackWrongPasswordOrUsername, nil
		} else {
			return nil, "", fmt.Errorf("an unexpected error occurred when attempting to retrieve user: %s. Error: %v", email, err)
		}
	}
	return &user, "", nil
}


func PersistNewUser(email string) error {
    _, err := ConnectionPool.Exec(
        `INSERT INTO athlete (email) 
        VALUES (?)`,
        email,
    )
    if err != nil {
        return fmt.Errorf("error, when attempting to execute sql statement. Error: %v", err)
    }
	return nil
}
