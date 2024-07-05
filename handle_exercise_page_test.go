package main

import (
	"reflect"
	"testing"
)

func Test_generateTimeOptions(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		expect := TimeOptions{
			{
				Label: "0:05",
				Value: 5,
			},
			{
				Label: "0:10",
				Value: 10,
			},
			{
				Label: "0:15",
				Value: 15,
			},
			{
				Label: "0:20",
				Value: 20,
			},
			{
				Label: "0:25",
				Value: 25,
			},
			{
				Label: "0:30",
				Value: 30,
			},
			{
				Label: "0:35",
				Value: 35,
			},
			{
				Label: "0:40",
				Value: 40,
			},
			{
				Label: "0:45",
				Value: 45,
			},
			{
				Label: "0:50",
				Value: 50,
			},
			{
				Label: "0:55",
				Value: 55,
			},
			{
				Label: "1:00",
				Value: 60,
			},
			{
				Label: "1:05",
				Value: 65,
			},
		}
		got := generateTimeOptions(5, 65)
		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expected: %+v, but got: %+v", expect, got)
		}
	})
}
