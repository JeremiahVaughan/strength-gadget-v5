import styles from './icon-cog-wheel-triple.module.scss';
import IconCogWheel from "../icon-cog-wheel/icon-cog-wheel";

export function IconCogWheelTriple() {
  return (
    <div className={styles['container']}>
        <div className={`${styles['bottom-wheels']} ${styles['left-wheel']}`}>
            <IconCogWheel cogwheelStyle={'left'}/>
        </div>
        <div className={styles['top-wheel']}>
            <IconCogWheel cogwheelStyle={'top'}/>
        </div>
        <div className={`${styles['bottom-wheels']} ${styles['right-wheel']}`}>
            <IconCogWheel cogwheelStyle={'right'}/>
        </div>
    </div>
  );
}

export default IconCogWheelTriple;
