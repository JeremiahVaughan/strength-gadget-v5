package model

type User struct {
	Id            string
	Email         string
	EmailVerified bool
	PasswordHash  string
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
