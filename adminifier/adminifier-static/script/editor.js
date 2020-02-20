var Range, Search, editor;
(function (a) {

var ae = adminifier.editor = {};

var plugins = [
    'text-formatting',
    'save',
    'link',
    'page-options',
    'revision'
];

// PAGE EVENTS

// upon adding editorLoaded event, call if the editor is already loaded
Element.Events.editorLoaded = {
    onAdd: function (fn) {
        if (ae.editorLoaded)
            fn.call(this);
    }
};

document.addEvent('pageScriptsLoaded',          pageScriptsLoadedHandler);
document.addEvent('pageUnloaded',               pageUnloadedHandler);
document.addEvent('keyup',                      handleEscapeKey);
document.addEvent('editorLoaded',               editorLoadedHandler);
window.addEvent('resize',                       adjustCurrentPopup);

function pageScriptsLoadedHandler () {
    console.log('Editor script loaded');
    setupToolbar();
    window.onbeforeunload = confirmOnPageExit;
    
    // load required ace resources
    Range  = ace.require('ace/range').Range;
    Search = ace.require('ace/search').Search;

    // initialize editor
    editor = ace.edit('editor');
    ae.lastSavedData = editor.getValue();

    // render editor
    var themeName = adminifier.themeName || 'twilight';
    editor.setTheme('ace/theme/' + themeName);
    // editor.session.setMode('ace/mode/wikifier');
    editor.on('input', handleInput);
    setTimeout(function () { editor.resize(); }, 500);

    // listen for clicks to navigate away
    document.body.addEvent('click', clickOutHandler);

    // fire editor loaded event
    if (!ae.editorLoaded) {
        ae.editorLoaded = true;
        document.fireEvent('editorLoaded');
    }
}

function pageUnloadedHandler () {
    console.log('Unloading editor script');
    document.removeEvent('pageScriptsLoaded',   pageScriptsLoadedHandler);
    document.removeEvent('pageUnloaded',        pageUnloadedHandler);
    document.removeEvent('keyup',               handleEscapeKey);
    document.removeEvent('editorLoaded',        editorLoadedHandler);
    window.removeEvent('resize',                adjustCurrentPopup);
    document.body.removeEvent('click',          clickOutHandler);
    delete window.onbeforeunload;
    delete a.editor;
}

function editorLoadedHandler () {
    ae.updatePageTitle();
    ae.resetSelectionAtTopLeft();
    if (a.currentJSONMetadata && a.currentJSONMetadata.display_result)
        ae.handlePageDisplayResult(a.currentJSONMetadata.display_result);
    plugins.each(function (plugin) {
        a.loadScript('editor-plugins/' + plugin);
    });
}

// FIXME: issue #31
function clickOutHandler (e) {
    if (!ae.hasUnsavedChanges())
        return;
    var findParent = function (tagname, el) {
        if ((el.nodeName || el.tagName).toLowerCase() === tagname.toLowerCase())
            return el;
        while (el = el.parentNode) {
            if ((el.nodeName || el.tagName).toLowerCase() === tagname.toLowerCase())
                return el;
        }
        return null;
    }
    var from = findParent('a', e.target);
    if (from) {
        e.preventDefault();
        alert('You have unsaved changes.');
    }
}

// FIXME: issue #31
function confirmOnPageExit (e) {
    if (!ae.hasUnsavedChanges())
        return;
    var message = 'You have unsaved changes.';
    if (e) e.returnValue = message;
    return message;
}

// escape key pressed
function handleEscapeKey (e) {
    if (e.key != 'esc') return;
    console.log('handle escape');

    // if there's a popup, exit it maybe
    if (ae.currentPopup) closeCurrentPopup({
        unlessSticky: true,
        reason: 'Escape key'
    });
}

function handleInput () {

    // update undo
    var um = editor.session.getUndoManager();
    if (um.hasUndo())
        $('toolbar-undo').removeClass('disabled')
    else
        $('toolbar-undo').addClass('disabled');

    // update redo
    if (um.hasRedo())
        $('toolbar-redo').removeClass('disabled');
    else
        $('toolbar-redo').addClass('disabled');


    // current line
    var lineText = editor.session.getLine(editor.getSelectionRange().start.row);

    // we're changing the title.
    if (lineText.match(ae.expressions.pageTitle)) {
        ae.updatePageTitle(editor.selection.getLineRange());
    }
}

// EXPORTS

// search expressions for variables
ae.expressions = {
    pageTitle:      new RegExp('\\s*^@page\\.title:(.*)$'),
    keyValueVar:    new RegExp('^\\s*@page\\.(\\w+):(.*?)(;?)\\s*$'),
    boolVar:        new RegExp('^\\s*@page\\.(\\w+);\\s*$'),
    category:       new RegExp('^\\s*@category\\.([\\w\\.]+);\\s*$')
};

// toolbar click functions
ae.toolbarFunctions = {
    undo:   function () { editor.undo(); },
    redo:   function () { editor.redo(); }
};

// add toolbar function
ae.addToolbarFunctions = function (funcs) {

    // on domready, make the associated toolbar buttons available
    document.addEvent('domready', function () { 
        Object.keys(funcs).each(function (funcName) {
            var li = $$('.editor-toolbar li.hidden[data-action=' + funcName + ']');
            if (!li || !li[0]) return;
            li[0].setStyle('display', 'inline-block');
        });
    });

    // remember the responders
    Object.append(ae.toolbarFunctions, funcs);
}

// get current page filename
ae.getFilename = function () {
    return $('editor').getProperty('data-file');
};

// true if the file being edited is a model
ae.isModel = function () {
    return $('editor').getProperty('data-model') != null;
};

// returns the page title text, with any escapes accounted for.
// returns nothing if the @page.title can't be found.
// range is optional
ae.getPageTitle = function (range) {
    var found = ae.findPageVariable(ae.expressions.pageTitle, range);
    if (!found)
        return;
    return found.text;
};

// update the page title to whatever @page.title says
// range is optional
ae.updatePageTitle = function (range) {
    var title = ae.getPageTitle(range);
    if (typeof title == 'undefined')
        return;
    a.updatePageTitle(title.length ? title : ae.getFilename());
};

// bind keyboard shortcuts to functions
ae.addKeyboardShortcuts = function (cuts) {
    // c = [ windows shortcut, mac shortcut, action ]
    cuts.each(function (c) {
        editor.commands.addCommand({
            name: c[2],
            bindKey: {
                windows: c[0],
                mac:     c[1]
            },
            exec: ae.toolbarFunctions[ c[2] ]
        });
    });
};

// returns true if there are unsaved changes
ae.hasUnsavedChanges = function () {
    return editor.getValue() != ae.lastSavedData;
};

// close a popup box
ae.closePopup = function (box, opts) {
    if (!ae.currentPopup)
        return;
    if (ae.currentPopup != box) {
        console.warn(
            'Attempted to close a box other than the current one',
            box,
            ae.currentPopup
        );
        return;
    }
    closeCurrentPopup(opts);
};

// find a page variable matching the provided expression within the given range
// range is optional, defaults to the whole file
ae.findPageVariable = function (exp, range) {
    var search = new Search().set({
        needle: exp,
        regExp: true,
        range:  range,
        backwards: true // prefer the last occurrence
    });
    return; // FIXME: search.find() is making editor hang.
    var found = search.find(editor.session);
    if (!found)
        return;
    var string = editor.session.getTextRange(found);
    var escaped    = false,
        inTitle    = false,     inName    = false,
        foundText  = '',        foundName = '',
        startIndex = 0,         endIndex  = 0;

    for (var i = 0; ; i++) {
        var char = string[i];

        // made it to the end without finding unescaped semicolon
        if (typeOf(char) == 'null') {
            endIndex = i;
            break;
        }

        // escapes
        var escaped = string[i - 1] == '\\';
        if (char == '\\' && !escaped) {
            continue;
        }

        // now we're in the title
        if (!startIndex && char == ':') {
            inName  = false;
            inTitle = true;
            startIndex = i + 1;
            continue;
        }

        if (!startIndex && char == '.' && !inName) {
            inName = true;
            continue;
        }

        // if we're in the title but no text has been found,
        // this is just spacing before the actual title
        if (inTitle && !foundText.length && char == ' ') {
            startIndex++;
            continue;
        }

        // ending the title
        if (inTitle && !escaped && char == ';') {
            endIndex = i;
            break;
        }

        // ending a bool var
        if (inName && char == ';')
            break;

        if (inTitle)
            foundText += char;
        else if (inName)
            foundName += char;
    }

    // offset on the line
    startIndex += found.start.column;
    endIndex   += found.start.column;

    return {
        name: foundName,
        text: foundText,
        range: new Range(found.start.row, startIndex, found.end.row, endIndex)
    };
};

// insert a blank line at the current position,
// unless there is already a blank line there
ae.insertBlankLineMaybe = function () {

    // select the line following the var insertion.
    editor.selection.selectLine();

    // if there is text on the line, insert a blank line before it.
    if (editor.getSelectedText().trim().length) {
        var r = editor.selection.getRange();
        editor.selection.setRange(new Range(
            r.start.row, 0,
            r.start.row, 0
        ));
        editor.insert('\n');
        return true;
    }

    return false;
};

// remove extra newlines at the current position
// returns number of lines removed
ae.removeExtraNewlines = function () {
    var oldLength = editor.session.getLength();
    var oldRange  = editor.selection.getRange();

    // not on a newline currently
    var i = new Range(
        oldRange.end.row, 0,
        oldRange.end.row, Infinity
    );
    if (editor.session.getTextRange(i).trim().length)
        return;

    // remove newlines
    editor.selection.selectLine();
    while (editor.selection.getRange().start.row != 0) {
        var text = editor.getSelectedText();

        // the line has length, so this is the end of it.
        if (text.trim().length) {
            editor.insert('\n' + text);
            break;
        }

        editor.removeLines();
        editor.selection.selectLine();
    }

    // put the cursor back where it was
    editor.selection.setRange(oldRange);
    return oldLength - editor.session.getLength();
};

// remove all lines within the provided ranges
ae.removeLinesInRanges = function (ranges) {
    if (!ranges || !ranges.length)
        return;

    var rows = [], smallest = Infinity, biggest = -1;
    ranges.each(function (i) {
        rows[i.start.row] = true;
        rows[i.end.row]   = true;
        biggest  = Math.max(biggest,  i.start.row, i.end.row);
        smallest = Math.min(smallest, i.start.row, i.end.row);
    });

    smallest--;
    var lastLine;
    for (var i = biggest; i >= smallest; i--) {

        // if the row does not exist, this is the end of a continuous range
        if (!rows[i]) {
            if (typeof lastLine != 'undefined') {
                editor.session.doc.removeFullLines(i + 1, lastLine);
                lastLine = undefined;
            }
            continue;
        }

        if (typeof lastLine == 'undefined')
            lastLine = i;
    }
};

// move the cursor to the top left position
ae.resetSelectionAtTopLeft = function () {
    editor.selection.setRange(new Range(0, 0, 0, 0));
    editor.focus();
};

// create and return a new popup box
ae.createPopupBox = function (li) {
    var box = new Element('div', { class: 'editor-popup-box' });
    if (li && li.hasClass('right'))
        box.addClass('right');
    return box;
};

// present a popup box
ae.displayPopupBox = function (box, height, li) {
    
    // can't open the li
    if (!openLi(li))
        return false;
    
    // add to body
    document.body.appendChild(box);
    
    // set as current popup, initial adjustment
    ae.currentPopup = box;
    box.store('li', li);
    adjustCurrentPopup();
    
    // animate open
    box.set('morph', { duration: 150 });
    if (height == 'auto') {
        height = window.innerHeight - parseInt(box.getStyle('top'));
        box.set('morph', {
            onComplete: function () { box.setStyle('height', 'auto'); }
        });
        box.morph({ height: height + 'px' });
    }
    else if (typeof height == 'number')
        box.morph({ height: height + 'px' });
    else
        box.setStyle('height', height);
        
    return true;
};

// start the loading animation on a toolbar item
ae.setLiLoading = function (li, loading, noCircle) {
    var i = li.getElement('i');
    if (loading) {
        if (!openLi(li))
            return;
        li.addClass('sticky');
        if (!noCircle) {
            i.store('oldClasses', i.get('class'));
            i.set('class', 'fa fa-circle-notch fa-spin');
        }
        return true;
    }
    li.removeClass('sticky');
    if (!noCircle)
        i.set('class', i.retrieve('oldClasses'));
    if (!ae.currentPopup)
        closeCurrentLi();
    return true;
};

// find an appropriate range for selection based on the cursor position.
// this is used for text formatting
ae.getSelectionRanges = function () {

    // find the current selection
    var selectRange = editor.selection.getRange();
    var originalRange = selectRange;

    // if there is no actual selection (just a cursor position),
    // use the word range. but only if it's in a word (check strlen).
    // also check if it strictly non-word chars, such as a symbol.
    if (selectRange.isEmpty()) {
        var wordRange = editor.selection.getWordRange();
        var word = editor.session.getTextRange(wordRange).trim();
        if (word.length && !word.match(/^\W*$/))
            selectRange = wordRange;
    }

    return {
        original: originalRange,
        select: selectRange
    }
};

// this is useful for replacing a range of text
// with something that surrounds it, such as [b]...[/b]
// because it re-selects the original selection
// after performing the operation: [b]<sel>...</sel>[/b]
ae.replaceSelectionRangeAndReselect = function (ranges, leftOffset, newText) {
    var selectRange = ranges.select,
        originalRange = ranges.original;

    // replace the text
    editor.session.replace(selectRange, newText);

    // return to the original selection
    editor.selection.setRange(new Range(
        originalRange.start.row,
        originalRange.start.column + leftOffset,
        originalRange.end.row,
        originalRange.end.column + leftOffset
    ));
};

var brackets = {
    ")": "(",
    "(": ")",
    "]": "[",
    "[": "]",
    "{": "}",
    "}": "{"
};

// find nearest bracket range for the given opening bracket type
// this uses private ace methods, so it might break
// getBracketRange('{') or [ or (
ae.getBracketRange = function (openingBracket, pos) {
    if (!pos)
        pos = editor.getCursorPosition();
    pos.column++; // fixes if it's at the actual position of the bracket
    var openPos  = editor.session.$findOpeningBracket(brackets[openingBracket], pos);
    var closePos = editor.session.$findClosingBracket(openingBracket, pos);
    if (!openPos || !closePos) return;
    return new Range(
        openPos.row,
        openPos.column + 1, // don't include the first bracket itself
        closePos.row,
        closePos.column
    );
};

// returns info about the innermost block at the given position.
// returns undefined if we are at the document level.
// if no position is provided, uses the current cursor position.
ae.getBlockInfo = function (pos) {
    var bracketRange = ae.getBracketRange('{', pos);
    if (!bracketRange) return;
    
    // get the text for the line, dropping anything at/after the bracket
    var lineText = editor.session.getLine(bracketRange.start.row);
    lineText = lineText.substr(0, bracketRange.start.column - 1);

    // find the block type and title
    var chars = lineText.split(''),
        charsScanned = 0,
        inBlockName  = 0,
        blockType    = '',
        blockName    = '';
        
    // scan backwards
    while (chars.length) {
        var char = chars.pop();
        charsScanned++;
        
        // entering block name
        if (char == ']') {
            if (!inBlockName++)
                continue;
            // if it was at 0, we just entered the block name
        }
        
        // exiting block name
        if (char == '[') {
            if (!--inBlockName)
                continue;
            // if it was 1 or more, we're still in the block name
        }
        
        // we are in the block name, so add this character to the front
        if (inBlockName)
            blockName = char + blockName;
        
        // this char is part of the block type at this point
        else if (char.match(/[\w\-\$\.]/)) {
            blockType = char + blockType;
            continue;
        }
        
        // tilde can terminate the block type
        else if (char == '~' && blockType.length)
            continue;
            
        // space between things
        else if (char.match(/\s/))
            continue;
            
        // give up
        else {
            charsScanned--;
            break;
        }
    }
    
    // split classes by dot
    var blockClasses = blockType.split('.');
    blockType = blockClasses.shift();
    
    return {
        bracketRange:   bracketRange,
        type:           blockType,
        name:           blockName,      // for some block types, this may
        classes:        blockClasses    // include formatting codes
    };
};

ae.getBlocksInfo = function (pos) {
    if (!pos)
        pos = editor.getCursorPosition();
    var blocks = [], current;
    while (current = ae.getBlockInfo(pos)) {
        blocks.push(current);
        pos = current.bracketRange.start;
        pos.column -= 1;
    }
    return blocks;
};

// returns the toolbar item for the given action name
ae.liForAction = function (action) {
    return document.getElement('li[data-action="' + action + '"]');
};

// adopts an element to an invisible container in the DOM so that dimension
// properties become available before displaying it
ae.fakeAdopt = a.fakeAdopt;

// sets the selection to the page title variable, if it can be found
ae.selectPageTitle = function () {
    var found = ae.findPageVariable(ae.expressions.pageTitle);
    if (!found)
        return;
    editor.selection.setRange(found.range);
};

// updates warnings and errors on page save
ae.handlePageDisplayResult = function (res) {
    if (!res)
        return;

    // annotate errors and warnings
    var annotations = [];
    var addAnnotation = function (str, type, isDraft) {
        var match = str.match(/^Line (\d+):(\d+):([\s\S]+)/);
        if (!match) {
            if (!isDraft)
                alert(str);
            return;
        };
        var row = match[1] - 1,
            col = match[2],
            errorText = match[3].trim().replace(/\n/g, '\u2424');
        annotations.push({
            row:    row,
            column: col,
            text:   errorText,
            type:   type
        });
    };

    // warnings
    if (res.warnings != null)
        res.warnings.each(function (msg) { addAnnotation(msg, "warning"); });

    // error
    if (res.error != null)
        addAnnotation(res.error, "error", res.draft);

    if (annotations.length)
        editor.session.setAnnotations(annotations);
    else
        editor.session.clearAnnotations();
};

// returns a function for wrapping the current selection with formatting tags
// e.g. ae.wrapTextFunction('b')
ae.wrapTextFunction = wrapTextFunction;
function wrapTextFunction (type) {
    return function () {

        var r = ae.getSelectionRanges();
        var selectRange = r.select,
            originalRange = r.original;
        editor.selection.setRange(selectRange);

        // dtermine the new text
        var terminator  = type.length > 1 ? '' : type;
        var leftSide    = '[' + type + ']';
        var newText     = leftSide + editor.getSelectedText() + '[/' + terminator + ']';

        // replace the text and select the original text
        ae.replaceSelectionRangeAndReselect(r, leftSide.length, newText);

        closeCurrentPopup();
    };
}

// INTERNAL POPUP FUNCTIONS

// close the current popup box
function closeCurrentPopup (opts) {
    var box = ae.currentPopup;
    if (!box)
        return;
    if (!opts)
        opts = {};

    // check if sticky
    if (opts.unlessSticky && box.hasClass('sticky')) {
        console.log('Keeping popup open: Sticky');
        return;
    }

    // check if mouse is over it.
    // note this will only work if the box has at least one child with
    // the hover selector active
    if (opts.unlessActive && box.getElement(':hover')) {
        console.log('Keeping popup open: Active');

        // once the mouse exits, close it
        box.addEvent('mouseleave', function () {
            ae.closePopup(box, opts);
        });

        return;
    }

    // Safe point - we will close the box.
    if (opts.reason)
        console.log('Closing popup: ' + opts.reason);
        
    // run destroy callback
    if (ae.onPopupDestroy) {
        ae.onPopupDestroy(box);
        delete ae.onPopupDestroy;
    }

    closeCurrentLi();
    box.set('morph', {
        duration: 150,
        onComplete: function () {
            ae.currentPopup = null;
            if (box) box.destroy();
            if (opts.afterHide) opts.afterHide();
        }
    });
    
    if (box.getStyle('height') == 'auto')
        box.setStyle('height', window.innerHeight - parseInt(box.getStyle('top')));
    box.morph({ height: '0px' });
}

// move a popup when the window resizes
function adjustCurrentPopup () {
    
    // no popup open
    var box = ae.currentPopup;
    if (!box)
        return;
        
    var li = box.retrieve('li');
    var rect = li.getBoundingClientRect();
            
    // adjust top no matter what
    box.setStyle('top', rect.top + li.offsetHeight);
    
    // this is a fixed box; don't change left or right
    if (box.hasClass('fixed'))
        return;
    
    // set left or right
    box.setStyle('left',
        box.hasClass('right') ?
        rect.right - 300 :
        rect.left
    );
};

// close current popup on click outside
function bodyClickPopoverCheck (e) {

    // no popup is displayed
    if (!ae.currentPopup)
        return;

    // clicked within the popup
    if (e.target == ae.currentPopup || ae.currentPopup.contains(e.target))
        return;

    // the target says not to do this
    if (e.target && e.target.hasClass('no-close-popup'))
        return;

    closeCurrentPopup({
        unlessSticky: true,
        unlessActive: true,
        reason: 'Clicked outside the popup'
    });
}

// INTERNAL TOOLBAR ITEM FUNCTIONS

// animated open of a toolbar item
function openLi (li) {

    // if a popup is open, ignore this.
    if (ae.currentPopup)
        return false;

    // if another one is animating, force it to instantly finish
    if (ae.currentLi && ae.currentLi != li) {
        if (ae.currentLi.hasClass('sticky'))
            return;
        if (ae.currentLi != li)
            closeCurrentLi();
    }
    
    // redo the closing morph so that the timing is
    // perfectly lined up with the opening of the new one
    if (ae.closingLi) {
        var morph = ae.closingLi.get('morph');
        morph.cancel();
        ae.closingLi.morph({
            width: '15px',
            backgroundColor: '#696969'
        });
        delete ae.closingLi;
    }

    // animate this one
    li.morph({
        width: '100px',
        backgroundColor: '#2096ce',
        onComplete: function () { delete ae.openingLi; }
    });
    ae.openingLi = li;
    ae.currentLi = li;
    return true;
}

// animated close the current toolbar item
function closeCurrentLi () {
    if (!ae.currentLi)
        return;
    if (ae.currentLi.hasClass('sticky'))
        return;
    ae.currentLi.morph({
        width: '15px',
        backgroundColor: '#696969',
        onComplete: function () { delete ae.closingLi; }
    });
    ae.closingLi = ae.currentLi;
    delete ae.currentLi;
}

// add events for toolbar items
function setupToolbar () {
    document.body.addEvent('click', bodyClickPopoverCheck);

    // switch between buttons
    $$('ul.editor-toolbar li').each(function (li) {

        // hover animation
        li.set('morph', { duration: 150 });
        li.addEvent('mouseenter', function () {
            openLi(li);
        });

        // clicked
        li.addEvent('click', function (e) {

            // disabled or another popup is being displayed
            if (li.hasClass('disabled') || ae.currentPopup)
                return;

            // no action? not sure what to do
            var action = ae.toolbarFunctions[ li.getAttribute('data-action') ];
            if (!action)
                return;

            e.stopPropagation(); // don't let click out handler see this
            action();
        });

    });

    // leaving the toolbar, close it
    $$('ul.editor-toolbar').addEvent('mouseleave', function () {
        if (ae.currentLi && !ae.currentPopup)
            closeCurrentLi();
    });
}

})(adminifier);
