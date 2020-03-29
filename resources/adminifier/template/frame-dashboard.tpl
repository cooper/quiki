<meta
    data-nav="dashboard"
    data-title="Dashboard"
    data-icon="home"
    data-styles="dashboard"
    data-flags="buttons"
    data-buttons="date-selection"
    data-button-date-selection="{'title': 'Last 30 days', 'icon': 'calendar', 'func': 'displayDateSelector'}"
/>


{{if .Errors}}
<h2>Pages with Errors</h2>
These pages are not being served on the wiki due to errors.

<pre class="info">
{{- range .Errors -}}
<a href="edit-page?page={{.File}}">{{.File}}</a>:
{{- .Error.Position.Line}}:{{.Error.Position.Column}}: {{.Error.Message}}
{{end -}}
</pre>
{{end}}

{{if .Warnings}}
<h2>Pages with Warnings</h2>
These pages have warnings.

<pre class="info">
{{- range .Warnings -}}
{{- $file := .File -}}
{{- range .Warnings -}}
<a href="edit-page?page={{$file}}">{{$file}}</a>:
{{- .Position.Line}}:{{.Position.Column}}: {{.Message}}
{{end -}}
{{end -}}
</pre>
{{end}}


<h2>Logs</h2>
<pre class="info">
{{.Logs}}
</pre>