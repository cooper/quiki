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

    data-buttons="create filter"
    data-button-create="{'title': 'New Model', 'icon': 'plus-circle', 'func': 'createModel'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"

    data-selection-buttons="move rename delete"
    data-button-move="{'title': 'Move', 'icon': 'folder', 'func': 'moveSelected', 'hide': true}"
    data-button-rename="{'title': 'Rename', 'icon': 'file-signature', 'func': 'renameSelected', 'hide': true}"
    data-button-delete="{'title': 'Delete', 'icon': 'trash', 'func': 'deleteSelected', 'hide': true}"
/>

{{ template "common-file-list.tpl" . }}

<template id="tmpl-create-model">
    <form action="func/create-model" method="post">
        <label for="name">Model Name:</label>
        <input type="text" name="title" />
        <input type="submit" value="Create" />
    </form>
</template>