import {AxiosError} from 'axios';
import {useEffect, useState} from 'react';
import styles from './exercise-page.module.scss';
import {useNavigate} from "react-router-dom";
import {getAxiosInstance, incrementLocalWorkoutProgressIndex, sendRequestWithRetry} from "../../utils";
import ToolBar from "../../components/tool-bar/tool-bar";
import {Exercise} from "../../model/exercise";
import LoadingMessage from "../../loading-message/loading-message";
import {
    measurementTypeMile,
    measurementTypePounds,
    measurementTypeRepetition,
    measurementTypeSecond
} from "../../model/meaurement_type";
import {ExerciseDisplay} from "../../exercise/exercise-display";
import {WorkoutPhase} from "../../model/workout_phase";
import {Workout} from "../../model/workout";
import Button from "../../components/button/button";
import {ExerciseType} from "../../model/exercise_type";


export function ExercisePage() {
    const [workout, setWorkout] = useState<Workout>(new Workout())
    const [workoutComplete, setWorkoutComplete] = useState<boolean>(false)
    const [loading, setLoading] = useState(false)
    const [asyncLoading, setAsyncLoading] = useState(false)
    const [iframeLoaded, setIframeLoaded] = useState(false)
    const [currentStep, setCurrentStep] = useState<Exercise>(new Exercise())
    const [selectionMode, setSelectionMode] = useState<boolean>(true)

    const handleIframeLoad = () => {
        setIframeLoaded(true);
    };
    const navigate = useNavigate();

    function getCurrentWorkout() {
        getAxiosInstance().get("/getCurrentWorkout")
            .then(
                (response) => {
                    setLoading(false)
                    if (response.data) {
                        const workout: Workout = response.data
                        setWorkout(workout)
                        updateDisplayExercise(workout, workout.progressIndex)
                    }
                },
            ).catch(
            (e: AxiosError) => {
                setLoading(false)
                if (e?.response?.status === 401) {
                    navigate(
                        "/login",
                    )
                }
            }
        );
    }

    useEffect(() => {
        showLoadingMessage();
        getCurrentWorkout();
    }, [])

    function updateLocalLastCompletedMeasurement(workout: Workout, progressIndex: number[], lastCompletedMeasurement: number) {
        const currentWorkoutPhase = progressIndex.length - 1;
        const currentExerciseProgressIndex = progressIndex[currentWorkoutPhase];

        let exercises: Exercise[] = []
        let exercise: Exercise
        switch (currentWorkoutPhase) {
            case WorkoutPhase.WarmUp:
                exercises = workout.warmupExercises;
                break;
            case WorkoutPhase.Main:
                exercises = workout.mainExercises;
                break;
            case WorkoutPhase.CoolDown:
                exercises = workout.coolDownExercises;
                break;
        }

        exercise = exercises[currentExerciseProgressIndex]
        exercise = exercises[exercise.sourceExerciseSlotIndex]
        exercise.lastCompletedMeasurement = lastCompletedMeasurement
        exercises[exercise.sourceExerciseSlotIndex] = exercise

        switch (currentWorkoutPhase) {
            case WorkoutPhase.WarmUp:
                setWorkout(
                    {
                        ...workout,
                        warmupExercises: exercises
                    }
                )
                break;
            case WorkoutPhase.Main:
                setWorkout(
                    {
                        ...workout,
                        mainExercises: exercises
                    }
                )
                break;
            case WorkoutPhase.CoolDown:
                setWorkout(
                    {
                        ...workout,
                        coolDownExercises: exercises
                    }
                )
                break;
        }
    }

    function onCompleteExercise() {
        updateLocalLastCompletedMeasurement(
            workout,
            workout.progressIndex,
            currentStep.lastCompletedMeasurement
        )

        const incrementedWorkoutProgressIndex = incrementLocalWorkoutProgressIndex(workout, workout.progressIndex)
        updateDisplayExercise(workout, incrementedWorkoutProgressIndex)

        setAsyncLoading(true)
        sendRequestWithRetry("/recordIncrementedWorkoutStep", 'PUT', {
            incrementedProgressIndex: incrementedWorkoutProgressIndex,
            exerciseId: currentStep.id,
            lastCompletedMeasurement: currentStep.lastCompletedMeasurement,
            workoutId: workout.workoutId
        })
            .then(() => {
                setAsyncLoading(false);
                setWorkout({...workout, progressIndex: incrementedWorkoutProgressIndex})
                console.log("workout step completed")
            })
            .catch(
                (e: AxiosError) => {
                    setAsyncLoading(false)
                    if (e?.response?.status === 409) {
                        getCurrentWorkout()
                        return
                    }
                    if (e?.response?.status === 401) {
                        navigate(
                            "/login",
                        )
                    }
                }
            )
    }

    function onSwapExercise() {
        showLoadingMessage();
        getAxiosInstance().put("/swapExercise", {
            exerciseId: currentStep.id,
            workoutId: workout.workoutId
        }).then(
            (response) => {
                setLoading(false);
                const workout: Workout = response.data
                setWorkout(workout)
                updateDisplayExercise(workout, workout.progressIndex)
            },
        ).catch(
            (e: AxiosError) => {
                setLoading(false)
                if (e?.response?.status === 409) {
                    getCurrentWorkout()
                    return
                }
                if (e?.response?.status === 401) {
                    navigate(
                        "/login",
                    )
                }
            }
        );
    }

    function showLoadingMessage() {
        setIframeLoaded(false)
        setLoading(true)
    }

    function applyDefaultStartingValues(exercise: Exercise): Exercise {
        if (exercise.lastCompletedMeasurement) {
            return exercise
        }
        let startingValue = 0;
        if (exercise.exerciseType === ExerciseType.CoolDown) {
            startingValue = 30
        } else {
            switch (exercise.measurementType) {
                case measurementTypePounds:
                    startingValue = 5;
                    break;
                case measurementTypeRepetition:
                    startingValue = 3;
                    break;
                case measurementTypeSecond:
                    startingValue = 10;
                    break;
                case measurementTypeMile:
                    startingValue = 1;
                    break;
                default:
                    console.error("unexpected measurement type: ", exercise.measurementType);
                    break;
            }
        }
        return {
            ...exercise,
            lastCompletedMeasurement: startingValue
        }
    }

    function onEasier() {
        const incrementAmount = getIncrementAmount(currentStep.measurementType);
        const newValue = currentStep.lastCompletedMeasurement - incrementAmount;
        setCurrentStep({
            ...currentStep,
            lastCompletedMeasurement: newValue < 1 ?
                currentStep.lastCompletedMeasurement :
                newValue
        })
    }

    function onHarder() {
        setCurrentStep({
            ...currentStep,
            lastCompletedMeasurement: currentStep.lastCompletedMeasurement +
                getIncrementAmount(currentStep.measurementType)
        })
    }

    function getIncrementAmount(measurementType: string): number {
        let result = 0
        switch (measurementType) {
            case measurementTypePounds:
                result = 5
                break;
            case measurementTypeRepetition:
                result = 1
                break;
            case measurementTypeMile:
                result = 1
                break;
            case measurementTypeSecond:
                result = 5
                break;
            default:
                console.error("measurementType not accounted for: ", measurementType)
        }
        return result;
    }


    function updateDisplayExercise(workout: Workout, currentProgressIndex: number[]) {
        const workoutPhase = currentProgressIndex.length - 1
        const exerciseProgressIndex = currentProgressIndex[workoutPhase]
        let e: Exercise | null | undefined;
        let exercises: Exercise[] = []
        switch (workoutPhase) {
            case WorkoutPhase.WarmUp:
                exercises = workout.warmupExercises
                break;
            case WorkoutPhase.Main:
                exercises = workout.mainExercises
                break;
            case WorkoutPhase.CoolDown:
                exercises = workout.coolDownExercises
                break;
            default:
                setWorkoutComplete(true)
                return;
        }
        const exercise = exercises[exerciseProgressIndex]
        if (!exercise.id) {
            setSelectionMode(false)
            e = exercises[exercise.sourceExerciseSlotIndex]
        } else {
            setSelectionMode(true)
            e = exercise
        }
        const exerciseWithDefaultValuesApplied = applyDefaultStartingValues(e as Exercise);
        setCurrentStep(exerciseWithDefaultValuesApplied)
    }


    return (
        <div>
            <ToolBar/>
            <div className={styles['container']}>
                {loading ? <LoadingMessage/> :
                    workoutComplete ? <div className={styles['workout-complete-text']}>
                            <div>Workout Complete</div>
                        </div> :
                        <div>
                            {!iframeLoaded && <LoadingMessage/>}
                            <div style={{display: iframeLoaded ? 'block' : 'none'}}>
                                <ExerciseDisplay
                                    data={currentStep}
                                    onEasier={onEasier}
                                    onHarder={onHarder}
                                    handleIframeLoad={handleIframeLoad}
                                    selectionMode={selectionMode}
                                />
                                {/* pushing these buttons rapidly may cause undesired behavior
                                for the user if the backend becomes out of sync, so I am
                                disabling the buttons until the backend and client become in
                                sync again */}
                                {!asyncLoading && <div className={styles['buttons']}>
                                    {selectionMode ? <Button color="secondary"
                                                             loading={loading}
                                                             onClick={onSwapExercise}>
                                        No
                                    </Button> : <div className={styles['spacer']}/>}
                                    <Button triggerFromEnter={true} loading={loading}
                                            onClick={() => onCompleteExercise()}>
                                        {selectionMode ? "Yes" : "Complete"}
                                    </Button>
                                </div>}
                            </div>
                        </div>
                }
            </div>
        </div>
    );
}

export default ExercisePage;
