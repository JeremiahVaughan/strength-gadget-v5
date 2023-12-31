package model

type UserWorkoutDto struct {
	// ProgressIndex outer slice represents workout stage (e.g, warmup, main, coolDown). Inner slice represents exercise in the stage (e.g., barbell curls is exercise two in the main stage)
	ProgressIndex     [][]int            `json:"progressIndex"`
	WarmupExercises   []WarmupExercise   `json:"warmupExercises"`
	MainExercises     []MainExercise     `json:"mainExercises"`
	CoolDownExercises []CoolDownExercise `json:"coolDownExercises"`
}

func (u *UserWorkoutDto) Fill(userWorkout UserWorkout, dailyWorkout DailyWorkout, numberOfSetsPerSuperset, numberOfExercisesPerSuperset int) {
	u.ProgressIndex = userWorkout.ProgressIndex

	for _, exerciseIndex := range userWorkout.SlottedWarmupExercises {
		warmupExercise := WarmupExercise{
			dailyWorkout.CardioExercises[exerciseIndex],
		}
		u.WarmupExercises = append(u.WarmupExercises, warmupExercise)
	}

	numberOfMuscleGroupTargetMainExercises := len(dailyWorkout.MuscleCoverageMainExercises)
	numberOfMainExercises := len(userWorkout.SlottedMainExercises)
	totalSuperSets := numberOfMainExercises / numberOfExercisesPerSuperset
	for i := 0; i < totalSuperSets; i++ {
		superSetStepOffset := i * numberOfExercisesPerSuperset
		for j := 0; j < numberOfExercisesPerSuperset; j++ {
			var mainExercise MainExercise
			var exercise Exercise
			exerciseSlotIndex := superSetStepOffset + j
			exerciseIndex := userWorkout.SlottedMainExercises[exerciseSlotIndex]
			if exerciseSlotIndex < numberOfMuscleGroupTargetMainExercises {
				exercise = dailyWorkout.MuscleCoverageMainExercises[exerciseSlotIndex][exerciseIndex]
			} else {
				exercise = dailyWorkout.AllMainExercises[exerciseIndex]
			}
			mainExercise = MainExercise{
				SourceExerciseIndex: exerciseIndex,
				Exercise:            &exercise,
			}
			u.MainExercises = append(u.MainExercises, mainExercise)
		}
		// subtracting one to account for the source set
		for j := 0; j < numberOfSetsPerSuperset-1; j++ {
			for k := 0; k < numberOfExercisesPerSuperset; k++ {
				var mainExercise MainExercise
				exerciseIndex := userWorkout.SlottedMainExercises[superSetStepOffset+k]
				mainExercise = MainExercise{
					SourceExerciseIndex: exerciseIndex,
				}
				u.MainExercises = append(u.MainExercises, mainExercise)
			}
		}
	}

	for i, exercises := range dailyWorkout.CoolDownExercises {
		exerciseIndex := userWorkout.SlottedCoolDownExercises[i]
		coolDownExercise := CoolDownExercise{
			exercises[exerciseIndex],
		}
		u.CoolDownExercises = append(u.CoolDownExercises, coolDownExercise)
	}
}
