import TextBox from "../text-box/text-box";

export interface ConfirmPasswordControlProps {
    password: string;
    confirmPassword: string;
    setPassword: (password: string) => void;
    setConfirmPassword: (confirmPassword: string) => void;
    passwordValidationMessages: string[];
    passwordsDoNotMatchValidationMessages: string[];
}

export function ConfirmPasswordControl({password, confirmPassword, setPassword, setConfirmPassword, passwordValidationMessages, passwordsDoNotMatchValidationMessages}: ConfirmPasswordControlProps) {
    return (
        <div>
            <div>
                <TextBox aria-label='Password'
                         placeholder='password'
                         value={password}
                         validationErrorMessages={passwordValidationMessages}
                         onChange={(event) => setPassword(event.target.value)}
                         type='password'/>
            </div>
            <div>
                <TextBox aria-label='Confirm Password'
                         placeholder='confirm password'
                         value={confirmPassword}
                         validationErrorMessages={passwordsDoNotMatchValidationMessages}
                         onChange={(event) => setConfirmPassword(event.target.value)}
                         type='password'/>
            </div>
        </div>
    );
}

export default ConfirmPasswordControl;
