{{template "auth-base.tpl" .}}

{{define "title"}}{{.PageTitle}}{{end}}

{{define "form"}}
<form action="{{.RegisterAction}}" method="post">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
    <div class="auth-form-group">
        <label class="auth-label" for="username">username</label>
        <input class="auth-input" type="text" id="username" name="username" required autofocus value="{{.Username}}" />
        <div class="auth-help-text">choose a unique username for this wiki</div>
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="email">email address</label>
        <input class="auth-input" type="email" id="email" name="email" required value="{{.Email}}" />
        <div class="auth-help-text">used for account recovery and notifications</div>
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="password">password</label>
        <input class="auth-input" type="password" id="password" name="password" required />
        <div class="auth-help-text">choose a strong password</div>
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="password_confirm">confirm password</label>
        <input class="auth-input" type="password" id="password_confirm" name="password_confirm" required />
    </div>
    
    <button type="submit" class="auth-button">create account</button>
</form>
{{end}}

{{define "links"}}
<a href="{{.LoginURL}}">already have an account?</a> |
<a href="{{.HomeURL}}">back to {{.WikiName}}</a>
{{end}}
