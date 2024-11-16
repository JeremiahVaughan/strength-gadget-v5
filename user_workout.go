package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserWorkoutDto struct {
	ProgressIndex     WorkoutProgressIndex `json:"progressIndex"`
	Weekday           time.Weekday         `json:"weekday"`
	WorkoutId         int                  `json:"workoutId"`
	WarmupExercises   []Exercise           `json:"warmupExercises"`
	MainExercises     []Exercise           `json:"mainExercises"`
	CoolDownExercises []Exercise           `json:"coolDownExercises"`
}

// ChoosenExercisesMap is a type that represents which exercises have already been selected, key is the exerciseId, and the value holds the current measurement
type ChoosenExercisesMap map[int]int

// ExerciseMeasurementsMap key is exerciseId, value is current measurement value
type ExerciseMeasurementsMap map[int]int

type UserWorkout struct {
	Weekday                  time.Weekday         `json:"weekday"`
	ProgressIndex            WorkoutProgressIndex `json:"progressIndex,omitempty"`
	WorkoutRoutine           RoutineType          `json:"workoutRoutine"`
	WorkoutId                int                  `json:"workoutId"`
	SlottedWarmupExercises   []uint16             `json:"-"`
	SlottedMainExercises     []uint16             `json:"-"`
	SlottedCoolDownExercises []uint16             `json:"-"`
	// UserExerciseDataMap also is used to tell if an exercise has already been selected or not
	UserExerciseDataMap ChoosenExercisesMap `json:"-"`
	Exists              bool                `json:"-"`
}

const (
	userWorkoutKey              = "userWorkoutKey"
	WorkoutProgressIndexKey     = "workoutProgressIndexKey"
	slottedWarmupExercisesKey   = "slottedWarmupExercises"
	slottedMainExercisesKey     = "slottedMainExercises"
	slottedCoolDownExercisesKey = "slottedCoolDownExercises"
	UserExerciseUserDataKey     = "userExerciseUserData"
)


func serializeUniqueMember(score int, exerciseIndex uint16) string {
	return strconv.Itoa(score) + ":" + strconv.Itoa(int(exerciseIndex))
}

func deserializeUniqueMember(member string) (uint16, error) {
	parts := strings.Split(member, ":")
	if len(parts) != 2 {
		return 00, fmt.Errorf("invalid member format")
	}

	exerciseIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("error parsing exerciseIndex: %v", err)
	}

	return uint16(exerciseIndex), nil
}

// getNextAvailableExercise finds the next available exercise from the exercise pool based on the starting exercise index and the already slotted exercises.
// It returns the index of the next available exercise in the exercise pool.
// If the exercise pool doesn't contain any available exercises, it returns the starting exercise.
func getNextAvailableExercise(
	currentOffset int,
	randomPool []int,
	exercisePool []Exercise,
	alreadySlottedExercises ChoosenExercisesMap,
) (nextExercise Exercise, nextOffset int, err error) {
	exercisePoolSize := len(exercisePool)
	if exercisePoolSize == 0 {
		return Exercise{}, 0, fmt.Errorf("error, cannot have an empty exercise pool")
	}

	counter := currentOffset
	var selectedExercise Exercise
	for i := 0; i <= len(exercisePool); i++ {
		selectedExercise, err = getSelectedExercise(counter, randomPool, exercisePool)
		if err != nil {
			return Exercise{}, 0, fmt.Errorf("error, when getSelectedExercise() for getNextAvailableExercise(). Error: %v", err)
		}
		if isNewExercise(selectedExercise.Id, alreadySlottedExercises) {
			alreadySlottedExercises[selectedExercise.Id] = 0 // init to zero because exercise measurements are updated later
			break
		}
		counter++
	}
	return selectedExercise, counter, nil
}

