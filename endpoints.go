package main

const (
	LandingPage = "/"

	EndpointHealth             = "/health"
	EndpontSignUp              = "/signUp"
	EndpointLogin              = "/login"
	EndpointIsLoggedIn         = "/isLoggedIn"
	EndpointVerification       = "/verification"
	EndpointResendVerification = "/resendVerification"
	EndpointExercise           = "/exercise"
	EndpointWorkoutComplete    = "/workoutComplete"

	EndpointForgotPasswordPrefix = "/forgotPassword"
	EndpointEmail                = EndpointForgotPasswordPrefix + "/email"
	EndpointResetCode            = EndpointForgotPasswordPrefix + "/resetCode"
	EndpointNewPassword          = EndpointForgotPasswordPrefix + "/newPassword"

	EndpointLogout = "/logout"

	EndpointGetCurrentWorkout            = "/getCurrentWorkout"
	EndpointSwapExercise                 = "/swapExercise"
	EndpointRecordIncrementedWorkoutStep = "/recordIncrementedWorkoutStep"
	EndpointRunIntegrationTests          = "runIntegrationTests"

	EndpointRobots  = "/robots.txt"
	EndpointTerms   = "/terms"
	EndpointPrivacy = "/privacy"
	EndpointSiteMap = "/sitemap.xml"
	EndpointAbout   = "/about"
	EndpointBlog    = "/blog"
)
