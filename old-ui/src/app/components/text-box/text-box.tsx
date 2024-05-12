import {ChangeEvent, HTMLAttributes, ReactNode} from "react";
import styles from './text-box.module.scss';
import ErrorNotification from "../error-notification/error-notification";


interface Props extends HTMLAttributes<HTMLInputElement> {
    validationErrorMessages?: string[],
    value: string,
    type?: string,
    children?: ReactNode,
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
}


export function TextBox({
                            validationErrorMessages,
                            value,
                            type,
                            children,
                            onChange,
                            ...rest
                        }: Props) {


    const validationFeedbackContent = validationErrorMessages?.map(message => {
        return <ErrorNotification key={message} message={message}/>
    })

    let style = `${styles['field']}`

    if (validationErrorMessages && validationErrorMessages.length > 0) {
        style += ` ${styles['failed-validation']}`
    }

    return (
        <div>
            <div>
                <input
                    {...rest}
                    value={value}
                    type={type}
                    className={style}
                    onChange={onChange}
                />
            </div>
            <div>
                {validationFeedbackContent}
            </div>
            <div className={styles['spacer']}></div>
        </div>
    );
}


export default TextBox;
