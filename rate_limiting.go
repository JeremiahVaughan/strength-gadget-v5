package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nalgeon/redka"
	"strconv"
	"time"
)

func IncrementRateLimitingCount(ctx context.Context, key string, windowLengthInSecondsForTheNumberOfAllowedAttemptsBeforeLockout int) error {
	// we are not using redis increment because we don't want to risk the required subsequent expiration call failing and creating an immortal session
	value, err := RedisConnectionPool.Str().Get(key)
	if err != nil && !errors.Is(redka.ErrNotFound, err) {
		return fmt.Errorf("error, when getting current rate limit: %v", err)
	}
	var newValue string
	if errors.Is(redka.ErrNotFound, err) {
		newValue = "1"
	} else {
		parsedValue, err := strconv.Atoi(string(value))
		if err != nil {
			return fmt.Errorf("error, invalid rate limit value for IncrementRateLimitingCount(). Error: %v", err)
		}
		parsedValue++
		newValue = strconv.Itoa(parsedValue)
	}
	err = RedisConnectionPool.Str().SetExpires(
		key,
		newValue,
		time.Duration(windowLengthInSecondsForTheNumberOfAllowedAttemptsBeforeLockout)*time.Second,
	)
	if err != nil {
		return fmt.Errorf("error, when setting expiration time for IncrementRateLimitingCount(). Error: %v", err)
	}
	return nil
}

func HasRateLimitBeenReached(ctx context.Context, key string, attemptLimit int) (bool, error) {
	result, err := RedisConnectionPool.Str().Get(key)
	switch {
	// todo replace other places in this app where we are doing a separate call to test if the key exists. This pattern seen here is desired since its one call.
	case errors.Is(err, redka.ErrNotFound):
		// The key does not exist
		return false, nil
	case err != nil:
		return true, fmt.Errorf("error, when attempting to get key from redis. Error: %v", err)
	default:
		var parsedResult int
		parsedResult, err = strconv.Atoi(string(result))
		if err != nil {
			return true, fmt.Errorf("error, when attempting to parse result from redis. Error: %v", err)
		}
		return parsedResult >= attemptLimit, nil
	}
}
