package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/nalgeon/redka"
)

func FetchUserSession(r *http.Request) (*UserSession, error) {
	var result UserSession
	authCookie, err := r.Cookie(string(AuthSessionKey))
	if err != nil {
		// Missing auth session id means the user is not logged in
		result.Authenticated = false
		return &result, nil
	}

	workoutCookie, err := r.Cookie(string(WorkoutSessionKey))
	if err != nil {
		// Missing workout session id means the users workout session does not exist
		result.WorkoutSessionExists = false
		return &result, nil
	}
	result.UserId, err = strconv.ParseInt(workoutCookie.Value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error, when parsing workout session cookie. Error: %v", err)
	}

	// Prepare the pipeline commands
	_, err = RedisConnectionPool.Str().Get(authCookie.Value)
	if err != nil {
		if errors.Is(err, redka.ErrNotFound) {
			// Missing session id means the session has expired
			result.Authenticated = false
			return &result, nil
		}
		return nil, fmt.Errorf("error, redis call failed when fetching user id from session for FetchUserSession(). Error: %v", err)
	}
	result.Authenticated = true

	workoutGet, err := RedisConnectionPool.Str().Get(workoutCookie.Value)
	result.WorkoutSessionExists = !errors.Is(err, redka.ErrNotFound)
	if err != nil && result.WorkoutSessionExists {
		return nil, fmt.Errorf("error, when fetching workout session for FetchUserSession(). Error: %v", err)
	}

	if result.WorkoutSessionExists {
		err = json.Unmarshal(workoutGet, &result.WorkoutSession)
		if err != nil {
			return nil, fmt.Errorf("error, when unmarshalling workout session. Error: %v", err)
		}
		return &result, nil
	}

	return &result, nil
}

func GenerateSessionKey() string {
	keyLength := 32 // 256 bits

	// Allocate memory for the key
	key := make([]byte, keyLength)

	// Generate random bytes
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	// Convert to a base64 string
	sessionKey := base64.StdEncoding.EncodeToString(key)

	return sessionKey
}

func EmailIsValidFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
