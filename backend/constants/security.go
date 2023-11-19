package constants

const (

	// MockVerificationCode for local development we mock the verification code to avoid sending emails with real email servers
	MockVerificationCode = "ABCDEF"

	// SessionKey use for both locating the user session ID from the http cookie and locating the user struct in context
	SessionKey = "session_key"
)
