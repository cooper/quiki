<meta
    data-nav="sites"
    data-title="Sites"
    data-icon="globe-americas"
    data-scripts="admin/sites"
    data-flags="buttons"
    data-buttons="create"
    data-button-create="{'title': 'New Site', 'icon': 'plus-circle', 'func': 'createSite'}"
/>

<h1>Welcome, [[.User.DisplayName]]!</h1>

[[if not .Wikis]]
    No sites exist yet.
[[else]]
    <h2>Available Sites</h2>
    <ul>
    [[range $shortcode, $wi := .Wikis]]
        <li>
            [[$wi.Title]] -
            <a href="sites/[[$shortcode]]/dashboard">
                <i class="fa fa-edit"></i>
            </a>
            <a href="[[$wi.Opt.Root.Ext]]" target="_blank">
                <i class="fa fa-globe-americas"></i>
            </a>
        </li>
    [[end]]
    </ul>
[[end]]

<template id="tmpl-create-site">
    <form action="func/create-wiki" method="post">
        <label for="name">Site Name:</label>
        <input type="text" name="name" />
        <label for="template">Select Template:</label>
        <select name="base" id="base">
            [[range .BaseWikis]]
                <option value="[[.]]">[[.]]</option>
            [[end]]
        </select>
        <label for="template">Select Theme:</label>
        <select name="template" id="template">
            [[range .Templates]]
                <option value="[[.]]">[[.]]</option>
            [[end]]
        </select>
        <input type="submit" value="Create" />
    </form>
</template>