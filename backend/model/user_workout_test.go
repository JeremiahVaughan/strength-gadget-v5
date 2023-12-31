package model

import (
	"testing"
)

func TestGetNextAvailableExercise(t *testing.T) {
	exercisePool := []Exercise{{Id: "ex1"}, {Id: "ex2"}, {Id: "ex3"}, {Id: "ex4"}}
	testCases := []struct {
		name                    string
		startingExercise        uint16
		alreadySlottedExercises map[string]bool
		want                    uint16
	}{
		{
			name:                    "no_exercise_taken",
			startingExercise:        3,
			alreadySlottedExercises: map[string]bool{},
			want:                    3, // As no exercise is taken it will start from first
		},
		{
			name:                    "first_exercise_taken",
			startingExercise:        0,
			alreadySlottedExercises: map[string]bool{"ex1": true},
			want:                    1, // As the first exercise is taken it will start from second
		},
		{
			name:                    "first_and_second_exercise_taken",
			startingExercise:        0,
			alreadySlottedExercises: map[string]bool{"ex1": true, "ex2": true},
			want:                    2, // As the first and second exercises are taken it will start from third
		},
		{
			name:                    "all_exercises_taken",
			startingExercise:        0,
			alreadySlottedExercises: map[string]bool{"ex1": true, "ex2": true, "ex3": true, "ex4": true},
			want:                    0, // As all exercise are taken it will start from first
		},
		{
			name:                    "available_exercise_is_before_starting_exercise",
			startingExercise:        1,
			alreadySlottedExercises: map[string]bool{"ex2": true, "ex3": true, "ex4": true},
			want:                    0, // As all exercise are taken it will start from first
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := getNextAvailableExercise(tc.startingExercise, exercisePool, tc.alreadySlottedExercises)
			if got != tc.want {
				t.Errorf("getNextAvailableExercise() = %v; want %v", got, tc.want)
			}
		})
	}
}
