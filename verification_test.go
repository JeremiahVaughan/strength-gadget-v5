package main

import (
	"testing"
)

func Test_generateEmailVerificationCode(t *testing.T) {
	tests := []struct {
		name    string
		want    int
		wantErr bool
	}{
		{
			name:    "Happy path",
			want:    6,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateEmailVerificationCode()
			if (err != nil) != tt.wantErr {
				t.Errorf("generateEmailVerificationCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("generateEmailVerificationCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTwentyFourHoursEarlier(t *testing.T) {
	type args struct {
		theTime int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "Happy ",
			args: args{
				theTime: 1674924407,
			},
			want: 1674838007,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEarlierTime(tt.args.theTime, 86400); got != tt.want {

				t.Errorf("getEarlierTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasVerificationLimitBeenReached(t *testing.T) {
	type args struct {
		currentCount int
		allowedCount int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Yes more",
			args: args{
				currentCount: 8,
				allowedCount: 7,
			},
			want: true,
		},
		{
			name: "No less",
			args: args{
				currentCount: 6,
				allowedCount: 7,
			},
			want: false,
		},
		{
			name: "No same",
			args: args{
				currentCount: 7,
				allowedCount: 7,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasVerificationLimitBeenReached(tt.args.currentCount, tt.args.allowedCount); got != tt.want {
				t.Errorf("hasVerificationLimitBeenReached() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isCodeExpired(t *testing.T) {
	type args struct {
		code        *VerificationCode
		currentTime int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Is expired 1",
			args: args{
				code: &VerificationCode{
					Expires: 567,
				},
				currentTime: 589,
			},
			want: true,
		},
		{
			name: "Is not expired 1",
			args: args{
				code: &VerificationCode{
					Expires: 567,
				},
				currentTime: 567,
			},
			want: false,
		},
		{
			name: "Is not expired 2",
			args: args{
				code: &VerificationCode{
					Expires: 567,
				},
				currentTime: 45,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCodeExpired(tt.args.code, tt.args.currentTime); got != tt.want {
				t.Errorf("isCodeExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_doesProvidedCodeMatchExpectedCode(t *testing.T) {
	type args struct {
		provided string
		expected string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Does match 1",
			args: args{
				provided: "FSD5KC",
				expected: "FSD5KC",
			},
			want: true,
		},
		{
			name: "Does match 2",
			args: args{
				provided: "FSD5KC",
				expected: "fsd5kc",
			},
			want: true,
		},
		{
			name: "Does match 3",
			args: args{
				provided: "FSD5KC",
				expected: "fsd5KC",
			},
			want: true,
		},
		{
			name: "Does not match 1",
			args: args{
				provided: "FSD5KC",
				expected: "UJ8JZS",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doesProvidedCodeMatchExpectedCode(tt.args.provided, tt.args.expected); got != tt.want {
				t.Errorf("doesProvidedCodeMatchExpectedCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
