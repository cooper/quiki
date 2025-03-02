<!doctype html>
<html>
<head>
    <meta charset="utf-8" data-wredirect="login" />
    <title>quiki login</title>
    <link rel="icon" type="image/png" href="/static/favicon.png" />
    <style>
        body {
            background-color: #333;
            font-family: sans-serif;
            text-align: center;
        }
        #login-window {
            width: 400px;
            padding: 30px;
            border: 1px solid #999;
            background-color: white;
            margin: 50px auto;
        }
    </style>
</head>
<body>
    <div id="login-window">
        <div style="text-align: center; margin-bottom: 20px;">
            <h1>quiki</h1>
        </div>
        <form action="func/login?redirect={{.Redirect}}" method="post">
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
        </form>
    </div>
</body>
</html>