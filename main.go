package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

//go:embed static/robots.txt
var robotsTxt []byte

//go:embed static/privacy
var privacy []byte

//go:embed static/terms
var terms []byte

//go:embed static/sitemap.xml
var sitemap []byte

//go:embed static/about
var about []byte

//go:embed static/blog
var blog []byte

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
		EndpointHealth:          HandleHealth,
		EndpointExercise:        HandleExercisePage,
		EndpointWorkoutComplete: HandleWorkoutComplete,
		EndpointLogout:          HandleLogout,
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

	staticEndpoints := map[string]http.HandlerFunc{
		EndpointRobots:  ServeRobotsFile,
		EndpointTerms:   ServeTermsFile,
		EndpointPrivacy: ServePrivacyFile,
		EndpointSiteMap: ServieSiteMapFile,
		EndpointAbout:   ServieAboutFile,
		EndpointBlog:    ServieBlogFile,
	}

	for k, v := range staticEndpoints {
		mux.Handle(k, setCacheControl(v, 86400*7)) // Cache for one week
	}

	HttpServer.Handler = mux

	log.Printf("initialization complete")
	err = HttpServer.ListenAndServeTLS("", "") // certs are already present in the tls config
	if err != nil {
		return fmt.Errorf("error, when attempting to start server: %v", err)
	}

	return nil
}

func ServeRobotsFile(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(robotsTxt)
	HandleUnexpectedError(w, fmt.Errorf("error, when serving robots.txt file. Error: %v", err))
}

func ServeTermsFile(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(terms)
	HandleUnexpectedError(w, fmt.Errorf("error, when serving terms file. Error: %v", err))
}

func ServePrivacyFile(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(privacy)
	HandleUnexpectedError(w, fmt.Errorf("error, when serving privacy file. Error: %v", err))
}

func ServieSiteMapFile(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(sitemap)
	HandleUnexpectedError(w, fmt.Errorf("error, when serving sitemap.xml file. Error: %v", err))
}

func ServieAboutFile(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(about)
	HandleUnexpectedError(w, fmt.Errorf("error, when serving about file. Error: %v", err))
}

func ServieBlogFile(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(blog)
	HandleUnexpectedError(w, fmt.Errorf("error, when serving blog file. Error: %v", err))
}
