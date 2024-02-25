package model

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
