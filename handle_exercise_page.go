package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func HandleExercisePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userSession, err := FetchUserSession(r)
	if err != nil {
		err = fmt.Errorf("error, when FetchUserSession() for HandleExercisePage(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}
	if !userSession.Authenticated {
		// HX-Redirect only works if the page has already been loaded so we have to use full redirect instead
		http.Redirect(w, r, EndpointLogin, http.StatusFound)
		return
	}

	progressIndexString := r.URL.Query().Get("progressIndex")
	if progressIndexString != "" {
		userSession.WorkoutSession.ProgressIndex, err = strconv.Atoi(progressIndexString)
		if err != nil {
			err = fmt.Errorf("error, when parsing progress index for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	}

	var nextExercise ExerciseDisplay
	switch r.Method {
	case http.MethodGet:
		if !userSession.WorkoutSessionExists { // todo also handle the case where the user clicks the button to start a new workout
			userSession.WorkoutSession, err = createNewWorkout(ctx, userSession.UserId)
			if err != nil {
				err = fmt.Errorf("error, when createNewWorkout() for HandleExercisePage(). Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
		}
		nextExercise, err = getNextExercise(
			ctx,
			userSession,
		)
		if err != nil {
			err = fmt.Errorf("error, when getNextExercise() for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		nextExercise.Exercise.LastCompletedMeasurement, err = getDefaultCompletedMeasurement(nextExercise.Exercise)
		if err != nil {
			err = fmt.Errorf("error, when applyDefaultStartingValues() for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	case http.MethodPut:
		err = r.ParseForm()
		if err != nil {
			err = fmt.Errorf("error, when parsing form after no button was clicked. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		req := swapExerciseRequest{
			ExerciseId: r.FormValue("exerciseId"),
			WorkoutId:  r.FormValue("workoutId"),
		}
		err = validateSwapExerciseRequest(&req)
		if err != nil {
			err = fmt.Errorf("error, when validating request for swapping exercise. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		// currentWorkout, err = SwapExercise(
		// 	r.Context(),
		// 	RedisConnectionPool,
		// 	ConnectionPool,
		// 	req.ExerciseId,
		// 	req.WorkoutId,
		// 	NumberOfSetsInSuperSet,
		// 	NumberOfExerciseInSuperset,
		// 	CurrentSupersetExpirationTimeInHours,
		// )
		// if err != nil {
		// 	err = fmt.Errorf("error, when validating request for swapping exercise. Error: %v", err)
		// 	HandleUnexpectedError(w, err)
		// 	return
		// }
		// nextExercise, err = getNextExercise(currentWorkout)
		// if err != nil {
		// 	err = fmt.Errorf("error, when getNextExercise() for HandleExercisePage(). Error: %v", err)
		// 	HandleUnexpectedError(w, err)
		// 	return
		// }
	case http.MethodPost:
		err = r.ParseForm()
		if err != nil {
			err = fmt.Errorf("error, when parsing form for post request for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		if err != nil {
			err = fmt.Errorf("error, when unmarshalling the progress index for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		// var lcm int
		// lcm, err = strconv.Atoi(r.FormValue("lastCompletedMeasurement"))
		// if err != nil {
		// 	err = fmt.Errorf("error, when parsing lastCompletedMeasurement form value for HandleExercisePage(). Error: %v", err)
		// 	HandleUnexpectedError(w, err)
		// 	return
		// }
		// _ = RecordIncrementedWorkoutStepRequest{
		// 	ProgressIndex:            progressIndex,
		// 	ExerciseId:               r.FormValue("exerciseId"),
		// 	LastCompletedMeasurement: lcm,

		// 	// WorkoutId is used to help prevent client and server sync issues
		// 	WorkoutId: r.FormValue("workoutId"),
		// }
		// nextExercise, err = incrementExerciseProgressIndex(r.Context(), req)
		// if err != nil {
		// 	err = fmt.Errorf("error, when incrementExerciseProgressIndex() for HandleExercisePage(). Error: %v", err)
		// 	HandleUnexpectedError(w, err)
		// 	return
		// }
	}

	err = userSession.WorkoutSession.saveToRedis(ctx, userSession.UserId)
	if err != nil {
		err = fmt.Errorf("error, when WorkoutSession.saveToRedis() for HandleExercisePage(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}

	switch r.Header.Get("HX-Trigger") {
	case nextExercise.Yes.Id:
		err = templateMap["exercise-page.html"].ExecuteTemplate(w, "content", nextExercise)
		if err != nil {
			err = fmt.Errorf("error, when executing exercise page template for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	default:
		err = templateMap["exercise-page.html"].ExecuteTemplate(w, "base", nextExercise)
		if err != nil {
			err = fmt.Errorf("error, when executing exercise page template for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	}
}

func createNewWorkout(ctx context.Context, userId int64) (WorkoutSession, error) {
	var err error
	newWorkout := WorkoutSession{
		CurrentWorkoutSeed: generateWorkoutSeed(userId),
	}

	newWorkout.CurrentWorkoutRoutine, err = fetchCurrentWorkoutRoutine(ctx, ConnectionPool, userId)
	if err != nil {
		return WorkoutSession{}, fmt.Errorf("error, when attempting to fetchCurrentWorkoutRoutine() for createNewWorkout(). Error: %v", err)
	}

	dr := DailyWorkoutRandomIndices{}
	switch newWorkout.CurrentWorkoutRoutine {
	case LOWER:
		dr.randomizeWorkoutIndices(lowerWorkout, newWorkout)
	case CORE:
		dr.randomizeWorkoutIndices(coreWorkout, newWorkout)
	case UPPER:
		dr.randomizeWorkoutIndices(upperWorkout, newWorkout)
	default:
		return WorkoutSession{}, fmt.Errorf("error, unexpected workout routine type: %v", newWorkout.CurrentWorkoutRoutine)
	}
	newWorkout.RandomizedIndices = dr

	newWorkout.CurrentOffsets = generateStartingOffsets(newWorkout.RandomizedIndices.MainMuscleGroups)
	newWorkout.CurrentExerciseMeasurements = make(ExerciseUserDataMap)

	err = newWorkout.saveToRedis(ctx, userId)
	if err != nil {
		return WorkoutSession{}, fmt.Errorf("error, when WorkoutSession.saveToRedis() for createNewWorkout(). Error: %v", err)
	}

	return newWorkout, nil
}

func generateStartingOffsets(mainMuscleGroups []int) DailyWorkoutOffsets {
	return DailyWorkoutOffsets{
		MainExercises: make([]int, len(mainMuscleGroups)),
	}
}

func (d *DailyWorkoutRandomIndices) randomizeWorkoutIndices(dailyWorkout AvailableWorkoutExercises, newWorkout WorkoutSession) {
	r := rand.New(rand.NewSource(newWorkout.CurrentWorkoutSeed))
	d.ShuffleCardioExercises(r, dailyWorkout, newWorkout)
	d.ShuffleMuscleCoverageMainExercises(r, dailyWorkout, newWorkout)
	d.ShuffleCoolDownExercises(r, dailyWorkout, newWorkout)
}

func generateWorkoutSeed(userId int64) int64 {
	t := time.Now().UTC()
	year := t.Year()
	month := t.Month()
	day := t.Day()
	return int64(year) + int64(month) + int64(day) + userId
}

// func incrementExerciseProgressIndex(ctx context.Context, req RecordIncrementedWorkoutStepRequest) (*Exercise, error) {
// 	var err error
// 	var currentWorkout *UserWorkoutDto
// 	err = validateRecordIncrementedWorkoutStepRequest(&req)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, when validateRecordIncrementedWorkoutStepRequest() for incrementExerciseProgressIndex(). Error: %v", err)
// 	}

// 	currentWorkout, err = GetCurrentWorkout(
// 		ctx,
// 		RedisConnectionPool,
// 		ConnectionPool,
// 		NumberOfSetsInSuperSet,
// 		NumberOfExerciseInSuperset,
// 		GetSuperSetExpiration(),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, when GetCurrentWorkout() for incrementExerciseProgressIndex(). Error: %v", err)
// 	}

// 	currentWorkout.ProgressIndex = getIncrementedProgressIndex(currentWorkout)

// 	err = RecordIncrementedWorkoutStep(ctx, req)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, RecordIncrementedWorkoutStep() for incrementExerciseProgressIndex(). Error: %v", err)
// 	}

// 	var nextExercise *Exercise
// 	nextExercise, err = getNextExercise(currentWorkout)
// 	if err != nil {
// 		return nil, fmt.Errorf("error, when getNextExercise() for incrementExerciseProgressIndex(). Error: %v", err)
// 	}

// 	return nextExercise, nil
// }

func getIncrementedProgressIndex(currentWorkout *UserWorkoutDto) WorkoutProgressIndex {
	progressIndex := currentWorkout.ProgressIndex
	workoutPhase := WorkoutPhase(len(progressIndex) - 1)
	exerciseProgressIndex := progressIndex[workoutPhase]

	exercisesInPhase := 0
	switch workoutPhase {
	case WorkoutPhaseWarmUp:
		exercisesInPhase = len(currentWorkout.WarmupExercises)
	case WorkoutPhaseMain:
		exercisesInPhase = len(currentWorkout.MainExercises)
	case WorkoutPhaseCoolDown:
		exercisesInPhase = len(currentWorkout.CoolDownExercises)
	}

	if exerciseProgressIndex+1 == exercisesInPhase {
		progressIndex = append(progressIndex, 0)
	} else {
		progressIndex[workoutPhase]++
	}
	return progressIndex
}

func getTimeLabel(time int) string {
	minute := time / 60
	seconds := time % 60
	return fmt.Sprintf("%d:%d", minute, seconds)
}

func getNextExercise(
	ctx context.Context,
	userSession *UserSession,
) (ExerciseDisplay, error) {
	var workoutExercises AvailableWorkoutExercises
	var err error
	timeSelectionCap := 180
	timeInterval := 5
	exercise := ExerciseDisplay{
		NextProgressIndex: userSession.WorkoutSession.ProgressIndex + 1,
		SelectMode:        true,
		Yes: Button{
			Id:    "yes",
			Label: "yes",
			Color: PrimaryButtonColor,
			Type:  ButtonTypeSubmit,
		},
		Complete: Button{
			Id:    "complete",
			Label: "complete",
			Color: PrimaryButtonColor,
			Type:  ButtonTypeSubmit,
		},
		No: Button{
			Id:    "no",
			Label: "no",
			Color: SecondaryButtonColor,
			Type:  ButtonTypeSubmit,
		},
	}
	exercise.TimeOptions = make(TimeOptions, timeSelectionCap/timeInterval)
	for i := timeInterval; i <= timeSelectionCap; i += timeInterval {
		exercise.TimeOptions[i] = TimeOption{
			Label: getTimeLabel(i),
			Value: i,
		}
	}
	switch userSession.WorkoutSession.CurrentWorkoutRoutine {
	case LOWER:
		workoutExercises = lowerWorkout
	case CORE:
		workoutExercises = coreWorkout
	case UPPER:
		workoutExercises = upperWorkout
	default:
		return ExerciseDisplay{}, fmt.Errorf("error, unexpected workout routine type: %v", userSession.WorkoutSession.CurrentWorkoutRoutine)
	}

	counter := 0
	if userSession.WorkoutSession.ProgressIndex == counter {
		exercise.Exercise, userSession.WorkoutSession.CurrentExerciseMeasurements, err = getNextAvailableExercise(
			userSession.WorkoutSession.CurrentOffsets.CardioExercise,
			userSession.WorkoutSession.RandomizedIndices.CardioExercises,
			workoutExercises.CardioExercises,
			userSession.WorkoutSession.CurrentExerciseMeasurements,
		)
		if err != nil {
			return ExerciseDisplay{}, fmt.Errorf("error, when getNextAvailableExercise() for getNextExercise() during cardio selection mode. Error: %v", err)
		}
		return exercise, nil
	}

	for i, r := range userSession.WorkoutSession.RandomizedIndices.MainMuscleGroups {
		counter++
		if userSession.WorkoutSession.ProgressIndex == counter {
			exercise.Exercise, userSession.WorkoutSession.CurrentExerciseMeasurements, err = getNextAvailableExercise(
				userSession.WorkoutSession.CurrentOffsets.MainExercises[i],
				userSession.WorkoutSession.RandomizedIndices.MainExercises[r],
				workoutExercises.MainExercises[r],
				userSession.WorkoutSession.CurrentExerciseMeasurements,
			)
			if err != nil {
				return ExerciseDisplay{}, fmt.Errorf("error, when getNextAvailableExercise() for getNextExercise() during main selection mode. Error: %v", err)
			}
			return exercise, nil
		}
	}

	exercise.SelectMode = false

	if userSession.WorkoutSession.ProgressIndex == counter+1 { // Fetching current measurements right after leaving selection mode
		for _, r := range userSession.WorkoutSession.RandomizedIndices.CoolDownMuscleGroups {
			exercise.Exercise, err = getSelectedExercise(
				0, // stretches don't require selection at the moment
				userSession.WorkoutSession.RandomizedIndices.CoolDownExercises[r],
				workoutExercises.CoolDownExercises[r],
			)
			if err != nil {
				return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() when fetching completed measurements for cooldown exercises. Error: %v", err)
			}
			userSession.WorkoutSession.CurrentExerciseMeasurements[exercise.Exercise.Id] = ExerciseUserData{}
		}
		err = fetchExerciseMeasurements(
			ctx,
			ConnectionPool,
			userSession.UserId,
			userSession.WorkoutSession.CurrentExerciseMeasurements,
		)
		if err != nil {
			return ExerciseDisplay{}, fmt.Errorf("error, when fetchExerciseMeasurements() for getNextExercise(). Error: %v", err)
		}
	}

	counter++
	if userSession.WorkoutSession.ProgressIndex == counter {
		exercise.Exercise, err = getSelectedExercise(
			userSession.WorkoutSession.CurrentOffsets.CardioExercise,
			userSession.WorkoutSession.RandomizedIndices.CardioExercises,
			workoutExercises.CardioExercises,
		)
		if err != nil {
			return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() during cardio workout mode. Error: %v", err)
		}
		return exercise, nil
	}

	for i := 0; i < NumberOfSetsInSuperSet; i++ {
		for j, r := range userSession.WorkoutSession.RandomizedIndices.MainMuscleGroups {
			counter++
			if userSession.WorkoutSession.ProgressIndex == counter {
				exercise.Exercise, err = getSelectedExercise(
					userSession.WorkoutSession.CurrentOffsets.MainExercises[j],
					userSession.WorkoutSession.RandomizedIndices.MainExercises[r],
					workoutExercises.MainExercises[r],
				)
				if err != nil {
					return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() during main workout mode. Error: %v", err)
				}
				return exercise, nil
			}
		}
	}

	for _, r := range userSession.WorkoutSession.RandomizedIndices.CoolDownMuscleGroups {
		counter++
		if userSession.WorkoutSession.ProgressIndex == counter {
			exercise.Exercise, err = getSelectedExercise(
				0, // stretches don't require selection
				userSession.WorkoutSession.RandomizedIndices.CoolDownExercises[r],
				workoutExercises.CoolDownExercises[r],
			)
			if err != nil {
				return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() during stretching workout mode. Error: %v", err)
			}
			return exercise, nil
		}
	}

	exercise.WorkoutCompleted = true
	return exercise, nil
}

func getDefaultCompletedMeasurement(exercise Exercise) (int, error) {
	if exercise.LastCompletedMeasurement != 0 {
		return exercise.LastCompletedMeasurement, nil
	}
	startingValue := 0
	if exercise.ExerciseType == ExerciseTypeCoolDown {
		startingValue = 30
	} else {
		switch exercise.MeasurementType {
		case MeasurementTypePounds:
			startingValue = 5
		case MeasurementTypeRepetition:
			startingValue = 3
		case MeasurementTypeSecond:
			startingValue = 10
		case MeasurementTypeMile:
			startingValue = 1
		default:
			return 0, fmt.Errorf("error, unexpected measurement type: %d", exercise.MeasurementType)
		}
	}
	return startingValue, nil
}
