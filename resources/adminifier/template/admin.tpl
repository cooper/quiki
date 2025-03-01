[[ template "header.tpl" . ]]
<script>
var adminifier = {
    adminRoot:      '[[.AdminRoot]]',
    staticRoot:     '[[.Static]]',
    wikiRoot:       '[[.AdminRoot]]',
    wikiName:       '[[.Title]]',
    homePage:       'sites'
};
</script>
[[ template "scripts.tpl" . ]]
</head>
<body>

<div id="top-bar">
    <span class="top-title account-title"><a href="#"><i class="fa fa-user"></i> [[.User.DisplayName]]</a></span>
    <input id="top-search" type="text" placeholder="Quick Search..." />
    <span class="top-title wiki-title">[[.Title]]</span>
    <span id="page-title" class="top-title page-title"><i class="fa fa-home"></i> <span></span></span>
</div>

<div id="navigation-sidebar">
    <ul id="navigation">
        <li data-nav="sites"><a class="frame-click" href="[[.AdminRoot]]/sites"><i class="fa fa-globe-americas"></i> <span>Sites</span></a></li>
        <li data-nav="routes"><a class="frame-click" href="[[.AdminRoot]]/routes"><i class="fa fa-route"></i> <span>Routes</span></a></li>
        <li data-nav="help"><a class="frame-click" href="[[.AdminRoot]]/help"><i class="fa fa-question-circle"></i> <span>Help</a></li>
        <li><a href="[[.AdminRoot]]/logout"><i class="fa fa-arrow-circle-left"></i> <span>Logout</span></a></li>
    </ul>
</div>

<div id="content">
</div>

<template id="tmpl-notification">
    <h3>{%= o.title %}</h3>
    <i class="fa fa-{%= o.icon %}"></i>
    <div>{%= o.message %}</div>
</template>

<template id="tmpl-login-check">
    <div id="login-window-check"><i class="fa fa-check-circle fa-5x center"></i></div>
</template>

</body>
</html>
