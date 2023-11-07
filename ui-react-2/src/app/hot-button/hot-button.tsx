import styles from './hot-button.module.scss';
import {ReactNode, useEffect, useRef} from "react";

/* eslint-disable-next-line */
export interface HotButtonProps {
    children: ReactNode;
    onClick: () => void;
}


export function HotButton({children, onClick}: HotButtonProps) {
    const fireContainerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        function createParticles(container: HTMLElement | null, num: number, leftSpacing: number) {
            if (!container) return;

            for (let i = 0; i < num; i += 1) {
                const particle = document.createElement("div");
                particle.style.left = `calc((100% - 2em) * ${i / leftSpacing})`;
                particle.className = styles['particle'];
                particle.style.animationDelay = 4 * Math.random() + "s";
                container.appendChild(particle);
            }
        }

        createParticles(fireContainerRef.current, 60, 60);
    }, []);
    return (
        <div className={`${styles['button-container']} no-select`}>
            <div ref={fireContainerRef} className={styles['fire-container']}></div>
                    <button onClick={onClick} className={`${styles['button']} ${styles['fire']}`} type="button">{children}</button>
        </div>

    );
}

export default HotButton;
