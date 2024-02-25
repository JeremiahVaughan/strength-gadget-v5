package service

import (
	"strengthgadget.com/m/v2/model"
	"testing"
)

func TestGenerateQueryForExerciseUserData(t *testing.T) {
	tests := []struct {
		name             string
		exerciseUserData map[string]model.ExerciseUserData
		user             *model.User
		wantSQL          string
		wantArgs         []any
	}{
		{
			name: "SingleItemMap",
			exerciseUserData: map[string]model.ExerciseUserData{
				"exer1": {Measurement: 100},
			},
			user: &model.User{Id: "1"},
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
			exerciseUserData: map[string]model.ExerciseUserData{
				"exer1": {Measurement: 100},
				"exer2": {Measurement: 150},
				"exer3": {Measurement: 150},
			},
			user: &model.User{Id: "1"},
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
