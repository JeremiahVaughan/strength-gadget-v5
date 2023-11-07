package model

type ForgotPassword struct {
	Email       string `json:"email"`
	ResetCode   string `json:"resetCode"`
	NewPassword string `json:"newPassword"`
}
