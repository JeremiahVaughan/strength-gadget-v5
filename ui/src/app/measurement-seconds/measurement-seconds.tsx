import styles from './measurement-seconds.module.scss';
import TimeDisplay from "../time-display/time-display";
import {useEffect, useState} from "react";

/* eslint-disable-next-line */
export interface MeasurementSecondsProps {
    currentMeasurement: number
}

export function MeasurementSeconds({currentMeasurement}: MeasurementSecondsProps) {
    const [seconds, setSeconds] = useState(currentMeasurement);
    const [countingDown, setCountingDown] = useState(false);
    const [countDownStarted, setCountDownStarted] = useState(false);

    const startCountDown = () => {
        setCountingDown(true);
        setCountDownStarted(true);
    };

    useEffect(() => {
        setSeconds(currentMeasurement);
    }, [currentMeasurement]);

    useEffect(() => {
        let timer: NodeJS.Timeout;
        if (countingDown && seconds > 0) {
            timer = setInterval(() => {
                setSeconds((prevSeconds) => prevSeconds - 1);
            }, 1000);
        } else { // @ts-ignore
            if (!countingDown && timer) {
                clearInterval(timer);
            }
        }

        return () => {
            clearInterval(timer);
        };
    }, [countingDown, seconds]);

    return (
        <div className={styles['container']}>
            <TimeDisplay onClick={() => startCountDown()} seconds={seconds} started={countDownStarted}/>
        </div>
    );
}

export default MeasurementSeconds;
