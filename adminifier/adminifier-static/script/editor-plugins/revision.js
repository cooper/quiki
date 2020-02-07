(function (a) {

document.addEvent('editorLoaded', loadedHandler);
document.addEvent('pageUnloaded', unloadedHandler);

var ae;
function loadedHandler () {
    ae = a.editor;

    // add toolbar functions
    Object.append(ae.toolbarFunctions, {
        view:       openPageInNewTab,
        revisions:  displayRevisionViewer
    });
    
    // disable view button for models
    if (ae.isModel())
        ae.liForAction('view').addClass('disabled');
        
    // load diff2html
    a.loadScript('diff2html');
}

function unloadedHandler () {
    document.removeEvent('editorLoaded', loadedHandler);
    document.removeEvent('pageUnloaded', unloadedHandler);
}

// VIEW PAGE BUTTON

function openPageInNewTab () {
    if (ae.isModel())
        return;
    var root = a.wikiPageRoot;
    var pageName = ae.getFilename().replace(/\.page$/, '');
    window.open(root + pageName);
}

// REVISION VIEWER

function displayRevisionViewer () {
    
    // make the li stay open until finish()
    var li = ae.liForAction('revisions');
    ae.setLiLoading(li, true);

    // create the box
    var box = ae.createPopupBox(li);
    box.setStyles({ right: 0, bottom: 0 });
    box.addClass('fixed');
    box.innerHTML = tmpl('tmpl-revision-viewer', {});
    var container = box.getElement('#editor-revisions');
    
    // populate and display it
    var finish = function (data) {
        ae.setLiLoading(li, false);
        if (!box)
            return;
        if (!data.success) {
            alert(data.error);
            return;
        }
        if (!data.revs) {
            alert('No revisions');
            return;
        }
        data.revs.each(function (rev) {
            var row = new Element('div', {
                class: 'editor-revision-row',
                'data-commit': rev.id
            });
            row.innerHTML = tmpl('tmpl-revision-row', rev);
            row.addEvent('click', function (e) {
                handleDiffClick(box, row, e);
            });
            container.appendChild(row);
        });
        ae.displayPopupBox(box, 'auto', li);
    };

    // request revision history
    var req = new Request.JSON({
        url: 'functions/page-revisions.php' + (ae.isModel() ? '?model' : ''),
        onSuccess: finish,
        onFailure: function () {
            finish({ error: 'Failed to fetch revision history' });
        },
    }).post({
        page: ae.getFilename()
    });
}

function handleDiffClick (box, row, e) {
    var msg = row.getElement('b').get('text').trim();
    var prevRow = row.getNext();
    
    // display overlay
    var overlay = new Element('div', { class: 'editor-revision-overlay' });
    overlay.appendChild(row.clone().addClass('preview'));
    overlay.innerHTML += tmpl('tmpl-revision-overlay', {});
    box.appendChild(overlay);
    
    // button clicks
    var funcs = [
        
        // view on wiki
        function () {
            alert('Unimplemented');
        },
        
        // view source
        function () {
            alert('Unimplemented');
        },
        
        // diff current
        function () {
            if (!row.getPrevious()) {
                alert('This is the current version');
                return;
            }
            displayDiffViewer(
                box,
                row.get('data-commit'),
                null,
                msg,
                'current'
            );
        },
        
        // diff previous
        function () {
            if (!prevRow) {
                alert('This is the oldest revision');
                return;
            }
            displayDiffViewer(
                box,
                prevRow.get('data-commit'),
                row.get('data-commit'),
                msg,
                'previous'
            );
        },
        
        // revert
        function () {
            alert('Unimplemented');
        },
        
        // restore
        function () {
            alert('Unimplemented');
        },
        
        // back
        function () {
            overlay.setStyle('display', 'none');
            setTimeout(function () { overlay.destroy(); }, 100);
        }
    ];
    overlay.getElements('.editor-revision-diff-button').each(function (but, i) {
        but.addEvent('click', funcs.shift());
    });
}

// DIFF VIEWER

function displayDiffViewer (box, from, to, message, which) {
    box.addClass('sticky');
    var finish = function (data) {
        
        // something wrong happened
        if (!data.success) {
            alert(data.error);
            return;
        }
        
        // no differences
        if (typeof data.diff == 'undefined') {
            alert('No changes');
            return;
        }
        
        // run diff2html
        var diffHTML, diffWindow;
        var runDiff = function (split) {
            diffHTML = Diff2Html.getPrettyHtml(data.diff, {
                outputFormat: split ? 'side-by-side' : 'line-by-line'
            });
            if (diffWindow) diffWindow.content.innerHTML = diffHTML;
        };
        runDiff();
        
        // create a modal window to show the diff in
        diffWindow = new ModalWindow({
            icon:           'clone',
            title:          "Compare '" + message + "' to " + which,
            padded:         true,
            html:           diffHTML,
            width:          '90%',
            doneText:       'Done',
            id:             'editor-diff-window',
            autoDestroy:    true,
            onDone:         function () {
                setTimeout(function () { box.removeClass('sticky'); }, 100);
            },
        });
        
        // switch modes
        var but, split;
        but = diffWindow.addButton('Split view', function () {
            if (split) {
                runDiff(false);
                split = false;
                but.set('text', 'Split view');
                return;
            }
            runDiff(true);
            split = true;
            but.set('text', 'Unified view');
        });
        
        diffWindow.show();
    };

    // request revision history
    var req = new Request.JSON({
        url: 'functions/page-diff.php' + (ae.isModel() ? '?model' : ''),
        onSuccess: finish,
        onFailure: function () {
            finish({ error: 'Failed to fetch page diff' });
        },
    }).post({
        page: ae.getFilename(),
        from: from,
        to: to
    });
}

})(adminifier);
