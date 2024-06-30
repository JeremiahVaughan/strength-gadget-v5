package main

import (
	"fmt"
	"net/http"
)

func HandleForgotPasswordEmail(w http.ResponseWriter, r *http.Request) {
	fields := &ForgotPasswordFields{
		Email: TextInput{
			Id:          "email",
			Label:       "Email",
			Placeholder: "email",
		},
		Submit: Button{
			Id:    "submit",
			Label: "submit",
			Type:  "submit",
			Color: PrimaryButtonColor,
		},
	}
	var err error
	if r.Method == http.MethodGet {
		err = returnForgotPasswordEmailPage(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnForgotPasswordEmailPage() for HandleForgotPasswordEmail(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}

	fields, err = resetUserPassword(r, fields)
	if err != nil {
		err = fmt.Errorf("error, when resetUserPassword() for HandleForgotPasswordEmail(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !fields.ValidForm {
		err = returnForgotPasswordEmailForm(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnForgotPasswordEmailForm() for HandleForgotPasswordEmail() due to invalid form. Error: %v", err)
			HandleUnexpectedError(w, err)
		}
	}

	url := fmt.Sprintf("%s?email=%s", EndpointResetCode, fields.Email.Value)
	w.Header().Set("HX-Redirect", url)
}

func returnForgotPasswordEmailPage(w http.ResponseWriter, fields *ForgotPasswordFields) error {
	err := templateMap["forgot-password-email-page.html"].ExecuteTemplate(w, "base", fields)
	if err != nil {
		return fmt.Errorf("error, when attempting to execute template. Error: %v", err)
	}

	return nil
}

func returnForgotPasswordEmailForm(w http.ResponseWriter, fields *ForgotPasswordFields) error {
	err := templateMap["forgot-password-email-form.html"].ExecuteTemplate(w, "forgotPasswordEmailForm", fields)
	if err != nil {
		return fmt.Errorf("error, when attempting to execute template. Error: %v", err)
	}

	return nil
}

func resetUserPassword(r *http.Request, fields *ForgotPasswordFields) (*ForgotPasswordFields, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("error, when parsing from for resetUserPassword(). Error: %v", err)
	}

	fields.Email.Value = r.FormValue(fields.Email.Id)

	if !EmailIsValidFormat(fields.Email.Value) {
		fields.Email.ErrorMsg = ErrorInvalidEmailAddress
		return fields, nil
	}

	user, e, err := GetUser(r.Context(), fields.Email.Value)
	if err != nil {
		return nil, fmt.Errorf("error, when GetUser() for resetUserPassword(). Error: %v", err)
	}
	if e != "" {
		fields.Email.ErrorMsg = ErrorUserFeedbackWrongPasswordOrUsername
		return fields, nil
	}

	verificationCodeRateLimitReached, err := HasVerificationCodeRateLimitBeenReached(r.Context(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to check if verification code limit has been reached for user for password reset: %s. Error: %v", fields.Email.Value, err)
	}

	if verificationCodeRateLimitReached {
		fields.Email.ErrorMsg = ErrorEmailVerificationRateLimitReached
		return fields, nil
	}

	err = SendForgotPasswordEmail(r.Context(), user)
	if err != nil {
		return nil, fmt.Errorf("error, failed to send forgot password email: %v", err)
	}

	fields.ValidForm = true
	return fields, nil
}
