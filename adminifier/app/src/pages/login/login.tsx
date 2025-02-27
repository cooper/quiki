import Input from '../../components/ui/input';
import styles from './login.module.css';

const Login = () => {
    return (
        <div class={styles.login}>
            <div style="text-align: center; margin-bottom: 20px;">
                <h1 class={styles.h1}>quiki</h1>
            </div>
            <form action="func/login" method="post">
                <table class="w-full">
                    <tbody>
                        <tr>
                            <td class="left">Username</td>
                            <td><Input type="text" name="username" /></td>
                        </tr>
                        <tr>
                            <td class="left">Password</td>
                            <td><Input type="password" name="password" /></td>
                        </tr>
                        <tr>
                            <td><Input type="submit" name="submit" value="Login" /></td>
                        </tr>
                    </tbody>
                </table>
            </form>
        </div>
    );
}

export default Login;