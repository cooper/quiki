<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <title>{{.VisibleTitle}}</title>
    <link rel="stylesheet" type="text/css" href="{{.StaticRoot}}/style.css" />
    <link rel="stylesheet" type="text/css" href="/styles/wiki.css" />
    {{with .PageCSS}}
    <style type="text/css">
    {{.}}
    </style>
    {{end}}
    <script type="text/javascript" src="{{.StaticRoot}}/retina.min.js"></script>
</head>

<body>
<div id="container">

    <div id="header">
        <ul id="navigation">
            <li><a href="/">Main page</a></li>
        </ul>
        <a href="/">
            <img src="/file/logo.png" alt="Wiki" />
        </a>
    </div>

    <div id="content">
