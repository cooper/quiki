(function (a, exports) {
    
if (!FileList || !a.json)
    return;

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
    var createWindow = new ModalWindow({
        icon:           'plus-circle',
        title:          'Create page',
        html:           'Coming soon',
        padded:         true,
        id:             'create-page-window',
        autoDestroy:    true,
        onDone:         null
    });
    createWindow.show();
}

})(adminifier, window);
