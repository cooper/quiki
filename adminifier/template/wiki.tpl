<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<title>WIKI TITLE</title>
<link type="text/css" rel="stylesheet" href="{{.Static}}/style/adminifier.css" />
<link type="text/css" rel="stylesheet" href="{{.Static}}/style/navigation.css" />
<link type="text/css" rel="stylesheet" href="{{.Static}}/style/notifications.css" />
<link href="//fonts.googleapis.com/css?family=Open+Sans:300,400,600" rel="stylesheet" type="text/css" />
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" />
<script type="text/javascript">

var adminifier = {
    adminRoot:      null,
    wikiShortName:  null,
    wikiName:       null,
    wikiPageRoot:   null,
    themeName:      null,
    autosave:       3000000

};

</script>
<script type="text/javascript" src="{{.Static}}/ext/mootools.js"></script>
<script type="text/javascript" src="{{.Static}}/ext/tmpl.min.js"></script>
<script type="text/javascript" src="{{.Static}}/script/adminifier.js"></script>
<script type="text/javascript" src="{{.Static}}/script/notifications.js"></script>
<script type="text/javascript" src="{{.Static}}/script/modal-window.js"></script>

</head>
<body>

<div id="top-bar">
    <span class="top-title account-title"><a href="#"><i class="fa fa-user"></i> USERNAME</a></span>
    <span class="top-title top-button"><a class="frame-click" href="#"><i class="fa fa-git-square"></i> master</a></span>
    <span class="top-title top-button"><a class="frame-click" href="{{.Root}}/create-page"><i class="fa fa-plus-circle"></i> New page</a></span>
    <input id="top-search" type="text" placeholder="Quick Search..." />
    <span class="top-title wiki-title">WIKI TITLE</span>
    <span id="page-title" class="top-title page-title"><i class="fa fa-home"></i> <span></span></span>
</div>

<div id="navigation-sidebar">
    <ul id="navigation">
        <li data-nav="dashboard"><a class="frame-click" href="{{.Root}}/dashboard"><i class="fa fa-home"></i> <span>Dashboard</span></a></li>
        <li data-nav="pages"><a class="frame-click" href="{{.Root}/pages"><i class="fa fa-file-text"></i> <span>Pages</span></a></li>
        <li data-nav="categories"><a class="frame-click" href="{{.Root}/categories"><i class="fa fa-list"></i> <span>Categories</span></a></li>
        <li data-nav="images"><a class="frame-click" href="{{.Root}/images"><i class="fa fa-picture-o"></i> <span>Images</span></a></li>
        <li data-nav="models"><a class="frame-click" href="{{.Root}/models"><i class="fa fa-cube"></i> <span>Models</span></a></li>
        <li data-nav="settings"><a class="frame-click" href="{{.Root}/settings"><i class="fa fa-cog"></i> <span>Settings</a></li>
        <li data-nav="help"><a class="frame-click" href="{{.Root}}/help"><i class="fa fa-question-circle"></i> <span>Help</a></li>
        <li><a href="logout"><i class="fa fa-arrow-circle-left"></i> <span>Logout</span></a></li>
    </ul>
</div>

<div id="content">
</div>

{{/* ALL OF THE JS TEMPLATES */}}

</body>
</html>
