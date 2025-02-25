(function (a) {

document.addEvent('editorLoaded', loadedHandler);
document.addEvent('pageUnloaded', unloadedHandler);

var ae;
function loadedHandler () {
    ae = a.editor;

    // add toolbar functions
    ae.addToolbarFunctions({
        save:       displaySaveHelper,
        delete:     displayDeleteConfirmation
    });

    if (ae.isConfig())
        ae.liForAction('delete').addClass('disabled');

    // add keyboard shortcut
    ae.addKeyboardShortcuts([
        [ 'Ctrl-S', 'Command-S', 'save']
    ]);
    
    // start the autosave timer
    resetAutosaveInterval();
}

function unloadedHandler () {
    document.removeEvent('editorLoaded', loadedHandler);
    document.removeEvent('pageUnloaded', unloadedHandler);
    clearAutosaveInterval();
}

// DELETE CONFIRMATION

function displayDeleteConfirmation () {
    var li  = ae.liForAction('delete');
    var box = ae.createPopupBox(li);
    ae.fakeAdopt(box);

    box.innerHTML = tmpl('tmpl-delete-confirm', {});

    // button text events
    var btn = $('editor-delete-button');
    var shouldChange = function () {
        return !btn.hasClass('progress') &&
               !btn.hasClass('success')  &&
               !btn.hasClass('failure');
    };
    btn.addEvent('mouseenter', function () {
        if (shouldChange()) btn.innerHTML = 'Delete this page';
    });
    btn.addEvent('mouseleave', function () {
        if (shouldChange()) btn.innerHTML = 'Are you sure?';
    });

    // delete page function
    var deletePage = function () {

        // prevent box from closing for now
        box.addClass('sticky');

        // "deleting..."
        $('editor-delete-wrapper').innerHTML = tmpl('tmpl-save-spinner', {});
        btn.innerHTML = 'Deleting page';
        btn.addClass('progress');

        // success callback
        var success = function () {

            // switch to checkmark
            var i = btn.parentElement.getElement('i');
            i.removeClass('fa-spinner');
            i.removeClass('fa-spin');
            i.addClass('fa-check-circle');

            // update button
            btn.addClass('success');
            btn.removeClass('progress');
            btn.innerHTML = 'File deleted';

            // redirect
            setTimeout(function () {
                window.location = a.adminRoot
            }, 3000);
        };

        var fail = function () {

            // switch to /!\
            var i = btn.parentElement.getElement('i');
            i.removeClass('fa-spinner');
            i.removeClass('fa-spin');
            i.addClass('fa-exclamation-triangle');

            // update button
            btn.addClass('failure');
            btn.removeClass('progress');
            btn.innerHTML = 'Delete failed';

            setTimeout(function () {
                ae.closePopup(box);
            }, 1500);
        };

        // delete request
        var req = new Request.JSON({
            url: 'func/delete-' + (ae.isModel() ? 'model' : 'page'),
            onSuccess: function (data) {

                // deleted without error
                if (data.success)
                    success();

                // delete error
                else
                    fail('Unknown error');
            },
            onFailure: function () { fail('Request error') },
        }).post({
            page: ae.getFilename()
        });

    };

    // on click, delete page
    btn.addEvent('click', deletePage);

    ae.displayPopupBox(box, 120, li);
}

// SAVE COMMIT HELPER

function displaySaveHelper () {
    return _saveHelper(false);
}

function _saveHelper () {
    var li  = ae.liForAction('save');
    var box = ae.createPopupBox(li);
    ae.fakeAdopt(box);

    box.innerHTML = tmpl('tmpl-save-helper', {
        file: ae.getFilename()
    });

    var closeBoxSoon = function () {
        setTimeout(function () {

            // make the box no longer sticky, so that when the user
            // clicks away, it will disappear now
            box.removeClass('sticky');

            // close the popup only if the mouse isn't over it
            ae.closePopup(box, {
                unlessActive: true,
                afterHide: function () {
                    li.getElement('span').set('text', 'Save');
                }
            });

        }, 3000);
    };

    // save changes function
    var saveChanges = function () {

        // already saving
        var message = $('editor-save-message');
        if (!message)
            return;

        var saveData = editor.getValue();

        // prevent box from closing for now
        box.addClass('sticky');
        var message = message.getProperty('value');

        // "saving..."
        $('editor-save-wrapper').innerHTML = tmpl('tmpl-save-spinner', {});
        var btn = $('editor-save-commit');
        btn.innerHTML = 'Comitting changes';
        btn.addClass('progress');

        // save failed callback
        var fail = function (msg) {
            alert('Save failed: ' + msg);

            // switch to /!\
            var i = btn.parentElement.getElement('i');
            i.removeClass('fa-spinner');
            i.removeClass('fa-spin');
            i.addClass('fa-exclamation-triangle');

            // update button
            btn.addClass('failure');
            btn.removeClass('progress');
            btn.innerHTML = 'Save failed';

            closeBoxSoon();
        };

        // successful save callback
        var success = function (data) {
            console.log(data);
            ae.lastSavedData = saveData;

            // switch to checkmark
            var i = btn.parentElement.getElement('i');
            i.removeClass('fa-spinner');
            i.removeClass('fa-spin');
            i.addClass('fa-check-circle');

            // update button
            btn.removeClass('progress');
            btn.addClass(data.displayError ? 'warning' : 'success');
            var text = data.unchanged ?
                'File unchanged' : 'Saved ' + (data.revLatestHash || '').substr(0, 7);
            if (data.displayError)
                text += ' with errors';
            btn.innerHTML = text;

            // show the page display error
            ae.handleWarningsAndError(data.warnings, data.displayError);

            closeBoxSoon();
        };

        // save request
        saveRequest(saveData, message, success, fail);
    };

    // display it
    if (!ae.displayPopupBox(box, 120, li))
        return;

    // on click or enter, save changes
    $('editor-save-commit').addEvent('click', saveChanges);
    $('editor-save-message').onEnter(saveChanges);
    $('editor-save-message').focus();
}

function autosave () {
    if (ae.currentPopup) return; // FIXME
    
    // make it apparent that autosave is occurring
    var li = ae.liForAction('save');
    ae.setLiLoading(li, true, true);
    li.getElement('span').set('text', 'Autosave');
    
    // on fail or success, close the li
    var done = function () {
        setTimeout(function () {
            ae.setLiLoading(li, false, true);
        }, 2000);
        setTimeout(function () {
            li.getElement('span').set('text', 'Save');
        }, 2500);
    };
    
    // attempt to save
    var saveData = editor.getValue();
    saveRequest(saveData, 'Autosave', function (data) { // success
        done();
        ae.lastSavedData = saveData;
        ae.handlePageDisplayResult(data.result);
    }, function (msg) { // failure
        done();
        alert('Save failed: ' + msg);
    });
}

function saveRequest (saveData, message, success, fail) {

    // do the request
    new Request.JSON({
        url: 'func/write-' + (ae.isConfig() ? 'config' : ae.isModel() ? 'model' : 'page'),
        secure: true,
        onSuccess: function (data) {

            // updated without error
            if (data.success)
                success(data);

            // revision error

            // nothing changed
            else if (data.revError && data.revError.match(/no changes|nothing to commit/)) {
                data.unchanged = true;
                success(data);
            }

            // git error
            else if (data.revError)
                fail(data.revError);

            // other error
            else if (data.reason)
                fail(data.reason);

            // not sure
            else
                fail("Unknown error");
        },
        onError: function () {
            fail('Bad JSON reply');
        },
        onFailure: function (data) {
            fail('Request error');
        },
    }).post({
        name:       ae.getFilename(),
        content:    saveData,
        message:    message
    });

    // reset the autosave timer
    resetAutosaveInterval();
}

var autosaveInterval;
function resetAutosaveInterval () {
    clearAutosaveInterval();
    if (a.autosave) {
        autosaveInterval = setInterval(autosave, a.autosave);
    }
}

function clearAutosaveInterval () {
    if (autosaveInterval != null)
        clearInterval(autosaveInterval);
}

})(adminifier);
