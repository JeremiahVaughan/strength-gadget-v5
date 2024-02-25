package model

import "strengthgadget.com/m/v2/constants"

type ExerciseUserData struct {
	Measurement           int                             `json:"measurement"`
	DailyWorkoutSlotIndex int                             `json:"dailyWorkoutSlotIndex"`
	DailyWorkoutSlotPhase constants.DailyWorkoutSlotPhase `json:"dailyWorkoutSlotPhase"`
}
