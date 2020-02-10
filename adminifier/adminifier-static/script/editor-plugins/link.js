(function (a) {

document.addEvent('editorLoaded', loadedHandler);
document.addEvent('pageUnloaded', unloadedHandler);

var ae;
function loadedHandler () {
    ae = a.editor;

    // add toolbar functions
    ae.addToolbarFunctions({
        link:       displayLinkHelper
    });

    // add keyboard shortcut
    ae.addKeyboardShortcuts([
        [ 'Ctrl-K', 'Command-K', 'link']
    ]);
}

function unloadedHandler () {
    document.removeEvent('editorLoaded', loadedHandler);
    document.removeEvent('pageUnloaded', unloadedHandler);
}

// LINK HELPER

function displayLinkHelper () {
    var li  = ae.liForAction('link');
    var box = ae.createPopupBox(li);
    ae.fakeAdopt(box);

    box.innerHTML = tmpl('tmpl-link-helper', {});

    // first input, second input, button title, left delimiter, right delimiter
    $('editor-link-type-internal').store('linkInfo', [
        'Display text', 'Page target', 'Insert page link',
        'My Page',
        '[', ']'
    ]);
    $('editor-link-type-external').store('linkInfo', [
        'Display text', 'External page target', 'Insert external page link',
        'Some Page',
        '!', '!'
    ]);
    $('editor-link-type-category').store('linkInfo', [
        'Display text', 'Category target', 'Insert category link',
        'News',
        '~', '~'
    ]);
    $('editor-link-type-url').store('linkInfo', [
        'Display text', 'URL target', 'Insert URL',
        'http://www.example.com',
        '$', '$'
    ]);

    // switch between link types
    var activeType = $('editor-link-type-internal');
    $$('.editor-link-type').each(function (type) {
        type.addEvent('click', function () {

            // set the active type
            if (type.hasClass('active')) return;
            activeType.removeClass('active');
            type.addClass('active');
            activeType = type;

            // update the text
            var info = type.retrieve('linkInfo');
            $('editor-link-title1').innerHTML = info[1];
            $('editor-link-title2').innerHTML = info[0];
            $('editor-link-insert').innerHTML = info[2];
            $('editor-link-target').setProperty('placeholder', info[3]);

        });
    });

    // selected text = display text
    // choose a word if there is no actual selection
    var r = ae.getSelectionRanges();
    editor.selection.setRange(r.select);
    var selected = editor.session.getTextRange(r.select);

    if (selected.trim().length) {
        $('editor-link-display').setProperty('value', selected);
        $('editor-link-target').setProperty('value', selected);
    }

    // insert link function
    var insertLink = function () {
        var displayText = $('editor-link-display').getProperty('value'),
            targetText  = $('editor-link-target').getProperty('value'),
            leftDel     = activeType.retrieve('linkInfo')[4],
            rightDel    = activeType.retrieve('linkInfo')[5];

        // make sure requirements were met
        if (!displayText.length || !targetText.length) {
            alert('Please specify display text and target');
            return;
        }

        // one or two parts, depending on if display == target
        var inner = displayText;
        if (displayText.toLowerCase() != targetText.toLowerCase())
            inner += ' | ' + targetText;

        var complete = '[' + leftDel + ' ' + inner + ' ' + rightDel + ']';
        editor.insert(complete);
        ae.closePopup(box);
    };

    // insert on click or enter
    $('editor-link-insert').addEvent('click', insertLink);
    $('editor-link-target').onEnter(insertLink);
    $('editor-link-display').onEnter(insertLink);

    ae.displayPopupBox(box, 215, li);
    $('editor-link-target').focus();
}

})(adminifier);
