import styles from './exercise-display.module.scss';
import CoolButton from "../cool-button/cool-button";
import ExerciseMeasurement from "../exercise-measurement/exercise-measurement";
import HotButton from "../hot-button/hot-button";
import {Exercise} from "../model/exercise";

/* eslint-disable-next-line */
export interface ExerciseProps {
    data: Exercise
    handleIframeLoad: () => void
    onEasier: () => void
    onHarder: () => void
    selectionMode: boolean
}

export function ExerciseDisplay({
                                    data,
                                    handleIframeLoad,
                                    onEasier,
                                    onHarder,
                                    selectionMode = false,
                                }: ExerciseProps) {
    return (
        <div className={styles['gif-container']}>
            <div className={styles['gif']}>
                <div style={{
                    top: 0,
                    left: 0,
                    width: '100vw',
                    height: '55vh',
                    position: "absolute"
                }}></div>
                <iframe src={`https://giphy.com/embed/${data.demonstrationGiphyId}`}
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
                <a href={`https://giphy.com/gifs/${data.demonstrationGiphyId}`}
                   style={{color: "white"}}>
                    via GIPHY
                </a>
            </p>
            {!selectionMode && <div className={styles['buttons']}>
                <CoolButton onClick={onEasier}>-</CoolButton>
                <ExerciseMeasurement exercise={data}/>
                <div>
                    <HotButton onClick={onHarder}>+</HotButton>
                    <div className={styles['button-spacer']}/>
                </div>
            </div>}
        </div>
    );
}

export default ExerciseDisplay;
