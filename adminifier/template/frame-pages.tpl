{{.JSON}}

<meta
    data-nav="pages"
    data-title="Pages"
    data-icon="file-text"
    data-buttons="create filter actions"
    data-scripts="file-list file-list/pages pikaday"

    data-styles="file-list pikaday"
    data-flags="no-margin search buttons"
    data-search="fileSearch"
    data-sort="{{.Order}}"
    data-button-create="{'title': 'New page', 'icon': 'plus-circle', 'href': '{{.Root}}/create-page'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"
    data-button-actions="{'title': 'With selected...', 'icon': 'magic', 'func': 'displayActionMenu', 'hide': true}"
/>