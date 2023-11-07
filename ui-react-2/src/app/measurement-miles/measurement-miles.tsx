import styles from './measurement-miles.module.scss';

/* eslint-disable-next-line */
export interface MeasurementMilesProps {
    currentMeasurement: number
}

export function MeasurementMiles({currentMeasurement}: MeasurementMilesProps) {
  return (
    <div className={styles['container']}>
      <b>{currentMeasurement}</b> {currentMeasurement === 1 ? "mile" : "miles"}
    </div>
  );
}

export default MeasurementMiles;
