package main

import (
	"fmt"
	"net/http"
)

func HandleAlreadyAuthenticated(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := templateMap["already_auth_page.html"].ExecuteTemplate(w, "base", nil)
		if err != nil {
			err = fmt.Errorf("error, when attempting to execute template for HandleAlreadyAuthenticated(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	case http.MethodPost:
		redirectExercisePage(w, r)
	}
}