func getSelectedExercise(
	currentOffset int,
	randomPool []int,
	exercisePool []Exercise,
) (Exercise, error) {
	exercisePoolLength := len(exercisePool)
	if exercisePoolLength == 0 {
		return Exercise{}, errors.New("error, exercisePool cannot be empty")
	}
	randomPoolLength := len(randomPool)
	if randomPoolLength == 0 {
		return Exercise{}, errors.New("error, randomPool cannot be empty")
	}
	if randomPoolLength != exercisePoolLength {
		return Exercise{}, fmt.Errorf("error, randomPoolLength %d does not equal exercisePoolLength %d", randomPoolLength, exercisePoolLength)
	}
	selectedIndex := currentOffset % exercisePoolLength
	if selectedIndex >= randomPoolLength {
		return Exercise{}, fmt.Errorf(
			`error, selectedIndex is out of bounds with the randomPool. 
			selectedIndex: %d,
			currentOffset: %d,
			randomPoolLength: %d,
			randomPool: %+v,
			exercisePoolLength: %d,
			exercisePool: %+v,`,
			selectedIndex,
			currentOffset,
			randomPoolLength,
			randomPool,
			exercisePoolLength,
			exercisePool,
		)
	}
	actualIndex := randomPool[selectedIndex]
	if actualIndex >= exercisePoolLength {
		return Exercise{}, fmt.Errorf(
			`error, actualIndex is out of bounds for given exercise pool. 
			actualIndex: %d,
			currentOffset: %d,
			randomPoolLength: %d,
			randomPool: %+v,
			exercisePoolLength: %d,
			exercisePool: %+v,`,
			actualIndex,
			currentOffset,
			randomPoolLength,
			randomPool,
			exercisePoolLength,
			exercisePool,
		)
	}
	return exercisePool[actualIndex], nil
}

func isNewExercise(selectedExerciseId int, alreadySlottedExercises ChoosenExercisesMap) bool {
	_, alreadyExists := alreadySlottedExercises[selectedExerciseId]
	return !alreadyExists
}

func calculateNumberOfSets(workout AvailableWorkoutExercises, exercisesPerSuperSet int) int {
	length := len(workout.MainExercises)
	result := length / exercisesPerSuperSet

	if length%exercisesPerSuperSet != 0 {
		result += 1
	}
	return result
}

func GetUserKey(userId int64, key string) string {
	uid := strconv.FormatInt(userId, 10)
	return uid + ":" + key
}

func fetchCurrentWorkoutRoutine(ctx context.Context, db *pgxpool.Pool, userId int64) (RoutineType, error) {
	var result RoutineType
	err := db.QueryRow(
		ctx,
		`SELECT current_routine
        FROM public.athlete
        WHERE id = $1`,
		userId,
	).Scan(
		&result,
	)
	if err != nil {
		return 0, fmt.Errorf("error, when attempting to execute sql statement: %v", err)
	}
	return result, nil
}

func fetchExerciseMeasurements(
	ctx context.Context,
	db *pgxpool.Pool,
	userId int64,
	choosenExercises ChoosenExercisesMap,
) (currentMeasurements ChoosenExercisesMap, err error) {
	placeholders := make([]string, len(choosenExercises))
	var args []interface{}
	args = append(args, userId) // user id will be our first argument

	exerciseIds := make([]any, len(choosenExercises))
	i := 0
	for exerciseId := range choosenExercises {
		exerciseIds[i] = exerciseId
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		i++
	}

	args = append(args, exerciseIds...)

	query := fmt.Sprintf(
		`SELECT exercise_id, measurement 
        FROM last_completed_measurement 
        WHERE user_id = $1 
            AND exercise_id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error, when attempting to retrieve records. Error: %v", err)
	}

	for rows.Next() {
		var exerciseId int
		var measurement int
		err = rows.Scan(
			&exerciseId,
			&measurement,
		)
		if err != nil {
			return nil, fmt.Errorf("error, when scanning database rows: %v", err)
		}
		choosenExercises[exerciseId] = measurement
	}

	// add placeholders for measurements that haven't been persisted yet
	for exerciseId := range choosenExercises {
		_, ok := choosenExercises[exerciseId]
		if !ok {
			choosenExercises[exerciseId] = 0
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error, when iterating through database rows: %v", err)
	}
	return choosenExercises, nil
}

func fetchExerciseMeasurement(
	ctx context.Context,
	db *pgxpool.Pool,
	userId string,
	exerciseId string,
) (int, error) {
	var exerciseMeasurement int
	err := db.QueryRow(
		ctx,
		`SELECT measurement 
		 FROM last_completed_measurement
		 WHERE user_id = $1
		   AND exercise_id = $2`,
		userId,
		exerciseId,
	).Scan(
		&exerciseMeasurement,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// means the user hasn't done this exercise before for a measurement to be
			// recorded
			return 0, nil
		} else {
			return 0, fmt.Errorf("error, when attempting to execute sql statement: %v", err)
		}
	}
	return exerciseMeasurement, nil
}
