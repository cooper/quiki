<!doctype html>
<html>
<head>
    <meta charset="utf-8" data-wredirect="login" />
    <title>{{template "title" .}}</title>
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
        
        {{template "form" .}}
        
        <div class="auth-links">
            {{template "links" .}}
        </div>
    </div>
</body>
</html>
