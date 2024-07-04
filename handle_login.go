package main

import (
	"fmt"
	"net/http"
)

type LoginFields struct {
	Email     TextInput
	Password  TextInput
	Submit    Button
	FormValid bool
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	fields := &LoginFields{
		Email: TextInput{
			Id:          "email",
			Label:       "Email",
			Placeholder: "email",
			Type:        "email",
		},
		Password: TextInput{
			Id:          "password",
			Label:       "Password",
			Placeholder: "password",
			Type:        "password",
		},
		Submit: Button{
			Id:    "submit",
			Label: "Submit",
			Color: PrimaryButtonColor,
			Type:  "submit",
		},
	}

	var err error
	if r.Method == http.MethodGet {
		err = returnLoginPage(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnLoginPage() for HandleLogin(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}
	var user *User
	fields, user, err = attemptLogin(w, r, fields)
	if err != nil {
		err = fmt.Errorf("error, when attemptLogin() for HandleLogin(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !fields.FormValid {
		err = returnLoginForm(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when attempting returnLoginForm() for HandleLogin(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}

	authCookie, workoutCookie, err := startNewSession(r.Context(), user.Id)
	if err != nil {
		err = fmt.Errorf("error, when persisting session key upon login: %v", err)
		HandleUnexpectedError(w, err)
		return
	}

    // this cookie holds the auth session id use for authentication
	http.SetCookie(w, authCookie)

    // this cookie holds the userId so the workout session can be retrieved
	http.SetCookie(w, workoutCookie)

    redirectToExercisePage(w, r, nil)
}

func returnLoginForm(w http.ResponseWriter, fields *LoginFields) error {
	err := templateMap["login-form.html"].ExecuteTemplate(w, "loginForm", fields)
	if err != nil {
		return fmt.Errorf("error, when executing template login-form.html. Error: %v", err)
	}

	return nil
}

func returnLoginPage(w http.ResponseWriter, fields *LoginFields) error {
	err := templateMap["login-page.html"].ExecuteTemplate(w, "base", fields)
	if err != nil {
		return fmt.Errorf("error, when executing template login-form.html. Error: %v", err)
	}

	return nil
}

func attemptLogin(w http.ResponseWriter, r *http.Request, fields *LoginFields) (*LoginFields, *User, error) {
	ctx := r.Context()
	err := r.ParseForm()
	if err != nil {
		return nil, nil, fmt.Errorf("error, when attempting to parse form for attemptLogin(). Error: %v", err)
	}

	fields.Email.Value = r.FormValue(fields.Email.Id)
	fields.Password.Value = r.FormValue(fields.Password.Id)

	var loginAttemptLockout bool
	loginAttemptLockout, err = HasLoginAttemptRateLimitBeenReached(ctx, fields.Email.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when HasLoginAttemptRateLimitBeenReached() for attemptLogin(). Error: %v", err)
	}

	if loginAttemptLockout {
		fields.Email.ErrorMsg = ErrorLoginAttemptRateLimitReached
		return fields, nil, nil
	}

	err = IncrementLoginAttemptCount(r.Context(), fields.Email.Value)
	if err != nil {
		return fields, nil, fmt.Errorf("error, when IncrementLoginAttemptCount() for attemptLogin(). Error: %v", err)
	}

	user, e, err := GetUser(ctx, fields.Email.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when GetUser() for attemptLogin(). Error: %v", err)
	}
	if e != "" {
		fields.Email.ErrorMsg = e
		return fields, nil, nil
	}

	salt, err := getSalt(user.PasswordHash)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when getSalt() for attemptLogin(). Error: %v", err)
	}

	hashedInput, err := GenerateSecureHash(fields.Password.Value, salt)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when GenerateSecureHash() for attemptLogin(). Error: %v", err)
	}

	// todo compare legacy versions of the hash should new versions of Argon2 come out or if the parameters change.
	// todo also, update the hash if a legacy version of the hash is detected. -- is this even possible? -- Maybe I can verify the hash with the legacy version of Argon2 somehow then if it matches. Create with the new version
	e = verifyValidUserProvidedPassword(*hashedInput, user.PasswordHash)
	successfulAttempt := e == ""
	deferErr := RecordAccessAttempt(ctx, user, successfulAttempt, LoginAttemptType)
	if deferErr != nil {
		HandleUnexpectedError(w, fmt.Errorf("error, when attempting to record login attempt for user %s. Defer error: %v. Original error: %v", user.Email, deferErr, err))
	}
	if e != "" {
		fields.Password.ErrorMsg = e
		return fields, nil, nil
	}

	if !user.EmailVerified {
		// todo need to provide user with feedback that they need to verify their email
		w.Header().Set("HX-Redirect", getVerificationRedirectAddress(fields.Email.Value))
		return nil, nil, nil
	}

	fields.FormValid = true
	return fields, user, nil
}
