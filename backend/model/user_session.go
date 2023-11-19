package model

type UserSession struct {
	UserId        string
	SessionKey    string
	Authenticated bool
}
