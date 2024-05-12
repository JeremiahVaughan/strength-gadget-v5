package main

type TextInput struct {
	Id          string
	Value       string
	Label       string
	Placeholder string
	Type        string
	ErrorMsg    userErr
	Disabled    bool
}
