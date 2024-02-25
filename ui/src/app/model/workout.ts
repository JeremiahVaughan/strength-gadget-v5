import {Exercise} from "./exercise";

export class Workout {
    progressIndex: number[] = []
    warmupExercises: Exercise[] = []
    mainExercises: Exercise[] = []
    coolDownExercises: Exercise[] = []
}