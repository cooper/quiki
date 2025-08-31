{{template "auth-base.tpl" .}}

{{define "title"}}{{.PageTitle}}{{end}}

{{define "form"}}
<form action="{{.LoginAction}}" method="post">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
    <div class="auth-form-group">
        <label class="auth-label" for="username">username</label>
        <input class="auth-input" type="text" id="username" name="username" required autofocus />
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="password">password</label>
        <input class="auth-input" type="password" id="password" name="password" required />
    </div>
    
    <button type="submit" class="auth-button">sign in</button>
</form>
{{end}}

{{define "links"}}
{{if .AllowRegister}}
    <a href="{{.RegisterURL}}">create account</a> |
{{end}}
<a href="{{.HomeURL}}">back to {{.WikiName}}</a>
{{end}}
