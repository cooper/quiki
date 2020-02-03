<!doctype html>
<html>
<head>
    <meta charset="utf-8" data-wredirect="login" />
    <title><?= $config->wiki_name ?> login</title>
    <link rel="icon" type="image/png" href="images/favicon.png" />
    <link href='http://fonts.googleapis.com/css?family=Open+Sans:300,400,600' rel='stylesheet' type='text/css' />
    <link type="text/css" rel="stylesheet" href="style/main.css" />
    <style type="text/css">
        body {
            background-color: #333;
            font-family: 'Open Sans', sans-serif;
            text-align: center;
        }
        #logo {
            width: 300px;
            border: none;
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
            <h1>Wiki name or logo</h1>
        </div>
        <form action="functions/login" method="post">
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

{{ template "tmpl-login-window.tpl" . }}

</body>
</html>