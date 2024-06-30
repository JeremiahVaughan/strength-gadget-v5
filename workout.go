package main

import (
	"fmt"
	"strconv"
	"strings"
)

// WorkoutProgressIndex Outer slice is the workout phase (e.g., warmup, main, cool-down).
// Inner slice is the index within the current workout phase (e.g., main exercise step 4)
type WorkoutProgressIndex []int

func (w WorkoutProgressIndex) IsValid() bool {
	return len(w) >= 1 && len(w) <= 4
}

func (w WorkoutProgressIndex) marshal() string {
	var theStrings []string
	for _, i := range w {
		theStrings = append(theStrings, strconv.Itoa(i))
	}
	return strings.Join(theStrings, ",")
}

func (w WorkoutProgressIndex) demarshal(index string) (WorkoutProgressIndex, error) {
	theStrings := strings.Split(index, ",")
	var result WorkoutProgressIndex
	for _, s := range theStrings {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("error, invalid string. Error: %v", err)
		}
		result = append(result, i)
	}
	return result, nil
}

type RecordIncrementedWorkoutStepRequest struct {
	ProgressIndex            WorkoutProgressIndex
	ExerciseId               string
	LastCompletedMeasurement int

	// WorkoutId is used to help prevent client and server sync issues
	WorkoutId string
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
