(function (a, exports) {

window.addEvent('resize', resize);
document.addEvent('pageUnloaded', pageUnloaded)
setTimeout(resize, 100);

function pageUnloaded () {
    window.removeEvent('resize', resize);
    document.removeEvent('pageUnloaded', pageUnloaded);
    closeFilter();

    // undo our content size adjustments
    $('content').setStyle('width', 'auto');
}

var FileList = exports.FileList = new Class({
    
    Implements: [Options, Events],
    
    options: {
        selection: true,    // allow rows to be selected
        columns: [],        // ordered list of column names
        columnData: {},     // object of column data, column names as keys
        root: 'files'       // type of files being listed
        // isTitle      true for the widest column
        // sort         sort letter
        // fixer        transformation to apply to text before displaying it
    },
    
    initialize: function (opts) {
        this.entries = [];
        this.showColumns = {};
        this.setOptions(opts);
    },
    
    // add an entry
    addEntry: function (entry) {
        var self = this;
        Object.each(entry.columns, function (val, col) {
            
            // skip if zero-length string
            if (typeof val == 'string' && !val.length)
                return;
            
            // convert date to object
            switch (self.getColumnData(col, 'dataType')) {
                case 'date':
                    val = new Date(val);
                    break;
            }
            
            // overwrite with transform
            entry.columns[col] = val;
            
            // if we made it to here, show the column
            self.showColumns[col] = true;
        });
        this.entries.push(entry);
    },
    
    // a list of visible column numbers
    getVisibleColumns: function () {
        var self = this;
        return this.options.columns.filter(function (col) {
            return self.showColumns[col];
        });
    },
    
    // getColumnData(column number, data key) returns value at that key
    // getColumnData(column number) returns entire object
    getColumnData: function (col, key) {
        if (!this.options.columnData[col])
            return;
        if (typeof key != 'undefined')
            return this.options.columnData[col][key];
        return this.options.columnData[col];
    },
    
    getValuesForColumn: function (col) {
        return this.entries.map(function (entry) {
            return entry.columns[col];
        });
    },

    getSelection: function () {
        return this.entries.filter(function (entry) {
            return entry.checkbox && entry.checkbox.checked;
        });
    },
    
    redraw: function () {
        var container = this.container;
        
        // not drawn yet
        if (!container) {
            console.log('Cannot redraw() without previous draw()');
            return;
        }
        
        // destroy previous injections
        this.container.getElements('.file-list, .file-list-no-entries').each(function (el) {
            el.destroy();
        });
        
        // re-draw
        delete this.container;
        this.draw(container);
    },
    
    // draw the table in the specified place
    draw: function (container) {
        var self = this;
        
        // already drawn
        if (self.container) {
            console.log('Cannot draw() file list again; use redraw()');
            return;
        }
        
        self.container = container;
        
        // if a filter is applied, filter
        var visibleEntries = self.entries;
        if (self.filter)
            visibleEntries = visibleEntries.filter(self.filter);
        
        // create table
        var table = self.table = new Element('table', { 'class': 'file-list' });
        table.store('file-list', self);
        
        // TABLE HEADING

        // may be cached
        var thead = self.thead;
        var selectAllInput;
        if (thead) {
            table.appendChild(thead);
            selectAllInput = thead.getElement('input[type=checkbox]');
        }

        // create heading
        else {

            thead = self.thead = new Element('thead');
            var theadTr = new Element('tr');
            thead.appendChild(theadTr);
            table.appendChild(thead);
            
            // checkbox column for table head
            var checkTh = new Element('th', { 'class': 'checkbox' });
            selectAllInput = new Element('input', { type: 'checkbox' });
            checkTh.appendChild(selectAllInput);
            if (self.options.selection)
                theadTr.appendChild(checkTh);
            
            // other columns
            self.getVisibleColumns().each(function (col) {
                
                // column is title?
                var className = self.getColumnData(col, 'isTitle') ?
                    'title' : 'info';
                var th = new Element('th', { 'class': className });
                var anchor = a.addFrameClickHandler(new Element('a', { text: col, 'class': 'frame-click' }));
                th.appendChild(anchor);
                theadTr.appendChild(th);
                
                // sort method
                var sort = self.getColumnData(col, 'sort');
                if (sort) {
                    th.set('data-sort', sort);
                    var setSort = sort + '-';
                    if (a.data.sort == setSort)
                        setSort = sort + encodeURIComponent('+');
                    anchor.set('href', updateURLParameter(window.location.href, 'sort', setSort));
                }
            });

            // duplicate that will be covered
            thead.appendChild(theadTr.clone()); 
        }

        // on change, check/uncheck all
        if (self.options.selection) selectAllInput.addEvent('change', function () {
            var checked = selectAllInput.checked;
            tbody.getElements('input[type=checkbox]').each(function (box) {
                box.checked = checked;
            });
            self.updateActions();
        });

        // TABLE BODY

        var tbody = new Element('tbody');
        table.appendChild(tbody);
        
        // Render message if no entries found
        if (visibleEntries.length === 0) {
            var noEntriesMessage = new Element('p', {
                class: 'file-list-no-entries',
                style: 'padding: 20px;',
                text: 'No ' + self.options.root + ' found.'
            });
            container.appendChild(noEntriesMessage);
            return;
        }
        
        // checkbox column for table body
        var checkTd = new Element('td', { 'class': 'checkbox' });
        checkTd.appendChild(new Element('input', { type: 'checkbox' }));
        
        // add each entry as a table row
        visibleEntries.each(function (entry) {
            
            // row may be cached
            var tr = entry.tr;
            if (tr) {
                tbody.appendChild(tr);
                return;
            }
            
            // have to create row
            tr = entry.tr = new Element('tr');
            
            // checkbox
            if (self.options.selection) {
                tr.appendChild(checkTd.clone());
                entry.checkbox = tr.getElement('input');
            }
            
            // visible data columns
            self.getVisibleColumns().each(function (col) {
                
                // column is title?
                var isTitle   = self.getColumnData(col, 'isTitle');
                var className = isTitle ? 'title' : 'info';
                var td = new Element('td', { 'class': className });
                
                // anchor or span
                var textContainer;
                if (entry.link && isTitle)
                    textContainer = a.addFrameClickHandler(new Element('a', {
                        href: entry.link, 'class': 'frame-click' }));
                else
                    textContainer = new Element('span');
                td.appendChild(textContainer);
                
                if (isTitle) {

                    // add states to title cell
                    entry.infoState.each(function (name) {
                        var span = new Element('span', {
                            text:   name,
                            class: 'file-info ' + name.toLowerCase()
                        });
                        td.appendChild(span);
                    });

                    if (entry.data) {
                        // add description to title cell
                        var preview = entry.data.desc || entry.data.preview;
                        if (preview) td.appendChild(new Element('span', {
                            text:   preview,
                            class: 'file-preview'
                        }));

                        // set td title
                        if (Browser.name == 'safari')
                            td.set('title', preview);
                    }
                }
                                
                // apply fixer
                var text  = entry.columns[col];
                var fixer = self.getColumnData(col, 'fixer');
                if (fixer) text = fixer(text);
                                
                // set text if it has length
                if (typeof text == 'string' && text.length)
                    textContainer.set('text', text);

                // add folder icon if it is dir
                if (entry.isDir && isTitle) {
                    var icon = new Element('i', { 'class': 'fa fa-folder', style: 'margin-right: 5px;' });
                    icon.inject(textContainer, 'top');
                }
                    
                // apply tooltip fixer
                var tooltip = entry.tooltips[col];
                fixer = self.getColumnData(col, 'tooltipFixer');
                if (fixer) tooltip = fixer(entry.columns[col]);
                    
                // tooltip
                if (typeof tooltip == 'string' && tooltip.length)
                    textContainer.set('title', tooltip);
                    
                tr.appendChild(td);
            });
            tbody.appendChild(tr);
        });

        // handle checkbox click for each row
        tbody.getElements('input[type=checkbox]').each(function (cb) {
            cb.addEvent('change', function () {

                // if all are checked, check the "select all" box
                if (cb.checked && tbody.getElements('input[type=checkbox]').every(function (c) {
                    return c.checked;
                })) {
                    selectAllInput.checked = true;
                }

                // if unchecking, uncheck the "select all" box
                else if (!cb.checked)
                    selectAllInput.checked = false;

                self.updateActions();
            });
        });

        // update sort
        self.updateSortMethod(a.data.sort);

        // update "select all" state and action buttons
        self.updateSelectAll();
        self.updateActions();

        container.appendChild(table);
    },

    // reflect the current sort method
    updateSortMethod: function (sort) {
        var table = this.table;
            
        // destroy existing
        var existing = table.getElement('.sort-icon');
        if (existing)
            existing.destroy();
        
        // no sort specified
        if (!sort)
            return;
        
        var split = sort.split('');
        sort = split[0];
        var order = split[1];
        var char = order == '+' ? 'caret-up' : 'caret-down';
        var th = table.getElement('th[data-sort="' + sort + '"] a');
        if (th) th.innerHTML +=
            ' <i class="sort-icon fa fa-' + char +
            '" style="padding-left: 3px; width: 1em;"></i>';
    },

    // update "select all" state
    updateSelectAll: function () {
    
        // selection is disabled
        if (!this.options.selection)
            return;

        var tbody = this.table.getElement('tbody'),
            thead = this.table.getElement('thead');
        var selectAllInput = thead.getElement('input[type=checkbox]');
        var checkboxes = tbody.getElements('input[type=checkbox]');

        // no checkboxes are there, so unselect
        if (!checkboxes.length) {
            selectAllInput.checked = false;
            return;
        }

        // if all are checked, check the "select all" box
        if (checkboxes.every(function (c) {
            return c.checked;
        }))
            selectAllInput.checked = true;

        // otherwise, if any are unchecked, uncheck "select all"
        else if (checkboxes.some(function (c) {
            return !c.checked;
        }))
            selectAllInput.checked = false;

    },

    // update action buttons
    updateActions: function () {
        var someSelected = this.table.getElement('tbody').
        getElements('input[type=checkbox]').some(function (c) {
            return c.checked;
        });
        var display = someSelected ? 'inline-block' : 'none';
        $$('.top-button.action').each(function (btn) {
            btn.setStyle('display', display);
        });
    }
});

var FileListEntry = exports.FileListEntry = new Class({
    Implements: [Options],

    options: { values: {} },

    initialize: function (values) {
        this.columns    = {};
        this.tooltips   = {};
        this.infoState  = [];

        // data for entry
        if (values.data) {
            this.data = values.data;
            delete values.data;
        }

        // add values
        this.setValues(values);
    },
    
    setValue: function (key, value) {
        if (typeof value == 'undefined')
            return;
        if (typeof value == 'string' && !value.length)
            return;
        this.columns[key] = value;
    },
    
    setValues: function (obj) {
        if (!obj) return;
        var self = this;
        Object.each(obj, function(value, key) {
            self.setValue(key, value);
        });
    },
    
    setInfoState: function (name, state) {
        if (state && !this.infoState.contains(name))
            this.infoState.push(name);
        else
            this.infoState.erase(name);
    }
});

function getList () {
    var list = document.getElement('.file-list');
    if (!list)
        return;
    list = list.retrieve('file-list');
    if (!list)
        return;
    list.filter = defaultFilter;
    return list;
}

function defaultFilter (entry) {
    return filterFilter(entry) && quickSearch(entry);
}

var searchText;

exports.fileSearch = fileSearch;
function fileSearch (text) {
    var list = getList();
    if (!list)
        return;
    searchText = text;
    list.redraw();
}

function quickSearch (entry) {
    $('top-search').removeClass('invalid');
    $('top-search').set('title', '');
    
    // quicksearch not enabled
    if (typeof searchText != 'string' || !searchText.length)
        return true;
        
    var matched = 0;

    Object.values(entry.columns).each(function (val) {
        if (typeof val != 'string') {
            if (val.toString)
                val = val.toString();
            else
                return;
        }
        try {
            if (val.match(new RegExp(searchText, 'i')))
                matched++;
        }
        catch (e) {
            $('top-search').addClass('invalid');
            $('top-search').set('title', e.message);
        }
    });
    return !!matched;
}

exports.dateToPreciseHR = dateToPreciseHR;
function dateToPreciseHR (d) {
    if (!d)
        return '';
    if (typeof d == 'string' || typeof d == 'number')
        d = new Date(d);
    if (!d)
        d = new Date(); 
    return d.toString();
}

exports.dateToHRTimeAgo = dateToHRTimeAgo;
function dateToHRTimeAgo (time) {
    if (!time)
        return '';
    switch (typeof time) {
        case 'number':
            break;
        case 'string':
            time = +new Date(time);
            break;
        case 'object':
            if (time.constructor === Date)
                time = time.getTime();
            break;
        default:
            time = +new Date(0);
    }
    var time_formats = [
        [60,            'seconds',                          1           ],
        [120,           '1 minute ago', '1 minute from now'             ],
        [3600,          'minutes',                          60          ],
        [7200,          '1 hour ago', '1 hour from now'                 ],
        [86400,         'hours',                            3600        ],
        [172800,        'Yesterday', 'Tomorrow'                         ],
        [604800,        'days',                             86400       ],
        [1209600,       'Last week', 'Next week'                        ],
        [2419200,       'weeks',                            604800      ],
        [4838400,       'Last month', 'Next month'                      ],
        [29030400,      'months',                           2419200     ],
        [58060800,      'Last year', 'Next year'                        ],
        [2903040000,    'years',                            29030400    ],
        [5806080000,    'Last century', 'Next century'                  ],
        [58060800000,   'centuries',                        2903040000  ]
    ];
    var seconds = (+new Date() - time) / 1000,
        token = 'ago',
        list_choice = 1;
    if (seconds == 0)
        return 'Just now';
    if (seconds < 0) {
        seconds = Math.abs(seconds);
        token = 'from now';
        list_choice = 2;
    }
    var i = 0, format;
    while (format = time_formats[i++])
    if (seconds < format[0]) {
        if (typeof format[2] == 'string')
            return format[list_choice];
        else
            return Math.floor(seconds / format[2]) +
            ' ' + format[1] + ' ' + token;
    }
    return time;
}

function resize () {
    var filterEditor = document.getElement('.filter-editor');
    var width = window.innerWidth - $('navigation-sidebar').offsetWidth;
    if (filterEditor)
        width -= filterEditor.offsetWidth;
    width += 'px';
    $('content').setStyle('width', width);
    $$('table.file-list thead tr:first-child').each(function (tr) {
        tr.setStyle('width', width);
        tr.setStyle('opacity', 1);
    });
}

// filter button toggle
exports.displayFilter = displayFilter;
function displayFilter () {
    var list = getList();
    if (!list)
        return;
        
    // if filter is already displayed, close it
    if (document.getElement('.filter-editor')) {
        closeFilter();
        return;
    }

    // TODO: warn that opening the filter will clear the current selection
    
    // make filter button active
    $('top-button-filter').addClass('active');
    
    // create filter editor
    var filterEditor = new Element('div', {
        class:  'filter-editor',
        html:   tmpl('tmpl-filter-editor', {})
    });

    // filter editor close button
    filterEditor.getElement('.filter-editor-title a').addEvent('click', closeFilter);
    
    // add each column
    list.options.columns.each(function (col) {
        
        // find data type; fall back to text
        var template = 'tmpl-filter-text',
            dataType = list.getColumnData(col, 'dataType');
        if (dataType && $('tmpl-filter-' + dataType))
            template = 'tmpl-filter-' + dataType;
        
        // create row
        var row = new Element('div', {
            class:      'filter-row',
            html:       tmpl(template, { column: col }),
            'data-col': col
        });
        
        // on click, show the inner part
        var inner = row.getElement('.filter-row-inner');
        var check = row.getElement('input[type=checkbox]');
        var input = row.getElement('input[type=text]');
        check.addEvent('change', function () {
            var d = check.checked ? 'block' : 'none';
            inner.setStyle('display', d);
            input.focus();
            row.set('data-enabled', check.checked ? true : '');
            list.redraw();
        });

        // on radio button change, focus the text input
        inner.getElements('input[type=radio]').each(function (radio) {
            radio.addEvent('change', function () {
                input.focus();
            });
        });
        
        var textInput = row.getElement('input[type=text]');
        var onEnterOrClick = function () {
            textInput.set('value', textInput.value.trim());
            
            // no length
            if (!textInput.value.length) {
                textInput.set('value', '');
                return;
            }
            
            // check if entry exists
            var maybeDuplicate = inner.getElements('.filter-item')
                .filter(function (item) {
                    return item.get('data-text') == textInput.value
                })[0];
                
            // it does, and it has the same mode
            if (maybeDuplicate && maybeDuplicate.get('data-mode') == mode) {
                textInput.set('value', '');
                return;
            }
            
            // it does, but the mode is different. overwrite with new mode
            else if (maybeDuplicate)
                maybeDuplicate.destroy();
            
            var mode = inner.getElements('input[type=radio]')
                .filter(function (rad) { return rad.checked })
                .get('data-mode');
                        
            var item = new Element('div', {
                class:  'filter-item',
                html:   tmpl('tmpl-filter-item', {
                    mode:   mode,
                    item:   textInput.value
                }),
                'data-mode': mode,
                'data-text': textInput.value
            });
            
            // on delete button click, delete
            item.getElement('i[class~="fa-minus-circle"]').addEvent('click',
            function () {
                item.destroy();
                list.redraw();
            });
            
            // add the item
            inner.appendChild(item);
            textInput.set('value', '');
            
            list.redraw();
        };
        
        // on enter or click, add item
        textInput.onEnter(onEnterOrClick);
        inner.getElement('i[class~="fa-plus-circle"]').addEvent('click',
            onEnterOrClick);
        
        // fake adopt for pikaday
        filterEditor.appendChild(row);
        a.fakeAdopt(filterEditor);
        
        // if this is a date, enable pikaday
        textInput.set('placeholder', 'Enter text...');
        if (dataType == 'date') {
            textInput.set('placeholder', 'Pick a date...');
            
            // determine date range
            var orderedDates = list.getValuesForColumn(col).sort(function(a, b) {
                return a - b;
            });
            var firstDate = new Date(),
                lastDate  = new Date();
            if (orderedDates[0])
                firstDate = orderedDates[0];
            if (orderedDates.getLast())
                lastDate = orderedDates.getLast();
            
            // create date picker
            var picker = new Pikaday({
                field:       textInput,
                firstDay:    0,
                minDate:     firstDate,
                maxDate:     lastDate,
                defaultDate: lastDate,
                yearRange:   [firstDate.getFullYear(), lastDate.getFullYear()]
            });
            
            // date select, add to list of accepted values
            textInput.addEvent('change', function () {
                onEnterOrClick();
                setTimeout(function () {
                    picker.setDate(null);
                }, 200);
            });
        }
    });
    
    // add each info state
    list.entries.map(function (e) {
        return e.infoState
    }).flatten().unique().each(function (stateName) {
        var row = new Element('div', {
            class:          'filter-row',
            html:           tmpl('tmpl-filter-state', { stateName: stateName }),
            'data-state':   stateName
        });
        var check = row.getElement('input[type=checkbox]');
        check.addEvent('change', function () {
            row.set('data-enabled', check.checked ? true : '');
            list.redraw();
        });
        filterEditor.appendChild(row);
    });
    
    document.body.adopt(filterEditor);
    resize();    
}

function getFilterRules (row) {
    var inner = row.getElement('.filter-row-inner');
    return inner.getElements('.filter-item').map(function (item) {
        return [ item.get('data-mode'), item.get('data-text') ];
    });
}

function filterFilter (entry) {
    
    // filters not enabled
    var filterEditor = document.getElement('.filter-editor');
    if (!filterEditor)
        return true;
        
    var allFuncsMustPass = [];
    filterEditor.getElements('.filter-row').each(function (row) {
        var someFuncsMustPass = [];
        var col = row.get('data-col');

        // row isn't enabled
        if (!row.get('data-enabled'))
            return;

        // info state
        if (!col) {
            var state = row.get('data-state');
            allFuncsMustPass.push(function (entry) {
                return entry.infoState.contains(state);
            });
            return;
        }
        
        // column
        getFilterRules(row).each(function (rule) {
            var right = entry.columns[col];
            
            // if no value is present, always fail
            if (typeof right == 'undefined')
            someFuncsMustPass.push(function () {
                return false;
            });
            
            // contains text
            else if (rule[0] == 'Contains')
            someFuncsMustPass.push(function (entry) {
                return right.toString().toLowerCase()
                    .contains(rule[1].toLowerCase());
            });

            // matches regex
            else if (rule[0] == 'Matches')
            someFuncsMustPass.push(function (entry) {
                if (typeof right != 'string') {
                    if (right.toString)
                        right = right.toString();
                    else
                        return;
                }
                try {
                    if (right.match(new RegExp(rule[1], 'i')))
                        return true;
                }
                catch (e) {
                    // TODO: somehow tell user regex is invalid but just once
                }
            });
            
            // equals
            else if (rule[0] == 'Is')
            someFuncsMustPass.push(function (entry) {
                
                // date
                if (typeOf(right) == 'date') {
                    var left = rule[1];
                    if (typeOf(left) != 'date') left = new Date(left);
                    left.setHours(0, 0, 0, 0);  // lose precision
                    right.setHours(0, 0, 0, 0);
                    return left.getTime() == right.getTime();
                }
                
                // string
                return right.toString().toLowerCase() == rule[1].toLowerCase();
            });
            
            // date less than
            else if (rule[0] == 'Before' && typeOf(right) == 'date')
            someFuncsMustPass.push(function (entry) {
                var left = rule[1];
                if (typeOf(left) != 'date') left = new Date(left);
                return left > right;
            });
        
            // date greater than
            else if (rule[0] == 'After' && typeOf(right) == 'date')
            someFuncsMustPass.push(function (entry) {
                var left = rule[1];
                if (typeOf(left) != 'date') left = new Date(left);
                return left < right;
            });
            
            // only successful if one or more of someFuncsMustPass passes
            allFuncsMustPass.push(function (entry) {
                return someFuncsMustPass.some(function (func) {
                    return func(entry);
                });
            });
        });
    });
    
    // only successful if every allFuncsMustPass passes
    return allFuncsMustPass.every(function (func) {
        return func(entry);
    });
}

function closeFilter () {
    var list = getList();
    if (!list) return;
    delete list.filter;

    // filter not active
    if (!document.getElement('.filter-editor'))
        return;
        
    // release the filter button in the top bar
    if ($('top-button-filter'))
        $('top-button-filter').removeClass('active');
        
    // disable the filter
    list.redraw();
    
    // destroy the editor
    document.getElement('.filter-editor').destroy();
    
    resize();
}

function updateURLParameter(url, param, paramVal){
    var newAdditionalURL = "";
    var tempArray = url.split("?");
    var baseURL = tempArray[0];
    var additionalURL = tempArray[1];
    var temp = "";
    if (additionalURL) {
        tempArray = additionalURL.split("&");
        for (var i=0; i<tempArray.length; i++){
            if(tempArray[i].split('=')[0] != param){
                newAdditionalURL += temp + tempArray[i];
                temp = "&";
            }
        }
    }

    var rows_txt = temp + "" + param + "=" + paramVal;
    return baseURL + "?" + newAdditionalURL + rows_txt;
}

// FILE ACTIONS

exports.deleteSelected = function (but) {
    var box = a.createPopupBox(but);

    var deleteBut = new Element('div', {
        id:         'file-delete-button',
        'class':    'popup-large-button',
        text:       'Are you sure?'
    });

    var list = getList();
    var selection = list.getSelection();

    // button text events
    var shouldChange = function () {
        return !deleteBut.hasClass('progress') &&
            !deleteBut.hasClass('success')  &&
            !deleteBut.hasClass('failure');
    };
    deleteBut.addEvent('mouseenter', function () {
        if (shouldChange()) deleteBut.innerHTML = 'Delete ' + selection.length + ' items';
    });
    deleteBut.addEvent('mouseleave', function () {
        if (shouldChange()) deleteBut.innerHTML = 'Are you sure?';
    });
    
    // button click
    deleteBut.addEvent('click', function () {
        var fileNames = selection.map(function (entry) {
            return entry.data.file;
        });
        deleteFiles(fileNames);
    });
    
    box.adopt(deleteBut);
    a.displayPopupBox(box, 40, but);
}

function deleteFiles (fileNames) {
    console.log("DeleteFiles", fileNames);
}

})(adminifier, window);
