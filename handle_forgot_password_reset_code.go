package main

import (
	"fmt"
	"net/http"
)

func HandleForgotPasswordResetCode(w http.ResponseWriter, r *http.Request) {
	fields := &ForgotPasswordFields{
		Email: TextInput{
			Id:          "email",
			Label:       "Email",
			Placeholder: "email",
			Type:        "email",
			Disabled:    true,
		},
		ResetCode: TextInput{
			Id:          "resetCode",
			Label:       "Reset Code",
			Placeholder: "reset code",
			Type:        "text",
		},
		Submit: Button{
			Id:    "submit",
			Label: "submit",
			Type:  "submit",
			Color: PrimaryButtonColor,
		},
	}
	fields.Email.Value = r.URL.Query().Get(fields.Email.Id)

	var err error
	if r.Method == http.MethodGet {
		err = returnForgotPasswordResetCodePage(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnForgotPasswordResetCodePage() for HandleForgotPasswordResetCode(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}

	fields, err = confirmUserProvidedCode(r, fields)
	if err != nil {
		err = fmt.Errorf("error, when confirmUserProvidedCode() for HandleForgotPasswordResetCode(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !fields.ValidForm {
		err = returnForgotPasswordResetCodeForm(w, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnForgotPasswordResetCodeForm() for HandleForgotPasswordResetCode() for invalid form. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	}

	// todo consider not hardcoding ids anywhere like below email or resetCode
	url := fmt.Sprintf("%s?email=%s&resetCode=%s", EndpointNewPassword, fields.Email.Value, fields.ResetCode.Value)
	w.Header().Set("HX-Redirect", url)
}

func returnForgotPasswordResetCodePage(w http.ResponseWriter, fields *ForgotPasswordFields) error {
	err := templateMap["forgot-password-reset-code-page.html"].ExecuteTemplate(
		w,
		"base",
		fields,
	)
	if err != nil {
		return fmt.Errorf("error, when executing template. Error: %v", err)
	}

	return nil
}

func returnForgotPasswordResetCodeForm(w http.ResponseWriter, fields *ForgotPasswordFields) error {
	err := templateMap["forgot-password-reset-code-form.html"].ExecuteTemplate(
		w,
		"forgotPasswordResetCodeForm",
		fields,
	)
	if err != nil {
		return fmt.Errorf("error, when executing template. Error: %v", err)
	}

	return nil
}

func confirmUserProvidedCode(r *http.Request, fields *ForgotPasswordFields) (*ForgotPasswordFields, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("error, when parsing form for HandleForgotPasswordResetCode(). Error: %v", err)
	}

	fields.ResetCode.Value = r.FormValue(fields.ResetCode.Id)

	e := emailIsValid(fields.Email.Value)
	if e != "" {
		fields.ResetCode.ErrorMsg = e
		return fields, nil
	}

	if fields.ResetCode.Value == "" {
		fields.ResetCode.ErrorMsg = ErrorMissingResetCode
		return fields, nil
	}

	var user *User
	user, e, err = GetUser(r.Context(), fields.Email.Value)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to get user for password reset code: %v", err)
	}
	if e != "" {
		fields.ResetCode.ErrorMsg = e
		return fields, nil
	}

	e, err = IsVerificationCodeValid(r.Context(), user, fields.ResetCode.Value, PasswordResetAttemptType)
	if err != nil {
		return nil, fmt.Errorf("error, when IsVerificationCodeValid() for confirmUserProvidedCode(). Error: %v", err)
	}
	if e != "" {
		fields.ResetCode.ErrorMsg = e
		return fields, nil
	}

	fields.ValidForm = true
	return fields, nil
}
