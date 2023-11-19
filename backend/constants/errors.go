package constants

import (
	"net/http"
	"strengthgadget.com/m/v2/model"
)

var (
	ErrorUserFeedbackAccessDenied = model.UserFeedbackError{
		Message:      "access denied",
		ResponseCode: http.StatusForbidden,
	}
	ErrorUserFeedbackWrongPasswordOrUsername = model.UserFeedbackError{
		Message:      "incorrect username or password",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorUnexpectedTryAgain = model.UserFeedbackError{
		Message:      "an unexpected problem has occurred; please try again",
		ResponseCode: http.StatusInternalServerError,
	}

	ErrorEmailVerificationRateLimitReached = model.UserFeedbackError{
		Message:      "too many email verifications have been sent recently, please try again later",
		ResponseCode: http.StatusBadRequest,
	}

	ErrorPasswordResetCodeRateLimitReached = model.UserFeedbackError{
		Message:      "too many password reset codes have been sent recently, please try again later",
		ResponseCode: http.StatusBadRequest,
	}

	ErrorLoginAttemptRateLimitReached = model.UserFeedbackError{
		Message:      "too many login attempts were made recently, please try again later",
		ResponseCode: http.StatusUnauthorized,
	}
	ErrorEmailAlreadyExists = model.UserFeedbackError{
		Message:      "An account with this email already exists.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorEmailDoesNotExists = model.UserFeedbackError{
		Message:      "Could not find an account with this email.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorMissingEmailAddress = model.UserFeedbackError{
		Message:      "email address not provided",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorInvalidEmailAddress = model.UserFeedbackError{
		Message:      "invalid email address format",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorUnverifiedEmailAddress = model.UserFeedbackError{
		Message:      "you must verify your email address before logging in",
		ResponseCode: http.StatusForbidden,
	}
	ErrorUserNotLoggedIn = model.UserFeedbackError{
		Message:      "you must login to access this functionality",
		ResponseCode: http.StatusUnauthorized,
	}
	ErrorPasswordMustBeAtLeastTwelveCharsLong = model.UserFeedbackError{
		Message:      "your password must contain at least 12 characters",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorPasswordCannotContainAllNumbers = model.UserFeedbackError{
		Message:      "your password cannot contain all numbers",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorPasswordWasNotProvided = model.UserFeedbackError{
		Message:      "must provide a password",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorVerificationCodeHasExpired = model.UserFeedbackError{
		Message:      "verification code expired.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorVerificationCodeIsInvalid = model.UserFeedbackError{
		Message:      "Invalid verification code. Please try again.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorPasswordResetCodeIsInvalid = model.UserFeedbackError{
		Message:      "Invalid password reset code. Please try again.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorVerificationNoLongerRequired = model.UserFeedbackError{
		Message:      "This email is already verified",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorLimitReachedOnVerificationAttempts = model.UserFeedbackError{
		Message:      "Too many recent attempts, please try again tomorrow",
		ResponseCode: http.StatusForbidden,
	}
)
