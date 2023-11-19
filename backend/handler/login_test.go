package handler

import (
	"net/http"
	"strengthgadget.com/m/v2/constants"
	"strengthgadget.com/m/v2/model"
	"testing"
)

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
			if err := verifyValidUserProvidedPassword(tt.args.hashedInput, tt.args.userHash); (err != nil) != tt.wantErr {
				t.Errorf("verifyValidUserProvidedPassword() error = %v, wantErr %v", err, tt.wantErr)
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
		name      string
		args      args
		want      string
		wantErr   bool
		wantedErr model.UserFeedbackError
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
			want:      "",
			wantErr:   true,
			wantedErr: constants.ErrorUnexpectedTryAgain,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSalt(tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSalt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.wantedErr != err.UserFeedbackError {
				t.Errorf("getUser() error = %v, wantErr %v, wantedErr %v, but got %v", err, tt.wantErr, tt.wantedErr, err.UserFeedbackError)
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
