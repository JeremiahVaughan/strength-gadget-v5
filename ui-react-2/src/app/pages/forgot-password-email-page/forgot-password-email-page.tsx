import styles from './forgot-password-email-page.module.scss';
import TextBox from "../../components/text-box/text-box";
import Button from "../../components/button/button";
import ErrorNotification from "../../components/error-notification/error-notification";
import {getAxiosInstance, isEmailValid} from "../../utils";
import {FormEvent, useContext, useEffect, useState} from 'react';
import {exercise, forgotPasswordEmail, forgotPasswordResetCode} from "../../constants/nav";
import BaseUrlContext from "../../context/base-url";
import {useNavigate} from "react-router-dom";
import {AxiosError} from 'axios';
import FormHeader from "../../form-header/form-header";


export function ForgotPasswordEmailPage() {
    const [submitErrorMessage, setSubmitErrorMessage] = useState('')
    const [email, setEmail] = useState('')
    const [submitButtonClicked, setSubmitButtonClicked] = useState(false)
    const [loading, setLoading] = useState(false)
    const baseUrl = useContext(BaseUrlContext)
    const navigate = useNavigate();

    useEffect(() => {
        getAxiosInstance().get("/isLoggedIn")
            .then(
                () => {
                    navigate(
                        exercise
                    )
                },
            )

    }, [navigate])

    let emailValidationMessages: string[] = []
    if (('' !== email || submitButtonClicked)) {
        emailValidationMessages = isEmailValid(email)
    }

    const onSubmit = () => {
        setSubmitButtonClicked(true)

        if (isEmailValid(email).length !== 0) {
            return
        }

        setLoading(true)
        getAxiosInstance().post(baseUrl + forgotPasswordEmail, {
                email,
            },
        ).then(
            () => {
                setLoading(false)
                navigate(forgotPasswordResetCode + "?email=" + email)
            },
        ).catch(
            (e: AxiosError) => {
                setLoading(false)
                const responseMessage = e?.response?.data as string;
                setSubmitErrorMessage(responseMessage)
            }
        )
    }

    return (
        <div className={styles['container']}>
            <div className={styles['content']}>
                <FormHeader>Forgot Password</FormHeader>
                <TextBox aria-label='username'
                         validationErrorMessages={emailValidationMessages}
                         value={email}
                         onChange={(event: FormEvent<HTMLInputElement>) => setEmail(event.currentTarget.value)}
                         placeholder='email'/>
                <div className={styles['buttons']}>
                    <Button triggerFromEnter={true} loading={loading} onClick={onSubmit}>Submit</Button>
                </div>
                <div className={styles['spacer']}></div>
                <ErrorNotification message={submitErrorMessage}/>
            </div>
        </div>
    );
}

export default ForgotPasswordEmailPage;
