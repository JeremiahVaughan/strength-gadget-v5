package model

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
