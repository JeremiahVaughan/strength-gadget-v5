package model

// WorkoutProgressIndex Outer slice is the workout phase (e.g., warmup, main, cool-down).
// Inner slice is the index within the current workout phase (e.g., main exercise step 4)
type WorkoutProgressIndex []int

func (w WorkoutProgressIndex) IsValid() bool {
	return len(w) >= 1 && len(w) <= 4
}

type RecordIncrementedWorkoutStepRequest struct {
	IncrementedProgressIndex WorkoutProgressIndex `json:"incrementedProgressIndex"`
	ExerciseId               string               `json:"exerciseId"`
	LastCompletedMeasurement int                  `json:"lastCompletedMeasurement"`

	// WorkoutId is used to help prevent client and server sync issues
	WorkoutId string `json:"workoutId"`
}
