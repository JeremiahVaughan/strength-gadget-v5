package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type UserSlottedExercises struct {
	SlottedExercisesMap   map[string]Exercise `json:"slottedExercisesMap"`
	SlottedExercisesSlice []Exercise          `json:"slottedExercisesSlice"`
}

func (use *UserSlottedExercises) FromRedis(ctx context.Context, client *redis.Client, userId string) error {
	// Assuming the key for each UserSlottedExercises in Redis is "UserSlottedExercises:<userId>"
	jsonStr, err := client.Get(ctx, UserExercisesSlottedPrefix+userId).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("error, when fetching slotted exercises from redis. User: %s. Error: %v", userId, err)
	}

	err = json.Unmarshal([]byte(jsonStr), use)
	if err != nil {
		return fmt.Errorf("error, when unmarshalling slotted exercises from redis. User: %s. Error: %v", userId, err)
	}

	return nil
}

func (use *UserSlottedExercises) ToRedis(ctx context.Context, client *redis.Client, userId string, exp time.Duration) error {
	// Marshalling the UserSlottedExercises struct to a JSON string
	jsonStr, err := json.Marshal(use)
	if err != nil {
		return fmt.Errorf("error, when marshalling slotted exercises to redis. User: %s. Error: %v", userId, err)
	}

	// Storing the JSON string to Redis under the key "UserSlottedExercises:<userId>"
	err = client.Set(ctx, UserExercisesSlottedPrefix+userId, jsonStr, exp).Err()
	if err != nil {
		return fmt.Errorf("error, when storing slotted exercises to redis. User: %s. Error: %v", userId, err)
	}

	return nil
}
