(function (a) {

document.addEvent('editorLoaded', loadedHandler);
document.addEvent('pageUnloaded', unloadedHandler);

var ae;
function loadedHandler () {
    ae = a.editor;

    // this function is only available for quiki pages and models
    if (!ae.isPage() && !ae.isModel() && !ae.isReadOnly())
        return;

    // add toolbar function
    ae.addToolbarFunctions({ options: displayPageOptionsWindow });
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

    // TODO: add support for markdown stuff too. if none of these, the options
    // button should never have been displayed on the toolbar or been disabled.

    var title    = ae.isModel() ? 'Modal options' : 'Page options';
    var template = ae.isModel() ? 'tmpl-model-options' : 'tmpl-page-options';

    // create the options window
    var optionsWindow = new ModalWindow({
        icon:           'cog',
        title:          title,
        html:           tmpl(template, foundOptsValues),
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
    var doc = editor.session.doc;

    // the header is the options, then a blank line, then the categories
    var header = generatePageOptions(newOpts);
    if (newCats.length) {
        if (header.length)
            header += '\n';
        newCats.sort().each(function (catName) {
            header += '@category.' + catName + ';\n';
        });
    }

    // remove blank lines from the top of the file so that exactly one blank
    // line will separate the header from the content. this also takes care
    // of blank lines left behind by removeLinesInRanges().
    while (doc.getLength() > 1 && !doc.getLine(0).trim().length)
        doc.removeFullLines(0, 0);

    // nothing to insert
    if (!header.length)
        return;

    // separate the header from the content, if there is any
    if (doc.getValue().trim().length)
        header += '\n';

    // inject the new lines at the beginning. the selection follows along.
    doc.insert({ row: 0, column: 0 }, header);
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
    var string = '',
        prefix = ae.isModel() ? '@model.' : '@page.';

    // true if the option is set
    var isSet = function (optName) {
        var value = opts[optName];
        if (typeOf(value) == 'null')
            return false;
        if (typeOf(value) == 'string' && !value.length)
            return false;
        if (typeOf(value) == 'boolean' && !value)
            return false;
        return true;
    };

    // these three always go at the top, in this order.
    // then the rest alphabetically, with booleans last.
    var names = ['title', 'author', 'created'].filter(isSet);
    names.combine(Object.keys(opts).filter(isSet).sort(function (a, b) {
        var aBool = opts[a] === true,
            bBool = opts[b] === true;
        if (aBool != bBool)
            return aBool ? 1 : -1;
        return a.localeCompare(b);
    }));

    // values line up two columns past the longest name
    var biggest = 0;
    names.each(function (optName) {
        if (opts[optName] !== true)
            biggest = Math.max(biggest, optName.length);
    });

    names.each(function (optName) {
        var value = opts[optName];
        string += prefix + optName;

        // non-boolean value
        if (value !== true) {
            string += ':' + Array(biggest - optName.length + 3).join(' ');

            // escape semicolons; see parsePageVariable in editor.js
            string += value.replace(/;/g, '\\;');
        }

        string += ';\n';
    });
    return string;
}

})(adminifier);
