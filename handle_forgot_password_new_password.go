package main

import (
	"fmt"
	"net/http"
)

func HandleForgotPasswordNewPassword(w http.ResponseWriter, r *http.Request) {
	fields := &ForgotPasswordFields{
		Email: TextInput{
			Id:          "email",
			Label:       "Email",
			Placeholder: "email",
			Type:        "email",
		},
		ResetCode: TextInput{
			Id:          "resetCode",
			Label:       "Reset Code",
			Placeholder: "reset code",
			Type:        "text",
		},
		NewPassword: TextInput{
			Id:          "password",
			Label:       "New Password",
			Placeholder: "new password",
			Type:        "password",
		},
		PasswordMatch: TextInput{
			Id:          "passwordMatch",
			Label:       "Confirm Password",
			Placeholder: "confirm password",
			Type:        "password",
		},
		Submit: Button{
			Id:    "submit",
			Label: "submit",
			Type:  "submit",
			Color: PrimaryButtonColor,
		},
	}

	fields.Email.Value = r.URL.Query().Get(fields.Email.Id)
	fields.ResetCode.Value = r.URL.Query().Get(fields.ResetCode.Id)

	var err error
	if r.Method == http.MethodGet {
		err = returnForgotPasswordNewPasswordPage(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnForgotPasswordNewPasswordPage() for HandleForgotPasswordNewPassword(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}
	// todo ensure one password is promoting to update the correct credentials
	// todo add resend reset code button to reset password page
	var user *User
	fields, user, err = forgotPasswordNewPassword(r, w, fields)
	if err != nil {
		err = fmt.Errorf("error, when forgotPasswordNewPassword() for HandleForgotPasswordNewPassword(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !fields.ValidForm {
		err = returnForgotPasswordNewPasswordForm(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnForgotPasswordNewPasswordForm() for HandleForgotPasswordNewPassword(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}

	authCookie, workoutCookie, err := startNewSession(r.Context(), user.Id)
	if err != nil {
		err = fmt.Errorf("error, when persisting session key upon login: %v", err)
		HandleUnexpectedError(w, err)
	}

	http.SetCookie(w, authCookie)
	http.SetCookie(w, workoutCookie)

    redirectExercisePage(w, r, nil)
}

func returnForgotPasswordNewPasswordForm(w http.ResponseWriter, fields *ForgotPasswordFields) error {
	err := templateMap["forgot-password-new-password-form.html"].ExecuteTemplate(w, "forgotPasswordNewPasswordForm", fields)
	if err != nil {
		return fmt.Errorf("error, when executing template for returnForgotPasswordNewPasswordForm(). Error: %v", err)
	}

	return nil
}

func returnForgotPasswordNewPasswordPage(w http.ResponseWriter, fields *ForgotPasswordFields) error {
	err := templateMap["forgot-password-new-password-page.html"].ExecuteTemplate(w, "base", fields)
	if err != nil {
		return fmt.Errorf("error, when executing template for returnForgotPasswordNewPasswordPage(). Error: %v", err)
	}

	return nil
}

func forgotPasswordNewPassword(
	r *http.Request,
	w http.ResponseWriter,
	fields *ForgotPasswordFields,
) (*ForgotPasswordFields, *User, error) {
	ctx := r.Context()
	r.ParseForm()
	fields.NewPassword.Value = r.FormValue(fields.NewPassword.Id)
	fields.PasswordMatch.Value = r.FormValue(fields.PasswordMatch.Id)

	e := emailIsValid(fields.Email.Value)
	if e != "" {
		fields.NewPassword.ErrorMsg = e
		return fields, nil, nil
	}

	user, e, err := GetUser(ctx, fields.Email.Value)
	if err != nil {
		return fields, nil, fmt.Errorf("error, when GetUser() for forgotPasswordNewPassword() for user: %s. Error: %v", fields.Email.Value, err)
	}
	if e != "" {
		fields.NewPassword.ErrorMsg = e
		return fields, nil, nil
	}

	if fields.NewPassword.Value != fields.PasswordMatch.Value {
		fields.PasswordMatch.ErrorMsg = ErrorUserPasswordAndMatchPasswordDidNotMatch
		return fields, nil, nil
	}

	e, err = IsVerificationCodeValid(
		ctx,
		user,
		fields.ResetCode.Value,
		PasswordResetAttemptType,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when IsVerificationCodeValid() for forgotPasswordNewPassword(). Error: %v", err)
	}

	successfulAttempt := e == ""
	err2 := RecordAccessAttempt(ctx, user, successfulAttempt, PasswordResetAttemptType)
	if err2 != nil {
		return nil, nil, fmt.Errorf("error, when attempting to record password reset attempt for user %s. Defer error: %v. Original error: %v", user.Email, err2, err)
	}
	if !successfulAttempt {
		// todo provide feedback that the user needs to attempt to get another password reset code sent because the current one expired
		w.Header().Set("HX-Redirect", getVerificationRedirectAddress(fields.Email.Value))
		return fields, nil, nil
	}

	e, err = PasswordIsAcceptable(fields.NewPassword.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when attempting to validate password for forgotPasswordNewPassword(). Error: %v", err)
	}
	if e != "" {
		fields.NewPassword.ErrorMsg = e
		return fields, nil, nil
	}

	var hashedPassword *string
	hashedPassword, err = ObtainHashFromPassword(fields.NewPassword.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when attempting to hash password for forgotPasswordNewPassword(). Error: %v", err)
	}

	err = UpdateUserPassword(ctx, user, hashedPassword)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when attempting to update password for forgotPasswordNewPassword(). Error: %v", e)
	}

	fields.ValidForm = true
	return fields, user, nil
}
