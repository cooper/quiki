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

if (a.json.results)
a.json.results.each(function (pageData) {
    var entry = new FileListEntry({
        data:       pageData,
        Title:      pageData.title || pageData.file_ne || pageData.file,
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

pageList.draw($('content'));

exports.createPage = function () {
    var modal = new ModalWindow({
        icon:           'plus-circle',
        title:          'New Page',
        html:           'Coming soon',
        padded:         true,
        id:             'create-page-window',
        autoDestroy:    true,
        onDone:         null
    });
    modal.show();
}

exports.createFolder = function () {
    var modal = new ModalWindow({
        icon:           'folder',
        title:          'New Folder',
        html:           'Coming soon',
        padded:         true,
        id:             'create-folder-window',
        autoDestroy:    true,
        onDone:         null
    });
    modal.show();
}

})(adminifier, window);
