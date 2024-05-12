import styles from './exercise-measurement.module.scss';
import {Exercise} from "../model/exercise";
import {
    measurementTypeMile,
    measurementTypePounds,
    measurementTypeRepetition,
    measurementTypeSecond
} from "../model/meaurement_type";
import MeasurementPounds from "../measurement-pounds/measurement-pounds";
import MeasurementReps from "../measurement-reps/measurement-reps";
import MeasurementSeconds from "../measurement-seconds/measurement-seconds";
import MeasurementMiles from "../measurement-miles/measurement-miles";

export interface ExerciseMeasurementProps {
    exercise: Exercise
}

export function ExerciseMeasurement({exercise}: ExerciseMeasurementProps) {
    let displayMeasurement;
    switch (exercise.measurementType) {
        case measurementTypePounds:
            displayMeasurement = <MeasurementPounds
                currentMeasurement={exercise.lastCompletedMeasurement !== 0 ? exercise.lastCompletedMeasurement : 5}/>
            break;
        case measurementTypeRepetition:
            displayMeasurement = <MeasurementReps
                currentMeasurement={exercise.lastCompletedMeasurement !== 0 ? exercise.lastCompletedMeasurement : 3}/>
            break;
        case measurementTypeSecond:
            displayMeasurement = <MeasurementSeconds
                currentMeasurement={exercise.lastCompletedMeasurement !== 0 ? exercise.lastCompletedMeasurement : 10}/>
            break;
        case measurementTypeMile:
            displayMeasurement = <MeasurementMiles
                currentMeasurement={exercise.lastCompletedMeasurement !== 0 ? exercise.lastCompletedMeasurement : 1}/>
            break;
        default:
            console.error("unexpected measurement type: ", exercise.measurementType)
            break;
    }

    return (
        <div className={`${styles['container']} no-select`}>
            {displayMeasurement}
        </div>
    );
}

export default ExerciseMeasurement;
