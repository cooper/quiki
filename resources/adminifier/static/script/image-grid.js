(function (a, exports) {
    
if (!a.currentJSONMetadata)
    return;
    
var container = new Element('div', { class: 'image-grid' });
$('content').appendChild(container);

if (a.currentJSONMetadata.results)
a.currentJSONMetadata.results.each(function (imageData) {
    imageData.root = adminifier.wikiRoot;

    // the larger dimension dictates the image size
    imageData.dimension = imageData.width < imageData.height ?
        'height' : 'width';
    
    // max is 250 for either
    imageData.dimValue = imageData[imageData.dimension];
    if (imageData.dimValue > 250)
        imageData.dimValue = 250;

    // retina scaling
    imageData.dimValue *= retinaDensity();
    
    var div = new Element('div', {
        class: 'image-grid-item',
        html:   tmpl('tmpl-image-grid-item', imageData)
    });
    container.appendChild(div);
});

function retinaDensity() {
    if (!window.matchMedia) return;
    if (window.devicePixelRatio < 1) return;
    
    // 3x
    var mq = window.matchMedia('only screen and (min--moz-device-pixel-ratio: 2.25), only screen and (-o-min-device-pixel-ratio: 2.6/2), only screen and (-webkit-min-device-pixel-ratio: 2.25), only screen  and (min-device-pixel-ratio: 2.25), only screen and (min-resolution: 2.25dppx)');
    if (mq && mq.matches)
        return 3;
        
    // 2x
    mq = window.matchMedia('only screen and (min--moz-device-pixel-ratio: 1.25), only screen and (-o-min-device-pixel-ratio: 2.6/2), only screen and (-webkit-min-device-pixel-ratio: 1.25), only screen  and (min-device-pixel-ratio: 1.25), only screen and (min-resolution: 1.25dppx)');
    if (mq && mq.matches)
        return 2;
        
    return 1;
}

})(adminifier, window);
