import styles from './error-notification.module.scss';

export interface ErrorNotificationProps {
    message: string
}

export function ErrorNotification({message}: ErrorNotificationProps) {
  return (
      <div className={styles['input-validation']}>{message}</div>
  );
}

export default ErrorNotification;
