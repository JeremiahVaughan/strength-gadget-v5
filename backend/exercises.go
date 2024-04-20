package main

import (
	"fmt"
	"math/rand"
	"strings"
)

const (
	Weightlifting = "6bdb3624-bed1-41a9-bf8c-7b1066411446"
	Calisthenics  = "8ffe7196-4e3d-4439-ae19-3159ad5387bd"
	Cardio        = "982d0b18-a67c-401a-95f2-ddb702ba80b5"
	WarmUp        = "ce6133be-2bd8-48e9-adbb-05f03ad7b4f9"
	CoolDown      = "db085937-cd84-406a-b9db-34f9e091816b"
)

type SuperSet struct {
	Exercises              []Exercise `json:"exercise"`
	CurrentExercisePointer int        `json:"currentExercisePointer"`
	SetCompletionCount     int        `json:"completionCount"`
	SuperSetProgress
}

type SuperSetProgress struct {
	WorkoutComplete bool `json:"workoutComplete"`
}

type ExerciseUserData struct {
	Measurement           int                   `json:"measurement"`
	DailyWorkoutSlotIndex int                   `json:"dailyWorkoutSlotIndex"`
	DailyWorkoutSlotPhase DailyWorkoutSlotPhase `json:"dailyWorkoutSlotPhase"`
}

type ExerciseType string

type Exercise struct {
	Id                       string       `json:"id,omitempty"`
	Name                     string       `json:"name,omitempty"`
	DemonstrationGiphyId     string       `json:"demonstrationGiphyId,omitempty"`
	LastCompletedMeasurement int          `json:"lastCompletedMeasurement,omitempty"`
	MeasurementType          string       `json:"measurementType,omitempty"`
	ExerciseType             ExerciseType `json:"exerciseType,omitempty"`
	MuscleGroupId            string       `json:"-"`
	RoutineType              RoutineType  `json:"-"`

	// SourceExerciseSlotIndex will be used to reference the selected exercise's CurrentExerciseSlotIndex when not in selection mode
	CurrentExerciseSlotIndex int `json:"currentExerciseSlotIndex"`
	SourceExerciseSlotIndex  int `json:"sourceExerciseSlotIndex"`
}

func hasMuscleGroupWorkedSessionLimitBeenReached(totalMuscleGroupsCount int, count int) bool {
	// Adding one before division if totalMuscleGroupsCount be odd to handle ceiling
	halfMuscleGroups := totalMuscleGroupsCount / 2
	if totalMuscleGroupsCount%2 != 0 {
		halfMuscleGroups++
	}

	return halfMuscleGroups <= count
}

func markPreviousExerciseAsCompleted(currentSuperset *SuperSet, numberOfAvailableMuscleGroups int, numberOfExerciseInSuperset int) *SuperSet {
	numberOfActiveExercises := len(currentSuperset.Exercises)
	currentExerciseNumber := currentSuperset.CurrentExercisePointer + 1
	if currentExerciseNumber == numberOfExerciseInSuperset || (numberOfAvailableMuscleGroups == 0 && numberOfActiveExercises == currentExerciseNumber) {
		currentSuperset.CurrentExercisePointer = 0
		currentSuperset.SetCompletionCount++
	} else {
		currentSuperset.CurrentExercisePointer++
	}
	return currentSuperset
}

func getExerciseArgsAndInsertValues(exerciseIds []string) (string, []any) {
	var exercisesArgsSlice []string
	var insertValues []any
	for i, exerciseId := range exerciseIds {
		exercisesArgsSlice = append(exercisesArgsSlice, fmt.Sprintf("$%d", i+1))
		insertValues = append(insertValues, exerciseId)
	}
	return strings.Join(exercisesArgsSlice, ", "), insertValues
}

func selectRandomMuscleGroup(availableMuscleGroups []MuscleGroup) *MuscleGroup {
	muscleGroupCount := len(availableMuscleGroups)
	if muscleGroupCount == 0 {
		return nil
	}
	result := availableMuscleGroups[rand.Intn(muscleGroupCount)]
	return &result
}
