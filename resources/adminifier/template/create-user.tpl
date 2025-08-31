<!doctype html>
<html>
<head>
    <meta charset="utf-8" data-wredirect="login" />
    <title>{{.Title}}</title>
    <link rel="icon" type="image/png" href="{{.Static}}/image/favicon.png" />
    <link rel="stylesheet" href="{{.SharedStatic}}/auth.css" />
</head>
<body class="auth-page">
    <div class="auth-container">
        <div class="auth-logo">
            {{if .WikiLogo}}
                <img src="{{.Static}}/{{.WikiLogo}}" alt="{{.WikiName}}" />
            {{else}}
                <h1>{{.WikiName}}</h1>
            {{end}}
        </div>
        
        <h2 class="auth-heading">{{.Heading}}</h2>
        
        {{if .Error}}
            <div class="auth-error">{{.Error}}</div>
        {{end}}
        
        {{if .Success}}
            <div class="auth-success">{{.Success}}</div>
        {{end}}
        
        <form action="func/create-user" method="post">
            <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
            <div class="auth-form-group">
                <label class="auth-label" for="display">Display Name</label>
                <input class="auth-input" type="text" name="display" id="display" required autofocus />
            </div>
            
            <div class="auth-form-group">
                <label class="auth-label" for="email">Email</label>
                <input class="auth-input" type="email" name="email" id="email" required />
            </div>
            
            <div class="auth-form-group">
                <label class="auth-label" for="username">Username</label>
                <input class="auth-input" type="text" name="username" id="username" required pattern="[A-Za-z0-9]+" />
            </div>
            
            <div class="auth-form-group">
                <label class="auth-label" for="password">Password</label>
                <input class="auth-input" type="password" name="password" id="password" required />
            </div>
            
            <div class="auth-form-group">
                <label class="auth-label" for="token">Setup Token</label>
                <input class="auth-input" type="text" name="token" id="token" required />
                <div class="auth-help-text">Enter the setup token from the startup logs</div>
            </div>
            <button type="submit" class="auth-button">Create User</button>
        </form>
        
    </div>
</body>
</html>
