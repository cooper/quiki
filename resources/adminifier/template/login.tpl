<!doctype html>
<html>
<head>
    <meta charset="utf-8" data-wredirect="login" />
    <title>quiki login</title>
    <link rel="icon" type="image/png" href="{{.Static}}/image/favicon.png" />
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
            margin: 15% auto;
        }
        input[type="text"], input[type="password"] {
            width: 100%;
            padding: 8px;
            margin: 5px 0;
            box-sizing: border-box;
            border: 1px solid #ccc;
            border-radius: 4px;
        }
        input[type="submit"] {
            width: 100%;
            background-color: #4f6079;
            color: white;
            padding: 12px 18px;
            margin-top: 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        input[type="submit"]:hover {
            background-color: #2096ce;
        }
        table {
            width: 100%;
        }
        img {
            width: 100px;
        }
    </style>
</head>
<body>
    <div id="login-window">
        <div style="text-align: center; margin-bottom: 20px;">
            <img src="{{.Static}}/image/favicon.png" alt="quiki" />
        </div>
        <form action="func/login?redirect={{.Redirect}}" method="post">
            <table>
                <tr>
                    <td class="left">Username</td>
                    <td><input type="text" name="username" required /></td>
                </tr>
                <tr>
                    <td class="left">Password</td>
                    <td><input type="password" name="password" required /></td>
                </tr>
                <tr>
                    <td colspan="2"><input type="submit" name="submit" value="Login" /></td>
                </tr>
            </table>
        </form>
    </div>
</body>
</html>