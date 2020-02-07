(function (a) {

document.addEvent('editorLoaded', loadedHandler);
document.addEvent('pageUnloaded', unloadedHandler);

var colorList = [
    'AliceBlue', 'AntiqueWhite', 'Aqua', 'Aquamarine', 'Azure', 'Beige',
    'Bisque', 'Black', 'BlanchedAlmond', 'Blue', 'BlueViolet',
    'Brown', 'BurlyWood', 'CadetBlue', 'Chartreuse', 'Chocolate', 'Coral',
    'CornflowerBlue', 'Cornsilk', 'Crimson', 'Cyan', 'DarkBlue', 'DarkCyan',
    'DarkGoldenRod', 'DarkGray', 'DarkGreen', 'DarkKhaki', 'DarkMagenta',
    'DarkOliveGreen', 'DarkOrange', 'DarkOrchid', 'DarkRed', 'DarkSalmon',
    'DarkSeaGreen', 'DarkSlateBlue', 'DarkSlateGray', 'DarkTurquoise',
    'DarkViolet', 'DeepPink', 'DeepSkyBlue', 'DimGray', 'DodgerBlue',
    'FireBrick', 'FloralWhite', 'ForestGreen', 'Fuchsia', 'Gainsboro',
    'GhostWhite', 'Gold', 'GoldenRod', 'Gray', 'Green', 'GreenYellow',
    'HoneyDew', 'HotPink', 'IndianRed', 'Indigo', 'Ivory', 'Khaki',
    'Lavender', 'LavenderBlush', 'LawnGreen', 'LemonChiffon', 'LightBlue',
    'LightCoral', 'LightCyan', 'LightGoldenRodYellow', 'LightGray',
    'LightGreen', 'LightPink', 'LightSalmon', 'LightSeaGreen',
    'LightSkyBlue', 'LightSlateGray', 'LightSteelBlue', 'LightYellow',
    'Lime', 'LimeGreen', 'Linen', 'Magenta', 'Maroon', 'MediumAquaMarine',
    'MediumBlue', 'MediumOrchid', 'MediumPurple', 'MediumSeaGreen',
    'MediumSlateBlue', 'MediumSpringGreen', 'MediumTurquoise',
    'MediumVioletRed', 'MidnightBlue', 'MintCream', 'MistyRose', 'Moccasin',
    'NavajoWhite', 'Navy', 'OldLace', 'Olive', 'OliveDrab', 'Orange',
    'OrangeRed', 'Orchid', 'PaleGoldenRod', 'PaleGreen', 'PaleTurquoise',
    'PaleVioletRed', 'PapayaWhip', 'PeachPuff', 'Peru', 'Pink', 'Plum',
    'PowderBlue', 'Purple', 'Red', 'RosyBrown', 'RoyalBlue', 'SaddleBrown',
    'Salmon', 'SandyBrown', 'SeaGreen', 'SeaShell', 'Sienna', 'Silver',
    'SkyBlue', 'SlateBlue', 'SlateGray', 'Snow', 'SpringGreen', 'SteelBlue',
    'Tan', 'Teal', 'Thistle', 'Tomato', 'Turquoise', 'Violet', 'Wheat',
    'White', 'WhiteSmoke', 'Yellow', 'YellowGreen'
];

var ae;
function loadedHandler () {
    ae = a.editor;

    // add toolbar functions
    Object.append(ae.toolbarFunctions, {
        font:       displayFontSelector,
        bold:       ae.wrapTextFunction('b'),
        italic:     ae.wrapTextFunction('i'),
        underline:  ae.wrapTextFunction('u'),
        strike:     ae.wrapTextFunction('s')
    });

    // add keyboard shortcuts
    ae.addKeyboardShortcuts([
        [ 'Ctrl-B', 'Command-B',    'bold'      ],
        [ 'Ctrl-I', 'Command-I',    'italic'    ],
        [ 'Ctrl-U', 'Command-U',    'underline' ]
    ]);
    
    a.loadScript('colorpicker');
}

function unloadedHandler () {
    document.removeEvent('editorLoaded', loadedHandler);
    document.removeEvent('pageUnloaded', unloadedHandler);
    $$('.colorpicker-asset').each(function (script) {
        script.destroy();
    });
}

// TEXT COLOR SELECTOR

function displayFontSelector () {
        
    // create box
    var li  = ae.liForAction('font');
    var box = ae.createPopupBox(li);
    box.innerHTML = tmpl('tmpl-color-helper', {});
    ae.fakeAdopt(box); // for injectInto
    ae.setLiLoading(li, true);
    
    // create color picker
    var cp = new DynamicColorPicker({
        injectInto: box.getElement('#editor-color-hex')
    });
    
    // on close, destroy color picker
    ae.onPopupDestroy = function () {
        cp.picker.destroy();
    };
    
    // create crayon picker.
    var container = box.getElement('#editor-color-names');
    ae.fakeAdopt(container); // for getComputedStyle()
    
    // create color elements
    colorList.each(function (colorName) {
        var div = new Element('div', {
            styles: { backgroundColor: colorName },
            class: 'editor-color-cell'
        });

        // separate the name into words
        div.innerHTML = tmpl('tmpl-color-name', {
            colorName: colorName.replace(/([A-Z])/g, ' $1')
        });
        container.appendChild(div);

        // compute and set the appropriate text color
        var color = new Color(getComputedStyle(div, null).getPropertyValue('background-color'));
        div.setStyle('color', getContrastYIQ(color.hex.substr(1)));

        // add click event
        div.addEvent('click', ae.wrapTextFunction(colorName));

    });

    // add events for toggling between hex/list
    var btnPicker = $$('.editor-color-type')[0],
        btnList   = $$('.editor-color-type')[1];
    btnPicker.addEvent('click', function () {
        btnPicker.addClass('active');
        btnList.removeClass('active');
        $('editor-color-names').setStyle('display', 'none');
        $('editor-color-hex').setStyle('display', 'block');
        box.morph({ width: '395px' });
        cp.picker.show();
    });
    btnList.addEvent('click', function () {
        btnList.addClass('active');
        btnPicker.removeClass('active');
        $('editor-color-hex').setStyle('display', 'none');
        $('editor-color-names').setStyle('display', 'block');
        box.morph({ width: '300px' });
        cp.picker.hide();
    });


    // put it where it belongs
    container.parentElement.removeChild(container);
    box.appendChild(container);

    // delay showing the box until the color picker is loaded
    DynamicColorPicker.autoLoad(function () {
        
        // on click, insert
        $('colorpicker-preview').addEvent('click', function () {
            var color = cp.picker.color.hex;
            return ae.wrapTextFunction('#' + color)();
        });
        
        // prevent the popup from closing due to a click on the positioner
        cp.container.addEvents({
            mouseover: function () {
                box.addClass('sticky');
            },
            mouseout: function () {
                box.removeClass('sticky');
            }
        });
        $$('.colorpicker-arrow').each(function (arrow) {
            arrow.addClass('no-close-popup');
        });
        
        // show the box with the picker
        box.setStyle('width', '395px');
        ae.setLiLoading(li, false);
        ae.displayPopupBox(box, 290, li);
        cp.show();
    }, 100);
}

function getContrastYIQ (hexColor) {
    var r = parseInt(hexColor.substr(0, 2), 16);
    var g = parseInt(hexColor.substr(2, 2), 16);
    var b = parseInt(hexColor.substr(4, 2), 16);
    var yiq = ((r * 299) + (g * 587) + (b * 114)) / 1000;
    return (yiq >= 128) ? '#000' : '#fff';
}

})(adminifier);
