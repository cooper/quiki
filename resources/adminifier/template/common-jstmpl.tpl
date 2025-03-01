<template id="tmpl-notification">
    <h3>{%= o.title %}</h3>
    <i class="fa fa-{%= o.icon %}"></i>
    <div>{%= o.message %}</div>
</template>

<template id="tmpl-login-check">
    <div id="login-window-check"><i class="fa fa-check-circle fa-5x center"></i></div>
</template>

<template id="tmpl-login-window">
    <table>
        <tr>
            <td class="left">Username</td>
            <td><input type="text" name="username" /></td>
        </tr>
        <tr>
            <td class="left">Password</td>
            <td><input type="password" name="password" /></td>
        </tr>
        <tr>
            <td><input type="submit" name="submit" value="Login" /></td>
        </tr>
    </table>
</template>