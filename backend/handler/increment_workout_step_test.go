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
			},
			wantErr: true,
		},
		{
			name: "Invalid ProgressIndex",
			req: &model.RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: model.WorkoutProgressIndex{},
				ExerciseId:               "exerciseId1",
			},
			wantErr: true,
		},
		{
			name: "Valid Request",
			req: &model.RecordIncrementedWorkoutStepRequest{
				IncrementedProgressIndex: model.WorkoutProgressIndex{3},
				ExerciseId:               "exerciseId1",
			},
			wantErr: false,
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
