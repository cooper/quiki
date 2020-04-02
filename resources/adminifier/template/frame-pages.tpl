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

    data-buttons="create filter"
    data-button-create="{'title': 'New page', 'icon': 'plus-circle', 'func': 'createPage'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"

    data-selection-buttons="move rename delete"
    data-button-move="{'title': 'Move', 'icon': 'folder', 'func': 'moveSelected', 'hide': true}"
    data-button-rename="{'title': 'Rename', 'icon': 'file-signature', 'func': 'renameSelected', 'hide': true}"
    data-button-delete="{'title': 'Delete', 'icon': 'trash', 'func': 'deleteSelected', 'hide': true}"
/>