package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"strengthgadget.com/m/v2/config"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/handler"
	custom_middleware "strengthgadget.com/m/v2/middleware"
	"strengthgadget.com/m/v2/model"
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

	defer config.ConnectionPool.Close()

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

	if os.Getenv(constants.ModeKey) == constants.WorkoutGen {
		err = generateDailyWorkout(ctx)
		if err != nil {
			log.Fatalf("error, when generateDailyWorkout() for main(). Error: %v", err)
		}
	} else {
		err = serveAthletes(ctx)
		if err != nil {
			log.Fatalf("error, when serveAthletes() for main(). Error: %v", err)
		}
	}
}

func generateDailyWorkout(ctx context.Context) error {
	allExercises, err := service.FetchAllExercises(ctx)
	if err != nil {
		return fmt.Errorf("error, when FetchAllExercises() for generateDailyWorkout(). Error: %v", err)
	}

	// third key is muscle group id, the value is the exercises that target the muscle group
	exerciseMap := make(map[model.RoutineType]map[model.ExerciseType]map[string][]model.Exercise)
	for _, exercise := range allExercises {
		// Initialize nested maps and slices if they do not exist yet
		if exerciseMap[exercise.RoutineType] == nil {
			exerciseMap[exercise.RoutineType] = make(map[model.ExerciseType]map[string][]model.Exercise)
		}
		if exerciseMap[exercise.RoutineType][exercise.ExerciseType] == nil {
			exerciseMap[exercise.RoutineType][exercise.ExerciseType] = make(map[string][]model.Exercise)
		}

		// Categorize the exercise
		exerciseMap[exercise.RoutineType][exercise.ExerciseType][exercise.MuscleGroupId] = append(exerciseMap[exercise.RoutineType][exercise.ExerciseType][exercise.MuscleGroupId], exercise)
	}

	lowerWorkout, err := generateDailyWorkoutValue(exerciseMap, model.LOWER)
	if err != nil {
		return fmt.Errorf("error, when generateDailyWorkoutValue(model.LOWER) for generateDailyWorkout(). Error: %v", err)
	}
	coreWorkout, err := generateDailyWorkoutValue(exerciseMap, model.CORE)
	if err != nil {
		return fmt.Errorf("error, when generateDailyWorkoutValue(model.CORE) for generateDailyWorkout(). Error: %v", err)
	}
	upperWorkout, err := generateDailyWorkoutValue(exerciseMap, model.UPPER)
	if err != nil {
		return fmt.Errorf("error, when generateDailyWorkoutValue(model.UPPER) for generateDailyWorkout(). Error: %v", err)
	}
	err = config.RedisConnectionPool.HSet(ctx, model.DailyWorkoutKey,
		getDailyWorkoutKey(model.LOWER), lowerWorkout,
		getDailyWorkoutKey(model.CORE), coreWorkout,
		getDailyWorkoutKey(model.UPPER), upperWorkout,
	).Err()
	if err != nil {
		return fmt.Errorf("error, when attempting to create daily_workout redis hash. Error: %v", err)
	}

	err = config.RedisConnectionPool.Expire(ctx, model.DailyWorkoutHashKeyPrefix, 36*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error, then attempting to set expiration for daily_workout redis hash. Error: %v", err)
	}
	return nil
}

func generateDailyWorkoutValue(exerciseMap map[model.RoutineType]map[model.ExerciseType]map[string][]model.Exercise, rt model.RoutineType) ([]byte, error) {
	dailyWorkout := model.DailyWorkout{}
	var cardioExercises []model.Exercise
	cardioExercises = []model.Exercise{}
	targetRoutine := exerciseMap[rt]
	for _, v := range exerciseMap[model.ALL][model.Cardio] {
		for _, e := range v {
			cardioExercises = append(cardioExercises, e)
		}
	}
	dailyWorkout.CardioExercises = cardioExercises
	dailyWorkout.ShuffleCardioExercises()

	weightLiftingExercises := targetRoutine[model.Weightlifting]
	calisthenicExercises := targetRoutine[model.Calisthenics]
	combinedExercises := make(map[string][]model.Exercise)
	// First, copy everything from weightLiftingExercises to combinedExercises
	for key, exercises := range weightLiftingExercises {
		combinedExercises[key] = append(combinedExercises[key], exercises...)
	}
	// Then, append exercises from calisthenicExercises to combinedExercises
	for key, exercises := range calisthenicExercises {
		combinedExercises[key] = append(combinedExercises[key], exercises...)
	}
	for _, exercises := range combinedExercises {
		dailyWorkout.MuscleCoverageMainExercises = append(dailyWorkout.MuscleCoverageMainExercises, exercises)
	}
	dailyWorkout.ShuffleMuscleCoverageMainExercises()

	// key is exercise id
	allExercisesMap := make(map[string]model.Exercise)
	for _, exercises := range combinedExercises {
		for _, e := range exercises {
			allExercisesMap[e.Id] = e
		}
	}
	for _, exercise := range allExercisesMap {
		dailyWorkout.AllMainExercises = append(dailyWorkout.AllMainExercises, exercise)
	}
	dailyWorkout.ShuffleMainExercises()

	coolDownExercises := targetRoutine[model.CoolDown]
	for _, exercises := range coolDownExercises {
		dailyWorkout.CoolDownExercises = append(dailyWorkout.CoolDownExercises, exercises)
	}
	dailyWorkout.ShuffleCoolDownExercises()

	bytes, err := json.Marshal(dailyWorkout)
	if err != nil {
		return nil, fmt.Errorf("error, when marshalling daily workout to json. Error: %v", err)
	}
	return bytes, nil
}

func getDailyWorkoutKey(rt model.RoutineType) string {
	return fmt.Sprintf("%s%d", model.DailyWorkoutHashKeyPrefix, rt)
}

func serveAthletes(ctx context.Context) error {
	err := service.ProcessSchemaChanges(ctx, databaseFiles)
	if err != nil {
		return fmt.Errorf("error, when attempting to process schema changes: %v", err)
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

	log.Printf("initialization complete")
	config.HttpServer.Handler = r
	err = config.HttpServer.ListenAndServeTLS("", "") // certs are already present in the tls config
	if err != nil {
		return fmt.Errorf("error, when attempting to start server: %v", err)
	}
	return nil
}
