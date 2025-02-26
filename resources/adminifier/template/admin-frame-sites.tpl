<meta
    data-nav="sites"
    data-title="Sites"
    data-icon="globe-americas"
    data-scripts="admin/sites"
    data-flags="buttons"
    data-buttons="create"
    data-button-create="{'title': 'New Site', 'icon': 'plus-circle', 'func': 'createSite'}"
/>

<h1>Welcome, {{.User.DisplayName}}!</h1>

{{if not .Wikis}}
    No sites exist yet.
{{else}}
    <h2>Available Sites</h2>
    <ul>
    {{range $shortcode, $wi := .Wikis}}
        <li><a href="sites/{{$shortcode}}/dashboard">{{$wi.Title}}</a></li>
    {{end}}
    </ul>
{{end}}

<div id="tmpl-create-site" class="display-none">
    <form action="func/create-wiki" method="post">
        <label for="name">Site Name:</label>
        <input type="text" name="name" />
        <label for="template">Select Template:</label>
        <select name="base" id="base">
            {{range .BaseWikis}}
                <option value="{{.}}">{{.}}</option>
            {{end}}
        </select>
        <label for="template">Select Theme:</label>
        <select name="template" id="template">
            {{range .Templates}}
                <option value="{{.}}">{{.}}</option>
            {{end}}
        </select>
        <input type="submit" value="Create" />
    </form>
</div>