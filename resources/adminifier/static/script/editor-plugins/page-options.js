(function (a) {

document.addEvent('editorLoaded', loadedHandler);
document.addEvent('pageUnloaded', unloadedHandler);

var ae;
function loadedHandler () {
    ae = a.editor;

    // add toolbar function
    ae.addToolbarFunctions({
        options:    displayPageOptionsWindow
    });
}

function unloadedHandler () {
    document.removeEvent('editorLoaded', loadedHandler);
    document.removeEvent('pageUnloaded', unloadedHandler);
}

// PAGE OPTIONS

function displayPageOptionsWindow () {
    if ($('editor-options-window'))
        return;

    // find and store the current values
    var foundOpts = findPageOptions();
    var foundCats = findPageCategories();
    var foundOptsValues = Object.map(foundOpts.found, function (value) {
        return value.value;
    });

    // TODO: check explicitly for whether it's a quiki markup page or model.
    // maybe add support for markdown stuff too. if none of these, the options
    // button should never have been displayed on the toolbar or been disabled.

    // create the options window
    var optionsWindow = new ModalWindow({
        icon:           'cog',
        title:          ae.isModel() ? 'Model options' : 'Page options',
        html:           tmpl(ae.isModel() ?
            'tmpl-model-options' : 'tmpl-page-options', foundOptsValues),
        padded:         true,
        id:             'editor-options-window',
        autoDestroy:    true,
        onDone:         updatePageOptions
    });

    // update page title as typing
    var titleInput = optionsWindow.content.getElement('input.title');
    titleInput.addEvent('input', function () {
        var title = titleInput.getProperty('value');
        a.updatePageTitle(title.length ? title : ae.getFilename());
    });

    // category list
    if (!ae.isModel()) {

        // get categories from list
        var getCategories = optionsWindow.getCategories = function () {
            var cats = optionsWindow.content.getElements('tr.category');
            return cats.map(function (tr) {
                return tr.retrieve('safeName');
            });
        };

        // add category to liust
        var addCategoryTr = optionsWindow.content.getElement('.add-category');
        var addCategory = function (catName) {

            // category already exists
            var safeName = a.safeName(catName);
            if (getCategories().contains(safeName))
                return;

            // is it a main page?
            var match = catName.match(/(.*)\.main$/), visibleName;
            if (match)
                visibleName = match[1] + ' (main page)';
            else
                visibleName = catName;

            // create the row
            var tr = new Element('tr', {
                class: 'category',
                html:  tmpl('tmpl-page-category', { catName: visibleName })
            });

            // on click, delete the category
            tr.addEvent('click', function () {
                this.destroy();
            });

            tr.store('safeName', safeName);
            tr.inject(addCategoryTr, 'before');
        };

        // on enter of category input, add the category.
        addCategoryTr.getElement('input').onEnter(function () {
            var catName = this.get('value').trim();
            if (!catName.length)
                return;
            addCategory(catName);
            this.set('value', '');
        });

        // add initial categories
        foundCats.found.each(addCategory);
    }

    // store state in the options window
    optionsWindow.foundOptsValues = foundOptsValues;
    optionsWindow.foundOpts = foundOpts;
    optionsWindow.foundCats = foundCats;

    // show it
    optionsWindow.show();
}

function updatePageOptions () {

    // editor is read-only
    if (ae.isReadOnly())
        return;

    // replace old option values with new ones
    var container = this.container;
    var newOpts = Object.merge({},
        this.foundOptsValues,
        Object.filter(Object.map({
            title:      [ 'input.title',    'value'     ],
            author:     [ 'input.author',   'value'     ],
            draft:      [ 'input.draft',    'checked'   ]
        }, function (value) {
            var el = container.getElement(value[0]);
            if (!el) return;
            return el.get(value[1]);
        }), function (value) {
            return typeof value != 'undefined';
        })
    );

    // get new categories
    var newCats = this.getCategories ? this.getCategories() : [];

    var removeRanges = [this.foundOpts.ranges, this.foundCats.ranges].flatten();
    ae.removeLinesInRanges(removeRanges);
    insertPageOptions(newOpts, newCats);
}

function insertPageOptions (newOpts, newCats) {

    // this will actually be passed user input
    var optsString = generatePageOptions(newOpts);

    // inject the new lines at the beginning
    var pos = { row: 0, column: 0 };
    pos = editor.session.insert(pos, optsString);

    // after inserting, the selection will be the line following
    // the insertion at column 0.

    // now check for categories
    if (newCats.length) {
        pos = editor.session.insert(pos, '\n');
        newCats.sort().each(function (catName) {
            pos = editor.session.insert(
                pos,
                '@category.' + catName + ';\n'
            );
        });
    }

    // above this point, the selection has not been affected

    // set the current selection to the insert position
    var oldRange = editor.selection.getRange();
    editor.selection.setRange(new Range(
        pos.row, 0,
        pos.row, 0
    ));

    // remove extra newlines at the new position; set the selection to where
    // it was originally by shifting up by number of lines removed
    var removed = ae.removeExtraNewlines();
    editor.selection.setRange(new Range(
        oldRange.start.row - removed, oldRange.start.column,
        oldRange.end.row   - removed, oldRange.end.column
    ));
}

function pageVariableFromRange (range, exp, bool) {
    var text  = editor.session.getTextRange(range);
    var match = ae.findPageVariable(exp, range);
    if (!match)
        return;
    return {
        name:   match.name,
        text:   match.text,
        value:  bool ? true : match.text.trim(),
        range:  match.range
    };
};

function findVariables (found, exp, bool) {
    var search = new Search().set({ needle: exp, regExp: true });

    // find each thing
    var ranges = search.findAll(editor.session);
    ranges.each(function (i) {
        var res = pageVariableFromRange(i, exp, bool);
        if (res) found[res.name] = res;
    });

    return ranges;
}

function findPageOptions () {
    var found = {}, ranges = [];
    ranges.combine(findVariables(found, ae.expressions.keyValueVar));
    ranges.combine(findVariables(found, ae.expressions.boolVar, true));
    return {
        found:  found,
        ranges: ranges
    };
}

function findPageCategories () {
    var found = {}, ranges = [];
    ranges.combine(findVariables(found, ae.expressions.category, true));
    return {
        found:  Object.keys(found),
        ranges: ranges
    };
}

function generatePageOptions (opts) {
    var string  = '',
        biggest = 9,
        done    = {};
    var updateBiggest = function (length, ret) {
        var maybeBigger = length + 5;
        if (maybeBigger > biggest)
            biggest = maybeBigger;
        return ret;
    };

    // these three always go at the top, in this order
    ['title', 'author', 'created'].append(
        Object.keys(opts).sort(function (a, b) {
        var aBool = opts[a] === true,
            bBool = opts[b] === true;

        // both bool, fallback to alphabetical
        if (aBool && bBool)
            return a.localeCompare(b);

        // one bool, it comes last
        if (bBool && !aBool)
            return updateBiggest(a.length, -1);
        if (aBool && !bBool)
            return updateBiggest(b.length, 1);

        // neither bool, fallback to alphabetical
        updateBiggest(Math.max(a.length, b.length));
        return a.localeCompare(b);

    })).each(function (optName) {
        if (done[optName])
            return;
        done[optName] = true;

        // not present
        var value = opts[optName];
        if (typeOf(value) == 'null')
            return;
        if (typeOf(value) == 'string' && !value.length)
            return;
        if (typeOf(value) == 'boolean' && !value)
            return;

        string += '@page.' + optName;

        // non-boolean value
        if (value !== true) {
            string += ':';

            // add however many spaces to make it line up
            if (optName.length < biggest)
                string += Array(biggest - optName.length).join(' ');

            // escape semicolons
            value = value.replace(/;/g, '\\;');

            string += value;
        }

        string += ';\n';
    });
    return string;
}

})(adminifier);
