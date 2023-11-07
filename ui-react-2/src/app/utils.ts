import axios, {AxiosInstance} from "axios";
import {useLocation} from "react-router-dom";

export function getBaseApiUrlFromHostname() {
    const hostname = window.location.hostname;
    if (hostname.includes('localhost')) {
        // local
        return `http://localhost:4200/api`;
    } else if (hostname.includes('staging.')) {
        // staging
        return 'https://api.staging.strengthgadget.com/api';
    } else {
        // production
        return 'https://api.strengthgadget.com/api';
    }
}

export function isPasswordValid(password: string): string[] {
    const passwordValidationMessages: string[] = []
    if (password.length < 12) {
        passwordValidationMessages.push("password is too short")
    }

    const testPasswordFormat = /^\d+$/
    if (testPasswordFormat.test(password)) {
        passwordValidationMessages.push("cannot be all numbers")
    }
    return passwordValidationMessages
}

export function isEmailValid(email: string): string[] {
    const emailValidationMessages: string[] = []
    const testEmailFormat = /^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/
    if (!testEmailFormat.test(email)) {
        emailValidationMessages.push("invalid email")
    }
    return emailValidationMessages
}

//  todo either keep this strategy remove it in favor of the existing one
const localAxiosInstance: AxiosInstance = axios.create({
    baseURL: getBaseApiUrlFromHostname(),
    withCredentials: true
});

const axiosInstance: AxiosInstance = axios.create({
    baseURL: getBaseApiUrlFromHostname(),
    withCredentials: true
});

export const getAxiosInstance = (): AxiosInstance => {
    console.log("Application mode: ", process.env.NODE_ENV)
    return process.env.NODE_ENV === 'development' ? localAxiosInstance : axiosInstance
}

export function useQuery() {
    return new URLSearchParams(useLocation().search);
}