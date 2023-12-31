package model

import "testing"

func TestGetNextRoutine(t *testing.T) {
	// Test transitioning from LOWER to UPPER
	if result := getNextRoutine(LOWER); result != CORE {
		t.Errorf("expected %d, but got %d", CORE, result)
	}

	// Test transitioning from CORE to LOWER
	if result := getNextRoutine(CORE); result != UPPER {
		t.Errorf("expected %d, but got %d", UPPER, result)
	}

	// Test transitioning from UPPER to CORE
	if result := getNextRoutine(UPPER); result != LOWER {
		t.Errorf("expected %d, but got %d", LOWER, result)
	}
}
