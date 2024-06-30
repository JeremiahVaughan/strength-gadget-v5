package main

type WorkoutPhase int

const (
	WorkoutPhaseWarmUp WorkoutPhase = iota
	WorkoutPhaseMain
	WorkoutPhaseCoolDown
	WorkoutPhaseCompleted
)

