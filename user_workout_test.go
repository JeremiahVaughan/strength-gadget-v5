package main

import (
	"reflect"
	"testing"
)

func TestSerializeUniqueMember(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		index    uint16
		expected string
	}{
		{
			name:     "base case",
			score:    100,
			index:    1,
			expected: "100:1",
		},
		{
			name:     "negative score",
			score:    -100,
			index:    1,
			expected: "-100:1",
		},
		{
			name:     "zero score",
			score:    0,
			index:    3,
			expected: "0:3",
		},
		{
			name:     "large index",
			score:    200,
			index:    65535,
			expected: "200:65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeUniqueMember(tt.score, tt.index)
			if result != tt.expected {
				t.Fatalf("Expected %s but got %s", tt.expected, result)
			}
		})
	}
}

func TestDeserializeUniqueMember(t *testing.T) {
	tt := []struct {
		name             string
		member           string
		expectedErr      bool
		expectedExercise uint16
	}{
		{"valid distinction", "100:120", false, 120},
		{"invalid format", "abcd", true, 0},
		{"no distinction", "100120", true, 0},
		{"empty", "", true, 0},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			exercise, err := deserializeUniqueMember(tc.member)
			if tc.expectedErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if exercise != tc.expectedExercise {
					t.Errorf("Expected exercise index %d, but got %d", tc.expectedExercise, exercise)
				}
			}
		})
	}
}

func TestGetNextAvailableExercise(t *testing.T) {
	exercisePool := []Exercise{{Id: 1}, {Id: 2}, {Id: 3}, {Id: 4}}
	testCases := []struct {
		name                    string
		startingOffset          int
		randomPool              []int
		alreadySlottedExercises ExerciseUserDataMap
		want                    Exercise
	}{
		{
			name:                    "no_exercise_taken",
			startingOffset:          0,
			randomPool:              []int{0, 1, 2, 3},
			alreadySlottedExercises: ExerciseUserDataMap{},
			want:                    Exercise{Id: 1}, // As no exercise is taken it will start from first
		},
		{
			name:                    "first_exercise_taken",
			startingOffset:          0,
			randomPool:              []int{0, 1, 2, 3},
			alreadySlottedExercises: ExerciseUserDataMap{1: ExerciseUserData{}},
			want:                    Exercise{Id: 2}, // As the first exercise is taken it will start from second
		},
		{
			name:                    "first_and_second_exercise_taken",
			startingOffset:          0,
			randomPool:              []int{0, 1, 2, 3},
			alreadySlottedExercises: ExerciseUserDataMap{1: ExerciseUserData{}, 2: ExerciseUserData{}},
			want:                    Exercise{Id: 3}, // As the first and second exercises are taken it will start from third
		},
		{
			name:                    "all_exercises_taken",
			startingOffset:          0,
			randomPool:              []int{0, 1, 2, 3},
			alreadySlottedExercises: ExerciseUserDataMap{1: ExerciseUserData{}, 2: ExerciseUserData{}, 3: ExerciseUserData{}, 4: ExerciseUserData{}},
			want:                    Exercise{Id: 1}, // As all exercise are taken so we will keep what we have
		},
		{
			name:                    "available_exercise_is_before_starting_exercise",
			startingOffset:          1,
			randomPool:              []int{0, 1, 2, 3},
			alreadySlottedExercises: ExerciseUserDataMap{2: ExerciseUserData{}, 3: ExerciseUserData{}, 4: ExerciseUserData{}},
			want:                    Exercise{Id: 1}, // As all exercise are taken it will start from first
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, _, _ := getNextAvailableExercise(
				tc.startingOffset,
				tc.randomPool,
				exercisePool,
				tc.alreadySlottedExercises,
			)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getNextAvailableExercise() got: %v; want: %v", got, tc.want)
			}
		})
	}
}
