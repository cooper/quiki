<meta
    data-nav="routes"
    data-title="Routes"
    data-icon="route"
/>

<p>
The following routes are registered on the quiki webserver.
</p>

<ul>
{{range .Routes}}
    <li>{{.Pattern}} - {{.Description}}</li>
{{end}}
</ul>