package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type AvailableWorkoutExercises struct {
	// Cardio is done first for initial warmup
	CardioExercises []Exercise

	// outer slice is for each target muscle group, inner slice is for
	// applicable exercises for the corresponding target muscle group
	MainExercises [][]Exercise

	// CoolDownExercises is used for stretching
	CoolDownExercises [][]Exercise
}

// DailyWorkoutOffsets used by the user to change exercises
type DailyWorkoutOffsets struct {
	CardioExercise int   `json:"cardioExercises"`
	MainExercises  []int `json:"mainExercises"`
}

// DailyWorkoutRandomIndices all exercises in a random order to help with variety
type DailyWorkoutRandomIndices struct {
	CardioExercises []int `json:"cardioExercises"`
	// MainMuscleGroups represents the randomness of MainExercises outer slice
	MainMuscleGroups []int `json:"mainMuscleGroups"`
	// MainExercises outerslice is not random, inner slices are
	MainExercises [][]int `json:"mainExercises"`
	// CoolDownMuscleGroups represents the randomness of CoolDownExercises outer slice
	CoolDownMuscleGroups []int `json:"coolDownMuscleGroups"`
	// CoolDownExercises outerslice is not random, inner slices are
	CoolDownExercises [][]int `json:"coolDownExercises"`
}

func (d *DailyWorkoutRandomIndices) ShuffleCardioExercises(
	r *rand.Rand,
	dailyWorkout AvailableWorkoutExercises,
	newWorkout WorkoutSession,
) {
	d.CardioExercises = make([]int, len(dailyWorkout.CardioExercises))
	for i := range d.CardioExercises {
		d.CardioExercises[i] = i
	}

	r.Shuffle(len(d.CardioExercises), func(i, j int) {
		d.CardioExercises[i], d.CardioExercises[j] = d.CardioExercises[j], d.CardioExercises[i]
	})
}

func (d *DailyWorkoutRandomIndices) ShuffleMuscleCoverageMainExercises(
	r *rand.Rand,
	dailyWorkout AvailableWorkoutExercises,
	newWorkout WorkoutSession,
) {
	d.MainMuscleGroups = make([]int, len(dailyWorkout.MainExercises))
	for i := range d.MainMuscleGroups {
		d.MainMuscleGroups[i] = i
	}

	// Shuffle the outer slice
	r.Shuffle(len(d.MainMuscleGroups), func(i, j int) {
		d.MainMuscleGroups[i], d.MainMuscleGroups[j] = d.MainMuscleGroups[j], d.MainMuscleGroups[i]
	})

	// Shuffle each inner slice
	d.MainExercises = make([][]int, len(dailyWorkout.MainExercises))
	for i := range d.MainExercises {
		d.MainExercises[i] = make([]int, len(dailyWorkout.MainExercises[i]))
		for j := range d.MainExercises[i] {
			d.MainExercises[i][j] = j
		}
		r.Shuffle(len(d.MainExercises[i]), func(a, b int) {
			d.MainExercises[i][b], d.MainExercises[i][a] = d.MainExercises[i][a], d.MainExercises[i][b]
		})
	}
}

func (d *DailyWorkoutRandomIndices) ShuffleCoolDownExercises(
	r *rand.Rand,
	dailyWorkout AvailableWorkoutExercises,
	newWorkout WorkoutSession,
) {
	d.CoolDownMuscleGroups = make([]int, len(dailyWorkout.CoolDownExercises))
	for i := range d.CoolDownMuscleGroups {
		d.CoolDownMuscleGroups[i] = i
	}

	// Shuffle the outer slice
	r.Shuffle(len(d.CoolDownMuscleGroups), func(i, j int) {
		d.CoolDownMuscleGroups[i], d.CoolDownMuscleGroups[j] = d.CoolDownMuscleGroups[j], d.CoolDownMuscleGroups[i]
	})

	// Shuffle each inner slice
	d.CoolDownExercises = make([][]int, len(dailyWorkout.CoolDownExercises))
	for i := range d.CoolDownExercises {
		d.CoolDownExercises[i] = make([]int, len(dailyWorkout.CoolDownExercises[i]))
		for j := range d.CoolDownExercises[i] {
			d.CoolDownExercises[i][j] = j
		}
		r.Shuffle(len(d.CoolDownExercises[i]), func(a, b int) {
			d.CoolDownExercises[i][b], d.CoolDownExercises[i][a] = d.CoolDownExercises[i][a], d.CoolDownExercises[i][b]
		})
	}
}

func getTomorrowsWeekday(today time.Weekday) time.Weekday {
	return (today + 1) % 7
}

func generateWorkoutExercises(exerciseMap map[RoutineType]map[ExerciseType]map[int][]Exercise, rt RoutineType) AvailableWorkoutExercises {
	dailyWorkout := AvailableWorkoutExercises{}
	var cardioExercises []Exercise
	cardioExercises = []Exercise{}
	targetRoutine := exerciseMap[rt]
	for _, v := range exerciseMap[ALL][ExerciseTypeCardio] {
		cardioExercises = append(cardioExercises, v...)
	}
	dailyWorkout.CardioExercises = cardioExercises

	weightLiftingExercises := targetRoutine[ExerciseTypeWeightlifting]
	calisthenicExercises := targetRoutine[ExerciseTypeCalisthenics]

	combinedExercises := make(map[int][]Exercise)
	// First, copy everything from weightLiftingExercises to combinedExercises
	for key, exercises := range weightLiftingExercises {
		combinedExercises[key] = append(combinedExercises[key], exercises...)
	}
	// Then, append exercises from calisthenicExercises to combinedExercises
	for key, exercises := range calisthenicExercises {
		combinedExercises[key] = append(combinedExercises[key], exercises...)
	}

	for _, exercises := range combinedExercises {
		dailyWorkout.MainExercises = append(dailyWorkout.MainExercises, exercises)
	}

	coolDownExercises := targetRoutine[ExerciseTypeCoolDown]
	for _, exercises := range coolDownExercises {
		dailyWorkout.CoolDownExercises = append(dailyWorkout.CoolDownExercises, exercises)
	}

	return dailyWorkout
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

func getDailyWorkoutHashKey(rt RoutineType) string {
	return fmt.Sprintf("%s%d", DailyWorkoutHashKeyPrefix, rt)
}
