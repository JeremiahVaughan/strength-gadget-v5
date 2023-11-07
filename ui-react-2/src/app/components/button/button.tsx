import styles from './button.module.scss';
import {ReactNode, useEffect, useRef} from "react";
import IconCogWheelTriple from "../icon-cog-wheel-triple/icon-cog-wheel-triple";

export type color =
    'primary' |
    'secondary' |
    'disabled'

/* eslint-disable-next-line */
export interface ButtonProps {
    children?: ReactNode,
    color?: color,
    onClick?: () => void
    triggerFromEnter?: boolean
    loading?: boolean
}

export function Button({
                           children,
                           color = 'primary',
                           onClick,
                           triggerFromEnter,
                           loading,
                           ...rest
                       }: ButtonProps) {

    const buttonRef = useRef<HTMLButtonElement | null>(null);
    useEffect(() => {
        if (!triggerFromEnter) {
            return
        }
        const handleKeyPress = (event: KeyboardEvent) => {
            if (event.key === 'Enter') {
                buttonRef.current?.click();
            }
        };

        window.addEventListener('keydown', handleKeyPress);

        return () => {
            window.removeEventListener('keydown', handleKeyPress);
        };
    }, [triggerFromEnter]);

    if (loading) {
        color = 'disabled'
    }
    return (
        <button ref={buttonRef}
                onClick={onClick}
                className={`${styles['button']} ${styles[color]} no-select`}
                type='button' {...rest}>
            {!loading ? children : <IconCogWheelTriple/>}
        </button>
    );
}

export default Button;
