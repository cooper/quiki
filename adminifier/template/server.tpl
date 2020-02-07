<h1>Welcome, {{.User.DisplayName}}!</h1>

Available sites:
<ul>
{{range $shortcode, $wi := .Wikis}}
    <li><a href="{{$shortcode}}/">{{$wi.Title}}</a></li>
{{end}}
</ul>
