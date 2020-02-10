(function (a) {

document.addEvent('editorLoaded', loadedHandler);
document.addEvent('pageUnloaded', unloadedHandler);

var ae;
function loadedHandler () {
    ae = a.editor;

    // add toolbar functions
    ae.addToolbarFunctions({
        save:       displaySaveHelper
    });

    // add keyboard shortcuts
    ae.addKeyboardShortcuts([
        [ 'Ctrl-S', 'Command-S',    'save'      ]
    ]);
}

function unloadedHandler () {
    document.removeEvent('editorLoaded', loadedHandler);
    document.removeEvent('pageUnloaded', unloadedHandler);
}

})(adminifier);
