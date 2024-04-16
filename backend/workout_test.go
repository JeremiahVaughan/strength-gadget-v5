package main

import (
	"testing"
)

func TestWorkoutProgressIndex_IsValid(t *testing.T) {
	type TestCase struct {
		name string
		w    WorkoutProgressIndex
		want bool
	}
	tests := []TestCase{
		{
			name: "empty slice",
			w:    WorkoutProgressIndex{},
			want: false,
		},
		{
			name: "one member slice",
			w:    WorkoutProgressIndex{1},
			want: true,
		},
		{
			name: "two members slice",
			w:    WorkoutProgressIndex{1, 2},
			want: true,
		},
		{
			name: "three members slice",
			w:    WorkoutProgressIndex{1, 2, 3},
			want: true,
		},
		{
			name: "four members slice (workout completed)",
			w:    WorkoutProgressIndex{1, 2, 3, 5},
			want: true,
		},
		{
			name: "five members slice",
			w:    WorkoutProgressIndex{1, 2, 3, 5, 6},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.w.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextRoutine(t *testing.T) {
	// Test transitioning from LOWER to UPPER
	if result := LOWER.GetNextRoutine(); result != CORE {
		t.Errorf("expected %d, but got %d", CORE, result)
	}

	// Test transitioning from CORE to LOWER
	if result := CORE.GetNextRoutine(); result != UPPER {
		t.Errorf("expected %d, but got %d", UPPER, result)
	}

	// Test transitioning from UPPER to CORE
	if result := UPPER.GetNextRoutine(); result != LOWER {
		t.Errorf("expected %d, but got %d", LOWER, result)
	}
}
