package main

import (
	"fmt"
	"net/http"
)

func HandleRunIntegrationTests(w http.ResponseWriter, r *http.Request) {
	t := IntegrationTestClient{}
	err := t.runIntegrationTests()
	if err != nil {
		msg := fmt.Sprintf("error, when runIntegrationTests() for HandleRunIntegrationTests(). Error: %s", err.Error())
		http.Error(w, msg, http.StatusInternalServerError)
	}
}
