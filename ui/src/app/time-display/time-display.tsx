import styles from './time-display.module.scss';

/* eslint-disable-next-line */
export interface TimeDisplayProps {
    seconds: number;
    started: boolean;
    onClick: () => void;
}

export function TimeDisplay({started, seconds, onClick}: TimeDisplayProps) {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    let style = "";
    if (started) {
        style = `${styles['timer-running']}`
    }

    if (seconds === 0) {
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
