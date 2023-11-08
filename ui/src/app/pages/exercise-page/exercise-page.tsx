import {AxiosError} from 'axios';
import {useEffect, useState} from 'react';
import styles from './exercise-page.module.scss';
import {useNavigate} from "react-router-dom";
import {getAxiosInstance} from "../../utils";
import ToolBar from "../../components/tool-bar/tool-bar";
import {Exercise} from "../../model/exercise";
import Button from "../../components/button/button";
import HotButton from "../../hot-button/hot-button";
import CoolButton from 'src/app/cool-button/cool-button';
import LoadingMessage from "../../loading-message/loading-message";
import ExerciseMeasurement from "../../exercise-measurement/exercise-measurement";
import {
    measurementTypeMile,
    measurementTypePounds,
    measurementTypeRepetition,
    measurementTypeSecond
} from "../../model/meaurement_type";


export function ExercisePage() {
    const [exercise, setExercise] = useState<Exercise>(new Exercise())
    const [supersetFull, setSupersetFull] = useState<boolean>(false)
    const [workoutComplete, setWorkoutComplete] = useState<boolean>(false)
    const [loading, setLoading] = useState(false)
    const [iframeLoaded, setIframeLoaded] = useState(false);

    const handleIframeLoad = () => {
        setIframeLoaded(true);
    };
    const navigate = useNavigate();

    useEffect(() => {
        showLoadingMessage();
        getAxiosInstance().get("/currentExercise")
            .then(
                (response) => {
                    setLoading(false)
                    if (response.data) {
                        setData(response.data)
                    } else {
                        onShuffleExercise();
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
    }, [])

    function onCompleteExercise() {
        showLoadingMessage();
        getAxiosInstance().get("/readyForNextExercise?measurement=" + exercise.lastCompletedMeasurement)
            .then((response) => {
                setLoading(false)
                setIframeLoaded(false)
                if (response.data) {
                    setData(response.data)
                }
            }).catch(
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

    function onShuffleExercise() {
        showLoadingMessage();
        getAxiosInstance().get("/shuffleExercise")
            .then(
                (response) => {
                    setLoading(false)
                    if (response.data) {
                        setData(response.data)
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

    function showLoadingMessage() {
        setIframeLoaded(false)
        setLoading(true)
    }

    function setData(data: any) {
        setSupersetFull(data.superSetFull)
        setWorkoutComplete(data.workoutComplete)
        let exercise: Exercise = new Exercise()
        if (data.exercise) {
            exercise = data.exercise
        }
        setExercise(applyDefaultStartingValues(exercise))
    }

    function applyDefaultStartingValues(exercise: Exercise): Exercise {
        let startingValue = 0
        switch (exercise.measurementType) {
            case measurementTypePounds:
                startingValue = 5
                break;
            case measurementTypeRepetition:
                startingValue = 3
                break;
            case measurementTypeSecond:
                startingValue = 10
                break;
            case measurementTypeMile:
                startingValue = 1
                break;
            default:
                console.error("unexpected measurement type: ", exercise.measurementType)
                break;
        }
        return {
            ...exercise,
            lastCompletedMeasurement: exercise.lastCompletedMeasurement !== 0 ?
                exercise.lastCompletedMeasurement :
                startingValue
        }
    }

    function onEasier() {
        let incrementAmount = getIncrementAmount(exercise.measurementType);
        let newValue = exercise.lastCompletedMeasurement - incrementAmount;
        setExercise({
            ...exercise,
            lastCompletedMeasurement: newValue < 1 ?
                exercise.lastCompletedMeasurement :
                newValue
        })
    }

    function onHarder() {
        setExercise({
            ...exercise,
            lastCompletedMeasurement: exercise.lastCompletedMeasurement + getIncrementAmount(exercise.measurementType)
        })
    }

    function getIncrementAmount(measurementType: string): number {
        let result: number = 0
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

    return (
        <div>
            <ToolBar/>
            <div key={exercise.id} className={styles['container']}>
                {loading ? <LoadingMessage/> :
                    workoutComplete ? <div className={styles['workout-complete-text']}>
                            <div>Workout Complete</div>
                        </div> :
                        <div>
                            {!iframeLoaded && <LoadingMessage/>}
                            <div className={styles['gif-container']}
                                 style={{display: iframeLoaded ? 'block' : 'none'}}>
                                <div className={styles['gif']}>
                                    <div style={{
                                        top: 0,
                                        left: 0,
                                        width: '100vw',
                                        height: '55vh',
                                        position: "absolute"
                                    }}></div>
                                    <iframe src={`https://giphy.com/embed/${exercise.demonstrationGiphyId}`}
                                            style={{
                                                width: '100vw',
                                                height: '55vh',
                                            }}
                                            frameBorder="0"
                                            className="giphy-embed"
                                            onLoad={handleIframeLoad}
                                            allowFullScreen>
                                    </iframe>
                                </div>
                                <p className={styles['giphy-link']}>
                                    <a href={`https://giphy.com/gifs/${exercise.demonstrationGiphyId}`}
                                       style={{color: "white"}}>
                                        via GIPHY
                                    </a>
                                </p>
                                <div className={styles['buttons-first-row']}>
                                    <CoolButton onClick={onEasier}>-</CoolButton>
                                    <ExerciseMeasurement exercise={exercise}/>
                                    <div>
                                        <HotButton onClick={onHarder}>+</HotButton>
                                        <div className={styles['button-spacer']}/>
                                    </div>

                                </div>
                                <div className={styles['buttons-second-row']}>
                                    {supersetFull ? <div className={styles['spacer']}/> :
                                        <Button color="secondary"
                                                loading={loading}
                                                onClick={onShuffleExercise}>
                                            Shuffle
                                        </Button>}
                                    <Button triggerFromEnter={true} loading={loading}
                                            onClick={() => onCompleteExercise()}>Complete</Button>
                                </div>
                            </div>
                        </div>
                }

            </div>
        </div>
    );
}

export default ExercisePage;
