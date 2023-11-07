package model

type SuperSet struct {
	Exercises              []Exercise `json:"exercise"`
	CurrentExercisePointer int        `json:"currentExercisePointer"`
	SetCompletionCount     int        `json:"completionCount"`
	SuperSetProgress
}

type SuperSetProgress struct {
	WorkoutComplete bool `json:"workoutComplete"`
}
