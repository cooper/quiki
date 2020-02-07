var ModalWindow = new Class({

    Implements: [Options, Events],

    options: {
        icon:           'cog',
        doneText:       'Done',
        autoDestroy:    false,
        sticky:         false
    },

    initialize: function (opts) {

        // fixed shaded region, modal window, padded content region, header
        this.container   = new Element('div',  { class: 'modal-container'   });
        this.modalWindow = new Element('div',  { class: 'modal-window'      });
        this.content     = new Element('div',  { class: 'modal-content'     });
        this.header      = new Element('div',  { class: 'modal-header'      });
        this.doneButton  = new Element('span', { class: 'modal-done-button' });
        this.header.adopt(this.doneButton);
        this.modalWindow.adopt(this.header, this.content);
        this.container.adopt(this.modalWindow);

        // done button click
        var _this = this;
        this.doneButton.addEvent('click', function () {
            _this.hide();
        });

        // store self in the container; this is OK because .destroy()
        // destroys the container
        this.container.store('modal', this);

        // initial options
        this.setOptions(opts);

        return this;
    },

    setOptions: function (opts) {
        Options.prototype.setOptions.call(this, opts);
        opts = this.options;

        // window width
        if (opts.width != null)
            this.modalWindow.setStyle('width', opts.width);

        // header text
        if (opts.title != null)
            this.setTitle(opts.title);

        // header icon
        if (opts.icon != null)
            this.setIcon(opts.icon);

        // done button text
        if (typeof opts.doneText != 'undefined')
            this.setDoneText(opts.doneText);

        // content inner HTML
        if (opts.html != null)
            this.content.innerHTML = opts.html;

        // add padding to content?
        if (opts.padded)
            this.content.addClass('padded');

        // ID for content div
        if (opts.id != null)
            this.content.set('id', opts.id);
    },

    setTitle: function (title) {
        var span = this.title;
        if (!span) {
            span = this.title = new Element('span');
            this.header.adopt(span);
        }
        span.innerText = ' ' + title;
    },

    setIcon: function (iconName) {
        var icon = this.icon;
        if (!icon) {
            icon = this.icon = new Element('i');
            this.icon.inject(this.doneButton, 'after');
        }
        icon.className = 'fa fa-' + iconName;
    },

    setDoneText: function (text) {
        if (typeOf(text) == 'null')
            this.doneButton.setStyle('display', 'none');
        else {
            this.doneButton.innerText = text;
            this.doneButton.setStyle('display', 'inline');
        }
    },
    
    addButton: function (text, onClick) {
        var but = new Element('span', {
            class: 'modal-done-button',
            text: text
        });
        but.inject(this.icon, 'before');
        if (onClick) but.addEvent('click', onClick);
        return but;
    },

    show: function (container) {
        if (!container)
            container = document.body;
        container.adopt(this.container);
        this.container.setStyle('display', 'block');
        this.shown = true;
    },

    hide: function (isDestroy) {
        if (!this.shown || this.options.sticky)
            return;
        delete this.shown;
        this.fireEvent('done');
        this.container.setStyle('display', 'none');
        if (this.options.autoDestroy && !isDestroy)
            this._destroy();
    },

    destroy: function (force) {
        if (this.options.sticky && !force)
            return;
        delete this.options.sticky;
        this.hide(true);
        this._destroy();
    },

    _destroy: function () {
        this.doneButton.destroy();
        this.container.destroy();
    }

});
