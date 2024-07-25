package main

import (
	"testing"
)

func TestGenerateQueryForExerciseUserData(t *testing.T) {
	tests := []struct {
		name                 string
		exerciseMeasurements ChoosenExercisesMap
		user                 *User
		wantSQL              string
		wantArgs             []any
	}{
		{
			name: "SingleItemMap",
			exerciseMeasurements: ChoosenExercisesMap{
				2: 100,
			},
			user: &User{Id: 1},
			wantSQL: `
		INSERT INTO last_completed_measurement (user_id, exercise_id, measurement)
		VALUES 
			($1, $2, $3)
		ON CONFLICT (user_id, exercise_id) DO UPDATE 
		SET 
			measurement = excluded.measurement`,
			wantArgs: []any{1, 2, 100},
		},
		{
			name: "MultipleItemsMap",
			exerciseMeasurements: ChoosenExercisesMap{
				1: 100,
				2: 150,
				3: 150,
			},
			user: &User{Id: 1},
			wantSQL: `
		INSERT INTO last_completed_measurement (user_id, exercise_id, measurement)
		VALUES 
			($1, $2, $3),
($4, $5, $6),
($7, $8, $9)
		ON CONFLICT (user_id, exercise_id) DO UPDATE 
		SET 
			measurement = excluded.measurement`,
			wantArgs: []any{1, 1, 100, 1, 2, 150, 1, 3, 150},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs := generateQueryForExerciseMeasurements(tt.exerciseMeasurements, tt.user.Id)
			if gotSQL != tt.wantSQL {
				t.Errorf("generateQueryForExerciseUserData() got SQL = %v, want %v", gotSQL, tt.wantSQL)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("generateQueryForExerciseUserData() got Args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
