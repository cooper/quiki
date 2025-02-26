(function (a, exports) {

exports.createSite = function () {
    var createWindow = new ModalWindow({
        icon:           'plus-circle',
        title:          'Create Site',
        html:           tmpl('tmpl-create-site', {}),
        padded:         true,
        id:             'create-page-window',
        autoDestroy:    true,
        doneText:      'Cancel',
        onDone:         null
    });
    createWindow.show();
};

})(adminifier, window);