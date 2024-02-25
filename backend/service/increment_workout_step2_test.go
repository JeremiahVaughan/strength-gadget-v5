package service

import (
	"strengthgadget.com/m/v2/model"
	"testing"
)

func TestHasWorkoutBeenCompleted(t *testing.T) {
	tests := []struct {
		name string
		arg  model.WorkoutProgressIndex
		want bool
	}{
		{
			name: "ValidIndexComplete",
			arg:  []int{1, 2, 3, 0},
			want: true,
		},
		{
			name: "ValidIndexNotComplete",
			arg:  []int{1, 2, 3},
			want: false,
		},
		{
			name: "InvalidEmptyIndex",
			arg:  []int{},
			want: false,
		},
		{
			name: "SingleItemIndex",
			arg:  []int{0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasWorkoutBeenCompleted(tt.arg); got != tt.want {
				t.Errorf("hasWorkoutBeenCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}
