(function (a) {

// regex to remove the wiki root 
var wikiRootRgx = new RegExp(adminifier.wikiRoot.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&'));

// whether this page's scripts have been loaded yet
var pageScriptsDone = false;

// workaround for html/template bug where http:// becomes http:/
if (a.wikiPageRoot && !a.wikiPageRoot.match(/:\/\//) && a.wikiPageRoot.match(/:\//))
    a.wikiPageRoot = a.wikiPageRoot.replace(/:\//, '://');

// this is for if pageScriptsDone event is added
// and the page scripts are already done
Element.Events.pageScriptsLoaded = {
	onAdd: function (fn) {
		if (pageScriptsDone)
            fn.call(this);
	}
};

// onEnter event
Element.implement('onEnter', function (func) {
    this.addEvent('keyup', function (e) {
        if (e.key != 'enter')
            return;
        func.call(this, e);
    });
});

// "space-separated values" splitter
function SSV (str) {
    if (typeof str != 'string' || !str.length)
        return [];
    return str.split(' ');
}

// convert hyphens to camel case
function camelCase (str) {
    return str.replace(/-([a-z])/g, function (g) { return g[1].toUpperCase(); })
}

// EXPORTS

// update page title
a.updatePageTitle = function (title, titleTagOnly) {
    console.log(title);
    if (!titleTagOnly)
        $('page-title').getElement('span').innerText = title;
    document.title = title + ' | ' + a.wikiName;
};

// update page icon
a.updateIcon = function (icon, b) {
    b = b ? 'b' : '';
	$('page-title').getElement('i').set('class', 'fa' + b + ' fa-' + icon);
};

// normalize page/category name
a.safeName = function (name) {
    return name.replace(/[^\w\.\-]/g, '_');
};

// load a script
a.loadScript = function (src) {
	a.loadScripts([src]);
};

// this is called when each script is loaded, so we can proceed to load the next
var scriptsToLoad = [], firedPageScriptsLoaded;
function scriptLoaded () {
	if (typeof jQuery != 'undefined')
		jQuery.noConflict();
	
	// there are still more to load
	if (scriptsToLoad.length) {
		loadNextScript();
		return;
	}
	
	// this was the last one
	$('content').setStyle('user-select', 'auto');
	a.updateIcon(a.data.icon, a.data.iconB);
	pageScriptsDone = true;
	if (firedPageScriptsLoaded) return;
	document.fireEvent('pageScriptsLoaded');
	firedPageScriptsLoaded = true;
}

// this is called when the next pending script should be loaded
function loadNextScript () {
	var script = scriptsToLoad.shift();
	if (!script)
		return;
	document.head.appendChild(script);
}

// load several scripts
a.loadScripts = function (srcs) {
    var sticky = false;
	srcs.each(function (src) {
        if (!src.length) return;
        
        // ace editor is special-- since it's large, load once and do not unload.
		if (src == 'ace') {
            if (window.ace)
                return;
            src = adminifier.staticRoot + '/ext/ace/src-min/ace.js';
            sticky = true;
        }

        // other distributions
		else if (src == 'pikaday')
			src = adminifier.staticRoot + '/ext/pikaday/pikaday.js';
		else if (src == 'jquery')
			src = '//cdnjs.cloudflare.com/ajax/libs/jquery/2.2.3/jquery.js';
		else if (src == 'diff2html')
			src = adminifier.staticRoot + '/ext/diff2html/dist/diff2html.js';
		else if (src == 'colorpicker')
			src = adminifier.staticRoot + '/ext/colorpicker/DynamicColorPicker.js';
		else if (src == 'prettify')
            src = 'https://cdn.rawgit.com/google/code-prettify/master/loader/run_prettify.js';
            
        // normal script
		else
            src = adminifier.staticRoot + '/script/' + src + '.js';

        // create script
        var script = new Element('script', { src: src });

        // remember it's dynamic so as to unload it later, UNLESS it's sticky
        if (!sticky)
            script.addClass('dynamic');

        // on load, call scriptLoaded
        script.addEvent('load', scriptLoaded);
        
        // add to the list to load
		scriptsToLoad.push(script);
	});
	
	// call once in case there are no scripts
	scriptLoaded();
}

// utility to adopt an element to an invisible container in the DOM so that dimension
// properties become available before displaying it
a.fakeAdopt = function (child) {
    var parent = $('fake-parent');
    if (!parent) {
        parent = new Element('div', {
            id: 'fake-parent',
            styles: { display: 'none' }
        });
        document.body.appendChild(parent);
    }
    parent.appendChild(child);
};

window.addEvent('resize',       adjustCurrentPopup);
window.addEvent('popstate',     loadURL);
document.addEvent('domready', 	loadURL);
document.addEvent('domready',	searchHandler);
document.addEvent('domready',   addDocumentFrameClickHandler);
document.addEvent('keyup',		handleEscapeKey);


// PAGE LOADING

// this adds the frame click handler to all .frame-click anchors in the DOM
function addDocumentFrameClickHandler () {
    addFrameClickHandler($$('a.frame-click'));
    document.body.addEvent('click', bodyClickPopoverCheck);
}

// export addFrameClickHandler so that dynamically created anchors from other scripts
// can be handled properly 
a.addFrameClickHandler = addFrameClickHandler;
function addFrameClickHandler (where) {
    if (typeOf(where) == 'element')
        where = [where];
    where.each(function (a) {
        a.addEvent('click', function (e) {
            e.preventDefault();
            var page = a.href.replace(/^.*\/\/[^\/]+/, '').replace(wikiRootRgx, '').replace(/^\//, '');
            history.pushState(page, '', adminifier.wikiRoot + '/' + page);
            loadURL();
        });
    });
    return where[0];
}

// load a page
function frameLoad (page) {
    if (!page || page == '/')
        page = adminifier.homePage;

	// same page
    if (a.currentPage == page)
        return;
		
	// unload old page
    document.fireEvent('pageUnloaded');
    a.currentPage = page;
    console.log("Loading " + page);

    var handleResponse = function (html) {

        // the page may start with JSON metadata...
        if (!html.indexOf('<!--JSON'))
            a.json = JSON.parse(html.split('\n', 3)[1]);
        else
            a.json = {};

        // set the content
        $('content').innerHTML = html;

        // apply click handlers
        addFrameClickHandler($('content').getElements('a.frame-click'));

        // find HTML metadata
        var meta = $('content').getElement('meta');
        if (meta) {
            var attrs = {};
            [].forEach.call(meta.attributes, function(attr) {
                if (/^data-/.test(attr.name)) {
                    var val = attr.value;
                    if (val.textContent)
                        val = val.textContent;
                    var name = camelCase(attr.name.replace(/^data-/, ''));
                    attrs[name] = val;
                }
            });

            //////////////////////////
            // Tools for all pages ///
            //////////////////////////
            //
            // data-redirect                    javascript frame redirect
            // data-wredirect                   window redirect
            // data-nav                         navigation item identifier to highlight
            // data-title                       page title for top bar
            // data-icon                        page icon name for top bar
            // data-scripts                     SSV script names w/o extensions
            // data-styles                      SSV css names w/o extensions
            // data-flags                       SSV page flags
            // data-search 		                name of function to call on search
            // data-buttons                     buttons to display in top bar
            // data-selection-buttons           same but for bulk actions
            // data-button-*                    data to define a button
            // data-cd                          current directory, for breadcrumbs
            //
            //////////////////////////////
            /// Used by specific pages ///
            //////////////////////////////
            //
            // data-sort                        sort option, used by file lists

            handlePageData(attrs);
        }
    };

	a.updateIcon('circle-notch fa-spin');
    var request = new Request({
        url: adminifier.wikiRoot + '/frame/' + page,
        onSuccess: handleResponse,
        onFailure: function (e) {
            handleResponse(e.response
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
            );
        }
    });
    request.get();
}

// extract the page from the current URL and load it
function loadURL() {
    var loc = window.location.pathname;
    frameLoad(loc.replace(wikiRootRgx, '') + window.location.search);
}

// page options
var flagOptions = {

    // disable margins on #content
    'no-margin': {
        init: function () {
            $('content').addClass('no-margin');
        },
        destroy: function () {
            $('content').removeClass('no-margin');
        }
    },

    // minimize the sidebar to just icons
    'compact-sidebar': {
        init: function () {
			document.getElement('span.wiki-title').tween('min-width', '75px');
            $('navigation-sidebar').tween('width', '50px');
            $('content').tween('margin-left', '50px');
            $$('#navigation-sidebar li a').each(function (a) {
                a.getElement('span').fade('out');
                a.addEvents({
                    mouseenter: handleCompactSidebarMouseenter,
                    mouseleave: handleCompactSidebarMouseleave
                });
            });
            setTimeout(function () { window.fireEvent('resize') }, 500);
        },
        destroy: function () {
			document.getElement('span.wiki-title').tween('min-width', '140px');
            $('navigation-sidebar').tween('width', '170px');
            $('content').tween('margin-left', '170px');
            $$('#navigation-sidebar li a').each(function (a) {
                a.getElement('span').fade('in');
                a.removeEvents({
                    mouseenter: handleCompactSidebarMouseenter,
                    mouseleave: handleCompactSidebarMouseleave
                });
            });
            $$('div.navigation-popover').each(function (p) {
                p.parentElement.eliminate('popover');
                p.parentElement.removeChild(p);
            });
            setTimeout(function () { window.fireEvent('resize') }, 500);
        }
    },

    // display the search bar
	search: {
		init: function () {
			$('top-search').set('value', '');
			$('top-search').setStyle('display', 'inline-block');
			searchUpdate();
		},
		destroy: function () {
			$('top-search').setStyle('display', 'none');
		}
    },
    
    // top bar buttons
	buttons: {
		init: function () {           
            SSV(a.data.buttons).each(function (btn) {
                var but = makeButton(btn);
                but.inject($$('.top-title').getLast(), 'after');
            });
            SSV(a.data.selectionButtons).each(function (btn) {
                var but = makeButton(btn);
                but.addClass('action');
                but.inject($('top-search'), 'after');
            });
		},
		destroy: function () {
			$$('.top-button.injected').each(function (but) {
				but.destroy();
			});
		}
	}
};

function makeButton (buttonID, where) {
    
    // find opts
    var buttonStuff = a.data[camelCase('button-' + buttonID)];
    if (!buttonStuff) {
        console.warn('Button "' + buttonID + '" is not configured');
        return;
    }
    
    // parse psuedo-JSON
    buttonStuff = JSON.decode(buttonStuff.replace(/'/g, '"'));
    if (!buttonStuff) {
        console.warn('Failed to parse JSON for button "' + buttonID + '"');
        return;
    }
    
    // create button
    var but = new Element('span', {
        id:		'top-button-' + buttonID,
        class: 	'top-title top-button injected'
    });
    
    // hide if default state is to hide
    if (buttonStuff.hide)
        but.setStyle('display', 'none');
    
    // title
    var anchor = new Element('a', {
        href: buttonStuff.href || buttonStuff.frameHref || '#'
    });
    anchor.set('text', buttonStuff.title);
    but.appendChild(anchor);
    
    // icon
    if (buttonStuff.icon) {
        anchor.set('text', ' ' + anchor.get('text'));
        var i = new Element('i', {
            'class': 'fa fa-' + buttonStuff.icon
        });
        i.inject(anchor, 'top');
    }
    
    if (buttonStuff.frameHref)
        addFrameClickHandler(anchor);

    // click click event if callback func is provided
    if (buttonStuff.func) anchor.addEvent('click', function (e) {
        e.preventDefault();
        e.stopPropagation(); // don't let click out handler see this

        // close current popup if there is one
        if (a.currentPopup) {
            closeCurrentPopup();
            return;
        }

        // call click func
        var func = window[buttonStuff.func];
        if (!func) {
            console.warn(
                'Button "' + buttonID + '" function ' +
                buttonStuff.func + ' does not exist'
            );
            return;
        }
        func(but);
    });

    return but;
}

// handle page data on request completion
function handlePageData (data) {
    pageScriptsDone = false;

    console.log(data);
    a.data = data;
    $('content').setStyle('user-select', 'none');

    // window redirect
    var target = data.wredirect;
    if (target) {
        console.log('Redirecting to ' + target);
        window.location = target;
        return;
    }

    // page redirect
    target = data.redirect;
    if (target) {
        console.log('Redirecting to ' + target);
        history.pushState(page, '', adminifier.wikiRoot + '/' + target);
        loadURL();
        return;
    }

    // page title and icon
    a.updatePageTitle(data.title);
    window.scrollTo(0, 0);
    // ^ not sure if scrolling necessary when setting display: none

    // highlight navigation item
    var li = $$('li[data-nav="' + data.nav + '"]')[0];
    if (li) {
        $$('li.active').each(function (li) { li.removeClass('active'); });
        li.addClass('active');
    }

    // destroy old scripts
    $$('script.dynamic').each(function (script) { script.destroy(); });

    // load new scripts
	firedPageScriptsLoaded = false;
    a.loadScripts(SSV(data.scripts));


    // destroy old styles
    $$('link.dynamic').each(function (link) { link.destroy(); });

    // inject new styles
    SSV(data.styles).each(function (style) {
        if (!style.length) return;
		var href;
		if (style == 'colorpicker')
			href = 'ext/colorpicker/colorpicker.css';
		else if (style == 'diff2html')
			href = 'ext/diff2html/dist/diff2html.css';
		else if (style == 'pikaday')
			href = 'ext/pikaday-custom.css';
		else
			href = 'style/' + style + '.css';
		
        var link = new Element('link', {
            href:   adminifier.staticRoot + '/' + href,
            class: 'dynamic',
            type:  'text/css',
            rel:   'stylesheet'
        });
        document.head.appendChild(link);
    });

    // deinitialize old page flags
    if (a.currentFlags) a.currentFlags.each(function (flag) {
        if (flag.destroy)
            flag.destroy();
    });

    // initialize new page flags
    a.currentFlags = [];
    SSV(data.flags).each(function (flagName) {
        var flag = flagOptions[flagName];
        if (!flag) return;
        a.currentFlags.push(flag);
        flag.init();
    });

    // derive breadcrumbs from cd
    var bc = $('breadcrumbs');
    if (bc) {
        bc.empty();
        var cd = data.cd;
        if (cd) {
            var crumbs = cd.split('/');
            if (!bc)
                return;
            crumbs.reverse().each(function (crumb, i) {
                var a = new Element('a', {
                    href: '../'.repeat(i) + crumb,
                    text: crumb
                });
                addFrameClickHandler(a);
                var icon = new Element('i', {
                    class: 'fa fa-angle-right'
                });
                a.inject(bc, 'top');
                icon.inject(bc, 'top');
            });
        }
    }
}

// escape key pressed
function handleEscapeKey (e) {
    if (e.key != 'esc')
        return;

    // if there's a popup, exit it maybe
    if (a.currentPopup) {
        closeCurrentPopup({
            unlessSticky: true,
            reason: 'Escape key'
        });
        return;
    }

    // if there is a modal window, close it
    var container = document.getElement('.modal-container');
    if (container)
        container.retrieve('modal').destroy();
}

// SEARCH

// domready event to add searchUpdate responder
function searchHandler () {
	$('top-search').addEvent('input', searchUpdate);
}

// callback for when search field changes
function searchUpdate () {
    var text = $('top-search').get('value');
    
    // the page with search function has since been destroyed
	if (!a.data.search)
        return;
        
    // the page provides no search handler
	var searchFunc = window[a.data.search];
	if (!searchFunc)
        return;
        
    // call the search handler provided by the page
	searchFunc(text);
}

// COMPACT SIDEBAR

// on mouseover, expand the text
function handleCompactSidebarMouseenter (e) {
    var a = e.target;
    var p = a.retrieve('popover');
    if (!p) {
        p = new Element('div', { class: 'navigation-popover' });
        p.innerHTML = a.getElement('span').innerHTML;
        a.appendChild(p);
        p.set('morph', { duration: 150 });
        a.store('popover', p);
    }
    a.setStyle('overflow', 'visible');
    p.setStyle('background-color', '#444');
    p.morph({
        width: '90px',
        paddingLeft: '10px'
    });
}

// on mouseleave, retract to just icon
function handleCompactSidebarMouseleave (e) {
    var a = e.target;
    var p = a.retrieve('popover');
    if (!p) return;
    p.setStyle('background-color', '#333');
    p.morph({
        width: '0px',
        paddingLeft: '0px'
    });
    setTimeout(function () {
        a.setStyle('overflow', 'hidden');
    }, 200);
}

// TOP BAR POPUP BOXES


// create and return a new popup box
a.createPopupBox = function (but) {
    var box = new Element('div', { class: 'popup-box' });
    if (but && but.hasClass('right'))
        box.addClass('right');
    return box;
};

// present a popup box
a.displayPopupBox = function (box, height, but) {
    
    // can't open the but
    if (!openBut(but))
        return false;
    
    // add to body
    document.body.appendChild(box);
    
    // set as current popup, initial adjustment
    a.currentPopup = box;
    box.store('but', but);
    adjustCurrentPopup();
    
    // animate open
    box.set('morph', { duration: 150 });
    if (height == 'auto') {
        height = window.innerHeight - parseInt(box.getStyle('top'));
        box.set('morph', {
            onComplete: function () { box.setStyle('height', 'auto'); }
        });
        box.morph({ height: height + 'px' });
    }
    else if (typeof height == 'number')
        box.morph({ height: height + 'px' });
    else
        box.setStyle('height', height);
        
    return true;
};

// close the current popup box
function closeCurrentPopup (opts) {
    var box = a.currentPopup;
    if (!box)
        return;
    if (!opts)
        opts = {};

    // check if sticky
    if (opts.unlessSticky && box.hasClass('sticky')) {
        console.log('Keeping popup open: Sticky');
        return;
    }

    // check if mouse is over it.
    // note this will only work if the box has at least one child with
    // the hover selector active
    if (opts.unlessActive && box.getElement(':hover')) {
        console.log('Keeping popup open: Active');

        // once the mouse exits, close it
        box.addEvent('mouseleave', function () {
            a.closePopup(box, opts);
        });

        return;
    }

    // Safe point - we will close the box.
    if (opts.reason)
        console.log('Closing popup: ' + opts.reason);
        
    // run destroy callback
    if (a.onPopupDestroy) {
        a.onPopupDestroy(box);
        delete a.onPopupDestroy;
    }

    closeCurrentBut();
    box.set('morph', {
        duration: 150,
        onComplete: function () {
            a.currentPopup = null;
            if (box) box.destroy();
            if (opts.afterHide) opts.afterHide();
        }
    });
    
    if (box.getStyle('height') == 'auto')
        box.setStyle('height', window.innerHeight - parseInt(box.getStyle('top')));
    box.morph({ height: '0px' });
}

// move a popup when the window resizes
function adjustCurrentPopup () {
    
    // no popup open
    var box = a.currentPopup;
    if (!box)
        return;
        
    var but = box.retrieve('but');
    var rect = but.getBoundingClientRect();
            
    // adjust top no matter what
    box.setStyle('top', rect.top + but.offsetHeight);
    
    // this is a fixed box; don't change left or right
    if (box.hasClass('fixed'))
        return;
    
    // set left or right
    box.setStyle('left',
        box.hasClass('right') ?
        rect.right - 300 :
        rect.left
    );
};

// close current popup on click outside
function bodyClickPopoverCheck (e) {

    // no popup is displayed
    if (!a.currentPopup)
        return;

    // clicked within the popup
    if (e.target == a.currentPopup || a.currentPopup.contains(e.target))
        return;

    // the target says not to do this
    if (e.target && e.target.hasClass('no-close-popup'))
        return;

    closeCurrentPopup({
        unlessSticky: true,
        unlessActive: true,
        reason: 'Clicked outside the popup'
    });
}

// set button as active
function openBut (but) {

    // if a popup is open, ignore this.
    if (a.currentPopup)
        return false;

    but.addClass('active');
    a.currentBut = but;
    return true;
}

// set current active button as inactive
function closeCurrentBut () {
    if (!a.currentBut)
        return;
    if (a.currentBut.hasClass('sticky'))
        return;
    a.currentBut.removeClass('active');
    delete a.currentBut;
}

})(adminifier);
