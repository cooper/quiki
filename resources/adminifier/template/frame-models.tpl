{{.JSON}}

<meta
    data-nav="models"
    data-title="Models"
    data-icon="cube"
    data-buttons="create filter actions"
    data-scripts="file-list file-list/models pikaday"

    data-styles="file-list pikaday"
    data-flags="no-margin search buttons"
    data-search="fileSearch"
    data-sort="{{.Order}}"
    data-button-create="{'title': 'New model', 'icon': 'plus-circle', 'href': '{{.Root}}/create-model'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"
    data-button-actions="{'title': 'With selected...', 'icon': 'magic', 'func': 'displayActionMenu', 'hide': true}"
/>