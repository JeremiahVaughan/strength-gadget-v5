import {Exercise} from "./exercise";

export class Workout {
    progressIndex: number[] = []
    workoutId: string = ""
    warmupExercises: Exercise[] = []
    mainExercises: Exercise[] = []
    coolDownExercises: Exercise[] = []
}