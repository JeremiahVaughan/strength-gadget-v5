package model

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

type ExerciseResponse struct {
	Exercise *Exercise `json:"exercise"`
	SuperSetProgress
}
