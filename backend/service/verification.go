package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/jackc/pgx/v4"
	"io"
	"log"
	"net/smtp"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strings"
	"time"
)

func sendEmailVerification(email string, verificationCode string, isPasswordReset bool) (err error) {
	from := config.RegistrationEmailFrom
	password := config.RegistrationEmailFromPassword // app password, not user password

	toEmailAddress := email
	to := []string{toEmailAddress}

	host := "smtp.gmail.com"
	port := "587"

	var subjectPurpose string
	if isPasswordReset {
		subjectPurpose = "Password Reset"
	} else {
		subjectPurpose = "Email Verification"
	}
	subject := fmt.Sprintf("strengthgadget.com %s", subjectPurpose)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	// generating html body with: https://premailer.dialect.ca/

	var bodyPurpose string
	var enterCodeLink string
	if isPasswordReset {
		bodyPurpose = "password reset"
		enterCodeLink = fmt.Sprintf("%s/forgotPassword/resetCode?email=%s", config.TrustedUiOrigins[0], email)
	} else {
		bodyPurpose = "email verification"
		enterCodeLink = fmt.Sprintf("%s/verification?email=%s", config.TrustedUiOrigins[0], email)
	}
	body := fmt.Sprintf("<div>Your %s code is: %s</div>"+
		"<div>Enter the code here: %s</div>", bodyPurpose, verificationCode, enterCodeLink)
	message := []byte(fmt.Sprintf("Subject: %s\n%s\n\n%s", subject, mime, body))

	decodedCert, err := base64.StdEncoding.DecodeString(config.EmailRootCa)
	if err != nil {
		return fmt.Errorf("error, when attempting to decode the gmail root CA: %v", err)
	}
	certPool := x509.NewCertPool()

	if !certPool.AppendCertsFromPEM(decodedCert) {
		return fmt.Errorf("Failed to append PEM.\n")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
		RootCAs:            certPool,
	}

	address := host + ":" + port
	c, err := smtp.Dial(address)
	if err != nil {
		return fmt.Errorf("Failed to connect: %v\n", err)
	}
	defer func(smtpClient *smtp.Client) {
		_ = smtpClient.Close()
	}(c)
	e := c.StartTLS(tlsConfig)
	if e != nil {
		return fmt.Errorf("error, when attempting to start TLS with email server. Error: %v", e)
	}

	a := smtp.PlainAuth("", from, password, host)
	e = c.Auth(a)
	if e != nil {
		return fmt.Errorf("error, unable to authenticate with email server. Error: %v", e)
	}
	e = c.Mail(from)
	if e != nil {
		return fmt.Errorf("error, when initiating email transaction: %v", e)
	}
	e = c.Rcpt(to[0])
	if e != nil {
		return fmt.Errorf("error, when attempting to tell the smtp server who is recieving the email: %v", e)
	}
	w, e := c.Data()
	if e != nil {
		return fmt.Errorf("error, when attempting to retrieve writer from the email server: %v", e)
	}
	_, e = w.Write(message)
	if e != nil {
		return fmt.Errorf("error, when attempting to write message to email server: %v", e)
	}
	e = w.Close()
	if e != nil {
		return fmt.Errorf("error, when attempting to close the email writer: %v", e)
	}
	return c.Quit()
}

