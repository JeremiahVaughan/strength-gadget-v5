import styles from './measurement-seconds.module.scss';
import TimeDisplay from "../time-display/time-display";
import {useEffect, useState} from "react";

export interface MeasurementSecondsProps {
    currentMeasurement: number
}

export function MeasurementSeconds({currentMeasurement}: MeasurementSecondsProps) {
    const [currentCountDownInSeconds, setCurrentCountDownInSeconds] = useState(currentMeasurement);
    const [countDownRunning, setCountDownRunning] = useState(false);

    const timerPressed = () => {
        if (countDownRunning) {
            setCurrentCountDownInSeconds(currentMeasurement)
            if (currentCountDownInSeconds !== 0) {
                setCountDownRunning(false);
            }
        } else {
            setCountDownRunning(true)
        }
    };

    useEffect(() => {
        setCurrentCountDownInSeconds(currentMeasurement);
        setCountDownRunning(false)
    }, [currentMeasurement]);

    useEffect(() => {
        let timer: NodeJS.Timeout | undefined;
        if (countDownRunning && currentCountDownInSeconds > 0) {
            timer = setInterval(() => {
                setCurrentCountDownInSeconds((prevSeconds) => prevSeconds - 1);
            }, 1000);
        } else {
            if (!countDownRunning && timer) {
                clearInterval(timer);
            }
        }

        return () => {
            clearInterval(timer);
        };
    }, [countDownRunning, currentCountDownInSeconds]);

    return (
        <div className={styles['container']}>
            <TimeDisplay onClick={() => timerPressed()}
                         countDownRunning={countDownRunning}
                         currentCountDownInSeconds={currentCountDownInSeconds}/>
        </div>
    );
}

export default MeasurementSeconds;
