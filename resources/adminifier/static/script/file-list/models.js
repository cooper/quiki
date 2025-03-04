(function (a, exports) {
    
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

var currentDir;

function nextDir(dir) {
    if (!currentDir)
        return dir;
    return currentDir + '/' + dir;
}

if (a.json.results) {

currentDir = a.json.results.cd;

a.json.results.dirs.each(function (dir) {
    var entry = new FileListEntry({ Title: dir });
    entry.isDir = true;
    entry.link = adminifier.wikiRoot + '/models/' + nextDir(dir) + location.search;
    modelList.addEntry(entry);
});
    
a.json.results.models.each(function (modelData) {
    var entry = new FileListEntry({
        Title:      modelData.title || modelData.file_ne || modelData.file,
        Author:     modelData.author,
        Created:    modelData.created,
        Modified:   modelData.modified
    });
    entry.link = adminifier.wikiRoot + '/edit-model?page=' + encodeURIComponent(modelData.file);
    modelList.addEntry(entry);
});

}

modelList.draw($('content'));

exports.createModel = function () {
    var modal = new ModalWindow({
        icon:           'plus-circle',
        title:          'New Model',
        html:           tmpl('tmpl-create-model', {}),
        padded:         true,
        id:             'create-model-window',
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
        html:           tmpl('tmpl-create-folder', { mode: 'model' }),
        padded:         true,
        id:             'create-folder-window',
        autoDestroy:    true,
        onDone:         null,
        doneText:       'Cancel',
    });
    modal.show();
}

})(adminifier, window);
