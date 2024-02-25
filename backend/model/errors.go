package model

import (
	"net/http"
)

var (
	ErrorUserFeedbackAccessDenied = UserFeedbackError{
		Message:      "access denied",
		ResponseCode: http.StatusForbidden,
	}
	ErrorUserFeedbackWrongPasswordOrUsername = UserFeedbackError{
		Message:      "incorrect username or password",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorUnexpectedTryAgain = UserFeedbackError{
		Message:      "an unexpected problem has occurred; please try again",
		ResponseCode: http.StatusInternalServerError,
	}

	ErrorCouldNotLocateUserWorkout = UserFeedbackError{
		Message:      "the current workout has expired",
		ResponseCode: http.StatusNotFound,
	}

	ErrorEmailVerificationRateLimitReached = UserFeedbackError{
		Message:      "too many email verifications have been sent recently, please try again later",
		ResponseCode: http.StatusBadRequest,
	}

	ErrorPasswordResetCodeRateLimitReached = UserFeedbackError{
		Message:      "too many password reset codes have been sent recently, please try again later",
		ResponseCode: http.StatusBadRequest,
	}

	ErrorLoginAttemptRateLimitReached = UserFeedbackError{
		Message:      "too many login attempts were made recently, please try again later",
		ResponseCode: http.StatusUnauthorized,
	}
	ErrorEmailAlreadyExists = UserFeedbackError{
		Message:      "An account with this email already exists.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorEmailDoesNotExists = UserFeedbackError{
		Message:      "Could not find an account with this email.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorMissingEmailAddress = UserFeedbackError{
		Message:      "email address not provided",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorInvalidEmailAddress = UserFeedbackError{
		Message:      "invalid email address format",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorUnverifiedEmailAddress = UserFeedbackError{
		Message:      "you must verify your email address before logging in",
		ResponseCode: http.StatusForbidden,
	}
	ErrorUserNotLoggedIn = UserFeedbackError{
		Message:      "you must login to access this functionality",
		ResponseCode: http.StatusUnauthorized,
	}
	ErrorPasswordMustBeAtLeastTwelveCharsLong = UserFeedbackError{
		Message:      "your password must contain at least 12 characters",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorPasswordCannotContainAllNumbers = UserFeedbackError{
		Message:      "your password cannot contain all numbers",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorPasswordWasNotProvided = UserFeedbackError{
		Message:      "must provide a password",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorVerificationCodeHasExpired = UserFeedbackError{
		Message:      "verification code expired.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorVerificationCodeIsInvalid = UserFeedbackError{
		Message:      "Invalid verification code. Please try again.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorPasswordResetCodeIsInvalid = UserFeedbackError{
		Message:      "Invalid password reset code. Please try again.",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorVerificationNoLongerRequired = UserFeedbackError{
		Message:      "This email is already verified",
		ResponseCode: http.StatusBadRequest,
	}
	ErrorLimitReachedOnVerificationAttempts = UserFeedbackError{
		Message:      "Too many recent attempts, please try again tomorrow",
		ResponseCode: http.StatusForbidden,
	}
)
