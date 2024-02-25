package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jackc/pgx/v4"
	"io"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
)

func GetEmailAndPassword(requestBody io.Reader) (*model.Credentials, *model.Error) {
	bytes, err := io.ReadAll(requestBody)
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("error, when attempting to read request body: %v", err),
			UserFeedbackError: model.ErrorUserFeedbackAccessDenied,
		}
	}

	var credentials model.Credentials
	err = json.Unmarshal(bytes, &credentials)
	if err != nil {
		return nil, &model.Error{
			InternalError:     fmt.Errorf("unable to decode credentials due to: %v", err),
			UserFeedbackError: model.ErrorUserFeedbackAccessDenied,
		}
	}
	return &credentials, nil
}

func GetAuthHeader(request events.APIGatewayProxyRequest) string {
	return request.Headers["Authentication"]
}

func GetUser(ctx context.Context, email string) (*model.User, *model.Error) {
	var user model.User
	row := config.ConnectionPool.QueryRow(
		ctx,
		"SELECT id,\n       password_hash,\n       email,\n    (SELECT EXISTS(SELECT 1\n        FROM access_attempt\n        WHERE user_id = \"user\".id\n          AND type = $1\n          AND access_granted = true)) as email_verified\nFROM \"user\"\nWHERE email = $2",
		constants.VerificationAttemptType,
		email,
	)
	err := row.Scan(
		&user.Id,
		&user.PasswordHash,
		&user.Email,
		&user.EmailVerified,
	)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return nil, &model.Error{
				InternalError:     fmt.Errorf("no user found for email: %v", err),
				UserFeedbackError: model.ErrorUserFeedbackWrongPasswordOrUsername,
			}
		} else {
			return nil, &model.Error{
				InternalError:     fmt.Errorf("an unexpected error occurred when attempting to retrieve user: %s. Error: %v", email, err),
				UserFeedbackError: model.ErrorUnexpectedTryAgain,
			}
		}
	}
	return &user, nil
}
