<meta
    data-nav="sites"
    data-title="Sites"
    data-icon="globe-americas"
/>

<h1>Welcome, {{.User.DisplayName}}!</h1>

<h2>Available Sites</h2>
{{if not .Wikis}}
    No sites exist yet
{{else}}
    <ul>
    {{range $shortcode, $wi := .Wikis}}
        <li><a href="{{$shortcode}}/dashboard">{{$wi.Title}}</a></li>
    {{end}}
    </ul>
{{end}}

<h2>Create New Site</h2>
<form action="func/create-wiki" method="post">
    <label for="name">Site Name:</label>
    <input type="text" name="name" />
    <label for="template">Select Template:</label>
    <select name="template" id="template">
        {{range .Templates}}
            <option value="{{.}}">{{.}}</option>
        {{end}}
    </select>
    <input type="submit" value="Create" />
</form>

<h2>User Actions</h2>

<a href="logout">Logout</a>