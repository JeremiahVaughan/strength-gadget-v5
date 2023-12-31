package model

const (
	EmailRateLimitPrefix = "email_rate_limit_"
	LoginAttemptPrefix   = "login_attempt_rate_limit_"

	CurrentSupersetPrefix             = "current_superset_for_"
	MuscleGroupsCompletedInSessionKey = "ex_completed"

	TotalMuscleGroupCountKey = "total_muscle_groups"

	CachedExercisePrefix  = "exercise_"
	CachedMuscleGroupsKey = "muscle_groups"

	// DailyWorkoutHashKeyPrefix is the key used for the daily workout.
	DailyWorkoutHashKeyPrefix = "daily_workout_"

	// DailyWorkoutKey is the key used for the daily workout.
	DailyWorkoutKey = "daily_workout"

	// UserUpdatedExerciseMeasurementUpdatesPrefix is the prefix used for keys related to the updates of exercise measurements for a user.
	UserUpdatedExerciseMeasurementUpdatesPrefix = "exercise_measurement_updates_"

	// UserExercisesSlottedPrefix is the prefix used for keys related to exercise slotted for a user.
	UserExercisesSlottedPrefix = "exercises_slotted_"
)
