var quiki = {};
(function (exports) {

document.addEvent('domready', function () {

    // jump to section
    hashLoad();

    // load image gallery if needed
    if ($$(".q-gallery")) {
        loadCSS("/static/ext/nanogallery2/css/nanogallery2.min.css");
        loadJS("/static/ext/nanogallery2/jquery.nanogallery2.min.js");
    }
});

window.addEvent('hashchange', hashLoad);

// redirect #some-section to #qa-some-section
function hashLoad() {
    var hash = window.location.hash;
    if (hash.lastIndexOf('#', 0) === 0)
        hash = hash.substring(1);
    var anchor = 'qa-' + hash;
    var el = $(anchor);
    if (el) {
        pos = el.getPosition();
        scrollTo(pos.x, pos.y);
    }
}

function loadJS (src) {
    var script = new Element('script', {
        src:  src,
        type: 'text/javascript'
    });
    document.head.appendChild(script);
}

function loadCSS (href) {
    var link = new Element('link', {
        href:  href,
        rel:  'stylesheet',
        type: 'text/css'
    });
    document.head.appendChild(link);
}

// javascript image sizing
exports.imageResize = function (img) {
    img.parentElement.parentElement.setStyle('width', img.offsetWidth + 'px');
    img.setStyle('width', '100%');
};

})(quiki);
