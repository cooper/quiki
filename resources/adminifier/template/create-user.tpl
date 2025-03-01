<!doctype html>
<html>
<head>
    <meta charset="utf-8" data-wredirect="login" />
    <title>quiki - create user</title>
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
        input:not([type=submit]):not(:placeholder-shown) {
            background-color: pink;
        }
        input:not([type=submit]):valid {
            background-color: inherit;
        }
    </style>
</head>
<body>
    <div id="login-window">
        <div style="text-align: center; margin-bottom: 20px;">
            <h1>quiki</h1>
            Create Initial User
        </div>
        <form action="func/create-user" method="post">
            <table>
                <tr>
                    <td class="left">Display name</td>
                    <td><input type="text" name="display" required placeholder=" " /></td>
                </tr>
                <tr>
                    <td class="left">Email</td>
                    <td><input type="email" name="email" required placeholder=" "  /></td>
                </tr>
                <tr>
                    <td class="left">Username</td>
                    <td><input type="text" name="username" required pattern="[A-Za-z0-9]+" placeholder=" " /></td>
                </tr>
                <tr>
                    <td class="left">Password</td>
                    <td><input type="password" name="password" required pattern="\S+" placeholder=" " /></td>
                </tr>
                <tr>
                    <td class="left">Token</td>
                    <td><input type="text" name="token" required placeholder=" " /></td>
                </tr>
                [[if .DefaultWikiPath]]
                <tr>
                    <td class="left">Wikis Path</td>
                    <td><input type="text" name="path" required value="[[.DefaultWikiPath]]" placeholder=" " /></td>
                </tr>
                [[end]]
                <tr>
                    <td><input type="submit" name="submit" value="Create" /></td>
                </tr>
            </table>
        </form>
    </div>
</body>
</html>