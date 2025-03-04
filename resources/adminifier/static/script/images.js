(function (a, exports) {

// this file is common to both the image grid and the image list.
// see also ./image-grid.js and ./file-list/images.js

exports.createFolder = function () {
    var modal = new ModalWindow({
        icon:           'folder',
        title:          'New Folder',
        html:           tmpl('tmpl-create-folder', { mode: 'image' }),
        padded:         true,
        id:             'create-folder-window',
        autoDestroy:    true,
        onDone:         null,
        doneText:       'Cancel',
    });
    modal.show();
}

})(adminifier, window);