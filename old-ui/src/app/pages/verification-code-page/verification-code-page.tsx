import styles from './verification-code-page.module.scss';
import TextBox from "../../components/text-box/text-box";
import axios, { AxiosError } from "axios";
import { useContext, useEffect, useState } from "react";
import BaseUrlContext from "../../context/base-url";
import { useNavigate } from "react-router-dom";
import Button from "../../components/button/button";
import ErrorNotification from "../../components/error-notification/error-notification";
import { getAxiosInstance, useQuery } from "../../utils";
import { exercise } from "../../constants/nav";
import FormHeader from "../../form-header/form-header";


export function VerificationCodePage() {
    const [verificationCode, setVerificationCode] = useState<string>('')
    const [submitButtonClicked, setSubmitButtonClicked] = useState(false)
    const [submitErrorMessage, setSubmitErrorMessage] = useState('')
    const baseUrl = useContext(BaseUrlContext)
    const navigate = useNavigate();
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

    const email = useQuery().get("email");

    const message = useQuery().get("message");

    useEffect(() => {
        if (message) {
            setSubmitErrorMessage(message)
        }
    }, [message])

    const validationMessages = [];
    if (('' !== verificationCode || submitButtonClicked) && '' === verificationCode) {
        validationMessages.push("must provide a verification code")
    }

    const onSubmit = () => {
        setSubmitButtonClicked(true)
        if ('' === verificationCode) {
            return
        }
        setLoadingSubmit(true)
        axios.post(baseUrl + "/verification", {
            email,
            code: verificationCode
        },
        ).then(
            () => {
                setLoadingSubmit(false)
                navigate(exercise)
            },
        ).catch(
            (e: AxiosError) => {
                setLoadingSubmit(false)
                setSubmitErrorMessage(e.response?.data as string)
            }
        )
    }

    const onResendCode = () => {
        setLoadingResendCode(true)
        axios.post(baseUrl + "/resendVerification", {
            email,
        },
        ).then(() => {
            setLoadingResendCode(false)
            setSubmitErrorMessage('')
            setVerificationCode('')
        }
            //     todo show message to check email that code has been resent
        ).catch(
            (e: AxiosError) => {
                setLoadingResendCode(false)
                setSubmitErrorMessage(e.response?.data as string)
                console.log(e)
            }
        );
    }


    return (
        <div className={styles['container']}>
            <div>
                <FormHeader>Confirm Email</FormHeader>
                <TextBox aria-label='Verification Code'
                    placeholder='verification code'
                    value={verificationCode}
                    validationErrorMessages={validationMessages}
                    onChange={(event) => setVerificationCode(event.target.value)} />
                <div className={styles['spacer']}></div>
                <div className={styles['buttons']}>
                    <Button color='secondary' loading={loadingResendCode} onClick={onResendCode}>Resend Code</Button>
                    <div className={styles['spacer']}></div>
                    <Button triggerFromEnter={true} loading={loadingSubmit} onClick={onSubmit}>Submit</Button>
                </div>
                <div className={styles['spacer']}></div>
                <ErrorNotification message={submitErrorMessage} />
            </div>
        </div>
    );
}

export default VerificationCodePage;
