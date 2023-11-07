import {
    findByRole,
    findByText,
    getByPlaceholderText,
    getByRole,
    queryByText,
    render,
} from '@testing-library/react';
import userEvent from '@testing-library/user-event'

import RegistrationPage from './registration-page';
import {BrowserRouter} from "react-router-dom";
import {createServer} from "../../../test/server";
import {environment} from "../../constants/env";

describe('RegistrationPage', () => {
    createServer([
        {
            path: environment.registrationUrl,
            res: ({req, res, context}) => {
                return {}
            }
        }
    ])
    function setup() {
        const {baseElement} = render(
            <BrowserRouter>
                <RegistrationPage/>
            </BrowserRouter>
        );
        return baseElement;
    }

    it('should render successfully', () => {
        const baseElement = setup();
        expect(baseElement).toBeTruthy();
    });

    it('should show an email, password, and confirm password input', () => {
        const baseElement = setup();

        const emailElement = getByRole(baseElement,
            'textbox',
            {
                name: /email/i
            })
        const passwordElement = getByPlaceholderText(baseElement,
            /^password$/i
        )
        const confirmPasswordElement = getByPlaceholderText(baseElement,
            /confirm password/i
        )

        expect(emailElement).toBeTruthy()
        expect(passwordElement).toBeTruthy()
        expect(confirmPasswordElement).toBeTruthy()
    })

    it('should show an submit and cancel button', () => {
        const baseElement = setup();

        const cancelButton = getByRole(baseElement,
            'button',
            {
                name: /cancel/i
            })
        const submitButton = getByRole(baseElement,
            'button',
            {
                name: /submit/i
            })

        expect(cancelButton).toBeTruthy()
        expect(submitButton).toBeTruthy()
    })

    it('should show feedback for an invalid email', () => {
        const baseElement = setup();
        const emailElement = getByPlaceholderText(baseElement,
            /^email$/i
        )
        userEvent.click(emailElement)
        userEvent.keyboard('notvalidemail123')

        const feedbackMessage = findByText(baseElement, /invalid email/i)
        expect(feedbackMessage).toBeTruthy()
    })

    it('should show feedback if passwords do not match', async () => {
        const baseElement = setup();
        const passwordElement = getByPlaceholderText(baseElement,
            /^password$/i
        )
        const confirmPasswordElement = getByPlaceholderText(baseElement,
            /confirm password/i
        )
        await userEvent.click(passwordElement)
        await userEvent.keyboard('9999999999password123')

        await userEvent.click(confirmPasswordElement)
        await userEvent.keyboard('9999999999password12')

        const feedbackMessage = await findByText(baseElement, /passwords do not match/i)
        expect(feedbackMessage).toBeTruthy()
    })

    it('should show feedback if password contains all numbers', async () => {
        const baseElement = setup()
        const passwordElement = getByPlaceholderText(baseElement,
            /^password$/i
        )

        await userEvent.click(passwordElement)
        await userEvent.keyboard('12345667754332234')
        const feedbackMessage = await findByText(baseElement, /cannot be all numbers/i)
        expect(feedbackMessage).toBeTruthy()
    })

    it('should show feedback if password is too short', async () => {
        const baseElement = setup()
        const passwordElement = getByPlaceholderText(baseElement,
            /^password$/i
        )

        await userEvent.click(passwordElement)
        await userEvent.keyboard('123')
        const feedbackMessage = await findByText(baseElement, /password is too short/i)
        expect(feedbackMessage).toBeTruthy()
    })

    it('should show feedback if password is blank', async () => {
        const baseElement = setup()
        const submitButton = getByRole(baseElement,
            'button',
            {
                name: /submit/i
            })
        await userEvent.click(submitButton)
        const feedbackMessage = await findByText(baseElement, /password is too short/i)
        expect(feedbackMessage).toBeTruthy()
    })

    it('should show feedback if submit button is clicked and no text was placed in the email field', async () => {
        const baseElement = setup();
        const submitButton = getByRole(baseElement,
            'button',
            {
                name: /submit/i
            })
        await userEvent.click(submitButton)
        const feedbackMessage = await findByText(baseElement, /invalid email/i)
        expect(feedbackMessage).toBeTruthy()
    })

    it('should not show any validation feedback should all validators be satisfied', async () => {
        const baseElement = setup();
        const emailElement = getByPlaceholderText(baseElement,
            /^email$/i
        )
        const passwordElement = getByPlaceholderText(baseElement,
            /^password$/i
        )
        const confirmPasswordElement = getByPlaceholderText(baseElement,
            /confirm password/i
        )

        await userEvent.click(emailElement)
        await userEvent.keyboard('goodEmail@gmail.com')

        await userEvent.click(passwordElement)
        await userEvent.keyboard('wkh-urz2ztz@wez4YBT')

        await userEvent.click(confirmPasswordElement)
        await userEvent.keyboard('wkh-urz2ztz@wez4YBT')

        // Waiting so feedback messages can appear
        await findByRole(baseElement, 'heading', {
            name: /sign up/i
        });

        const possibleFeedbackMessageInvalidEmail = queryByText(baseElement, /invalid email/i)
        expect(possibleFeedbackMessageInvalidEmail).toBeFalsy()

        const possibleFeedbackMessageCannotBeAllNumbers = queryByText(baseElement, /cannot be all numbers/i)
        expect(possibleFeedbackMessageCannotBeAllNumbers).toBeFalsy()

        const possibleFeedbackMessagePasswordIsTooShort = queryByText(baseElement, /password is too short/i)
        expect(possibleFeedbackMessagePasswordIsTooShort).toBeFalsy()

        const possibleFeedbackMessagePasswordsDoNotMatch = queryByText(baseElement, /passwords do not match/i)
        expect(possibleFeedbackMessagePasswordsDoNotMatch).toBeFalsy()
    })
});
