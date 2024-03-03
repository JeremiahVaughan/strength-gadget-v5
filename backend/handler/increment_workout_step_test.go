package handler

import (
	"strengthgadget.com/m/v2/model"
	"testing"
)

func TestValidateRecordIncrementedWorkoutStepRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *model.RecordIncrementedWorkoutStepRequest
		wantErr bool
	}{
		{
			name: "Empty ExerciseId",
			req: &model.RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: model.WorkoutProgressIndex{2},
				ExerciseId:               "",
				WorkoutId:                "86cf4fea-8a25-45a2-82fe-d9065537f9fb",
			},
			wantErr: true,
		},
		{
			name: "Invalid ProgressIndex",
			req: &model.RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: model.WorkoutProgressIndex{},
				ExerciseId:               "exerciseId1",
				WorkoutId:                "86cf4fea-8a25-45a2-82fe-d9065537f9fb",
			},
			wantErr: true,
		},
		{
			name: "Valid Request",
			req: &model.RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: model.WorkoutProgressIndex{3},
				ExerciseId:               "exerciseId1",
				WorkoutId:                "86cf4fea-8a25-45a2-82fe-d9065537f9fb",
			},
			wantErr: false,
		},
		{
			name: "Missing workout UUID",
			req: &model.RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: model.WorkoutProgressIndex{3},
				ExerciseId:               "exerciseId1",
			},
			wantErr: true,
		},
		{
			name: "Invalid workout UUID",
			req: &model.RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: model.WorkoutProgressIndex{3},
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
