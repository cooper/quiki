<h1>Welcome, {{.User.DisplayName}}!</h1>

Available sites:
<ul>
{{range $shortcode, $wi := .Wikis}}
    <li><a href="{{$shortcode}}/dashboard">{{$wi.Title}}</a></li>
{{end}}
</ul>
