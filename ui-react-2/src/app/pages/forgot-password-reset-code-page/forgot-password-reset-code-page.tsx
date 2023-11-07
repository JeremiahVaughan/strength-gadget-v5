import styles from './forgot-password-reset-code-page.module.scss';
import {getAxiosInstance, isEmailValid, useQuery} from "../../utils";
import TextBox from "../../components/text-box/text-box";
import Button from "../../components/button/button";
import ErrorNotification from "../../components/error-notification/error-notification";
import {useContext, useEffect, useState} from 'react';
import BaseUrlContext from "../../context/base-url";
import axios, {AxiosError} from "axios";
import {useNavigate} from "react-router-dom";
import {exercise, forgotPasswordEmail, forgotPasswordNewPassword} from "../../constants/nav";
import FormHeader from "../../form-header/form-header";


export function ForgotPasswordResetCodePage() {
    const email = useQuery().get("email");
    const [passwordResetCode, setPasswordResetCode] = useState<string>('')
    const [submitButtonClicked, setSubmitButtonClicked] = useState(false)
    const [submitErrorMessage, setSubmitErrorMessage] = useState('')
    const navigate = useNavigate();
    const baseUrl = useContext(BaseUrlContext)
    const [loadingResendCode, setLoadingResendCode] = useState(false)
    const [loadingSubmit, setLoadingSubmit] = useState(false)

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
    const onResendCode = () => {
        if (!email || isEmailValid(email).length !== 0) {
            return
        }

        setLoadingResendCode(true)
        getAxiosInstance().post(baseUrl + forgotPasswordEmail, {
                email,
            },
        ).then(() => {
                setLoadingResendCode(false)
                setSubmitErrorMessage('')
                setPasswordResetCode('')
            }
        ).catch(
            (e: AxiosError) => {
                setLoadingResendCode(false)
                const responseMessage = e?.response?.data as string;
                setSubmitErrorMessage(responseMessage)
            }
        )
    }

    const onSubmit = () => {
        setSubmitButtonClicked(true)
        if ('' === passwordResetCode) {
            return
        }
        setLoadingSubmit(true)
        axios.post(baseUrl + "/forgotPassword/resetCode", {
                email,
                resetCode: passwordResetCode
            },
        ).then(
            () => {
                setLoadingSubmit(false)
                navigate(forgotPasswordNewPassword + "?email=" + email + "&resetCode=" + passwordResetCode)
            },
        ).catch(
            (e: AxiosError) => {
                setLoadingSubmit(false)
                setSubmitErrorMessage(e.response?.data as string)
            }
        )
    }


    const validationMessages = [];
    if (('' !== passwordResetCode || submitButtonClicked) && '' === passwordResetCode) {
        validationMessages.push("must provide a password reset code")
    }


    return (
        <div className={styles['container']}>
            <div className={styles['content']}>
                <FormHeader>Reset Password</FormHeader>
                <p className={styles['email-sent-message']}>If your email <b>{email}</b> exists in our system, you will
                    receive a reset code shortly.</p>
                <TextBox aria-label='Password Reset Code'
                         placeholder='password reset code'
                         value={passwordResetCode}
                         validationErrorMessages={validationMessages}
                         onChange={(event) => setPasswordResetCode(event.target.value)}/>
                <div className={styles['spacer']}></div>
                <div className={styles['buttons']}>
                    <Button triggerFromEnter={true} loading={loadingResendCode} color="secondary"
                            onClick={onResendCode}>Resend Code</Button>
                    <div className={styles['spacer']}></div>
                    <Button triggerFromEnter={true} loading={loadingSubmit} onClick={onSubmit}>Submit</Button>
                </div>
                <div className={styles['spacer']}></div>
                <ErrorNotification message={submitErrorMessage}/>
            </div>
        </div>
    );
}

export default ForgotPasswordResetCodePage;
