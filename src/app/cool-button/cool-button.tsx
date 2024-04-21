import styles from './cool-button.module.scss';
import {ReactNode} from "react";

export interface CoolButtonProps {
    children: ReactNode;
    onClick: () => void;
}

export function CoolButton({children, onClick}: CoolButtonProps) {
    return (
        <button onClick={onClick} className={`${styles['snowy-button']} no-select`}>
            {children}
        </button>
    );
}

export default CoolButton;
