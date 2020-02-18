(function (a) {

// regex to remove the wiki root 
var wikiRootRgx = new RegExp((adminifier.wikiRoot + '/').replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&'));

// whether this page's scripts have been loaded yet
var pageScriptsDone = false;

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

// EXPORTS

// update page title
a.updatePageTitle = function (title, titleTagOnly) {
    if (!titleTagOnly)
        $('page-title').getElement('span').innerText = title;
    document.title = title + ' | ' + a.wikiName;
};

// update page icon
a.updateIcon = function (icon) {
	$('page-title').getElement('i').set('class', 'fa fa-' + icon);
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
	$('content').setStyle('user-select', 'all');
	a.updateIcon(a.currentData['data-icon']);
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

document.addEvent('domready', 	loadURL);
document.addEvent('domready',	searchHandler);
document.addEvent('domready',   addDocumentFrameClickHandler);
document.addEvent('keyup',		handleEscapeKey);

// PAGE LOADING

// this adds the frame click handler to all .frame-click anchors in the DOM
function addDocumentFrameClickHandler () {
    addFrameClickHandler($$('a.frame-click'));
}

// export addFrameClickHandler so that dynamically created anchors from other scripts
// can be handled properly 
a.addFrameClickHandler = addFrameClickHandler;
function addFrameClickHandler (where) {
    if (typeOf(where) == 'element')
        where = [where];
    where.each(function (a) {
        a.addEvent('click', function (e) {
            console.log("HANDLED IT FOR "+ a.href);
            e.preventDefault();
            var page = a.href.replace(/^.*\/\/[^\/]+/, '').replace(wikiRootRgx, '');
            history.pushState(page, '', adminifier.wikiRoot + '/' + page);
            loadURL();
        });
    });
    return where[0];
}

// load a page
function frameLoad (page) {

	// same page
    if (a.currentPage == page)
        return;
		
	// unload old page
    document.fireEvent('pageUnloaded');
    a.currentPage = page;
    console.log("Loading " + page);

	a.updateIcon('circle-o-notch fa-spin');
    var request = new Request({
        url: 'frame/' + page,
        onSuccess: function (html) {

            // the page may start with JSON metadata...
            if (!html.indexOf('<!--JSON')) {
                var json = JSON.parse(html.split('\n', 3)[1]);
                a.currentJSONMetadata = json;
            }

            // // some frames (editor) might request not to change the innerHTML until an event is fired
            // $('editor').setStyle('display', 'none');
            // if (typeof a.currentJSONMetadata == 'object' && a.currentJSONMetadata.wait) 
            //     document.addEvent(a.currentJSONMetadata.wait, function () {
            //         $('editor').setStyle('display', 'block');
            // });

            // set the content
            $('content').innerHTML = html;

            // apply click handlers
            addFrameClickHandler($('content').getElements('a'));

            // find HTML metadata
            var meta = $('content').getElement('meta');
            if (meta) {
                var attrs = meta.attributes;
				[].forEach.call(meta.attributes, function(attr) {
				    if (/^data-/.test(attr.name))
				        attrs[attr.name] = attr.value;
				});

                // // Tools for all pages
                // 'data-redirect',    // javascript frame redirect
                // 'data-wredirect',   // window redirect
                // 'data-nav',         // navigation item identifier to highlight
                // 'data-title',       // page title for top bar
                // 'data-icon',        // page icon name for top bar
                // 'data-scripts',     // SSV script names w/o extensions
                // 'data-styles',      // SSV css names w/o extensions
                // 'data-flags',       // SSV page flags
				// 'data-search', 		// name of function to call on search
				// 'data-buttons', 	// buttons to display in top bar
				//
                // // Used by specific pages
				//
                // 'data-sort'         // sort option, used by file lists

                handlePageData(attrs);
            }
        },
        onFail: function (html) {
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
			if (!a.currentData || !a.currentData['data-buttons'])
				return;
			SSV(a.currentData['data-buttons']).each(function (buttonID) {
				
				// find opts
				var buttonStuff = a.currentData['data-button-' + buttonID];
				if (!buttonStuff) {
					console.warn('Button ' + buttonID + ' is not configured');
					return;
				}
				
				// parse psuedo-JSON
				buttonStuff = JSON.decode(buttonStuff.replace(/'/g, '"'));
				if (!buttonStuff) {
					console.warn('Failed to parse JSON for button ' + buttonID);
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
					href: buttonStuff.href || '#'
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
				
				// click click event if callback func is provided
				if (buttonStuff.func) anchor.addEvent('click', function (e) {
					e.preventDefault();
					var func = window[buttonStuff.func];
					if (!func) {
						console.warn(
							'Button ' + buttonID + ' function ' +
							buttonStuff.func + ' does not exist'
						);
						return;
					}
					func();
				});
            
                // inject the button
				but.inject($$('.top-button').getLast(), 'after');
			});
		},
		destroy: function () {
			$$('.top-button.injected').each(function (but) {
				but.destroy();
			});
		}
	}
};

// handle page data on request completion
function handlePageData (data) {
    pageScriptsDone = false;

    console.log(data);
    a.currentData = data;
    $('content').setStyle('user-select', 'none');

    // window redirect
    var target = data['data-wredirect'];
    if (target) {
        console.log('Redirecting to ' + target);
        window.location = target;
        return;
    }

    // page redirect
    target = data['data-redirect'];
    if (target) {
        console.log('Redirecting to ' + target);
        history.pushState(page, '', adminifier.wikiRoot + '/' + target);
        loadURL();
        return;
    }

    // page title and icon
    a.updatePageTitle(data['data-title']);
    window.scrollTo(0, 0);
    // ^ not sure if scrolling necessary when setting display: none

    // highlight navigation item
    var li = $$('li[data-nav="' + data['data-nav'] + '"]')[0];
    if (li) {
        $$('li.active').each(function (li) { li.removeClass('active'); });
        li.addClass('active');
    }

    // destroy old scripts
    $$('script.dynamic').each(function (script) { script.destroy(); });

    // load new scripts
	firedPageScriptsLoaded = false;
    a.loadScripts(SSV(data['data-scripts']));


    // destroy old styles
    $$('link.dynamic').each(function (link) { link.destroy(); });

    // inject new styles
    SSV(data['data-styles']).each(function (style) {
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
    SSV(data['data-flags']).each(function (flagName) {
        var flag = flagOptions[flagName];
        if (!flag) return;
        a.currentFlags.push(flag);
        flag.init();
    });
}

// escape key pressed
function handleEscapeKey (e) {
    if (e.key != 'esc')
        return;
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
	if (!a.currentData || !a.currentData['data-search'])
        return;
        
    // the page provides no search handler
	var searchFunc = window[a.currentData['data-search']];
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

})(adminifier);
