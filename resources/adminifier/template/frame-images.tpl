{{.JSON}}
<meta
    data-nav="images"
    data-title="Images"
    data-icon="images"
    data-flags="no-margin search buttons"
    data-search="fileSearch"
    data-sort="{{.Order}}"
    data-cd="{{.Cd}}"


    data-button-upload="{'title': 'Upload', 'icon': 'upload', 'href': '{{.Root}}/upload-images'}"
    data-button-create-folder="{'title': 'New Folder', 'icon': 'folder', 'func': 'createFolder'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"

    data-selection-buttons="move rename delete"
    data-button-move="{'title': 'Move', 'icon': 'folder', 'func': 'moveSelected', 'hide': true}"
    data-button-rename="{'title': 'Rename', 'icon': 'file-signature', 'func': 'renameSelected', 'hide': true}"
    data-button-delete="{'title': 'Delete', 'icon': 'trash', 'func': 'deleteSelected', 'hide': true}"

{{if .List}}
    data-buttons="upload create-folder image-mode filter"
    data-button-image-mode="{'title': 'Grid', 'icon': 'th', 'frameHref': '{{.Root}}/images/{{.Cd}}'}"
    data-scripts="file-list file-list/images images pikaday"
    data-styles="file-list pikaday"
{{else}}
    data-buttons="upload create-folder image-mode"
    data-button-image-mode="{'title': 'List', 'icon': 'list', 'frameHref': '{{.Root}}/images/{{.Cd}}?mode=list'}"
    data-scripts="image-grid images pikaday"
    data-styles="image-grid pikaday"
{{end}}
/>

{{ template "common-file-list.tpl" . }}

<template id="tmpl-image-grid-item">
    <a href="{%= o.link %}" target="_blank">
        <img alt="{%= o.file %}" src="{%= o.link %}?{%= o.dimension %}={%= o.dimValue %}" />
        <span>{%= o.file %}</span>
    </a>
</template>

<template id="tmpl-image-grid-dir">
    <a href="{%= o.link %}">
        <span>
            <i class="fa fa-folder"></i>
            {%= o.name %}
        </span>
    </a>
</template>