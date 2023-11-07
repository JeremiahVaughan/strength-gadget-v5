import styles from './loading-message.module.scss';


export function LoadingMessage() {
    return (
        <div className={styles['loading-container']}>
            <div className={styles['spacer']}/>
            <div className={`${styles['loading-text']}`}>... Loading ...</div>
        </div>
    );
}

export default LoadingMessage;
