package model

type WarmupExercise struct {
	Exercise Exercise `json:"exercise"`
}

type MainExercise struct {
	SourceExerciseIndex uint16    `json:"sourceExercisePointer"`
	Exercise            *Exercise `json:"exercise"`
}

type CoolDownExercise struct {
	Exercise Exercise `json:"exercise"`
}

type Exercise struct {
	Id                       string       `json:"id"`
	Name                     string       `json:"name"`
	DemonstrationGiphyId     string       `json:"demonstrationGiphyId"`
	LastCompletedMeasurement int          `json:"lastCompletedMeasurement"`
	MeasurementType          string       `json:"measurementType"`
	ExerciseType             ExerciseType `json:"-"`
	MuscleGroupId            string       `json:"-"`
	RoutineType              RoutineType  `json:"-"`
}

type ExerciseResponse struct {
	Exercise *Exercise `json:"exercise"`
	SuperSetProgress
}
