{{template "auth-base.tpl" .}}

{{define "title"}}{{.PageTitle}}{{end}}

{{define "form"}}
<form action="{{.LoginAction}}" method="post">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
    <div class="auth-form-group">
        <label class="auth-label" for="username">Username</label>
        <input class="auth-input" type="text" id="username" name="username" required autofocus />
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="password">Password</label>
        <input class="auth-input" type="password" id="password" name="password" required />
    </div>
    
    <button type="submit" class="auth-button">Login</button>
</form>
{{end}}

{{define "links"}}
{{if .AllowRegister}}
    <a href="{{.RegisterURL}}">Create account</a> |
{{end}}
<a href="{{.HomeURL}}">Back to {{.WikiName}}</a>
{{end}}
