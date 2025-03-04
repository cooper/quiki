{{.JSON}}

<meta
    data-nav="pages"
    data-title="Pages"
    data-icon="file-alt"
    data-scripts="file-list file-list/pages pikaday"
    data-styles="file-list pikaday"
    data-flags="no-margin search buttons"
    data-search="fileSearch"
    data-sort="{{.Order}}"
    data-cd="{{.Cd}}"

    data-buttons="create-page create-folder filter"
    data-button-create-page="{'title': 'New Page', 'icon': 'plus-circle', 'func': 'createPage'}"
    data-button-create-folder="{'title': 'New Folder', 'icon': 'folder', 'func': 'createFolder'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"

    data-selection-buttons="move rename delete"
    data-button-move="{'title': 'Move', 'icon': 'folder', 'func': 'moveSelected', 'hide': true}"
    data-button-rename="{'title': 'Rename', 'icon': 'file-signature', 'func': 'renameSelected', 'hide': true}"
    data-button-delete="{'title': 'Delete', 'icon': 'trash', 'func': 'deleteSelected', 'hide': true}"
/>

{{ template "common-file-list.tpl" . }}

<template id="tmpl-create-page">
    <form action="{{.Root}}/func/create-page" method="post">
        <label for="name">Page Title:</label>
        <input type="text" name="title" />
        <input type="hidden" name="dir" value="{{.Cd}}" />
        <input type="submit" value="Create" />
    </form>
</template>
