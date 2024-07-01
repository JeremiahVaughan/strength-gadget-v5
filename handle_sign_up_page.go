package main

import (
	"fmt"
	"net/http"
)

type SignUpFields struct {
	Email           TextInput
	Password        TextInput
	ConfirmPassword TextInput
	Submit          Button
	ValidForm       bool
}

func HandleSignUp(w http.ResponseWriter, r *http.Request) {
	fields := SignUpFields{
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
		ConfirmPassword: TextInput{
			Id:          "passwordMatch",
			Label:       "Confirm Password",
			Placeholder: "confirm password",
			Type:        "password",
		},
		Submit: Button{
			Id:    "submit_button",
			Label: "Submit",
			Color: PrimaryButtonColor,
			Type:  ButtonTypeSubmit,
		},
	}

	var err error
	if r.Method == http.MethodGet {
		err = returnSignUpPage(w, &fields)
		if err != nil {
			HandleUnexpectedError(w, err)
			return
		}
		return

	}

	err = HandleRegister(w, r, &fields)
	if err != nil {
		err = fmt.Errorf("error, when HandleRegister() for HandleSignUp(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}

	w.Header().Set("HX-Redirect", getVerificationRedirectAddress(fields.Email.Value))
}

func returnSignUpForm(w http.ResponseWriter, f *SignUpFields) error {
	err := templateMap["sign-up-form.html"].ExecuteTemplate(w, "signUpForm", f)
	if err != nil {
		return fmt.Errorf("error, when attempting to execute template for returnSignUpPage(). Error: %v", err)
	}
	return nil
}

func returnSignUpPage(w http.ResponseWriter, f *SignUpFields) error {
	err := templateMap["sign-up-page.html"].ExecuteTemplate(w, "base", f)
	if err != nil {
		return fmt.Errorf("error, when attempting to execute template for returnSignUpPage(). Error: %v", err)
	}
	return nil
}

func HandleRegister(w http.ResponseWriter, r *http.Request, fields *SignUpFields) (err error) {
	err = r.ParseForm()
	if err != nil {
		return fmt.Errorf("error, unable to parse form for HandleRegister(). Error: %v", err)
	}

	fields.Email.Value = r.FormValue(fields.Email.Id)
	fields.Password.Value = r.FormValue(fields.Password.Id)
	fields.ConfirmPassword.Value = r.FormValue(fields.ConfirmPassword.Id)

	fields, err = validateRequest(r, fields)
	if err != nil {
		return fmt.Errorf("error, when validateRequest() for HandleRegister(). Error: %v", err)
	}

	if !fields.ValidForm {
		err = returnSignUpForm(w, fields)
		if err != nil {
			return fmt.Errorf("error, when returnSignUpForm() for HandleRegister(). Error: %v", err)
		}
		return nil
	}

	var hashedPassword *string
	hashedPassword, err = ObtainHashFromPassword(fields.Password.Value)
	if err != nil {
		return fmt.Errorf("error, when attempting to hash password for HandleRegister(). Error: %v", err)
	}

	err = PersistNewUser(r.Context(), fields.Email.Value, hashedPassword)
	if err != nil {
		return fmt.Errorf("error, when attempting to persist new user: %v", err)
	}

	//// todo add a way for users to change their email
	//// todo add terms of service and privacy policy checkbox
	//// todo for email verification go with verification code
	//// 		todo use all caps for ease of readability but accept caps or lowercase for the sake of user experience.
	//// 		todo use code verification because of the silly belief that clicking links in email is evil
	////				todo the silly believe makes people hesitate to click emails
	////  			todo the silly belief makes emails end up in spam folders
	//
	return nil
}

func validateRequest(r *http.Request, fields *SignUpFields) (*SignUpFields, error) {

	// todo add rate limiter
	e := emailIsValid(fields.Email.Value)
	if e != "" {
		fields.Email.ErrorMsg = e
		return fields, nil
	}

	// todo find a way for emails to become available again if verification takes too long. This prevents an edge case where the wrong email is accidentally used and prevents the person who actually owns that email from creating an account.
	// todo let user know in the UI right after submitting registration that they must verify within blank hours otherwise they will have to reregister.
	emailExists, err := emailAlreadyExists(r.Context(), fields.Email.Value)
	if err != nil {
		return nil, err
	}

	if *emailExists {
		fields.Email.ErrorMsg = ErrorEmailAlreadyExists
		return fields, nil
	}

	e, err = PasswordIsAcceptable(fields.Password.Value)
	if err != nil {
		return nil, fmt.Errorf("error, when PasswordIsAcceptable() for validateRequest(). Error: %v", err)
	}
	if e != "" {
		fields.Password.ErrorMsg = e
		return fields, nil
	}

	if fields.Password.Value != fields.ConfirmPassword.Value {
		fields.ConfirmPassword.ErrorMsg = ErrorUserPasswordAndMatchPasswordDidNotMatch
		return fields, nil
	}

	fields.ValidForm = true
	return fields, nil
}
