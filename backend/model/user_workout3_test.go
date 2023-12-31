package model

import (
	"testing"
)

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
