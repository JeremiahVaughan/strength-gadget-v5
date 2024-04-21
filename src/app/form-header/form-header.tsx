import styles from './form-header.module.scss';
import Header from "../components/header/header";
import {ReactNode} from "react";

export interface FormHeaderProps {
    children: ReactNode
}

export function FormHeader({children}: FormHeaderProps) {
  return (
    <div className={styles['header']}>
                    <Header>{children}</Header>
                </div>
  );
}

export default FormHeader;
