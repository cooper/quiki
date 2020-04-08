(function (a, exports) {

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

if (a.json.results)
a.json.results.each(function (imageData) {
    var dim = null;
    if (imageData.width && imageData.height)
        dim = imageData.width + 'x' + imageData.height;
    var entry = new FileListEntry({
        Filename:   imageData.file,
        Author:     imageData.author,
        Dimensions: dim,
        Created:    imageData.created,
        Modified:   imageData.modified
    });
    entry.link = adminifier.wikiRoot + '/func/image/' + imageData.file;
    imageList.addEntry(entry);
});

imageList.draw($('content'));

})(adminifier, window);