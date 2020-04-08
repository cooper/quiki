(function (a, exports) {
    
if (!FileList || !a.json)
    return;

var catList = new FileList({
    root: 'categories',
    columns: ['Title', 'Author', 'Created', 'Modified'],
    columnData: {
        Title:      { sort: 't', isTitle: true },
        Author:     { sort: 'a' },
        Created:    { sort: 'c', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' },
        Modified:   { sort: 'm', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' }
    }
});

if (a.json.results)
a.json.results.each(function (catData) {
    var entry = new FileListEntry({
        Title:      catData.title || catData.file_ne || catData.file,
        Author:     catData.author,
        Created:    catData.created,
        Modified:   catData.modified
    });
    entry.setInfoState('Generated', catData.generated);
    entry.setInfoState('Draft', catData.draft);
    entry.link = adminifier.wikiRoot + '/edit-category?cat=' + encodeURIComponent(catData.file);
    catList.addEntry(entry);
});

catList.draw($('content'));

})(adminifier, window);
