package model

type Exercise struct {
	Id                       string `json:"id"`
	Name                     string `json:"name"`
	DemonstrationGiphyId     string `json:"demonstrationGiphyId"`
	LastCompletedMeasurement int    `json:"lastCompletedMeasurement"`
	MeasurementType          string `json:"measurementType"`
}

type ExerciseResponse struct {
	Exercise *Exercise `json:"exercise"`
	SuperSetProgress
}
