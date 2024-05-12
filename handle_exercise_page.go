package main

import (
	"fmt"
	"net/http"
)

func HandleExercisePage(w http.ResponseWriter, r *http.Request) {
	loggedIn, err := isLoggedIn(r)
	if err != nil {
		err = fmt.Errorf("error, when checking if user is logged in for HandleExercisePage(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !loggedIn {
		http.Redirect(w, r, EndpointLogin, http.StatusFound) // HX-Redirect only works if the page has already been loaded so we have to use full redirect instead
		return
	}

	if r.Method == http.MethodGet {
		err := templateMap["exercise-page.html"].ExecuteTemplate(w, "base", nil)
		if err != nil {
			err = fmt.Errorf("error, when executing exercise page template for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		return
	}
}
