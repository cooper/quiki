(function (a, exports) {

// purge the generated page and image cache. the server regenerates
// everything afterward, so this is safe but may take a while on big sites.
exports.purgeCache = function (but) {
    if (!confirm('Purge the cache? Every page and image will be regenerated.'))
        return;

    var anchor = but.getElement('a');
    var original = anchor.innerHTML;
    var setLabel = function (text) {
        anchor.innerHTML = original.replace(/Purge cache/, text);
    };
    var finish = function (text) {
        setLabel(text);
        setTimeout(function () { anchor.innerHTML = original; }, 2000);
    };
    setLabel('Purging...');

    new Request.JSON({
        url: 'func/purge-cache',
        onSuccess: function (data) {
            if (data.success) {
                finish('Cache purged');
                return;
            }
            finish('Purge failed');
            alert(data.error || 'Unknown error');
        },
        onFailure: function () {
            finish('Purge failed');
        }
    }).post({});
};

})(adminifier, window);
