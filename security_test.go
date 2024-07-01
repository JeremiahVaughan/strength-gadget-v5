package main

import (
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func Test_generateSalt(t *testing.T) {
	tests := []struct {
		name       string
		wantLength int
		want1      error
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
			if !errors.Is(got1, tt.want1) {
				t.Errorf("generateSalt() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

// todo fix
//func TestGetEmailAndPassword(t *testing.T) {
//	testHeaderHappyPath := make(map[string]string)
//	testHeaderHappyPath["Authentication"] = "Zm9vOmRvZw=="
//
//	testHeaderUnhappyPathTooManyFields := make(map[string]string)
//	testHeaderUnhappyPathTooManyFields["Authentication"] = "Zm9vOmRvZzpzbG90aA=="
//
//	type args struct {
//		request events.APIGatewayProxyRequest
//	}
//	validEmail := "foo@gmail.com"
//	validPassword := "564873iideliuie"
//	validCredentials := Credentials{
//		Email:    validEmail,
//		Password: validPassword,
//	}
//	bytes, marshalErr := json.Marshal(&validCredentials)
//	if marshalErr != nil {
//		t.Errorf("error has occurred when marshalling test credentials: %v", marshalErr)
//	}
//
//	tests := []struct {
//		name    string
//		args    args
//		want    *string
//		want1   *string
//		wantErr bool
//	}{
//		{
//			name: "Happy Path",
//			args: args{
//				request: events.APIGatewayProxyRequest{
//					Body: string(bytes),
//				},
//			},
//			want:    &validEmail,
//			want1:   &validPassword,
//			wantErr: false,
//		},
//		{
//			name: "Unhappy path, missing credentials",
//			args: args{
//				request: events.APIGatewayProxyRequest{},
//			},
//			want:    nil,
//			want1:   nil,
//			wantErr: true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, got1, err := GetEmailAndPassword(tt.args.request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetEmailAndPassword() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetEmailAndPassword() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("GetEmailAndPassword() got1 = %v, want %v", got1, tt.want1)
//			}
//		})
//	}
//}

func TestHashPassword(t *testing.T) {
	type args struct {
		password string
		salt     string
	}
	result := "$argon2id$v=19$m=32768,t=3,p=1,l=32$bbFDuNOIT8LNpCkkE3c9jn5Nf26sTScLQwjmJqq4cr0=$iR3jRLGver2kGINumDXo1Al6Vlz2r9pG0FC9w0Pi8Q0="
	tests := []struct {
		name    string
		args    args
		want    *string
		wantErr bool
	}{
		{
			name: "Happy Path",
			args: args{
				password: "password123",
				salt:     "bbFDuNOIT8LNpCkkE3c9jn5Nf26sTScLQwjmJqq4cr0=",
			},
			want:    &result,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSecureHash(tt.args.password, tt.args.salt)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSecureHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateSecureHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	// todo write integration test for this
	//t.Run("Happy Path", func(t *testing.T) {
	//	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		r.Header["Authentication"] = []string{"cGl6emE6bWFu"}
	//	}))
	//	defer ts.Close()
	//
	//	headers := make(map[string]string)
	//	headers["Authentication"] = "cGl6emE6bWFu"
	//	_, err := handler(events.APIGatewayProxyRequest{
	//		Headers: headers,
	//	})
	//	if err != nil {
	//		t.Fatalf("Error failed to trigger with an invalid HTTP response: %v", err)
	//	}
	//})

	// todo write tests for each of the errors in auth.Error...
	//t.Run("Non 200 Response", func(t *testing.T) {
	//	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		w.WriteHeader(500)
	//	}))
	//	defer ts.Close()
	//
	//	//&& err.Error() != ErrNon200Response.Error()
	//	headers := make(map[string]string)
	//	headers["Authentication"] = "cGl6emE6bWFu"
	//	_, err := handler(events.APIGatewayProxyRequest{
	//		Headers: headers,
	//	})
	//	if err != nil {
	//		t.Fatalf("Error failed to trigger with an invalid HTTP response: %v", err)
	//	}
	//})

}

func Test_verifyValidUserProvidedPassword(t *testing.T) {
	type args struct {
		hashedInput string
		userHash    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if e := verifyValidUserProvidedPassword(tt.args.hashedInput, tt.args.userHash); (e != "") != tt.wantErr {
				t.Errorf("verifyValidUserProvidedPassword() error = %v, wantErr %v", e, tt.wantErr)
			}
		})
	}
}

// todo setup integration test for this
//func Test_getUserHash(t *testing.T) {
//	type args struct {
//		email *string
//	}
//	testEmail := "something@hotmail.com"
//	testEmailWithTypo := "somethingd@hotmail.com"
//	tests := []struct {
//		name      string
//		args      args
//		want      *User
//		wantErr   bool
//		wantedErr error
//	}{
//		{
//			name: "Happy Path",
//			args: args{
//				email: &testEmail,
//			},
//			want:    &User{
//				emailVerified: false,
//				passwordHash:  "",
//			},
//			wantErr: false,
//		}, {
//			name: "Unhappy Path user not found",
//			args: args{
//				email: &testEmailWithTypo,
//			},
//			want:      &User{
//				emailVerified: false,
//				passwordHash:  "",
//			},
//			wantErr:   true,
//			wantedErr: auth.ErrorUserNotFound,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := getUser(tt.args.email)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("getUser() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if tt.wantErr && tt.wantedErr != err.AppVersion {
//				t.Errorf("getUser() error = %v, wantErr %v, wantedErr %v, but got %v", err, tt.wantErr, tt.wantedErr, err.AppVersion)
//				return
//			}
//			if got != tt.want {
//				t.Errorf("getUser() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func Test_getSalt(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Happy Path",
			args: args{
				hash: "$argon2id$v=19$m=32768,t=3,p=1,l=32$bbFDuNOIT8LNpCkkE3c9jn5Nf26sTScLQwjmJqq4cr0=$iR3jRLGver2kGINumDXo1Al6Vlz2r9pG0FC9w0Pi8Q0=",
			},
			want:    "bbFDuNOIT8LNpCkkE3c9jn5Nf26sTScLQwjmJqq4cr0=",
			wantErr: false,
		},
		{
			name: "Unhappy Path",
			args: args{
				hash: "$bbFDuNOIT8LNpCkkE3c9jn5Nf26sTScLQwjmJqq4cr0=$iR3jRLGver2kGINumDXo1Al6Vlz2r9pG0FC9w0Pi8Q0=",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSalt(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSalt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSalt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleLogin(tt.args.w, tt.args.r)
		})
	}
}

//func Test_handler(t *testing.T) {
//	type args struct {
//		request Credentials
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    Credentials
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
			if got := EmailIsValidFormat(tt.args.email); got != tt.want {
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
		name    string
		args    args
		want    userErr
		wantErr bool
	}{
		{
			name: "Happy path",
			args: args{
				password: validPassword,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Cannot be all numbers",
			args: args{
				password: tooManyNumbersPassword,
			},
			want:    ErrorPasswordCannotContainAllNumbers,
			wantErr: false,
		},
		{
			name: "Must be at least 12 characters long 1",
			args: args{
				password: tooFewCharsPassword,
			},
			want:    ErrorPasswordMustBeAtLeastTwelveCharsLong,
			wantErr: false,
		},
		{
			name: "Must provide a password",
			args: args{
				password: "",
			},
			want:    ErrorPasswordWasNotProvided,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := PasswordIsAcceptable(tt.args.password)
			if err == nil && tt.wantErr {
				t.Errorf("got: %v but didn't expect", err)
			}
			if e != tt.want {
				t.Errorf("expected %s but got %s", tt.want, e)
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
		want userErr
	}{
		{
			name: "happy path",
			args: args{
				email: MockValidEmail,
			},
			want: "",
		},
		{
			name: "unhappy path, invalid email",
			args: args{
				email: MockInValidEmail,
			},
			want: ErrorInvalidEmailAddress,
		},
		{
			name: "Unhappy path, password not provided",
			args: args{
				email: missingEmail,
			},
			want: ErrorMissingEmailAddress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emailIsValid(tt.args.email)
			if got != tt.want {
				t.Errorf("emailIsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
