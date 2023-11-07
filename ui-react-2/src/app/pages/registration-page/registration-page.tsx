import styles from './registration-page.module.scss';
import {FormEvent, useContext, useEffect, useState} from "react";
import {Link, useNavigate} from "react-router-dom";
import TextBox from "../../components/text-box/text-box";
import Button from "../../components/button/button";
import axios, {AxiosError} from 'axios';
import ErrorNotification from "../../components/error-notification/error-notification";
import BaseUrlContext from "../../context/base-url";
import {getAxiosInstance, isEmailValid, isPasswordValid} from "../../utils";
import ConfirmPasswordControl from "../../components/confirm-password-control/confirm-password-control";
import {exercise} from "../../constants/nav";
import FormHeader from "../../form-header/form-header";

export function RegistrationPage() {
    const navigate = useNavigate();
    const [submitErrorMessage, setSubmitErrorMessage] = useState('')
    const [loading, setLoading] = useState(false)
    const baseUrl = useContext(BaseUrlContext)

    const emptyStringStartingState = ''
    const [email, setEmail] = useState(emptyStringStartingState)
    const [password, setPassword] = useState(emptyStringStartingState)
    const [confirmPassword, setConfirmPassword] = useState(emptyStringStartingState)
    const [submitButtonClicked, setSubmitButtonClicked] = useState(false)

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
    if ((emptyStringStartingState !== email || submitButtonClicked)) {
        emailValidationMessages = isEmailValid(email)
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
    const onSubmit = () => {
        setSubmitButtonClicked(true)
        if (mismatchedPasswordsNotValid || isEmailValid(email).length !== 0 || isPasswordValid(password).length !== 0) {
            return
        }

        setLoading(true)
        axios.post(baseUrl + "/register", {
                email,
                password
            },
        ).then(() => {
                setLoading(false)
                navigate(
                    "/verification?email=" + email,
                )
            },
        ).catch((e: AxiosError) => {
                setSubmitErrorMessage(e.response?.data as string)
                setLoading(false)
            }
        )
    }

    return (
        <div className={styles['container']}>
            <div>
                <FormHeader>Sign Up</FormHeader>
                <div>
                    <TextBox aria-label='Email'
                             validationErrorMessages={emailValidationMessages}
                             value={email}
                             onChange={(event: FormEvent<HTMLInputElement>) => setEmail(event.currentTarget.value)}
                             placeholder='email'/>
                </div>
                <ConfirmPasswordControl
                    password={password}
                    setPassword={setPassword}
                    confirmPassword={confirmPassword}
                    setConfirmPassword={setConfirmPassword}
                    passwordValidationMessages={passwordValidationMessages}
                    passwordsDoNotMatchValidationMessages={passwordsDoNotMatchValidationMessages}/>
                <div className={styles['spacer']}></div>
                <div className={styles['buttons']}>
                    <Link to="/login" className={styles['already-registered-text']}>Already have an account?</Link>
                    <Button triggerFromEnter={true} loading={loading} onClick={onSubmit}>Submit</Button>
                </div>
                <div className={styles['spacer']}></div>
                <ErrorNotification message={submitErrorMessage}/>
            </div>
        </div>
    );
}

export default RegistrationPage;
