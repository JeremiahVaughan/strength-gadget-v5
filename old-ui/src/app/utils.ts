import axios, { AxiosInstance, Method } from "axios";
import { useLocation } from "react-router-dom";
import { Workout } from "./model/workout";
import { WorkoutPhase } from "./model/workout_phase";

export function getBaseApiUrlFromHostname() {
    const hostname = window.location.hostname;
    if (hostname.includes('localhost')) {
        // local
        return `http://localhost/api`;
    } else if (hostname.includes('staging.')) {
        // staging
        return 'https://api.staging.strengthgadget.com/api';
    } else {
        // production
        return 'https://api.strengthgadget.com/api';
    }
}

export function isPasswordValid(password: string): string[] {
    const passwordValidationMessages: string[] = []
    if (password.length < 12) {
        passwordValidationMessages.push("password is too short")
    }

    const testPasswordFormat = /^\d+$/
    if (testPasswordFormat.test(password)) {
        passwordValidationMessages.push("cannot be all numbers")
    }
    return passwordValidationMessages
}

export function isEmailValid(email: string): string[] {
    const emailValidationMessages: string[] = []
    const testEmailFormat = /^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/
    if (!testEmailFormat.test(email)) {
        emailValidationMessages.push("invalid email")
    }
    return emailValidationMessages
}

//  todo either keep this strategy remove it in favor of the existing one
const localAxiosInstance: AxiosInstance = axios.create({
    baseURL: getBaseApiUrlFromHostname(),
    withCredentials: true
});

const axiosInstance: AxiosInstance = axios.create({
    baseURL: getBaseApiUrlFromHostname(),
    withCredentials: true
});

export const getAxiosInstance = (): AxiosInstance => {
    console.log("Application mode: ", process.env.NODE_ENV)
    return process.env.NODE_ENV === 'development' ? localAxiosInstance : axiosInstance
}

export function useQuery() {
    return new URLSearchParams(useLocation().search);
}


export async function sendRequestWithRetry(url: string, method: Method, data: any, retryCount = 3): Promise<any> {
    try {
        const response = await getAxiosInstance().request({ url, method, data });
        return response.data;
    } catch (err: any) {
        if (retryCount <= 0 || (err.response && err.response.status === 409)) {
            throw err;
        }
        await new Promise((resolve) => setTimeout(resolve, 1000)); // Wait for 1 second
        return sendRequestWithRetry(url, method, data, retryCount - 1);
    }
}


export function incrementLocalWorkoutProgressIndex(workout: Workout, oldProgressIndex: number[]): number[] {
    const oldWorkoutPhase = oldProgressIndex.length - 1;
    const oldExerciseProgressIndex = oldProgressIndex[oldWorkoutPhase];

    let exercisesInPhase = 0;
    switch (oldWorkoutPhase) {
        case WorkoutPhase.WarmUp:
            exercisesInPhase = workout.warmupExercises.length;
            break;
        case WorkoutPhase.Main:
            exercisesInPhase = workout.mainExercises.length;
            break;
        case WorkoutPhase.CoolDown:
            exercisesInPhase = workout.coolDownExercises.length;
            break;
    }

    if (oldExerciseProgressIndex + 1 === exercisesInPhase) {
        return [...oldProgressIndex, 0];
    } else {
        oldProgressIndex[oldWorkoutPhase]++
        return oldProgressIndex;
    }
}
