var TokenList = new Class({
    
    initialize: function (items, opts) {
        var _this = this;
        
        // container element options
        var elOpts = { class: 'token-list' };
        if (opts)
            elOpts = Object.merge(elOpts, opts);
        
        // create the container
        this.container = new Element('div', elOpts);
        
        // input field
        var input = this.input = new Element('input', {
            type:  'text',
            class: 'token-list-input'
        });
        this.input.addEvent('keyup', function (e) {
            if (e.key != 'enter') return;
            var value = input.getProperty('value').trim();
            if (value.length)
                _this.addItem(value);
            input.setProperty('value', '');
        });
        this.container.appendChild(this.input);
        
        // add initial items
        if (items)
            items.each(function (i) { _this.addItem(i); });
        
    },
    
    addItem: function (item) {
        var el   = new Element('div', { class: 'token-list-item'   }),
            but  = new Element('div', { class: 'token-list-delete' }),
            span = new Element('span');
        
        // show/hide delete
        el.addEvents({
            mouseenter: function () {
                but.setStyle('display', 'block');
            },
            mouseleave: function () {
                but.setStyle('display', 'none');
            }
        });
        
        // delete
        but.addEvent('click', function () {
            el.destroy();
        });

        span.setProperty('text', item);
        el.appendChild(but);
        el.appendChild(span)
        this.container.insertBefore(el, this.input);
    },
    
    getItems: function () {
        return this.container.getElements('span').map(function (i) {
            return i.getProperty('text');
        });
    },
    
    // this allows $(tokenField)
    toElement: function () {
        return this.container;
    }
    
});