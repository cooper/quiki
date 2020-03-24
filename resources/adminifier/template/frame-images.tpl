{{.JSON}}

<meta
    data-nav="images"
    data-title="Images"
    data-icon="images"
    data-buttons="upload image-mode filter actions"
    data-button-image-mode="{'title': 'Grid view', 'icon': 'th', 'href': '{{.Root}}/images'}"
    data-scripts="file-list file-list/images pikaday"

    data-styles="file-list pikaday"
    data-flags="no-margin search buttons"
    data-search="fileSearch"
    data-sort="{{.Order}}"
    data-button-upload="{'title': 'Upload images', 'icon': 'upload', 'href': '{{.Root}}/upload-images'}"
    data-button-filter="{'title': 'Filter', 'icon': 'filter', 'func': 'displayFilter'}"
    data-button-actions="{'title': 'With selected...', 'icon': 'magic', 'func': 'displayActionMenu', 'hide': true}"
/>