import {Exercise} from "./exercise";

export class Workout {
    progressIndex: number[] = []
    workoutId = ""
    warmupExercises: Exercise[] = []
    mainExercises: Exercise[] = []
    coolDownExercises: Exercise[] = []
}
