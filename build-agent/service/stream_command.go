package service

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

func StreamCmd(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}

	wg := &sync.WaitGroup{}
	errChan := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = streamOutput(stdout, "STDOUT"); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = streamOutput(stderr, "STDERR"); err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err = range errChan {
		if err != nil {
			return fmt.Errorf("error, when attempting to stream command: %v", err)
		}
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("command finished with error: %v", err)
	}
	return nil
}

func streamOutput(r io.Reader, prefix string) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("[%s] %s", prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("[%s] Error reading output: %v", prefix, err)
	}
	return nil
}
