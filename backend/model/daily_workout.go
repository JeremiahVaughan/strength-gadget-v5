package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
)

// todo address the issue where if someone is working out when the daily workout gets rotated
type DailyWorkout struct {
	// Cardio is done first for initial warmup
	CardioExercises []Exercise `json:"cardioExercises"`
	// outer slice is for each target muscle group, inner slice is for
	// applicable muscle groups for the corresponding target muscle group
	MuscleCoverageMainExercises [][]Exercise `json:"muscleCoverageExercises"`
	// AllMainExercises is to be used by filler exercises to reach 3 full super sets
	AllMainExercises []Exercise `json:"allExercises"`

	// CoolDownExercises is used for stretching
	CoolDownExercises [][]Exercise `json:"coolDownExercises"`
}

func (dw *DailyWorkout) FromRedis(ctx context.Context, client *redis.Client, key string) error {
	workoutJson, err := client.HGet(ctx, DailyWorkoutKey, key).Result()
	if err != nil {
		return fmt.Errorf("error retrieving DailyWorkout from Redis. Error: %v", err)
	}

	// Check if key exists (but expired or never set)
	if workoutJson == "" {
		return fmt.Errorf("error, DailyWorkout key expired or never set")
	}

	// Unmarshal
	err = json.Unmarshal([]byte(workoutJson), dw)
	if err != nil {
		return fmt.Errorf("error unmarshalling DailyWorkout. Error: %v", err)
	}

	return nil
}

func (d *DailyWorkout) ShuffleCardioExercises() {
	rand.Shuffle(len(d.CardioExercises), func(i, j int) {
		d.CardioExercises[i], d.CardioExercises[j] = d.CardioExercises[j], d.CardioExercises[i]
	})
}

func (d *DailyWorkout) ShuffleMuscleCoverageMainExercises() {
	// Shuffle the outer slice
	rand.Shuffle(len(d.MuscleCoverageMainExercises), func(i, j int) {
		d.MuscleCoverageMainExercises[i], d.MuscleCoverageMainExercises[j] = d.MuscleCoverageMainExercises[j], d.MuscleCoverageMainExercises[i]
	})

	// Shuffle each inner slice
	for _, exercises := range d.MuscleCoverageMainExercises {
		rand.Shuffle(len(exercises), func(i, j int) {
			exercises[i], exercises[j] = exercises[j], exercises[i]
		})
	}
}

func (d *DailyWorkout) ShuffleCoolDownExercises() {
	// Shuffle the outer slice
	rand.Shuffle(len(d.CoolDownExercises), func(i, j int) {
		d.CoolDownExercises[i], d.CoolDownExercises[j] = d.CoolDownExercises[j], d.CoolDownExercises[i]
	})

	// Shuffle each inner slice
	for _, exercises := range d.CoolDownExercises {
		rand.Shuffle(len(exercises), func(i, j int) {
			exercises[i], exercises[j] = exercises[j], exercises[i]
		})
	}
}
func (d *DailyWorkout) ShuffleMainExercises() {
	rand.Shuffle(len(d.AllMainExercises), func(i, j int) {
		d.AllMainExercises[i], d.AllMainExercises[j] = d.AllMainExercises[j], d.AllMainExercises[i]
	})
}
