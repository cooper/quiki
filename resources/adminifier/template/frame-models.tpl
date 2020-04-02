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
    data-button-create="{'title': 'New model', 'icon': 'plus-circle', 'href': '{{.Root}}/create-model'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"

    data-selection-buttons="move rename delete"
    data-button-move="{'title': 'Move', 'icon': 'folder', 'func': 'moveSelected', 'hide': true}"
    data-button-rename="{'title': 'Rename', 'icon': 'file-signature', 'func': 'renameSelected', 'hide': true}"
    data-button-delete="{'title': 'Delete', 'icon': 'trash', 'func': 'deleteSelected', 'hide': true}"
/>