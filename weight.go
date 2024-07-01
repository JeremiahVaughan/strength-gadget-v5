package main

import "math"

// Brzycki formula
func calculateOneRM(weight, reps float64) float64 {
	return weight * (36 / (37 - reps))
}

func calculateWeightForReps(oneRM float64, reps int) float64 {
	weight := oneRM * (float64(37-reps) / 36)

	return weight
}

func roundWeight(weight float64, weightIncrement int) int {
	weightRounded := int(math.Round(weight))
	remainder := weightRounded % weightIncrement
	baseWeight := weightRounded - remainder
	if (float64(weightIncrement) / 2) > float64(remainder) {
		return baseWeight
	} else {
		return baseWeight + weightIncrement
	}
}
