package handler

import (
	"context"
	"fmt"
	"net/http"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
)

func validateLogoutRequest(r *http.Request) (string, *model.Error) {
	cookie, err := r.Cookie(constants.SessionKey)
	if err != nil {
		return "", &model.Error{
			InternalError:     fmt.Errorf("error, no session_key provided in request when attempting to logout. Error: %v", err),
			UserFeedbackError: constants.ErrorUserNotLoggedIn,
		}
	}
	return cookie.Value, nil
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		service.GenerateResponse(w, &model.Error{
			InternalError:     fmt.Errorf("error, only POST method is supported"),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		})
		return
	}

	sessionKey, err := validateLogoutRequest(r)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}

	err = logout(r.Context(), w, sessionKey)
	if err != nil {
		service.GenerateResponse(w, err)
		return
	}
}

func logout(ctx context.Context, w http.ResponseWriter, sessionKey string) *model.Error {
	err := config.RedisConnectionPool.Del(ctx, sessionKey).Err()
	if err != nil {
		return &model.Error{
			InternalError:     fmt.Errorf("unable to delete session. Error: %v", err),
			UserFeedbackError: constants.ErrorUnexpectedTryAgain,
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:   constants.SessionKey,
		Value:  "",
		MaxAge: -1, // This deletes the cookie
	})
	return nil
}
