import styles from './header.module.scss';
import {ReactNode} from "react";

/* eslint-disable-next-line */
export interface HeaderProps {
    children?: ReactNode,
}

export function Header({children}: HeaderProps) {
  return (
      <div className={styles['header']}>
          <h2>{children}</h2>
      </div>
  );
}

export default Header;
