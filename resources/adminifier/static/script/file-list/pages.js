(function (a, exports) {

var pageList = new FileList({
    root: 'pages',
    columns: ['Title', 'Author', 'Created', 'Modified'],
    columnData: {
        Title:      { sort: 't', isTitle: true },
        Author:     { sort: 'a' },
        Created:    { sort: 'c', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' },
        Modified:   { sort: 'm', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' }
    }
});

var currentDir = a.json.results.cd;

function nextDir(dir) {
    if (!currentDir)
        return dir;
    return currentDir + '/' + dir;
}

if (a.json.results) {

a.json.results.dirs.each(function (dir) {
    var entry = new FileListEntry({ Title: dir });
    entry.isDir = true;
    entry.link = adminifier.wikiRoot + '/pages/' + nextDir(dir);
    pageList.addEntry(entry);
});

a.json.results.pages.each(function (pageData) {
    var entry = new FileListEntry({
        data:       pageData,
        Title:      pageData.title || pageData.base_ne || pageData.file_ne || pageData.file,
        Author:     pageData.author,
        Created:    pageData.created,
        Modified:   pageData.modified
    });
    if (pageData.redirect)
        pageData.desc = "Redirect \u00BB " + pageData.redirect;
    entry.setInfoState('Draft',     pageData.draft);
    entry.setInfoState('Redirect',  pageData.redirect);
    entry.setInfoState('External',  pageData.external);
    entry.setInfoState('Warnings',  pageData.warnings && pageData.warnings.length);
    entry.setInfoState('Error',     !!pageData.error);
    entry.link = adminifier.wikiRoot + '/edit-page?page=' + encodeURIComponent(pageData.file);
    pageList.addEntry(entry);
});
}

pageList.draw($('content'));

exports.createPage = function () {
    var modal = new ModalWindow({
        icon:           'plus-circle',
        title:          'New Page',
        html:           tmpl('tmpl-create-page', {}),
        padded:         true,
        id:             'create-page-window',
        autoDestroy:    true,
        onDone:         null,
        doneText:       'Cancel',
    });
    modal.show();
}

exports.createFolder = function () {
    var modal = new ModalWindow({
        icon:           'folder',
        title:          'New Folder',
        html:           tmpl('tmpl-create-folder', {}),
        padded:         true,
        id:             'create-folder-window',
        autoDestroy:    true,
        onDone:         null,
        doneText:       'Cancel',
    });
    modal.show();
}

})(adminifier, window);
