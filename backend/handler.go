package main

const (
	ApiPrefix                  = "/api"
	EndpointHealth             = "/health"
	EndpointRegister           = "/register"
	EndpointLogin              = "/login"
	EndpointIsLoggedIn         = "/isLoggedIn"
	EndpointVerification       = "/verification"
	EndpointResendVerification = "/resendVerification"

	EndpointForgotPasswordPrefix = "/forgotPassword"
	EndpointEmail                = "/email"
	EndpointResetCode            = "/resetCode"
	EndpointNewPassword          = "/newPassword"

	EndpointLogout = "/logout"

	EndpointGetCurrentWorkout            = "/getCurrentWorkout"
	EndpointSwapExercise                 = "/swapExercise"
	EndpointRecordIncrementedWorkoutStep = "/recordIncrementedWorkoutStep"
	EndpointRunIntegrationTests          = "runIntegrationTests"
)
