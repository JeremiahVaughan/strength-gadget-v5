package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func IncrementRateLimitingCount(ctx context.Context, key string, windowLengthInSecondsForTheNumberOfAllowedAttemptsBeforeLockout int) error {

	// The count check has to come from outside the transaction otherwise it always returns 0 for some reason.
	countExists, err := RedisConnectionPool.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("error, when getting current rate limit: %v", err)
	}

	// Must use a transaction to address the edge case where the current count is set to 1 and the expiration call fails. This would cause the user to have a counter that never expires.
	pipe := RedisConnectionPool.TxPipeline()
	err = pipe.Incr(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("error, when incrementing rate limit: %v", err)
	}
	if countExists == 0 {
		err = pipe.Expire(ctx, key, time.Duration(windowLengthInSecondsForTheNumberOfAllowedAttemptsBeforeLockout)*time.Second).Err()
		if err != nil {
			return fmt.Errorf("error, when expiring increment key for rate limit: %v", err)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error, when executing redis transaction: %v", err)
	}
	return nil
}

func HasRateLimitBeenReached(ctx context.Context, key string, attemptLimit int) (bool, error) {
	result, err := RedisConnectionPool.Get(ctx, key).Result()
	switch {
	// todo replace other places in this app where we are doing a separate call to test if the key exists. This pattern seen here is desired since its one call.
	case errors.Is(err, redis.Nil):
		// The key does not exist
		return false, nil
	case err != nil:
		return true, fmt.Errorf("error, when attempting to get key from redis. Error: %v", err)
	default:
		var parsedResult int
		parsedResult, err = strconv.Atoi(result)
		if err != nil {
			return true, fmt.Errorf("error, when attempting to parse result from redis. Error: %v", err)
		}
		return parsedResult >= attemptLimit, nil
	}
}
