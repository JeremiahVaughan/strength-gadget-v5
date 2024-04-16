package main

import (
	"testing"
)

func TestHasWorkoutBeenCompleted(t *testing.T) {
	tests := []struct {
		name string
		arg  WorkoutProgressIndex
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

func TestGenerateQueryForExerciseUserData(t *testing.T) {
	tests := []struct {
		name             string
		exerciseUserData map[string]ExerciseUserData
		user             *User
		wantSQL          string
		wantArgs         []any
	}{
		{
			name: "SingleItemMap",
			exerciseUserData: map[string]ExerciseUserData{
				"exer1": {Measurement: 100},
			},
			user: &User{Id: "1"},
			wantSQL: `
		INSERT INTO last_completed_measurement (user_id, exercise_id, measurement)
		VALUES 
			($1, $2, $3)
		ON CONFLICT (user_id, exercise_id) DO UPDATE 
		SET 
			measurement = excluded.measurement`,
			wantArgs: []any{"1", "exer1", 100},
		},
		{
			name: "MultipleItemsMap",
			exerciseUserData: map[string]ExerciseUserData{
				"exer1": {Measurement: 100},
				"exer2": {Measurement: 150},
				"exer3": {Measurement: 150},
			},
			user: &User{Id: "1"},
			wantSQL: `
		INSERT INTO last_completed_measurement (user_id, exercise_id, measurement)
		VALUES 
			($1, $2, $3),
($4, $5, $6),
($7, $8, $9)
		ON CONFLICT (user_id, exercise_id) DO UPDATE 
		SET 
			measurement = excluded.measurement`,
			wantArgs: []any{"1", "exer1", 100, "1", "exer2", 150, "1", "exer3", 150},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs := generateQueryForExerciseUserData(tt.exerciseUserData, tt.user)
			if gotSQL != tt.wantSQL {
				t.Errorf("generateQueryForExerciseUserData() got SQL = %v, want %v", gotSQL, tt.wantSQL)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("generateQueryForExerciseUserData() got Args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestHaveMainExercisesJustBeenCompleted(t *testing.T) {
	testCases := []struct {
		name                     string
		incrementedProgressIndex WorkoutProgressIndex
		want                     bool
	}{
		{
			"test when incrementedProgressIndex is empty",
			WorkoutProgressIndex{},
			false,
		},
		{
			"test when incrementedProgressIndex has length bigger than WorkoutPhaseCoolDown",
			WorkoutProgressIndex{1, 2, 0, 4, 5},
			false,
		},
		{
			"test when incrementedProgressIndex has length equal to WorkoutPhaseCoolDown but last value is not 0",
			WorkoutProgressIndex{1, 2, 1},
			false,
		},
		{
			"test when incrementedProgressIndex has length equal to WorkoutPhaseCoolDown and last value is 0",
			WorkoutProgressIndex{1, 2, 0},
			true,
		},
		{
			"test when incrementedProgressIndex has length smaller than WorkoutPhaseCoolDown",
			WorkoutProgressIndex{1, 0},
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

func TestValidateRecordIncrementedWorkoutStepRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *RecordIncrementedWorkoutStepRequest
		wantErr bool
	}{
		{
			name: "Empty ExerciseId",
			req: &RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: WorkoutProgressIndex{2},
				ExerciseId:               "",
				WorkoutId:                "86cf4fea-8a25-45a2-82fe-d9065537f9fb",
			},
			wantErr: true,
		},
		{
			name: "Invalid ProgressIndex",
			req: &RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: WorkoutProgressIndex{},
				ExerciseId:               "exerciseId1",
				WorkoutId:                "86cf4fea-8a25-45a2-82fe-d9065537f9fb",
			},
			wantErr: true,
		},
		{
			name: "Valid Request",
			req: &RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: WorkoutProgressIndex{3},
				ExerciseId:               "exerciseId1",
				WorkoutId:                "86cf4fea-8a25-45a2-82fe-d9065537f9fb",
			},
			wantErr: false,
		},
		{
			name: "Missing workout UUID",
			req: &RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: WorkoutProgressIndex{3},
				ExerciseId:               "exerciseId1",
			},
			wantErr: true,
		},
		{
			name: "Invalid workout UUID",
			req: &RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: WorkoutProgressIndex{3},
				ExerciseId:               "exerciseId1",
				WorkoutId:                "86cf4fea-8a25-45a2-82fe-d9065537f9f",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRecordIncrementedWorkoutStepRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRecordIncrementedWorkoutStepRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
