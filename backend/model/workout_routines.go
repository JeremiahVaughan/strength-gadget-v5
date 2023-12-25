package model

type RoutineType byte // Declare an alias for more descriptive code

const (
	LOWER RoutineType = iota
	CORE
	UPPER
	ALL
)

func getNextRoutine(current RoutineType) RoutineType {
	return (current + 1) % 3
}
