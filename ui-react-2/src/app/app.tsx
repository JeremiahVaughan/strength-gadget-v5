import {createBrowserRouter, RouterProvider,} from "react-router-dom";

import WelcomePage from "./pages/welcome-page/welcome-page";
import RegistrationPage from "./pages/registration-page/registration-page";
import LoginPage from "./pages/login-page/login-page";
import {getBaseApiUrlFromHostname} from "./utils";
import {useEffect, useState} from "react";
import BaseUrlContext from "./context/base-url";
import VerificationCodePage from "./pages/verification-code-page/verification-code-page";
import ExercisePage from "./pages/exercise-page/exercise-page";
import {
    exercise,
    forgotPasswordEmail, forgotPasswordNewPassword,
    forgotPasswordResetCode,
    home,
    login,
    register,
    verification
} from "./constants/nav";
import ForgotPasswordResetCodePage from "./pages/forgot-password-reset-code-page/forgot-password-reset-code-page";
import ForgotPasswordEmailPage from "./pages/forgot-password-email-page/forgot-password-email-page";
import ForgotPasswordNewPasswordPage from "./pages/forgot-password-new-password-page/forgot-password-new-password-page";

const router = createBrowserRouter([
    {
        path: home,
        element: <WelcomePage/>
    },
    {
        path: register,
        element: <RegistrationPage/>,
    },
    {
        path: verification,
        element: <VerificationCodePage/>,
    },
    {
        path: login,
        element: <LoginPage/>,
    },
    {
        path: forgotPasswordEmail,
        element: <ForgotPasswordEmailPage/>,
    },
    {
        path: forgotPasswordResetCode,
        element: <ForgotPasswordResetCodePage/>,
    },
    {
        path: forgotPasswordNewPassword,
        element: <ForgotPasswordNewPasswordPage/>
    },
    {
        path: exercise,
        element: <ExercisePage/>,
    },
]);


export function App() {
    const [baseUrl, setBaseUrl] = useState('');
    useEffect(() => {
        setBaseUrl(getBaseApiUrlFromHostname());
    }, []);
    return <BaseUrlContext.Provider value={baseUrl}>
        <RouterProvider router={router}/>
    </BaseUrlContext.Provider>
}


export default App;
