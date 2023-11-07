package handler

import (
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"strengthgadget.com/m/v2/service"
	"testing"
)

//func Test_handler(t *testing.T) {
//	type args struct {
//		request model.Credentials
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    model.Credentials
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := HandleRegister(tt.args.request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("handleRegister() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("handleRegister() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

// todo setup integration test for this
//func Test_emailAlreadyExists(t *testing.T) {
//	type args struct {
//		email *string
//	}
//	existingEmail := "existingEmail@hotmail.com"
//	existingEmailResult := true
//
//	nonExistingEmail := "nonExistingEmail@hotmail.com"
//	nonExistingEmailResult := false
//	tests := []struct {
//		name  string
//		args  args
//		want  *bool
//		want1 *auth.Error
//	}{
//		{
//			name: "Email exists path",
//			args: args{
//				email: &existingEmail,
//			},
//			want:  &existingEmailResult,
//			want1: nil,
//		},
//		{
//			name: "Email does not exist path",
//			args: args{
//				email: &nonExistingEmail,
//			},
//			want:  &nonExistingEmailResult,
//			want1: nil,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, got1 := emailAlreadyExists(tt.args.email)
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("emailAlreadyExists() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("emailAlreadyExists() got1 = %v, want %v", got1, tt.want1)
//			}
//		})
//	}
//}

func Test_emailIsValidFormat(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Email is valid",
			args: args{
				email: "rubberduck@gmail.com",
			},
			want: true,
		},
		{
			name: "Email is not valid 1",
			args: args{
				email: "rubberduckgmail.com",
			},
			want: false,
		},
		{
			name: "Email is somehow valid",
			args: args{
				email: "rubberduck@icantbelievethisisvalidwithoutadot",
			},
			want: true,
		},
		{
			name: "Email is not valid 3",
			args: args{
				email: "@gmail.com",
			},
			want: false,
		},
		{
			name: "",
			args: args{
				email: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.EmailIsValidFormat(tt.args.email); got != tt.want {
				t.Errorf("emailIsValidFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_passwordIsAcceptable(t *testing.T) {
	type args struct {
		password string
	}
	validPassword := "14897b4938463773"
	tooManyNumbersPassword := "14897456344564938463773"
	tooFewCharsPassword := "oeu963773"
	tests := []struct {
		name string
		args args
		want *model.UserFeedbackError
	}{
		{
			name: "Happy path",
			args: args{
				password: validPassword,
			},
			want: nil,
		},
		{
			name: "Cannot be all numbers",
			args: args{
				password: tooManyNumbersPassword,
			},
			want: &constants.ErrorPasswordCannotContainAllNumbers,
		},
		{
			name: "Must be at least 12 characters long 1",
			args: args{
				password: tooFewCharsPassword,
			},
			want: &constants.ErrorPasswordMustBeAtLeastTwelveCharsLong,
		},
		{
			name: "Must provide a password",
			args: args{
				password: "",
			},
			want: &constants.ErrorPasswordWasNotProvided,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.PasswordIsAcceptable(tt.args.password)
			if got != nil && got.UserFeedbackError != *tt.want {
				t.Errorf("passwordIsAcceptable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// todo setup an integration test for this
//func Test_sendEmailVerification(t *testing.T) {
//	type args struct {
//		email string
//	}
//	tests := []struct {
//		name string
//		args args
//		want interface{}
//	}{
//		{
//			name: "Happy Path",
//			args: args{
//				email: "piegarden@gmail.com",
//			},
//			want: nil,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := sendEmailVerification(tt.args.email, ); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("sendEmailVerification() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func Test_emailIsValid(t *testing.T) {
	type args struct {
		email string
	}
	missingEmail := ""
	tests := []struct {
		name string
		args args
		want model.UserFeedbackError
	}{
		{
			name: "Unhappy path, password not provided",
			args: args{
				email: missingEmail,
			},
			want: constants.ErrorMissingEmailAddress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emailIsValid(tt.args.email)
			if got != nil && tt.want != got.UserFeedbackError {
				t.Errorf("emailIsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
