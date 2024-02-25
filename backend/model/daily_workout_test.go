package model

import (
	"testing"
	"time"
)

func TestGetTomorrowsWeekday(t *testing.T) {
	tests := []struct {
		name     string
		today    time.Weekday
		expected time.Weekday
	}{
		{
			name:     "Monday",
			today:    time.Monday,
			expected: time.Tuesday,
		},
		{
			name:     "Tuesday",
			today:    time.Tuesday,
			expected: time.Wednesday,
		},
		{
			name:     "Wednesday",
			today:    time.Wednesday,
			expected: time.Thursday,
		},
		{
			name:     "Thursday",
			today:    time.Thursday,
			expected: time.Friday,
		},
		{
			name:     "Friday",
			today:    time.Friday,
			expected: time.Saturday,
		},
		{
			name:     "Saturday",
			today:    time.Saturday,
			expected: time.Sunday,
		},
		{
			name:     "Sunday",
			today:    time.Sunday,
			expected: time.Monday,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTomorrowsWeekday(tt.today); got != tt.expected {
				t.Errorf("getTomorrowsWeekday() = %v, want %v", got, tt.expected)
			}
		})
	}
}
