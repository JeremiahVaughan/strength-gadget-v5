import {incrementLocalWorkoutProgressIndex} from '../app/utils'
import {Workout} from "../app/model/workout";
import {expect} from "vitest";

describe('incrementLocalWorkoutProgressIndex function', () => {
    it('incrementing localWorkoutProgressIndex', () => {
        const workout = new Workout()
        workout.warmupExercises = [
            {
                exercise: {
                    id: "1",
                    name: "Jumping Jacks",
                    demonstrationGiphyId: "giphy1",
                    lastCompletedMeasurement: 10,
                    measurementType: "repetitions"
                }
            },
            {
                exercise: {
                    id: "2",
                    name: "High Knees",
                    demonstrationGiphyId: "giphy2",
                    lastCompletedMeasurement: 15,
                    measurementType: "repetitions"
                }
            },
            {
                exercise: {
                    id: "3",
                    name: "Arm Circles",
                    demonstrationGiphyId: "giphy3",
                    lastCompletedMeasurement: 20,
                    measurementType: "seconds"
                }
            }
        ]
        workout.mainExercises = [
            {
                sourceExerciseSlotIndex: 0,
                exercise: {
                    id: "4",
                    name: "Barbell Squat",
                    demonstrationGiphyId: "giphy4",
                    lastCompletedMeasurement: 5,
                    measurementType: "repetitions"
                }
            },
            {
                sourceExerciseSlotIndex: 1,
                exercise: {
                    id: "5",
                    name: "Bench Press",
                    demonstrationGiphyId: "giphy5",
                    lastCompletedMeasurement: 8,
                    measurementType: "repetitions"
                }
            },
            {
                sourceExerciseSlotIndex: 2,
                exercise: {
                    id: "6",
                    name: "Deadlift",
                    demonstrationGiphyId: "giphy6",
                    lastCompletedMeasurement: 6,
                    measurementType: "repetitions"
                }
            },
            {
                sourceExerciseSlotIndex: 0,
                exercise: null,
            },
            {
                sourceExerciseSlotIndex: 1,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 2,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 0,
                exercise: null,
            },
            {
                sourceExerciseSlotIndex: 1,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 2,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 0,
                exercise: null,
            },
            {
                sourceExerciseSlotIndex: 1,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 2,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 12,
                exercise: {
                    id: "7",
                    name: "Overhead Press",
                    demonstrationGiphyId: "giphy7",
                    lastCompletedMeasurement: 10,
                    measurementType: "repetitions"
                }
            },
            {
                sourceExerciseSlotIndex: 13,
                exercise: {
                    id: "8",
                    name: "Barbell Row",
                    demonstrationGiphyId: "giphy8",
                    lastCompletedMeasurement: 10,
                    measurementType: "repetitions"
                }
            },
            {
                sourceExerciseSlotIndex: 14,
                exercise: {
                    id: "9",
                    name: "Pull Up",
                    demonstrationGiphyId: "giphy9",
                    lastCompletedMeasurement: 10,
                    measurementType: "repetitions"
                }
            },
            {
                sourceExerciseSlotIndex: 12,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 13,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 14,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 12,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 13,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 14,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 12,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 13,
                exercise: null
            },
            {
                sourceExerciseSlotIndex: 14,
                exercise: null
            }
        ]
        workout.coolDownExercises = [
            {
                exercise: {
                    id: "10",
                    name: "Hamstring Stretch",
                    demonstrationGiphyId: "giphy10",
                    lastCompletedMeasurement: 30,
                    measurementType: "seconds"
                }
            },
            {
                exercise: {
                    id: "11",
                    name: "Quad Stretch",
                    demonstrationGiphyId: "giphy11",
                    lastCompletedMeasurement: 30,
                    measurementType: "seconds"
                }
            },
            {
                exercise: {
                    id: "12",
                    name: "Arm Stretch",
                    demonstrationGiphyId: "giphy12",
                    lastCompletedMeasurement: 20,
                    measurementType: "seconds"
                }
            }
        ]

        let oldProgressIndex = [0]
        let newProgressIndex = incrementLocalWorkoutProgressIndex(workout, oldProgressIndex)
        expect(newProgressIndex).toEqual([1])
        newProgressIndex = incrementLocalWorkoutProgressIndex(workout, newProgressIndex)
        expect(newProgressIndex).toEqual([2])

        for (let i = 0; i < 24; i++) {
            newProgressIndex = incrementLocalWorkoutProgressIndex(workout, newProgressIndex)
            expect(newProgressIndex).toEqual([2, i])
        }

        for (let i = 0; i < 3; i++) {
            newProgressIndex = incrementLocalWorkoutProgressIndex(workout, newProgressIndex)
            expect(newProgressIndex).toEqual([2, 23, i])
        }

        newProgressIndex = incrementLocalWorkoutProgressIndex(workout, newProgressIndex)
        expect(newProgressIndex).toEqual([2, 23, 2, 0])
    })
})