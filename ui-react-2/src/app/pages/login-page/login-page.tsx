import styles from './login-page.module.scss';
import TextBox from "../../components/text-box/text-box";
import {FormEvent, useContext, useEffect, useState} from "react";
import {getAxiosInstance, isEmailValid} from "../../utils";
import {Link, useNavigate} from "react-router-dom";
import Button from "../../components/button/button";
import {AxiosError} from "axios";
import BaseUrlContext from "../../context/base-url";
import ErrorNotification from "../../components/error-notification/error-notification";
import {exercise, forgotPasswordEmail, login, verification} from "../../constants/nav";
import FormHeader from "../../form-header/form-header";


export function LoginPage() {
    const [submitErrorMessage, setSubmitErrorMessage] = useState('')
    const baseUrl = useContext(BaseUrlContext)
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [submitButtonClicked, setSubmitButtonClicked] = useState(false)
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false)

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

    // todo add feedback if no password is provided. Need to take advantage of on blur hook somehow

    const onSubmit = () => {
        setSubmitButtonClicked(true)

        if (isEmailValid(email).length !== 0) {
            return
        }

        setLoading(true)
        getAxiosInstance().post(baseUrl + login, {
                email,
                password
            },
        ).then(
            () => {
                setLoading(false)
                navigate(
                    exercise,
                )
            },
        ).catch(
            (e: AxiosError) => {
                setLoading(false)
                const responseMessage = e?.response?.data as string;
                if (e?.response?.status === 403) {
                    navigate(
                        verification + "?email=" + email + "&message=" + responseMessage,
                    );
                }
                setSubmitErrorMessage(responseMessage)
            }
        )
    }


    return (
        <div className={styles['container']}>
            <div>
                <FormHeader>Login</FormHeader>
                <div>
                    <TextBox aria-label='username'
                             validationErrorMessages={emailValidationMessages}
                             value={email}
                             onChange={(event: FormEvent<HTMLInputElement>) => setEmail(event.currentTarget.value)}
                             placeholder='email'/>
                </div>
                <div>
                    <TextBox aria-label='password'
                             placeholder='password'
                             value={password}
                             validationErrorMessages={[]}
                             onChange={(event) => setPassword(event.target.value)}
                             type='password'/>
                </div>
                <div className={styles['spacer']}></div>
                <div className={styles['buttons']}>
                    <Link className={styles['forgot-password-link']} to={forgotPasswordEmail}>Forgot password?</Link>
                    <div className={styles['spacer']}></div>
                    <Button triggerFromEnter={true} loading={loading} onClick={onSubmit}>Submit</Button>
                </div>
                <div className={styles['spacer']}></div>
                <ErrorNotification message={submitErrorMessage}/>
                <div className={styles['divider']}>
                    <span className={styles['divider-text']}>or</span>
                </div>
                <Link to="/register" className={styles['new-account-text']}>Create new account</Link>
            </div>
        </div>
    );
}

export default LoginPage;
