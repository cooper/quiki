{{.JSON}}

<meta
    data-nav="models"
    data-title="Models"
    data-icon="cube"
    data-scripts="file-list file-list/models pikaday"
    data-styles="file-list pikaday"
    data-flags="no-margin search buttons"
    data-search="fileSearch"
    data-sort="{{.Order}}"
    data-cd="{{.Cd}}"

    data-buttons="create create-folder filter"
    data-button-create="{'title': 'New Model', 'icon': 'plus-circle', 'func': 'createModel'}"
    data-button-create-folder="{'title': 'New Folder', 'icon': 'folder', 'func': 'createFolder'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"

    data-selection-buttons="move rename delete"
    data-button-move="{'title': 'Move', 'icon': 'folder', 'func': 'moveSelected', 'hide': true}"
    data-button-rename="{'title': 'Rename', 'icon': 'file-signature', 'func': 'renameSelected', 'hide': true}"
    data-button-delete="{'title': 'Delete', 'icon': 'trash', 'func': 'deleteSelected', 'hide': true}"
/>

{{ template "common-file-list.tpl" . }}

<template id="tmpl-create-model">
    <form action="{{.Root}}/func/create-model" method="post">
        <label for="name">Model Name:</label>
        <input type="text" name="title" />
        <input type="hidden" name="dir" value="{{.Cd}}" />
        <input type="submit" value="Create" />
    </form>
</template>