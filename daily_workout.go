package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sort"
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

	// Shuffle Muscle Groups
	r.Shuffle(len(d.MainMuscleGroups), func(i, j int) {
		d.MainMuscleGroups[i], d.MainMuscleGroups[j] = d.MainMuscleGroups[j], d.MainMuscleGroups[i]
	})

	// Shuffle Exercises
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

func generateWorkoutExercises(
	exerciseMap map[RoutineType]map[ExerciseType]map[int][]Exercise,
	muscleGroupMap map[int]MuscleGroup,
	rt RoutineType,
) (AvailableWorkoutExercises, error) {
	var err error
	dailyWorkout := AvailableWorkoutExercises{}
	targetRoutine := exerciseMap[rt]
	for _, v := range exerciseMap[ALL][ExerciseTypeCardio] {
		dailyWorkout.CardioExercises = append(dailyWorkout.CardioExercises, v...)
	}
	dailyWorkout.CardioExercises, err = appendFocusMuscleGroups(dailyWorkout.CardioExercises, MuscleGroupCardio.Id, muscleGroupMap)
	if err != nil {
		return AvailableWorkoutExercises{}, fmt.Errorf("error, when appendFocusMuscleGroups() for generateWorkoutExercises() for main exercises. Error: %v", err)
	}
	sort.Sort(Exercises(dailyWorkout.CardioExercises))

	weightLiftingExercises := targetRoutine[ExerciseTypeWeightlifting]
	calisthenicExercises := targetRoutine[ExerciseTypeCalisthenics]
	combinedExercises := make(map[int][]Exercise)
	// First, copy everything from weightLiftingExercises to combinedExercises
	for muscleGroupId, exercises := range weightLiftingExercises {
		combinedExercises[muscleGroupId] = append(combinedExercises[muscleGroupId], exercises...)
	}
	// Then, append exercises from calisthenicExercises to combinedExercises
	for muscleGroupId, exercises := range calisthenicExercises {
		combinedExercises[muscleGroupId] = append(combinedExercises[muscleGroupId], exercises...)
	}
	for muscleGroupId, exercises := range combinedExercises {
		exercises, err = appendFocusMuscleGroups(exercises, muscleGroupId, muscleGroupMap)
		if err != nil {
			return AvailableWorkoutExercises{}, fmt.Errorf("error, when appendFocusMuscleGroups() for generateWorkoutExercises() for main exercises. Error: %v", err)
		}
		sort.Sort(Exercises(exercises))
		dailyWorkout.MainExercises = append(dailyWorkout.MainExercises, exercises)
	}
	sort.Sort(MuscleGroupExercises(dailyWorkout.MainExercises))

	coolDownExercises := targetRoutine[ExerciseTypeCoolDown]
	for muscleGroupId, exercises := range coolDownExercises {
		exercises, err = appendFocusMuscleGroups(exercises, muscleGroupId, muscleGroupMap)
		if err != nil {
			return AvailableWorkoutExercises{}, fmt.Errorf("error, when appendFocusMuscleGroups() for generateWorkoutExercises() for cool down exercises. Error: %v", err)
		}
		sort.Sort(Exercises(exercises))
		dailyWorkout.CoolDownExercises = append(dailyWorkout.CoolDownExercises, exercises)
	}
	sort.Sort(MuscleGroupExercises(dailyWorkout.CoolDownExercises))
	return dailyWorkout, nil
}

func appendFocusMuscleGroups(exercises []Exercise, muscleGroupId int, muscleGroups map[int]MuscleGroup) ([]Exercise, error) {
	for i, e := range exercises {
		mg, ok := muscleGroups[muscleGroupId]
		if !ok {
			return nil, errors.New("error, expected muscle group to exist in map but it did not")
		}
		e.FocusMuscleGroup = mg.Name
		exercises[i] = e
	}
	return exercises, nil
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
