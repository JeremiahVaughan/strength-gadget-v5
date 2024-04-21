import styles from './measurement-reps.module.scss';

export interface MeasurementRepsProps {
    currentMeasurement: number
}

export function MeasurementReps({currentMeasurement}: MeasurementRepsProps) {
  return (
    <div className={styles['container']}>
      <b>{currentMeasurement}</b> reps
    </div>
  );
}

export default MeasurementReps;
