package main

import (
	"fmt"
	"net/http"
)

type VerificationFields struct {
	Email            TextInput
	ConfirmationCode TextInput
	ConfirmPassword  TextInput
	Submit           Button
	ResendCode       Button
	ValidForm        bool
	ResentMessage    string
}

func HandleVerification(w http.ResponseWriter, r *http.Request) {
	fields := &VerificationFields{
		Email: TextInput{
			Id:          "email",
			Label:       "Email",
			Placeholder: "email",
			Type:        "email",
			Disabled:    true,
		},
		ConfirmationCode: TextInput{
			Id:          "confirmationCode",
			Label:       "Verification Code",
			Placeholder: "verification code",
			Type:        "text",
		},
		ResendCode: Button{
			Id:    "resendCodeButton",
			Label: "Resend Code",
			Color: SecondaryButtonColor,
			Type:  ButtonTypeRegular,
		},
		Submit: Button{
			Id:    "submitButton",
			Label: "Submit",
			Color: PrimaryButtonColor,
			Type:  ButtonTypeSubmit,
		},
	}

	err := r.ParseForm()
	if err != nil {
		err = fmt.Errorf("error, when parsing form for HandleVerification(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}

	fields.Email.Value = r.URL.Query().Get(fields.Email.Id)
	fields.ConfirmationCode.Value = r.FormValue(fields.ConfirmationCode.Id)

	if r.Method == http.MethodGet {
		err = returnVerificationPage(w, r, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnVerificationPage() for HandleVerification(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}

	if r.Method == http.MethodPut {
		fields, err = HandleResendVerification(w, r, fields)
		if err != nil {
			err = fmt.Errorf("error, when HandleResendVerification() for HandleVerification(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		if fields.ValidForm {
			fields.ResentMessage = "code has been sent again, please check your email"
		}
		err = returnVerificationForm(w, r, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnVerificationForm() for HandleVerification(). Error: %v", err)
			HandleUnexpectedError(w, err)
		}
		return
	}

	var user *User
	fields, user, err = handleVerificationAttempt(r, fields)
	if err != nil {
		err = fmt.Errorf("error, when handleVerificationAttempt() for HandleVerification(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !fields.ValidForm {
		err = returnVerificationForm(w, r, fields)
		if err != nil {
			err = fmt.Errorf("error, when returnVerificationForm() for handleVerificationAttempt() for error unexpected. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}
	authCookie, workoutCookie, err := startNewSession(r.Context(), user.Id)
	if err != nil {
		err = fmt.Errorf("error, when persisting session key upon verification completion: %v", err)
		HandleUnexpectedError(w, err)
		return
	}

	http.SetCookie(w, authCookie)
	http.SetCookie(w, workoutCookie)
	w.Header().Set("HX-Redirect", EndpointExercise)
}

func handleVerificationAttempt(r *http.Request, fields *VerificationFields) (*VerificationFields, *User, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, nil, fmt.Errorf("error, when attempting to parse form for handleVerificationAttempt(). Error: %v", err)
	}
	fields.Email.Value = r.URL.Query().Get(fields.Email.Id)
	fields.ConfirmationCode.Value = r.FormValue(fields.ConfirmationCode.Id)

	e := emailIsValid(fields.Email.Value)
	if e != "" {
		fields.ConfirmationCode.ErrorMsg = e
		return fields, nil, nil
	}

	// choosing to let empty codes not count against verification attempts since this doesn't give an attacker any advantage
	if fields.ConfirmationCode.Value == "" {
		fields.ConfirmationCode.ErrorMsg = ErrorVerificationCodeIsInvalid
		return fields, nil, nil
	}

	user, e, err := GetUser(r.Context(), fields.Email.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when GetUser() for handleVerificationAttempt(). Error: %v", err)
	}
	if e != "" {
		fields.ConfirmationCode.ErrorMsg = e
		return fields, nil, nil
	}

	if user.EmailVerified {
		// user email has already been verified
		fields.ConfirmationCode.ErrorMsg = ErrorVerificationCodeAlreadyVerified
		return fields, nil, nil
	}

	e, err = IsVerificationCodeValid(r.Context(), user, fields.ConfirmationCode.Value, VerificationAttemptType)
	if err != nil {
		return nil, nil, fmt.Errorf("error, when attempting to record verification attempt for IsVerificationCodeValid(). Error: %v", err)
	}

	successfulAttempt := e == ""
	err2 := RecordAccessAttempt(r.Context(), user, successfulAttempt, VerificationAttemptType)
	if err2 != nil {
		return nil, nil, fmt.Errorf("error, when attempting to record verification attempt for user %s. Defer error: %v. Original error: %v", user.Email, err2, err)
	}

	if e != "" {
		fields.ConfirmationCode.ErrorMsg = e
		return fields, nil, nil
	}

	fields.ValidForm = true
	return fields, user, nil
}

func returnVerificationPage(w http.ResponseWriter, r *http.Request, fields *VerificationFields) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("error, when attempting to parse form for returnVerificationPage(). Error: %v", err)
	}
	err = templateMap["verification-page.html"].ExecuteTemplate(w, "base", fields)
	if err != nil {
		return fmt.Errorf("error, when attempting to execute template for returnVerificationPage(). Error: %v", err)
	}
	return nil
}

func returnVerificationForm(w http.ResponseWriter, _ *http.Request, fields *VerificationFields) error {
	err := templateMap["verification-form.html"].ExecuteTemplate(w, "verificationForm", fields)
	if err != nil {
		return fmt.Errorf("error, when attempting to execute template for returnVerificationForm(). Error: %v", err)
	}
	return nil
}
