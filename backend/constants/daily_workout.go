package constants

type DailyWorkoutSlotPhase int

const (
	DailyWorkoutSlotPhaseWarmup DailyWorkoutSlotPhase = iota
	DailyWorkoutSlotPhaseMainFocused
	DailyWorkoutSlotPhaseMainFiller
	DailyWorkoutSlotPhaseCoolDown
)
