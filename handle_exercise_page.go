package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		if DebugMode == "true" {
			log.Printf("user session has expired, redirecting to login page")
		}
		// HX-Redirect only works if the page has already been loaded so we have to use full redirect instead
		http.Redirect(w, r, EndpointLogin, http.StatusSeeOther)
		return
	}

	if !userSession.WorkoutSessionExists { // todo also handle the case where the user clicks the button to start a new workout
		var currentWorkoutRoutine RoutineType
		currentWorkoutRoutine, err = fetchCurrentWorkoutRoutine(ctx, ConnectionPool, userSession.UserId)
		if err != nil {
			err = fmt.Errorf("error, when attempting to fetchCurrentWorkoutRoutine() for createNewWorkout(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		userSession.WorkoutSession, err = createNewWorkout(ctx, userSession.UserId, currentWorkoutRoutine)
		if err != nil {
			err = fmt.Errorf("error, when createNewWorkout() for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	}

	progressIndexString := r.URL.Query().Get("progressIndex")
	// pageFetchedAt keeps track of when the page was fetched. This supports the edge case where the user accidently restarts their
	// workout because they come back to a different tab of app or their auth session expired.
	pageFetchedAt := r.URL.Query().Get("pageFetchedAt")
	if progressIndexString == "" || pageFetchedAt == "" {
		// not having the progress index in the URL makes interactions too complex, so just always requiring it.
		alreadyAuthRedirect(w, r)
		return
	} else {
		var pfa int64
		pfa, err = strconv.ParseInt(pageFetchedAt, 10, 64)
		if err != nil {
			err = fmt.Errorf("error, when parsing pageFetchedAt for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		twoHoursAgo := time.Now().Unix() - 7200
		if pfa > twoHoursAgo {
			userSession.WorkoutSession.ProgressIndex, err = strconv.Atoi(progressIndexString)
			if err != nil {
				err = fmt.Errorf("error, when parsing progress index for HandleExercisePage(). Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
		}
	}

	var nextExercise ExerciseDisplay
	switch r.Method {
	case http.MethodGet:
		nextExercise, err = getExercise(
			ctx,
			userSession,
			false,
		)
		if err != nil {
			err = fmt.Errorf("error, when getNextExercise() for HandleExercisePage() when get. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	case http.MethodPut:
		if r.Header.Get("Hx-Trigger") == "no" {
			nextExercise, err = getExercise(
				ctx,
				userSession,
				true,
			)
			if err != nil {
				err = fmt.Errorf("error, when getNextExercise() for HandleExercisePage() when put. Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
		} else {
			err = r.ParseForm()
			if err != nil {
				err = fmt.Errorf("error, when parsing form for post request for HandleExercisePage(). Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
			lastCompletedMeasurement := r.FormValue("lastCompletedMeasurement")
			if lastCompletedMeasurement == "" {
				err = fmt.Errorf("error, must provide lastCompletedMeasurement for HandleExercisePage(). Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
			exerciseId := r.FormValue("exerciseId")
			if exerciseId == "" {
				err = fmt.Errorf("error, must provide exerciseId for HandleExercisePage(). Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
			var lcm int
			lcm, err = strconv.Atoi(lastCompletedMeasurement)
			if err != nil {
				err = fmt.Errorf("error, when converting lastCompletedMeasurement string to int. Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
			var eid int
			eid, err = strconv.Atoi(exerciseId)
			if err != nil {
				err = fmt.Errorf("error, when converting exerciseId string to int. Error: %v", err)
				HandleUnexpectedError(w, err)
				return
			}
			userSession.WorkoutSession.WorkoutMeasurements[eid] = lcm
		}
	case http.MethodPost:
		nextExercise, err = getExercise(
			ctx,
			userSession,
			false,
		)
		if err != nil {
			err = fmt.Errorf("error, when getNextExercise() for HandleExercisePage() when post. Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	}

	err = userSession.WorkoutSession.saveToRedis(ctx, userSession.UserId)
	if err != nil {
		err = fmt.Errorf("error, when WorkoutSession.saveToRedis() for HandleExercisePage(). Error: %v", err)
		HandleUnexpectedError(w, err)
		return
	}

	if nextExercise.WorkoutCompleted {
		if len(userSession.WorkoutSession.WorkoutMeasurements) == 0 {
			err = errors.New("error, no workout measurements found but were expecting them")
			HandleUnexpectedError(w, err)
			return
		}
		emq, args := generateQueryForExerciseMeasurements(
			userSession.WorkoutSession.WorkoutMeasurements,
			userSession.UserId,
		)
		_, err := ConnectionPool.Exec(
			ctx,
			emq,
			args...,
		)
		log.Printf("todo remove sql: %+v", args)
		if err != nil {
			err = fmt.Errorf("error, when persisting exercises measurements for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
		u := fmt.Sprintf("%s?currentWorkoutRoutine=%d", EndpointWorkoutComplete, userSession.WorkoutSession.CurrentWorkoutRoutine)
		w.Header().Set("HX-Redirect", u)
		return
	}

	switch r.Header.Get("HX-Trigger") {
	case nextExercise.Yes.Id, nextExercise.Complete.Id:
		err = templateMap["exercise-page.html"].ExecuteTemplate(w, "content", nextExercise)
		if err != nil {
			err = fmt.Errorf("error, when executing exercise page template for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	case "measurement":
		return
	default:
		err = templateMap["exercise-page.html"].ExecuteTemplate(w, "base", nextExercise)
		if err != nil {
			err = fmt.Errorf("error, when executing exercise page template for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	}
}

func createNewWorkout(ctx context.Context, userId int64, currentWorkoutRoutine RoutineType) (WorkoutSession, error) {
	var err error
	newWorkout := WorkoutSession{
		CurrentWorkoutSeed:    generateWorkoutSeed(userId),
		CurrentWorkoutRoutine: currentWorkoutRoutine,
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
	s := strconv.Itoa(seconds)
	if len(s) == 1 {
		s = fmt.Sprintf("%s%s", "0", s)
	}
	return fmt.Sprintf("%d:%s", minute, s)
}

func getWeightLabel(weight int) string {
	return fmt.Sprintf("%d lbs", weight)
}

func generateTimeOptions(timeInterval, timeSelectionCap int) MeasurementOptions {
	timeOptions := make(MeasurementOptions, timeSelectionCap/timeInterval)
	j := 0
	for i := timeInterval; i <= timeSelectionCap; i += timeInterval {
		timeOptions[j] = MeasurementOption{
			Label: getTimeLabel(i),
			Value: i,
		}
		j++
	}
	return timeOptions
}

func generateWeightOptions(weightInterval, timeSelectionCap int) MeasurementOptions {
	weightOptions := make(MeasurementOptions, timeSelectionCap/weightInterval)
	j := 0
	for i := weightInterval; i <= timeSelectionCap; i += weightInterval {
		weightOptions[j] = MeasurementOption{
			Label: getWeightLabel(i),
			Value: i,
		}
		j++
	}
	return weightOptions
}

func getExercise(
	ctx context.Context,
	userSession *UserSession,
	shuffle bool,
) (ExerciseDisplay, error) {
	var workoutExercises AvailableWorkoutExercises
	var err error
	exercise := ExerciseDisplay{
		ProgressIndex:     userSession.WorkoutSession.ProgressIndex,
		NextProgressIndex: userSession.WorkoutSession.ProgressIndex + 1,
		PageFetchedAt:     time.Now().Unix(),
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

	// choosenExercises are fetched once exercise selection has been completed
	choosenExercises := make(ChoosenExercisesMap)

	counter := 0
	if userSession.WorkoutSession.ProgressIndex == counter && shuffle {
		userSession.WorkoutSession.CurrentOffsets.CardioExercise++
	}
	exercise.Exercise, userSession.WorkoutSession.CurrentOffsets.CardioExercise, err = getNextAvailableExercise(
		userSession.WorkoutSession.CurrentOffsets.CardioExercise,
		userSession.WorkoutSession.RandomizedIndices.CardioExercises,
		workoutExercises.CardioExercises,
		choosenExercises,
	)
	if err != nil {
		return ExerciseDisplay{}, fmt.Errorf("error, when getNextAvailableExercise() for getNextExercise() during cardio selection mode. Error: %v", err)
	}

	if userSession.WorkoutSession.ProgressIndex == counter {
		return exercise, nil
	}

	for i, r := range userSession.WorkoutSession.RandomizedIndices.MainMuscleGroups {
		counter++
		if userSession.WorkoutSession.ProgressIndex == counter && shuffle {
			userSession.WorkoutSession.CurrentOffsets.MainExercises[i]++
		}
		exercise.Exercise, userSession.WorkoutSession.CurrentOffsets.MainExercises[i], err = getNextAvailableExercise(
			userSession.WorkoutSession.CurrentOffsets.MainExercises[i],
			userSession.WorkoutSession.RandomizedIndices.MainExercises[r],
			workoutExercises.MainExercises[r],
			choosenExercises,
		)
		if err != nil {
			return ExerciseDisplay{}, fmt.Errorf("error, when getNextAvailableExercise() for getNextExercise() during main selection mode. Error: %v", err)
		}
		if userSession.WorkoutSession.ProgressIndex == counter {
			return exercise, nil
		}
	}

	// select all default stetches
	for _, r := range userSession.WorkoutSession.RandomizedIndices.CoolDownMuscleGroups {
		exercise.Exercise, err = getSelectedExercise(
			0, // stretches don't require selection
			userSession.WorkoutSession.RandomizedIndices.CoolDownExercises[r],
			workoutExercises.CoolDownExercises[r],
		)
		if err != nil {
			return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() during stretching workout mode. Error: %v", err)
		}
		choosenExercises[exercise.Exercise.Id] = 0
	}

	exercise.SelectMode = false

	counter++
	exercise.Exercise, err = getSelectedExercise(
		userSession.WorkoutSession.CurrentOffsets.CardioExercise,
		userSession.WorkoutSession.RandomizedIndices.CardioExercises,
		workoutExercises.CardioExercises,
	)
	if err != nil {
		return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() during cardio workout mode. Error: %v", err)
	}
	if userSession.WorkoutSession.ProgressIndex == counter {
		exercise.CurrentSet = 1
		exercise.Exercise, userSession.WorkoutSession.WorkoutMeasurements, err = getCurrentMeasurement(
			ctx,
			userSession.WorkoutSession.WorkoutMeasurements,
			exercise.Exercise,
			userSession.UserId,
			choosenExercises,
		)
		return exercise, nil
	}

	for i := 0; i < NumberOfSetsInSuperSet; i++ {
		for j, r := range userSession.WorkoutSession.RandomizedIndices.MainMuscleGroups {
			counter++
			exercise.Exercise, err = getSelectedExercise(
				userSession.WorkoutSession.CurrentOffsets.MainExercises[j],
				userSession.WorkoutSession.RandomizedIndices.MainExercises[r],
				workoutExercises.MainExercises[r],
			)
			if err != nil {
				return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() during main workout mode. Error: %v", err)
			}
			if userSession.WorkoutSession.ProgressIndex == counter {
				exercise.CurrentSet = i + 1
				exercise.Exercise, userSession.WorkoutSession.WorkoutMeasurements, err = getCurrentMeasurement(
					ctx,
					userSession.WorkoutSession.WorkoutMeasurements,
					exercise.Exercise,
					userSession.UserId,
					choosenExercises,
				)
				return exercise, nil
			}
		}
	}

	for _, r := range userSession.WorkoutSession.RandomizedIndices.CoolDownMuscleGroups {
		counter++
		exercise.Exercise, err = getSelectedExercise(
			0, // stretches don't require selection
			userSession.WorkoutSession.RandomizedIndices.CoolDownExercises[r],
			workoutExercises.CoolDownExercises[r],
		)
		if err != nil {
			return ExerciseDisplay{}, fmt.Errorf("error, when getSelectedExercise() for getNextExercise() during stretching workout mode. Error: %v", err)
		}
		if userSession.WorkoutSession.ProgressIndex == counter {
			exercise.CurrentSet = 1
			exercise.Exercise, userSession.WorkoutSession.WorkoutMeasurements, err = getCurrentMeasurement(
				ctx,
				userSession.WorkoutSession.WorkoutMeasurements,
				exercise.Exercise,
				userSession.UserId,
				choosenExercises,
			)
			return exercise, nil
		}
	}

	exercise.WorkoutCompleted = true
	return exercise, nil
}

func getCurrentMeasurement(
	ctx context.Context,
	workoutMeasurements ChoosenExercisesMap,
	exercise Exercise,
	userId int64,
	choosenExercises ChoosenExercisesMap,
) (e Exercise, wm ChoosenExercisesMap, err error) {
	var ok bool
	_, ok = workoutMeasurements[exercise.Id]
	if !ok {
		workoutMeasurements, err = fetchExerciseMeasurements(
			ctx,
			ConnectionPool,
			userId,
			choosenExercises,
		)
	}
	exercise.LastCompletedMeasurement, ok = workoutMeasurements[exercise.Id]
	if !ok {
		return Exercise{}, nil, fmt.Errorf("error, when fetchExerciseMeasurements() for getCurrentMeasurement(). Expected measurement for exerciseId %d but did not get", exercise.Id)
	}
	exercise.LastCompletedMeasurement, err = getDefaultCompletedMeasurement(exercise)
	if err != nil {
		return Exercise{}, nil, fmt.Errorf("error, when getDefaultCompletedMeasurement() for getCurrentMeasurement(). Error: %v", err)
	}
	workoutMeasurements[exercise.Id] = exercise.LastCompletedMeasurement
	switch exercise.MeasurementType {
	case MeasurementTypePounds:
		exercise.MeasurementOptions = DefaultExerciseWeightOptions
	default:
		exercise.MeasurementOptions = DefaultExerciseTimeOptions
	}
	return exercise, workoutMeasurements, nil
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

// todo a session expiration will cause the workout to get restarted, need to fix this
func redirectExercisePage(
	w http.ResponseWriter,
	r *http.Request,
	userSession *UserSession,
) {
	var err error
	if userSession == nil {
		userSession, err = FetchUserSession(r)
		if err != nil {
			err = fmt.Errorf("error, when FetchUserSession() for HandleExercisePage(). Error: %v", err)
			HandleUnexpectedError(w, err)
			return
		}
	}
	var progressIndex int
	if userSession.WorkoutSessionExists {
		progressIndex = userSession.WorkoutSession.ProgressIndex
	}

	url := fmt.Sprintf("%s?progressIndex=%d&pageFetchedAt=%d", EndpointExercise, progressIndex, time.Now().Unix())
	w.Header().Set("HX-Redirect", url)
}

func alreadyAuthRedirect(
	w http.ResponseWriter,
	r *http.Request,
) {
	http.Redirect(w, r, EndpointAlreadyAuthenticated, http.StatusSeeOther)
}
