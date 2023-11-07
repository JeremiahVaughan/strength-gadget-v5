import styles from './tool-bar.module.scss';
import {getAxiosInstance} from "../../utils";
import {useNavigate} from "react-router-dom";


export function ToolBar() {
    const navigate = useNavigate();
    function handleLogout() {
        getAxiosInstance().post('/logout').then(() => {
            navigate('/login', {replace: true});
        }).catch(err => {
            console.log(err);
        })
    }
    return (
        <div className={styles['container']}>
            <div className={styles['authenticate-button']}
                 onClick={handleLogout}>logout</div>
        </div>
    );
}

export default ToolBar;
