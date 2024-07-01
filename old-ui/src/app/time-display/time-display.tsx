import styles from './time-display.module.scss';

export interface TimeDisplayProps {
    countDownRunning: boolean;
    currentCountDownInSeconds: number;
    onClick: () => void;
}

export function TimeDisplay({countDownRunning, currentCountDownInSeconds, onClick}: TimeDisplayProps) {
    const minutes = Math.floor(currentCountDownInSeconds / 60);
    const remainingSeconds = currentCountDownInSeconds % 60;
    let style = "";
    if (countDownRunning) {
        style = `${styles['timer-running']}`
    }

    if (currentCountDownInSeconds === 0) {
        style = `${styles['timer-finished']}`
    }

    const formatTime = (time: number): string => {
        return time < 10 ? `0${time}` : `${time}`;
    };

    return (
        <div onClick={onClick}>
            <h1 className={style}>
                {formatTime(minutes)}:{formatTime(remainingSeconds)}
            </h1>
        </div>
    );
}

export default TimeDisplay;
