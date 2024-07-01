import styles from './forgot-password-new-password-page.module.scss';
import {getAxiosInstance, isEmailValid, isPasswordValid, useQuery} from "../../utils";
import ConfirmPasswordControl from "../../components/confirm-password-control/confirm-password-control";
import {useContext, useEffect, useState} from 'react';
import Button from "../../components/button/button";
import axios, {AxiosError} from 'axios';
import {useNavigate} from "react-router-dom";
import BaseUrlContext from "../../context/base-url";
import ErrorNotification from "../../components/error-notification/error-notification";
import {exercise, login} from "../../constants/nav";
import FormHeader from "../../form-header/form-header";


export function ForgotPasswordNewPasswordPage() {
    const email = useQuery().get("email");
    const resetCode = useQuery().get("resetCode");
    const emptyStringStartingState = ''
    const [password, setPassword] = useState(emptyStringStartingState)
    const [confirmPassword, setConfirmPassword] = useState(emptyStringStartingState)
    const [submitErrorMessage, setSubmitErrorMessage] = useState('')

    const [submitButtonClicked, setSubmitButtonClicked] = useState(false)
    const [loading, setLoading] = useState(false)
    const navigate = useNavigate();
    const baseUrl = useContext(BaseUrlContext)

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

    const onSubmit = () => {
        setSubmitButtonClicked(true);
        if (!email || !resetCode || mismatchedPasswordsNotValid || isEmailValid(email).length !== 0 || isPasswordValid(password).length !== 0) {
            return
        }

        setLoading(true)
        axios.post(baseUrl + "/forgotPassword/newPassword", {
                email,
                newPassword: password,
                resetCode
            },
        ).then(() => {
                setLoading(false)
                navigate(
                    login,
                )
            },
        ).catch((e: AxiosError) => {
                setSubmitErrorMessage(e.response?.data as string)
                setLoading(false)
            }
        )
    }

    let passwordValidationMessages: string[] = []
    if ((emptyStringStartingState !== password || submitButtonClicked)) {
        passwordValidationMessages = isPasswordValid(password)
    }


    const passwordsDoNotMatchValidationMessages = []
    const mismatchedPasswordsNotValid = password !== confirmPassword;
    if ((emptyStringStartingState !== confirmPassword || submitButtonClicked) && mismatchedPasswordsNotValid) {
        passwordsDoNotMatchValidationMessages.push("passwords do not match")

    }
    return (
        <div className={styles['container']}>
            <div className={styles['content']}>
                <FormHeader>New Password</FormHeader>
                <ConfirmPasswordControl
                    password={password}
                    setPassword={setPassword}
                    confirmPassword={confirmPassword}
                    setConfirmPassword={setConfirmPassword}
                    passwordValidationMessages={passwordValidationMessages}
                    passwordsDoNotMatchValidationMessages={passwordsDoNotMatchValidationMessages}/>
                <div className={styles['buttons']}>
                    <div className={styles['spacer']}></div>
                    <Button triggerFromEnter={true} loading={loading} onClick={onSubmit}>Submit</Button>
                </div>
                <div className={styles['spacer']}></div>
                <ErrorNotification message={submitErrorMessage}/>
            </div>
        </div>
    );
}

export default ForgotPasswordNewPasswordPage;
