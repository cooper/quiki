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

// enable or disable a site in the server configuration
exports.toggleSite = function (name, enable) {
    var what = enable ? 'enable' : 'disable';
    if (!confirm('Really ' + what + ' the site "' + name + '"?'))
        return;
    new Request.JSON({
        url: 'func/' + what + '-wiki',
        onSuccess: function (data) {
            if (!data.success) {
                alert('Failed to ' + what + ' site: ' + (data.error || 'Unknown error'));
                return;
            }
            a.reloadFrame();
        },
        onFailure: function () {
            alert('Failed to ' + what + ' site: request error');
        }
    }).post({ name: name });
};

})(adminifier, window);