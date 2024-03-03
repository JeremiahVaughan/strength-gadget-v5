package model

import "time"

type UserWorkoutDto struct {
	ProgressIndex     WorkoutProgressIndex `json:"progressIndex"`
	Weekday           time.Weekday         `json:"weekday"`
	WorkoutId         string               `json:"workoutId"`
	WarmupExercises   []Exercise           `json:"warmupExercises"`
	MainExercises     []Exercise           `json:"mainExercises"`
	CoolDownExercises []Exercise           `json:"coolDownExercises"`
}

func (u *UserWorkoutDto) Fill(
	userWorkout UserWorkout,
	dailyWorkout DailyWorkout,
	numberOfSetsPerSuperset, numberOfExercisesPerSuperset int,
) {
	u.ProgressIndex = userWorkout.ProgressIndex
	u.Weekday = userWorkout.Weekday
	u.WorkoutId = userWorkout.WorkoutId

	currentExerciseSlotReference := 0 // isn't referenced by anything, this is just helpful for debugging
	for _, exerciseIndex := range userWorkout.SlottedWarmupExercises {
		// selection slot
		warmupExercise := dailyWorkout.CardioExercises[exerciseIndex]
		warmupExercise.CurrentExerciseSlotIndex = currentExerciseSlotReference
		warmupExercise.SourceExerciseSlotIndex = currentExerciseSlotReference
		u.WarmupExercises = append(u.WarmupExercises, warmupExercise)
		currentExerciseSlotReference++

		// work slot
		workingWarmupExercise := Exercise{
			CurrentExerciseSlotIndex: currentExerciseSlotReference,
			SourceExerciseSlotIndex:  currentExerciseSlotReference - 1,
		}
		u.WarmupExercises = append(u.WarmupExercises, workingWarmupExercise)
		currentExerciseSlotReference++
	}

	numberOfMuscleGroupTargetMainExercises := len(dailyWorkout.MuscleCoverageMainExercises)
	numberOfMainExercises := len(userWorkout.SlottedMainExercises)
	totalSuperSets := numberOfMainExercises / numberOfExercisesPerSuperset
	currentExerciseSlotReference = 0
	for i := 0; i < totalSuperSets; i++ {
		superSetSlottedExercisesOffset := i * numberOfExercisesPerSuperset
		// main exercise selection
		for j := 0; j < numberOfExercisesPerSuperset; j++ {
			var exercise Exercise
			exerciseSlotIndex := superSetSlottedExercisesOffset + j
			exerciseIndex := userWorkout.SlottedMainExercises[exerciseSlotIndex]
			if exerciseSlotIndex < numberOfMuscleGroupTargetMainExercises {
				exercise = dailyWorkout.MuscleCoverageMainExercises[exerciseSlotIndex][exerciseIndex]
			} else {
				exercise = dailyWorkout.AllMainExercises[exerciseIndex]
			}
			exercise.LastCompletedMeasurement = userWorkout.UserExerciseDataMap[exercise.Id].Measurement
			exercise.SourceExerciseSlotIndex = currentExerciseSlotReference
			exercise.CurrentExerciseSlotIndex = currentExerciseSlotReference
			u.MainExercises = append(u.MainExercises, exercise)
			currentExerciseSlotReference++
		}
		// conduct main exercises
		for m := 0; m < numberOfSetsPerSuperset; m++ {
			for k := 0; k < numberOfExercisesPerSuperset; k++ {
				mainExerciseSlotOffset := i * ((numberOfSetsPerSuperset + 1) * numberOfExercisesPerSuperset)
				userWorkoutDtoSlottedExercisesOffset := mainExerciseSlotOffset + k
				mainExercise := Exercise{
					CurrentExerciseSlotIndex: currentExerciseSlotReference,
					SourceExerciseSlotIndex:  userWorkoutDtoSlottedExercisesOffset,
				}
				u.MainExercises = append(u.MainExercises, mainExercise)
				currentExerciseSlotReference++
			}
		}
	}

	currentExerciseSlotReference = 0
	for i, exercises := range dailyWorkout.CoolDownExercises {
		// selection slot
		exerciseIndex := userWorkout.SlottedCoolDownExercises[i]
		coolDownExercise := exercises[exerciseIndex]
		coolDownExercise.CurrentExerciseSlotIndex = currentExerciseSlotReference
		coolDownExercise.SourceExerciseSlotIndex = currentExerciseSlotReference
		u.CoolDownExercises = append(u.CoolDownExercises, coolDownExercise)
		currentExerciseSlotReference++

		// work slot
		workingCoolDownExercise := Exercise{
			CurrentExerciseSlotIndex: currentExerciseSlotReference,
			SourceExerciseSlotIndex:  currentExerciseSlotReference - 1,
		}
		u.CoolDownExercises = append(u.CoolDownExercises, workingCoolDownExercise)
		currentExerciseSlotReference++
	}
}
