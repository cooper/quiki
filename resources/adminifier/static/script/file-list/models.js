(function (a, exports) {
    
if (!FileList || !a.currentJSONMetadata || !a.currentJSONMetadata)
    return;

var modelList = new FileList({
    root: 'models',
    columns: ['Title', 'Author', 'Created', 'Modified'],
    columnData: {
        Title:      { sort: 't', isTitle: true },
        Author:     { sort: 'a' },
        Created:    { sort: 'c', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' },
        Modified:   { sort: 'm', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'data' }
    }
});

if (a.currentJSONMetadata.results)
a.currentJSONMetadata.results.each(function (modelData) {
    var entry = new FileListEntry({
        Title:      modelData.title || modelData.file_ne || modelData.file,
        Author:     modelData.author,
        Created:    modelData.created,
        Modified:   modelData.modified
    });
    entry.link = adminifier.wikiRoot + '/edit-model?page=' + encodeURIComponent(modelData.file);
    modelList.addEntry(entry);
});

modelList.draw($('content'));

})(adminifier, window);
