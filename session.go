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

	"github.com/redis/go-redis/v9"
)

func FetchUserSession(r *http.Request) (*UserSession, error) {
	ctx := r.Context()

	var result UserSession
	authCookie, err := r.Cookie(string(AuthSessionKey))
	if err != nil {
		// Missing auth session id means the user is not logged in
		result.Authenticated = false
		return &result, nil
	}

	workoutCookie, err := r.Cookie(string(WorkoutSessionKey))
	if err != nil {
		// Missing workout session id means the user is not logged in
		result.Authenticated = false
		return &result, nil
	}
	result.UserId, err = strconv.ParseInt(workoutCookie.Value, 10, 64)

	rpl := RedisConnectionPool.Pipeline()

	// Prepare the pipeline commands
	authGet := rpl.Get(ctx, authCookie.Value)
	workoutGet := rpl.Get(ctx, string(workoutCookie.Value))

	// Execute the pipeline
	_, err = rpl.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("error, when executing Redis pipeline: %v", err)
	}

	err = authGet.Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Missing session id means the session has expired
			result.Authenticated = false
			return &result, nil
		}
		return nil, fmt.Errorf("error, redis call failed when fetching user id from session. Error: %v", err)
	}
	result.Authenticated = true

	var workoutSession string
	workoutSession, err = workoutGet.Result()
	result.WorkoutSessionExists = !errors.Is(err, redis.Nil)
	if err != nil && result.WorkoutSessionExists {
		return nil, fmt.Errorf("error, when fetching workoutSession from redis. Error: %v", err)
	}

	if result.WorkoutSessionExists {
		err = json.Unmarshal([]byte(workoutSession), &result.WorkoutSession)
		return &result, nil
	}

	return &result, nil
}

func completeWorkout() error {
	// todo save all exercise measurements (batch)
	// todo increment workout routine
	return nil
}

func startWorkout() error {
	// todo refresh user session so key does not expire during workout
	return nil
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
