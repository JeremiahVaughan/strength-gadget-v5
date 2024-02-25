package model

import (
	"context"
	"fmt"
)

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
