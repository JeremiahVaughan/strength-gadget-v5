package service

import "testing"

func Test_roundWeight(t *testing.T) {
	type args struct {
		weight          float64
		weightIncrement int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "1",
			args: args{
				weight:          37,
				weightIncrement: 5,
			},
			want: 35,
		},
		{
			name: "2",
			args: args{
				weight:          38,
				weightIncrement: 5,
			},
			want: 40,
		},
		{
			name: "3",
			args: args{
				weight:          40,
				weightIncrement: 5,
			},
			want: 40,
		},
		{
			name: "4",
			args: args{
				weight:          37.5,
				weightIncrement: 5,
			},
			want: 40,
		},
		{
			name: "5",
			args: args{
				weight:          37.6,
				weightIncrement: 5,
			},
			want: 40,
		},
		{
			name: "6",
			args: args{
				weight:          37.4,
				weightIncrement: 5,
			},
			want: 35,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roundWeight(tt.args.weight, tt.args.weightIncrement); got != tt.want {
				t.Errorf("roundWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
