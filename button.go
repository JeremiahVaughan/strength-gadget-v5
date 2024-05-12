package main

type ButtonColor string
type ButtonType string

const (
	PrimaryButtonColor   ButtonColor = "primary"
	SecondaryButtonColor ButtonColor = "secondary"
)

const (
	ButtonTypeRegular ButtonType = "button"
	ButtonTypeSubmit  ButtonType = "submit"
)

type Button struct {
	Id    string
	Label string
	Color ButtonColor
	Type  ButtonType
}
