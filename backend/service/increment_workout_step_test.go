package service

import (
	"strengthgadget.com/m/v2/model"
	"testing"
)

func TestHaveMainExercisesJustBeenCompleted(t *testing.T) {
	testCases := []struct {
		name                     string
		incrementedProgressIndex model.WorkoutProgressIndex
		want                     bool
	}{
		{
			"test when incrementedProgressIndex is empty",
			model.WorkoutProgressIndex{},
			false,
		},
		{
			"test when incrementedProgressIndex has length bigger than WorkoutPhaseCoolDown",
			model.WorkoutProgressIndex{1, 2, 0, 4, 5},
			false,
		},
		{
			"test when incrementedProgressIndex has length equal to WorkoutPhaseCoolDown but last value is not 0",
			model.WorkoutProgressIndex{1, 2, 1},
			false,
		},
		{
			"test when incrementedProgressIndex has length equal to WorkoutPhaseCoolDown and last value is 0",
			model.WorkoutProgressIndex{1, 2, 0},
			true,
		},
		{
			"test when incrementedProgressIndex has length smaller than WorkoutPhaseCoolDown",
			model.WorkoutProgressIndex{1, 0},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := haveMainExercisesJustBeenCompleted(tc.incrementedProgressIndex); got != tc.want {
				t.Errorf("haveMainExercisesJustBeenCompleted(%v) = %v, want %v", tc.incrementedProgressIndex, got, tc.want)
			}
		})
	}
}
