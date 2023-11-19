package main

import (
	"log"
	"strengthgadget.com/m/v2/test_tornado/model"
	"strengthgadget.com/m/v2/test_tornado/test_case"
	"sync"
)

var tests []model.TestCase

func initTestCases() {
	tests = append(
		tests,
		test_case.CanRegisterAndLogin,
		test_case.CanDoWorkout,
	)
}

func runTestCases() (string, error) {
	errChan := make(chan error, len(tests))
	var wg sync.WaitGroup

	for _, test := range tests {
		wg.Add(1)
		go func(t model.TestCase) {
			defer wg.Done()
			if err := t(); err != nil {
				errChan <- err
			}
		}(test)
	}

	wg.Wait()
	close(errChan)

	var responseErr error
	for err := range errChan {
		if err != nil {
			// Handle yer errors, ye scallywag!
			log.Printf("TEST FAILURE. Error: %v", err)
			responseErr = err
		}
	}
	var resultMessage string
	if responseErr == nil {
		resultMessage = "TEST SUITE SUCCESSFUL!"
	} else {
		resultMessage = "TEST SUITE FAILED"
	}
	log.Printf(resultMessage)
	return resultMessage, responseErr
}