func generateEmailVerificationCode() (string, error) {
	// table using only capitals since they are easier to read than lower case.
	// excluding these characters: 'I1O0' because they do not look unique making them harder to recognize.
	table := [...]byte{'2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
	codeLength := config.VerificationCodeLength
	result := make([]byte, codeLength)
	n, err := io.ReadAtLeast(rand.Reader, result, codeLength)
	if n != codeLength || err != nil {
		return "", fmt.Errorf("an unexpected error has occurred when generating random email verification code: %v", err)
	}
	for i := 0; i < len(result); i++ {
		result[i] = table[int(result[i])%len(table)]
	}
	return string(result), nil
}

func GenerateNewVerificationCode(ctx context.Context, tx pgx.Tx, userId string, email string, isPasswordReset bool) error {
	err := IncrementVerificationAttemptCount(ctx, email)
	if err != nil {
		return fmt.Errorf("error, when incrementing email rate limit: %v", err)
	}

	var verificationCode string
	verificationCode, err = generateEmailVerificationCode()
	if err != nil {
		return fmt.Errorf("error, when generating verification code: %v", err)
	}

	expiration := time.Now().Add(time.Minute * time.Duration(config.VerificationCodeValidityWindowInMin)).Unix()
	_, e := tx.Exec(
		ctx,
		"INSERT INTO \"verification_code\" (code, user_id, expires) VALUES ($1, $2, $3)",
		verificationCode,
		userId,
		expiration,
	)
	if e != nil {
		return fmt.Errorf("an error has occurred when attempting to generate a verification code. Error: %v", e)
	}

	e = sendEmailVerification(email, verificationCode, isPasswordReset)
	if e != nil {
		return fmt.Errorf("error, when sending email verification: %v", e)
	}
	return nil
}

func IncrementVerificationAttemptCount(ctx context.Context, email string) error {
	key := model.EmailRateLimitPrefix + email
	err := IncrementRateLimitingCount(ctx, key, config.WindowLengthInSecondsForTheNumberOfAllowedVerificationEmailsBeforeLockout)
	if err != nil {
		return fmt.Errorf("error, when attempting to increment rate limiting count for email. Email: %s. Error: %v", email, err)
	}
	return nil
}

func IsVerificationCodeValid(ctx context.Context, user *model.User, verificationRequest *model.VerificationRequest, verificationAttemptType string) *model.Error {
	count, err := getRecentVerificationCount(ctx, user)
	if err != nil {
		return err
	}

	if hasVerificationLimitBeenReached(*count, config.AllowedVerificationAttemptsWithTheExcessiveRetryLockoutWindow) {
		return &model.Error{
			InternalError:     fmt.Errorf("error, user %s has attempted verification too many times in a small window of time", user.Email),
			UserFeedbackError: constants.ErrorLimitReachedOnVerificationAttempts,
		}
	}

	var persistedCode *model.VerificationCode
	persistedCode, err = getMostRecentVerificationCode(ctx, user)
	if err != nil {
		return err
	}

	var userFeedbackError model.UserFeedbackError
	if verificationAttemptType == constants.PasswordResetAttemptType {
		userFeedbackError = constants.ErrorPasswordResetCodeIsInvalid
	} else {
		userFeedbackError = constants.ErrorVerificationCodeIsInvalid
	}
	codeMatch := doesProvidedCodeMatchExpectedCode(verificationRequest.Code, persistedCode.Code)
	if !codeMatch {
		return &model.Error{
			InternalError:     fmt.Errorf("error, user %s provided an invalid verification code", user.Email),
			UserFeedbackError: userFeedbackError,
		}
	}

	expired := isCodeExpired(persistedCode, time.Now().Unix())
	if expired {
		return &model.Error{
			InternalError:     fmt.Errorf("error, user %s provided an expired code", user.Email),
			UserFeedbackError: constants.ErrorVerificationCodeHasExpired,
		}
	}

	return nil
}

func RecordAccessAttempt(ctx context.Context, user *model.User, successfulAttempt bool, accessAttemptType string) error {
	retryBackoffLimit := 3
	var err error
	for i := 0; i < retryBackoffLimit; i++ {
		err = persistAccessAttemptInDatabase(ctx, user, successfulAttempt, accessAttemptType)
		if err != nil {
			log.Printf("error, failed to persistAccessAttemptInDatabase(). Retrying in 3 seconds: %v", err)
		} else {
			break
		}
		time.Sleep(time.Second * 3)
	}
	if err != nil {
		return fmt.Errorf("error, when after exausting retries to persistAccessAttemptInDatabase() for user: %s. Access type:  %s. Error: %v", user.Email, accessAttemptType, err)
	}
	return nil
}

func persistAccessAttemptInDatabase(ctx context.Context, user *model.User, successfulAttempt bool, accessAttemptType string) error {
	_, err := config.ConnectionPool.Exec(
		ctx,
		"INSERT INTO access_attempt (time, access_granted, type, user_id)\nVALUES ($1, $2, $3, $4)",
		time.Now().Unix(),
		successfulAttempt,
		accessAttemptType,
		user.Id,
	)
	if err != nil {
		return fmt.Errorf("error, when executing sql request for recording verification attempt: %v", err)
	}
	return nil
}

func doesProvidedCodeMatchExpectedCode(provided, expected string) bool {
	providedCode := []byte(strings.ToLower(provided))
	expectedCode := []byte(strings.ToLower(expected))
	if subtle.ConstantTimeCompare(providedCode, expectedCode) == 1 {
		return true
	} else {
		return false
	}
}

func isCodeExpired(code *model.VerificationCode, currentTime int64) bool {
	return currentTime > int64(code.Expires)
}

func getMostRecentVerificationCode(ctx context.Context, user *model.User) (*model.VerificationCode, *model.Error) {
	var verificationCode model.VerificationCode
	e := config.ConnectionPool.QueryRow(
		ctx,
		"SELECT id,\n       code,\n       user_id,\n       expires\nFROM verification_code\nWHERE user_id = $1\nORDER BY expires DESC\nLIMIT 1",
		user.Id,
	).Scan(
		&verificationCode.Id,
		&verificationCode.Code,
		&verificationCode.UserId,
		&verificationCode.Expires,
	)
	if e != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to execute sql statement for getMostRecentVerificationCode(): %v", e),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		}
	}
	return &verificationCode, nil
}

func hasVerificationLimitBeenReached(currentCount int, allowedCount int) bool {
	return currentCount > allowedCount
}

func getRecentVerificationCount(ctx context.Context, user *model.User) (*int, *model.Error) {
	currentTime := time.Now().Unix()
	startOfVerificationLimitWindow := getEarlierTime(currentTime, config.VerificationExcessiveRetryAttemptLockoutDurationInSeconds)
	row := config.ConnectionPool.QueryRow(
		ctx,
		"SELECT count(1)\nFROM access_attempt\nWHERE $1 < time\n  AND type = $2\n  AND user_id = (SELECT id FROM \"user\" WHERE email = $3)",
		startOfVerificationLimitWindow,
		constants.VerificationAttemptType,
		user.Email,
	)
	var count int
	queryErr := row.Scan(&count)
	if queryErr != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("error has occurred when attempting to scan result of access attempts made within the past 24 hours: %v", queryErr),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		}
	}
	return &count, nil
}

func getEarlierTime(theCurrentTime int64, secondsAgo int) int64 {
	return theCurrentTime - int64(secondsAgo)
}

func HasVerificationCodeRateLimitBeenReached(ctx context.Context, email string) (bool, error) {
	key := model.EmailRateLimitPrefix + email
	return HasRateLimitBeenReached(ctx, key, config.AllowedVerificationResendCodeAttemptsWithinOneHour)
}
