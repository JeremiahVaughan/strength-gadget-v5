package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
)

func IpFilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("error, when attempting to split the remote address into host and port: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		ip := net.ParseIP(clientIP)
		for _, block := range AllowedIpRanges {
			if block.Contains(ip) {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Printf("a request was filtered out because the source IP was not from Cloudflares whitelist. IP: %s", clientIP)
	})
}

func CheckForActiveSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userSession, err := FetchUserSession(r)
		if err != nil {
			err = fmt.Errorf("error, when FetchUserSession() for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		if !userSession.Authenticated {
			next.ServeHTTP(w, r)
			return
		}

        
		redirectToExercisePage(w, r, userSession, false)
	})
}

func setCacheControl(handler http.Handler, maxAge int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the Cache-Control header
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
		// Serve with the original handler
		handler.ServeHTTP(w, r)
	})
}
