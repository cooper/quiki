(function (a) {

    document.addEvent('editorLoaded', loadedHandler);
    document.addEvent('pageUnloaded', unloadedHandler);

    var ae;
    function loadedHandler () {
        ae = a.editor;
    
        // add toolbar functions
        ae.addToolbarFunctions({ emoji: displayEmojiSelector });
    }

    function unloadedHandler () {
        document.removeEvent('editorLoaded', loadedHandler);
        document.removeEvent('pageUnloaded', unloadedHandler);
    }

    var emojiList = {
        'stuck_out_tongue_winking_eye': '\u{1F61C}',
    }

    // EMOJI SELECTOR

    function displayEmojiSelector () {
            
        // create box
        var li  = ae.liForAction('emoji');
        var box = ae.createPopupBox(li);
        ae.fakeAdopt(box); // for injectInto
        ae.setLiLoading(li, true);

        // create color elements
        Object.each(emojiList, function (shortcode, char) {
            var div = new Element('div', { text: char });
            box.appendChild(div);

            // add click event
            // div.addEvent('click', 
        });

        box.setStyle('width', '395px');
        ae.setLiLoading(li, false);
        ae.displayPopupBox(box, 290, li);
    }
})(adminifier);
    
