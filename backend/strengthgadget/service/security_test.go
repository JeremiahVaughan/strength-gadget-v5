package service

import (
	"reflect"
	"strengthgadget.com/m/v2/model"
	"testing"
)

func Test_generateSalt(t *testing.T) {
	tests := []struct {
		name       string
		wantLength int
		want1      *model.Error
	}{
		{
			name:       "Happy path",
			wantLength: 172,
			want1:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := generateSalt()
			gotLength := len(*got)
			if gotLength != tt.wantLength {
				t.Errorf("generateSalt() gotLength = %v, want %v", gotLength, tt.wantLength)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("generateSalt() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
