package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templatesFiles embed.FS

//go:embed database/*
var databaseFiles embed.FS

var templateMap map[string]*template.Template

func main() {
	log.Printf("starting strengthgadget...")

	ctx := context.Background()
	err := InitConfig(ctx)
	if err != nil {
		log.Fatalf("error, attempting to initialize configuration: %v", err)
	}

	defer ConnectionPool.Close()
	err = sentry.Init(sentry.ClientOptions{
		Dsn: SentryEndpoint,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,

		Environment: Environment,
	})
	if err != nil {
		log.Fatalf("error, sentry.Init(): %s", err)
	}

	sentry.CaptureMessage(fmt.Sprintf("Strengthgadget backend has started in the %s environment", Environment))
	defer sentry.Flush(2 * time.Second)

	// pageTemplateMap key is the page template name (e.g., landing-page.html)
	templateMap = make(map[string]*template.Template)
	err = registerTemplates()
	if err != nil {
		log.Fatalf("error, when registerTemplates() for main(). Error: %v", err)
	}

	err = serveAthletes(ctx)
	if err != nil {
		log.Fatalf("error, when serveAthletes() for main(). Error: %v", err)
	}
}

func serveAthletes(ctx context.Context) error {
	err := ProcessSchemaChanges(ctx, databaseFiles)
	if err != nil {
		return fmt.Errorf("error, when attempting to process schema changes: %v", err)
	}

	// todo add a 404 not found page for invalid addresses
	mux := http.NewServeMux()

	// endpoints key is endpoint address
	endpoints := map[string]http.HandlerFunc{
		EndpointHealth:   HandleHealth,
		EndpointExercise: HandleExercisePage,
		EndpointWorkoutComplete: HandleWorkoutComplete,
		EndpointLogout:   HandleLogout,
		EndpointAlreadyAuthenticated: HandleAlreadyAuthenticated,
	}
	for k, v := range endpoints {
		// mux.Handle(k, IpFilterMiddleware(v)) // todo cloudflare IPs are not working, I think some are not whitelisted that should be.
		mux.Handle(k, v)
	}

	authEndPoints := map[string]http.HandlerFunc{
		LandingPage:          HandleLandingPage,
		EndpontSignUp:        HandleSignUp,
		EndpointLogin:        HandleLogin,
		EndpointVerification: HandleVerification,
		EndpointEmail:        HandleForgotPasswordEmail,
		EndpointResetCode:    HandleForgotPasswordResetCode,
		EndpointNewPassword:  HandleForgotPasswordNewPassword,
	}

	for k, v := range authEndPoints {
		// mux.Handle(k, IpFilterMiddleware(v)) // todo cloudflare IPs are not working, I think some are not whitelisted that should be.
		mux.Handle(k, CheckForActiveSession(v))
	}

	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("error, when attempting to create subdir from static files. Error: %v", err)
	}

	fileSystem := http.FS(static)
	fileServer := http.FileServer(fileSystem)
	fileServerStripPrefix := http.StripPrefix("/static/", fileServer)
	cachingAdded := setCacheControl(fileServerStripPrefix, 86400*7) // Cache for one week
	mux.Handle("/static/", cachingAdded)

	HttpServer.Handler = mux

	log.Printf("initialization complete")
	err = HttpServer.ListenAndServeTLS("", "") // certs are already present in the tls config
	if err != nil {
		return fmt.Errorf("error, when attempting to start server: %v", err)
	}

	return nil
}
