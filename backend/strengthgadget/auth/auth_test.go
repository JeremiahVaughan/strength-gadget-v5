package auth

import (
	"github.com/aws/aws-lambda-go/events"
	"reflect"
	"strengthgadget.com/m/v2/service"
	"testing"
)

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
			got, err := service.GenerateSecureHash(tt.args.password, tt.args.salt)
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

func TestGetAuthHeader(t *testing.T) {
	type args struct {
		request events.APIGatewayProxyRequest
	}

	testHeaderHappyPath := make(map[string]string)
	testHeaderHappyPath["Authentication"] = "Zm9vOmRvZw=="
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy Path",
			args: args{
				request: events.APIGatewayProxyRequest{
					Headers: testHeaderHappyPath,
				},
			},
			want: "Zm9vOmRvZw==",
		},
		{
			name: "Unhappy Path Missing Auth Header",
			args: args{
				request: events.APIGatewayProxyRequest{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAuthHeader(tt.args.request); got != tt.want {
				t.Errorf("GetAuthHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
