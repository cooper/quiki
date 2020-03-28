(function (a, exports) {
    
if (!FileList || !a.currentJSONMetadata)
    return;

var imageList = new FileList({
    root: 'images',
    columns: ['Filename', 'Author', 'Dimensions', 'Created', 'Modified'],
    columnData: {
        Filename:   { sort: 't', isTitle: true },
        Author:     { sort: 'a' },
        Dimensions: { sort: 'd' },
        Created:    { sort: 'c', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' },
        Modified:   { sort: 'm', fixer: dateToHRTimeAgo, tooltipFixer: dateToPreciseHR, dataType: 'date' }
    }
});

if (a.currentJSONMetadata.results)
a.currentJSONMetadata.results.each(function (imageData) {
    var entry = new FileListEntry({
        Filename:   imageData.file,
        Author:     imageData.author,
        Dimensions: imageData.width + 'x' + imageData.height,
        Created:    imageData.created,
        Modified:   imageData.modified
    });
    entry.link = adminifier.wikiRoot + '/func/image/' + imageData.file;
    imageList.addEntry(entry);
});

imageList.draw($('content'));

})(adminifier, window);

function imageModeToggle() {
    alert('Switching modes');
}
