package model

// todo validate the length of input so it can't crash your stuff
type VerificationRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerificationCode struct {
	Id      string
	Code    string
	Expires uint64
	UserId  string
}
