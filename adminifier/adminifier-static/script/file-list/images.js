(function (a, exports) {
    
if (!FileList || !a.currentJSONMetadata)
    return;

var imageList = new FileList({
    root: 'images',
    columns: ['Filename', 'Author', 'Created', 'Modified'],
    columnData: {
        Filename:   { sort: 'a', isTitle: true },
        Author:     { sort: 'u' },
        Created:    { sort: 'c', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' },
        Modified:   { sort: 'm', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' }
    }
});

if (a.currentJSONMetadata.results)
a.currentJSONMetadata.results.each(function (imageData) {
    var entry = new FileListEntry({
        Filename:   imageData.file,
        Author:     imageData.author,
        Created:    imageData.created,
        Modified:   imageData.mod_unix
    });
    entry.link = adminifier.wikiRoot + 'func/image?file=' + imageData.file; // TODO
    imageList.addEntry(entry);
});

imageList.draw($('content'));

})(adminifier, window);

function imageModeToggle() {
    alert('Switching modes');
}
