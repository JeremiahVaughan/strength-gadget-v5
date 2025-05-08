package main

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestGetTomorrowsWeekday(t *testing.T) {
	tests := []struct {
		name     string
		today    time.Weekday
		expected time.Weekday
	}{
		{
			name:     "Monday",
			today:    time.Monday,
			expected: time.Tuesday,
		},
		{
			name:     "Tuesday",
			today:    time.Tuesday,
			expected: time.Wednesday,
		},
		{
			name:     "Wednesday",
			today:    time.Wednesday,
			expected: time.Thursday,
		},
		{
			name:     "Thursday",
			today:    time.Thursday,
			expected: time.Friday,
		},
		{
			name:     "Friday",
			today:    time.Friday,
			expected: time.Saturday,
		},
		{
			name:     "Saturday",
			today:    time.Saturday,
			expected: time.Sunday,
		},
		{
			name:     "Sunday",
			today:    time.Sunday,
			expected: time.Monday,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTomorrowsWeekday(tt.today); got != tt.expected {
				t.Errorf("getTomorrowsWeekday() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func Test_ShuffleMuscleCoverageMainExercises(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		e := DailyWorkoutRandomIndices{
			MainMuscleGroups: []int{5, 0, 1, 2, 4, 3},
			MainExercises: [][]int{
				{
					0,
					1,
				},
				{
					8,
					9,
					3,
					7,
					6,
					10,
					5,
					4,
					2,
					11,
					1,
					0,
				},
				{
					9,
					5,
					0,
					7,
					4,
					10,
					2,
					8,
					1,
					6,
					3,
				},
				{
					1,
					0,
				},
				{
					1,
					4,
					0,
					2,
                    3,
					5,
				},
				{
					0,
				},
			},
		}
		d := DailyWorkoutRandomIndices{}
		u := WorkoutSession{
			CurrentWorkoutSeed: 1,
		}
		r := rand.New(rand.NewSource(u.CurrentWorkoutSeed))
		d.ShuffleMuscleGroups(r, lowerWorkout)
		d.ShuffleMuscleCoverageMainExercises(r, lowerWorkout)
		if !reflect.DeepEqual(e.MainMuscleGroups, d.MainMuscleGroups) {
			t.Errorf(`muscle groups are not in the correct order. 
				Expected: %+v, 
				Got: %+v`,
				e.MainMuscleGroups,
				d.MainMuscleGroups,
			)
		}
		if !reflect.DeepEqual(e.MainExercises, d.MainExercises) {
			t.Errorf(`exercise are not in the correct order. 
				Expected: %+v, 
				Got: %+v`,
				e.MainExercises,
				d.MainExercises,
			)
		}
		for i, exercises := range lowerWorkout.MainExercises {
			if len(exercises) != len(d.MainExercises[i]) {
				t.Errorf(`expecting outerslice to remain in the same order but it did not.
					Expected: %d,
					Got: %d`,
					len(exercises),
					len(d.MainExercises[i]),
				)
			}
		}
	})
}
