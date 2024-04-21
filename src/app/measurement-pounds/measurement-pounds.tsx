import styles from './measurement-pounds.module.scss';

export interface MeasurementPoundsProps {
    currentMeasurement: number
}

export function MeasurementPounds({currentMeasurement}: MeasurementPoundsProps) {
  return (
    <div className={styles['container']}>
       <b>{currentMeasurement}</b> lbs - <b>15</b> {currentMeasurement === 1 ? "rep" : "reps"}
    </div>
  );
}

export default MeasurementPounds;
