package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed database/*
var databaseFiles embed.FS

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

	if os.Getenv(ModeKey) == WorkoutGen {
		err = GenerateDailyWorkout(ctx, ConnectionPool, RedisConnectionPool, Environment)
		if err != nil {
			log.Fatalf("error, when GenerateDailyWorkout() for main(). Error: %v", err)
		}
	} else {
		err = serveAthletes(ctx)
		if err != nil {
			log.Fatalf("error, when serveAthletes() for main(). Error: %v", err)
		}
	}
}

func serveAthletes(ctx context.Context) error {
	err := ProcessSchemaChanges(ctx, databaseFiles)
	if err != nil {
		return fmt.Errorf("error, when attempting to process schema changes: %v", err)
	}

	r := chi.NewRouter()

	var corsMiddleware *cors.Cors
	log.Printf("CORS setup allowing traffic from: %s", TrustedUiOrigins)

	corsMiddleware = cors.New(cors.Options{
		MaxAge:           86400,
		AllowedOrigins:   TrustedUiOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		// todo look into more secure by excluding uneeded headers
		AllowedHeaders: []string{"*"},
	})

	r.Use(
		corsMiddleware.Handler,
		middleware.Logger,
		IpFilterMiddleware,
	)

	// The /api prefix is for the local proxy when developing locally. This proxy exists so authentication works locally.
	r.Route(ApiPrefix, func(r chi.Router) {
		// public routes
		r.Get(EndpointHealth, HandleHealth)
		r.Post(EndpointRegister, HandleRegister)
		r.Post(EndpointLogin, HandleLogin)
		r.Get(EndpointIsLoggedIn, HandleIsLoggedIn) // Currently needed for letting the browser know if it makes sense to be on the authentication pages or not.
		r.Post(EndpointVerification, HandleVerification)
		r.Post(EndpointResendVerification, HandleResendVerification)

		r.Route(EndpointForgotPasswordPrefix, func(r chi.Router) {
			r.Post(EndpointEmail, HandleForgotPasswordEmail)
			r.Post(EndpointResetCode, HandleForgotPasswordResetCode)
			r.Post(EndpointNewPassword, HandleForgotPasswordNewPassword)
		})

		// authentication required routes
		r.Group(func(r chi.Router) {
			r.Use(Authenticate)
			r.Post(EndpointLogout, HandleLogout)

			r.Get(EndpointGetCurrentWorkout, HandleGetCurrentWorkout)
			r.Put(EndpointSwapExercise, HandleSwapExercise)
			r.Put(EndpointRecordIncrementedWorkoutStep, HandleRecordIncrementedWorkoutStep)
		})
	})

	log.Printf("initialization complete")
	HttpServer.Handler = r
	err = HttpServer.ListenAndServeTLS("", "") // certs are already present in the tls config
	if err != nil {
		return fmt.Errorf("error, when attempting to start server: %v", err)
	}
	return nil
}
