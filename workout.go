package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	WorkoutPhaseWarmUp = iota
	WorkoutPhaseMain
	WorkoutPhaseCoolDown
	WorkoutPhaseCompleted
)

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

type RoutineType byte // Declare an alias for more descriptive code

const (
	LOWER RoutineType = iota
	CORE
	UPPER
	ALL
)

func (r RoutineType) GetNextRoutine() RoutineType {
	return (r + 1) % 3
}

func HandleGetCurrentWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "error, only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	result, err := GetCurrentWorkout(
		r.Context(),
		RedisConnectionPool,
		ConnectionPool,
		NumberOfSetsInSuperSet,
		NumberOfExerciseInSuperset,
		GetSuperSetExpiration(),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("error, failed to perform get current workout handler action: %v", err), http.StatusInternalServerError)
		return
	}
	responseData, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "error, failed to create JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseData)
	if err != nil {
		http.Error(w, "error, failed to write response", http.StatusInternalServerError)
		return
	}
}
