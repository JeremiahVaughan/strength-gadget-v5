package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
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

func (dw *DailyWorkout) FromRedis(ctx context.Context, client *redis.Client, key string, weekday time.Weekday) error {
	workoutJson, err := client.HGet(ctx, GetDailyWorkoutKey(weekday), key).Result()
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

func GenerateDailyWorkout(ctx context.Context, db *pgxpool.Pool, redisDb *redis.Client) error {
	allExercises, err := fetchAllExercises(ctx, db)
	if err != nil {
		return fmt.Errorf("error, when fetchAllExercises() for generateDailyWorkout(). Error: %v", err)
	}

	// third key is muscle group id, the value is the exercises that target the muscle group
	exerciseMap := make(map[RoutineType]map[ExerciseType]map[string][]Exercise)
	for _, exercise := range allExercises {
		// Initialize nested maps and slices if they do not exist yet
		if exerciseMap[exercise.RoutineType] == nil {
			exerciseMap[exercise.RoutineType] = make(map[ExerciseType]map[string][]Exercise)
		}
		if exerciseMap[exercise.RoutineType][exercise.ExerciseType] == nil {
			exerciseMap[exercise.RoutineType][exercise.ExerciseType] = make(map[string][]Exercise)
		}

		// Categorize the exercise
		exerciseMap[exercise.RoutineType][exercise.ExerciseType][exercise.MuscleGroupId] = append(exerciseMap[exercise.RoutineType][exercise.ExerciseType][exercise.MuscleGroupId], exercise)
	}

	lowerWorkout, err := generateDailyWorkoutValue(exerciseMap, LOWER)
	if err != nil {
		return fmt.Errorf("error, when generateDailyWorkoutValue(model.LOWER) for generateDailyWorkout(). Error: %v", err)
	}
	coreWorkout, err := generateDailyWorkoutValue(exerciseMap, CORE)
	if err != nil {
		return fmt.Errorf("error, when generateDailyWorkoutValue(model.CORE) for generateDailyWorkout(). Error: %v", err)
	}
	upperWorkout, err := generateDailyWorkoutValue(exerciseMap, UPPER)
	if err != nil {
		return fmt.Errorf("error, when generateDailyWorkoutValue(model.UPPER) for generateDailyWorkout(). Error: %v", err)
	}
	dailyWorkoutKey := GetDailyWorkoutKey(time.Now().Weekday())
	err = redisDb.HSet(ctx, dailyWorkoutKey,
		getDailyWorkoutHashKey(LOWER), lowerWorkout,
		getDailyWorkoutHashKey(CORE), coreWorkout,
		getDailyWorkoutHashKey(UPPER), upperWorkout,
	).Err()
	if err != nil {
		return fmt.Errorf("error, when attempting to create daily_workout redis hash. Error: %v", err)
	}

	err = redisDb.Expire(ctx, dailyWorkoutKey, 48*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("error, then attempting to set expiration for daily_workout redis hash. Error: %v", err)
	}

	healthPushUrl := os.Getenv("TF_VAR_daily_workout_generated_push_health")
	if healthPushUrl != "" {
		err = updateHealthCheck(healthPushUrl)
		if err != nil {
			return fmt.Errorf("error, when updateHealthCheck() for generateDailyWorkout(). Error: %v", err)
		}
	}
	return nil
}

func generateDailyWorkoutValue(exerciseMap map[RoutineType]map[ExerciseType]map[string][]Exercise, rt RoutineType) ([]byte, error) {
	dailyWorkout := DailyWorkout{}
	var cardioExercises []Exercise
	cardioExercises = []Exercise{}
	targetRoutine := exerciseMap[rt]
	for _, v := range exerciseMap[ALL][Cardio] {
		for _, e := range v {
			cardioExercises = append(cardioExercises, e)
		}
	}
	dailyWorkout.CardioExercises = cardioExercises
	dailyWorkout.ShuffleCardioExercises()

	weightLiftingExercises := targetRoutine[Weightlifting]
	calisthenicExercises := targetRoutine[Calisthenics]
	combinedExercises := make(map[string][]Exercise)
	// First, copy everything from weightLiftingExercises to combinedExercises
	for key, exercises := range weightLiftingExercises {
		combinedExercises[key] = append(combinedExercises[key], exercises...)
	}
	// Then, append exercises from calisthenicExercises to combinedExercises
	for key, exercises := range calisthenicExercises {
		combinedExercises[key] = append(combinedExercises[key], exercises...)
	}
	for _, exercises := range combinedExercises {
		dailyWorkout.MuscleCoverageMainExercises = append(dailyWorkout.MuscleCoverageMainExercises, exercises)
	}
	dailyWorkout.ShuffleMuscleCoverageMainExercises()

	// key is exercise id
	allExercisesMap := make(map[string]Exercise)
	for _, exercises := range combinedExercises {
		for _, e := range exercises {
			allExercisesMap[e.Id] = e
		}
	}
	for _, exercise := range allExercisesMap {
		dailyWorkout.AllMainExercises = append(dailyWorkout.AllMainExercises, exercise)
	}
	dailyWorkout.ShuffleMainExercises()

	coolDownExercises := targetRoutine[CoolDown]
	for _, exercises := range coolDownExercises {
		dailyWorkout.CoolDownExercises = append(dailyWorkout.CoolDownExercises, exercises)
	}
	dailyWorkout.ShuffleCoolDownExercises()

	bytes, err := json.Marshal(dailyWorkout)
	if err != nil {
		return nil, fmt.Errorf("error, when marshalling daily workout to json. Error: %v", err)
	}
	return bytes, nil
}

// GetDailyWorkoutKey is a function that takes a weekday as input and returns a string key for the daily workout.
// The weekday is used to handle the use-case where the daily workout may get changed while a user is working out.
// This user will continue to use the same daily workout since it is bound to the weekday.
// It uses the DailyWorkoutKey constant and the weekday to create the key in the format: "daily_workout:weekday"
func GetDailyWorkoutKey(weekday time.Weekday) string {
	return fmt.Sprintf("%s:%d", DailyWorkoutKey, weekday)
}

func updateHealthCheck(url string) error {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("error, when creating post request. ERROR: %v", err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if response != nil {
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				log.Printf("error, when attempting to close response body: %v", err)
			}
		}(response.Body)
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode > 299) {
		if response.StatusCode == http.StatusNotFound {
			log.Printf("recieved a 404 when attempting url: %s", request.URL)
		}
		var rb []byte
		rb, err = io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error, when reading error response body: %v", err)
		}
		return fmt.Errorf("error, when performing get request. ERROR: %v. RESPONSE CODE: %d. RESPONSE MESSAGE: %s", err, response.StatusCode, string(rb))
	}
	if err != nil {
		if response != nil {
			err = fmt.Errorf("error: %v. RESPONSE CODE: %d", err, response.StatusCode)
		}
		return fmt.Errorf("error, when performing post request. ERROR: %v", err)
	}

	return nil

}

func fetchAllExercises(ctx context.Context, db *pgxpool.Pool) ([]Exercise, error) {
	rows, err := db.Query(
		ctx,
		"SELECT e.id, e.name, e.demonstration_giphy_id, emg.muscle_group_id, e.exercise_type_id, mg.workout_routine\nFROM exercise e\nJOIN exercise_muscle_group emg on e.id = emg.exercise_id\nJOIN muscle_group mg on emg.muscle_group_id = mg.id",
	)
	defer rows.Close()

	if err != nil {
		return nil, fmt.Errorf("error, when attempting to retrieve records. Error: %v", err)
	}

	var exercises []Exercise
	for rows.Next() {
		var exercise Exercise
		err = rows.Scan(
			&exercise.Id,
			&exercise.Name,
			&exercise.DemonstrationGiphyId,
			&exercise.MuscleGroupId,
			&exercise.ExerciseType,
			&exercise.RoutineType,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when scanning database rows: %v", err)
		}
		exercises = append(exercises, exercise)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error, when iterating through database rows: %v", err)
	}
	return exercises, nil
}

func getDailyWorkoutHashKey(rt RoutineType) string {
	return fmt.Sprintf("%s%d", DailyWorkoutHashKeyPrefix, rt)
}
