package main

import (
	"fmt"
	"net/http"
)

func HandleLandingPage(w http.ResponseWriter, r *http.Request) {
	type Buttons struct {
		Login  Button
		SignUp Button
	}

	err := templateMap["landing-page.html"].ExecuteTemplate(w, "base", Buttons{
		Login: Button{
			Id:    "login",
			Label: "Login",
			Color: PrimaryButtonColor,
			Type:  ButtonTypeRegular,
		},
		SignUp: Button{
			Id:    "sign_up",
			Label: "Sign Up",
			Color: SecondaryButtonColor,
			Type:  ButtonTypeRegular,
		},
	})

	if err != nil {
		err = fmt.Errorf("error, when attempting to execute template for HandleLandingPage(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
}
