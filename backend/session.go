package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/mail"
)

func checkForValidActiveSession(r *http.Request) (*UserSession, *Error) {
	var result UserSession
	cookie, err := r.Cookie(SessionKey)
	if err != nil {
		// Missing session id means the user is not logged in
		result.Authenticated = false
		return &result, nil
	}

	// todo try to find a way to reduce the number of redis calls from 2 to 1 without making error handling complicated.
	sessionKey := cookie.Value
	sessionKeyExists, err := RedisConnectionPool.Exists(r.Context(), sessionKey).Result()
	if err != nil {
		return nil, &Error{
			InternalError:     fmt.Errorf("error, redis call failed when checking if session exists. Error: %v", err),
			UserFeedbackError: ErrorUnexpectedTryAgain,
		}
	}

	result.SessionKey = sessionKey
	if sessionKeyExists == 1 {
		var userId string
		userId, err = RedisConnectionPool.Get(r.Context(), sessionKey).Result()
		if err != nil {
			return nil, &Error{
				InternalError:     fmt.Errorf("error, redis call failed when fetching user id from session. Error: %v", err),
				UserFeedbackError: ErrorUnexpectedTryAgain,
			}
		}
		result.UserId = userId
		result.Authenticated = true
	} else {
		result.Authenticated = false
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
