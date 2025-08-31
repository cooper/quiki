{{template "auth-base.tpl" .}}

{{define "title"}}quiki setup wizard{{end}}

{{define "form"}}
<form action="func/create-user" method="post">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
    <div class="auth-form-group">
        <label class="auth-label" for="display">display name</label>
        <input class="auth-input" type="text" name="display" id="display" required autofocus />
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="email">email address</label>
        <input class="auth-input" type="email" name="email" id="email" required />
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="username">username</label>
        <input class="auth-input" type="text" name="username" id="username" required pattern="[A-Za-z0-9]+" />
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="password">password</label>
        <input class="auth-input" type="password" name="password" id="password" required />
    </div>
    
    <div class="auth-form-group">
        <label class="auth-label" for="token">setup token</label>
        <input class="auth-input" type="text" name="token" id="token" required />
        <div class="auth-help-text">enter the setup token from the server logs</div>
    </div>
    
    {{if .DefaultWikiPath}}
    <div class="auth-form-group">
        <label class="auth-label" for="path">wikis path</label>
        <input class="auth-input" type="text" name="path" id="path" required value="{{.DefaultWikiPath}}" />
    </div>
    {{end}}
    
    <button type="submit" class="auth-button">create user and login</button>
</form>
{{end}}

{{define "links"}}
{{end}}