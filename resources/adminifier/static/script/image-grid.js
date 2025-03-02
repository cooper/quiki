(function (a, exports) {

var dirContainer = new Element('div', { class: 'image-grid' });
var imageContainer = new Element('div', { class: 'image-grid' });
$('content').appendChild(dirContainer);
$('content').appendChild(imageContainer);

var currentDir = a.json.results.cd;

function nextDir(dir) {
    if (!currentDir)
        return dir;
    return currentDir + '/' + dir;
}

if (a.json.results && (a.json.results.dirs.length || a.json.results.images.length)) {

a.json.results.dirs.each(function (dir) {
    var div = new Element('div', {
        class: 'image-grid-item',
        html: tmpl('tmpl-image-grid-dir', {
            name: dir,
            link: adminifier.wikiRoot + '/images/' + nextDir(dir) + location.search,
        })
    });
    a.addFrameClickHandler(div.getElements('a'));
    dirContainer.appendChild(div);
});

a.json.results.images.each(function (imageData) {
    imageData.root = adminifier.wikiRoot;

    // the larger dimension dictates the image size
    imageData.dimension = imageData.width < imageData.height ?
        'height' : 'width';
    
    // max is 250 for either
    imageData.dimValue = imageData[imageData.dimension];
    if (imageData.dimValue > 250)
        imageData.dimValue = 250;

    // retina scaling is disabled in adminifier for performance
    // imageData.dimValue *= retinaDensity();
    
    var div = new Element('div', {
        class: 'image-grid-item',
        html:   tmpl('tmpl-image-grid-item', {
            ...imageData,
            link: adminifier.wikiRoot + '/func/image/' + imageData.file,
        })
    });
    imageContainer.appendChild(div);
});

} else {
    imageContainer.innerHTML = '<p style="padding: 20px;">No images found.</p>';
}

// retinaDensity is disabled in adminifier for performance
// function retinaDensity() {
//     if (!window.matchMedia) return;
//     if (window.devicePixelRatio < 1) return;
    
//     // 3x
//     var mq = window.matchMedia('only screen and (min--moz-device-pixel-ratio: 2.25), only screen and (-o-min-device-pixel-ratio: 2.6/2), only screen and (-webkit-min-device-pixel-ratio: 2.25), only screen  and (min-device-pixel-ratio: 2.25), only screen and (min-resolution: 2.25dppx)');
//     if (mq && mq.matches)
//         return 3;
        
//     // 2x
//     mq = window.matchMedia('only screen and (min--moz-device-pixel-ratio: 1.25), only screen and (-o-min-device-pixel-ratio: 2.6/2), only screen and (-webkit-min-device-pixel-ratio: 1.25), only screen  and (min-device-pixel-ratio: 1.25), only screen and (min-resolution: 1.25dppx)');
//     if (mq && mq.matches)
//         return 2;
        
//     return 1;
// }

})(adminifier, window);
