(function (a) {

var pingCount = 0;

// FIXME: Disabled for now.
// document.addEvent('domready', function () {
//     setInterval(pingServer, 20000);
//     pingServer();
// });

function pingServer () {
    new Request.JSON({
        url:        'func/events',
        secure:     true,
        onSuccess:  function (data) {
            
            // not connected, so ask for login
            if (!data.connected) {
                
                // if this is the first time we've tried, redirect to login page
                if (!pingCount) {
                    window.location.href = a.adminRoot + '/login';
                    return;
                }
                
                displayLoginWindow();
                return;
            }
            
            // show notifications
            if (data.notifications) data.notifications.each(function (noti) {
                displayNotification(noti);
            });
        },
        onError:    displayLoginWindow,
        onFailure:  displayLoginWindow
    }).get();
    pingCount++;
}

var icons = {
    user_logged_in: 'user'
};

var formats = {
    user_logged_in: function (noti) {
        if (noti.name != null && noti.name.length)
            str = noti.name + ' (' + noti.username + ')';
        else
            str = noti.username;
        return str + ' has just logged in.';
    }
};

function formatNotification (noti) {
    if (formats[noti.type])
        return formats[noti.type](noti);
    return noti.message;
}

var notificationQueue = [];
function displayNotification (noti) {

    // there is already a notification being displayed,
    // so add this to the queue.
    if ($('notification-popup')) {
        notificationQueue.push(noti);
        return;
    }

    var titleType = noti.type.replace(/_/g, ' ');
    titleType = titleType.charAt(0).toUpperCase() + titleType.slice(1);

    var popup = new NotificationPopup({
        title:          noti.title || titleType,
        message:        formatNotification(noti),
        icon:           icons[noti.icon] || 'info-circle',
        hideAfter:      noti.timeout || 10000,
        autoDestroy:    true,
        onDestroyed:    function () {
            var nextNoti = notificationQueue.shift();
            if (nextNoti) displayNotification(nextNoti);
        }
    });

    popup.show();
}

function displayLoginWindow () {
    if ($('login-window'))
        return;

    // create login modal window
    var loginWindow = new ModalWindow({
        icon:       'user',
        title:      'Login',
        html:       tmpl('tmpl-login-window', {}),
        padded:     true,
        sticky:     true,
        doneText:   null,
        id:         'login-window',
        width:      '400px'
    });

    // if some error we can't deal with occurs, redirect to real login
    var giveUp = function () {
        window.location = 'logout';
    };

    // attempt to login remotely
    var inputs = loginWindow.content.getElements('input');
    var attemptLogin = function () {
        
        // if the submit button is disabled, do not allow this
        if (loginWindow.content.getElement('input[type=submit]').get('disabled'))
            return;
        
        // disable the inputs temporarily
        inputs.each(function (i) { i.set('disabled', true); });

        var req = new Request.JSON({
            url: 'func/login',
            secure: true,
            onSuccess: function (data) {
                
                // incorrect credentials probably
                if (!data.success) {
                    inputs.each(function (i) {
                        
                        // enable submit button after 1 second
                        if (i.get('type') == 'submit') {
                            setTimeout(function() {
                                i.set('disabled', false);
                            }, 1000);
                        }
                        
                        // flash text fields and immediately enable them
                        else {
                            i.flash('#F78383', '#fff');
                            i.set('disabled', false);
                        }
                        
                    });
                    return;
                }
                loginWindow.content.innerHTML = tmpl('tmpl-login-check', {});
                setTimeout(function () {
                    loginWindow.destroy(true);
                }, 1000);
            },
            onError: giveUp,
            onFailure: giveUp
        });
        req.post({
            remote:   true,
            username: document.getElement('input[name=username]').get('value'),
            password: document.getElement('input[name=password]').get('value')
        });
    };

    // capture enters and clicks
    inputs.each(function (input) {
        if (input.get('type') == 'submit')
            input.addEvent('click', attemptLogin);
        else
            input.onEnter(attemptLogin);
    });

    loginWindow.show();
}

var NotificationPopup = window.NotificationPopup = new Class({

    Implements: [Options, Events],

    options: {
        title:          'Information',
        icon:           'info-circle',
        autoDestroy:    false,
        hideAfter:      0,
        sticky:         false
    },

    initialize: function (opts) {
        this.popup = new Element('div', { id: 'notification-popup' });
        this.setOptions(opts);
    },

    setOptions: function (opts) {
        Options.prototype.setOptions.call(this, opts);
        opts = this.options;
        var _this = this;
        if (opts.hideAfter)
            setTimeout(function () { _this.hide(); }, opts.hideAfter);
        this.popup.innerHTML = tmpl('tmpl-notification', opts);
    },

    show: function (container) {
        if (!container)
            container = document.body;
        container.adopt(this.popup);
        this.popup.setStyles({
            display: 'block',
            opacity: 0
        });
        this.popup.fade('in');
        this.shown = true;
    },

    hide: function (isDestroy) {
        if (!this.shown || this.options.sticky)
            return;
        delete this.shown;
        this.popup.fade('out');
        var _this = this;
        setTimeout(function () {
            _this.popup.setStyle('display', 'none');
            _this.fireEvent('done');
            if (_this.options.autoDestroy && !isDestroy)
                _this._destroy();
        }, 500);
    },

    destroy: function (force) {
        if (this.options.sticky && !force)
            return;
        delete this.options.sticky;
        this.hide(true);
        this._destroy();
    },

    _destroy: function () {
        this.popup.destroy();
        this.fireEvent('destroyed');
    }

});

})(adminifier);
