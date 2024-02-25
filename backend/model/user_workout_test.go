package model

import (
	"testing"
)

func TestGetNextAvailableExercise(t *testing.T) {
	exercisePool := []Exercise{{Id: "ex1"}, {Id: "ex2"}, {Id: "ex3"}, {Id: "ex4"}}
	testCases := []struct {
		name                    string
		startingExercise        uint16
		alreadySlottedExercises ExerciseUserDataMap
		want                    uint16
	}{
		{
			name:                    "no_exercise_taken",
			startingExercise:        3,
			alreadySlottedExercises: ExerciseUserDataMap{},
			want:                    3, // As no exercise is taken it will start from first
		},
		{
			name:                    "first_exercise_taken",
			startingExercise:        0,
			alreadySlottedExercises: ExerciseUserDataMap{"ex1": ExerciseUserData{}},
			want:                    1, // As the first exercise is taken it will start from second
		},
		{
			name:                    "first_and_second_exercise_taken",
			startingExercise:        0,
			alreadySlottedExercises: ExerciseUserDataMap{"ex1": ExerciseUserData{}, "ex2": ExerciseUserData{}},
			want:                    2, // As the first and second exercises are taken it will start from third
		},
		{
			name:                    "all_exercises_taken",
			startingExercise:        0,
			alreadySlottedExercises: ExerciseUserDataMap{"ex1": ExerciseUserData{}, "ex2": ExerciseUserData{}, "ex3": ExerciseUserData{}, "ex4": ExerciseUserData{}},
			want:                    0, // As all exercise are taken it will start from first
		},
		{
			name:                    "available_exercise_is_before_starting_exercise",
			startingExercise:        1,
			alreadySlottedExercises: ExerciseUserDataMap{"ex2": ExerciseUserData{}, "ex3": ExerciseUserData{}, "ex4": ExerciseUserData{}},
			want:                    0, // As all exercise are taken it will start from first
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := getNextAvailableExercise(tc.startingExercise, exercisePool, tc.alreadySlottedExercises, 5, 5)
			if got != tc.want {
				t.Errorf("getNextAvailableExercise() = %v; want %v", got, tc.want)
			}
		})
	}
}
