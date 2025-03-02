{{.JSON}}
<meta
    data-nav="images"
    data-title="Images"
    data-icon="images"
    data-flags="no-margin search buttons"
    data-search="fileSearch"
    data-sort="{{.Order}}"

    data-buttons="upload image-mode filter"
    data-button-upload="{'title': 'Upload', 'icon': 'upload', 'href': '{{.Root}}/upload-images'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"

    data-selection-buttons="move rename delete"
    data-button-move="{'title': 'Move', 'icon': 'folder', 'func': 'moveSelected', 'hide': true}"
    data-button-rename="{'title': 'Rename', 'icon': 'file-signature', 'func': 'renameSelected', 'hide': true}"
    data-button-delete="{'title': 'Delete', 'icon': 'trash', 'func': 'deleteSelected', 'hide': true}"

{{if .List}}
    data-button-image-mode="{'title': 'Grid view', 'icon': 'th', 'frameHref': '{{.Root}}/images'}"
    data-scripts="file-list file-list/images pikaday"
    data-styles="file-list pikaday"
{{else}}
    data-button-image-mode="{'title': 'List view', 'icon': 'list', 'frameHref': '{{.Root}}/images?mode=list'}"
    data-scripts="image-grid pikaday"
    data-styles="image-grid pikaday"
{{end}}
/>

{{ template "common-file-list.tpl" . }}

<template id="tmpl-image-grid-item">
    <a href="func/image/{%= o.file %}">
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