/** Wrapper for John Dyer's Photoshop-like color picker, that dynamically
 * creates the required HTML.
 *
 * http://johndyer.name/post/2007/09/26/PhotoShop-like-JavaScript-Color-Picker.aspx
 *
 * N.B:
 *
 * - The 'change' event will be fired when the color on the picker changes.
 *
 * - If you apply styles to the colorpicker container class
 *   (colorpicker-container), use !important to override the default styles.
 *
 * - It's possible to enable the colorpicker on any number of textfields, by
 *   calling DynamicColorPicker.auto(".class", [options]). The
 *   DynamicColorPicker instance will be stored in the text field's colorPicker
 *   (MooTools) property.
 *
 * - If MooTools More is included, DynamicColorPicker autoloads the rest of the
 *   required JS files, so you don't have to include them in your HTML file.
 */
(function (exports) {
    
var pickerPath = 'ext/colorpicker';

var DynamicColorPicker = exports.DynamicColorPicker = new Class({

    Implements: [Options, Events],

    options: {
        injectInto: document.body,
        textField: null,
        startMode: 'h',
        startHex: 'ff0000'
    },

    initialize: function(options) {
        this.setOptions(options);
        this.options.clientFilesPath = pickerPath + '/images/';
        this.container = new Element("div", {
            "class": "colorpicker-container"
        });
        var self = this;
        DynamicColorPicker.autoLoad(function() {
            self.container.inject(self.options.injectInto);
            self.setContainerHtml();
            self.picker = new Refresh.Web.ColorPicker('colorpicker', self.options);
            self.picker.addEvent('change', self._onColorChange.bind(self));
            self.open = false;
        });
    },

    setColor: function(color) {
        if (color.substring(0, 1) == "#") color = color.substring(1);
        this.picker._cvp._hexInput.value = color;
        this.picker._cvp.setValuesFromHex();
    },

    show: function() {
        this.picker.show();
        this.container.setStyles({ "display": "block" });
        this.open = true;
        if (this.options.textField) this.setColor(this.options.textField.value);
        this.picker.setColorMode(this.picker.ColorMode); // If we reset this after we show, it positions the selectors properly
        this.picker.updateVisuals();
    },

    hide: function() {
        this.container.setStyles({ "display": "none" });
        this.picker.hide();
        this.open = false;
    },

    toggle: function() {
        if (this.open) this.hide();
        else this.show();
    },

    _onColorChange: function() {
        var newHex = '#' + this.picker._cvp._hexInput.value;
        if (this.options.textField) this.options.textField.set('value', newHex);
        this.fireEvent('change', newHex);
    },

    setContainerHtml: function() {
        this.container.set('html', tmpl('tmpl-color-container', {}));
    }
});

DynamicColorPicker.auto = function(spec, options) {
    $$(spec).each(function(el) {
        var cp = new DynamicColorPicker(Object.extend(options || {}, { textField: el }));

        var button = new Element("img", { src: cp.options.clientFilesPath + "/colorwheel.png", styles: { marginLeft: 3, marginBottom: -5, cursor: "pointer" }})
            .injectAfter(el)
            .addEvent('click', function() {
                var p = el.getPosition();
                cp.toggle(p.x, p.y + el.getSize().y);
            });

        el.store('colorPicker', cp);
    });
};

// Try autoloading
DynamicColorPicker.autoLoad = (function() {
    var loadStage = 0; // 0 = not loaded, 1 = loading, 2 = loaded
    var callbacks = [];

    return function(onload) {
        // If loaded, immediately done
        if (loadStage == 2) { if (onload) onload(); return; }

        if (loadStage == 1) {
            if (onload) callbacks.push(onload);
            return;
        }

        // loadStage == 0
        
        // If loading not possible because of missing MooTools More, return immediately and fire event
        if (!window.Asset) { if (onload) onload(); return; }

        // Do loading
        loadStage = 1;
        callbacks.push(onload);

        // Otherwise load the files and fire the event when all four required files are loaded
        var filesLoaded = 0;
        var onFileLoaded = function() {
            filesLoaded++;
            if (filesLoaded == 4) loadStage = 2;
            if (loadStage == 2)
                callbacks.each(function(f) { f(); });
        };

        var path = pickerPath;
        Asset.javascript(path + "/ColorPicker.js", { onload: onFileLoaded }).addClass('colorpicker-asset');
        Asset.javascript(path + "/ColorValuePicker.js", { onload: onFileLoaded }).addClass('colorpicker-asset');
        Asset.javascript(path + "/ColorMethods.js", { onload: onFileLoaded }).addClass('colorpicker-asset');
        Asset.javascript(path + "/Slider.js", { onload: onFileLoaded }).addClass('colorpicker-asset');
    };
})();

})(window);
