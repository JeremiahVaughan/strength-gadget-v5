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
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/handler"
	custom_middleware "strengthgadget.com/m/v2/middleware"
	"strengthgadget.com/m/v2/service"
	"time"
)

//go:embed database/*
var databaseFiles embed.FS

func main() {
	log.Printf("starting strengthgadget...")

	ctx := context.Background()
	err := config.InitConfig(ctx)
	if err != nil {
		log.Fatalf("error, attempting to initialize configuration: %v", err)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn: config.SentryEndpoint,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,

		Environment: config.Environment,
	})
	if err != nil {
		log.Fatalf("error, sentry.Init: %s", err)
	}
	sentry.CaptureMessage(fmt.Sprintf("Strengthgadget backend has started in the %s environment", config.Environment))
	defer sentry.Flush(2 * time.Second)

	err = service.ProcessSchemaChanges(ctx, databaseFiles)
	if err != nil {
		log.Fatalf("error, when attempting to process schema changes: %v", err)
	}

	r := chi.NewRouter()

	var corsMiddleware *cors.Cors
	log.Printf("CORS setup allowing traffic from: %s", config.TrustedUiOrigins)

	corsMiddleware = cors.New(cors.Options{
		MaxAge:           86400,
		AllowedOrigins:   config.TrustedUiOrigins,
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		// todo look into more secure by excluding uneeded headers
		AllowedHeaders: []string{"*"},
	})

	r.Use(
		corsMiddleware.Handler,
		middleware.Logger,
		custom_middleware.IpFilterMiddleware,
	)

	// The /api prefix is for the local proxy when developing locally. This proxy exists so authentication works locally.
	r.Route(constants.ApiPrefix, func(r chi.Router) {
		// public routes
		r.Get(constants.Health, handler.HandleHealth)
		r.Post(constants.Register, handler.HandleRegister)
		r.Post(constants.Login, handler.HandleLogin)
		r.Get(constants.IsLoggedIn, handler.HandleIsLoggedIn) // Currently needed for letting the browser know if it makes sense to be on the authentication pages or not.
		r.Post(constants.Verification, handler.HandleVerification)
		r.Post(constants.ResendVerification, handler.HandleResendVerification)

		r.Route(constants.ForgotPasswordPrefix, func(r chi.Router) {
			r.Post(constants.Email, handler.HandleForgotPasswordEmail)
			r.Post(constants.ResetCode, handler.HandleForgotPasswordResetCode)
			r.Post(constants.NewPassword, handler.HandleForgotPasswordNewPassword)
		})

		// authentication required routes
		r.Group(func(r chi.Router) {
			r.Use(service.Authenticate)
			r.Post(constants.Logout, handler.HandleLogout)
			r.Get(constants.ReadyForNextExercise, handler.HandleReadyForNextExercise)
			r.Get(constants.CurrentExercise, handler.HandleFetchCurrentExercise)
			r.Get(constants.ShuffleExercise, handler.HandleShuffleExercise)
		})
	})

	defer config.ConnectionPool.Close()
	log.Printf("initialization complete")
	config.HttpServer.Handler = r
	err = config.HttpServer.ListenAndServeTLS("", "") // certs are already present in the tls config
	if err != nil {
		log.Fatalf("error, when attempting to start server: %v", err)
	}
}
