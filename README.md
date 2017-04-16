# quiki

a standalone web server for [wikifier](https://github.com/cooper/wikifier)

* [install](#install)
* [configure](#configure)
* [run](#run)
* [server configuration](#server-configuration)
  * [server\.http\.port](#serverhttpport)
  * [server\.http\.bind](#serverhttpbind)
  * [server\.dir\.template](#serverdirtemplate)
  * [server\.dir\.wikifier](#serverdirwikifier)
  * [server\.dir\.wiki](#serverdirwiki)
  * [server\.wiki\.[name]\.enable](#serverwikinameenable)
  * [server\.wiki\.[name]\.host](#serverwikinamehost)
  * [server\.wiki\.[name]\.config](#serverwikinameconfig)
  * [server\.wiki\.[name]\.password](#serverwikinamepassword)
* [wiki configuration](#wiki-configuration)
  * [name](#name)
  * [template](#template)
  * [main\_page](#main_page)

## install

install wikifier dependencies
```
cpanm GD Git::Wrapper HTTP::Date HTML::Strip HTML::Entities JSON::XS URI::Escape
```

install quiki
```
go get github.com/cooper/quiki
cd $GOPATH/src/github.com/cooper/quiki
git submodule update --init
```

## configure

```
cp quiki.conf.example quiki.conf
nano -w quiki.conf
```

## run

```
$GOPATH/bin/quiki quiki.conf
```

## server configuration

quiki works by running a wikiserver as a subprocess and communicating with it
via standard I/O. quiki and the underlying wikiserver share a configuration
file. In addition to the
[existing wikiserver options](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#wikifierserver-options),
quiki adds these:

### server.http.port

```
server.http.port: 8080;
```

__Required__. Port to run the HTTP server on.

### server.http.bind

```
@server.http.bind: 127.0.0.1;
```

_Optional_. Host to bind to. Defaults to all available hosts.

### server.dir.template

```
@server.dir.template: /home/www/wiki-templates;
```

_Optional_. Template search paths.

This is a comma-separated list of paths to look for templates when they are
specified by name in the wiki configuration. If you're running multiple wikis
that share a template, or if you are using the default template, this optional
is useful. Otherwise, you can just specify the absolute path to each wiki's
template in the [template](#template) directive.

If you are using a template packaged with quiki, such as the default one,
do something like this:
```
@gopath: /home/me/go;
@server.dir.template: [@gopath]/src/github.com/cooper/quiki/templates;
```

### server.dir.wiki

```
@server.dir.wiki: /home/www/wikis;
```

__Required__. Directory where wikis are stored.

Technically this is optional if you [provide](#serverwikinameconfig) an absolute
path to each wiki's configuration file.

### server.dir.wikifier

```
@server.dir.wikifier: /home/www/wikifier;
```

__Required__. Absolute path to the [wikifier](https://github.com/cooper/wikifier).

quiki needs this to run the wikiserver and to serve the static resources bundled
with wikifier.

### server.wiki.[name].enable

```
@server.wiki.mywiki.enable;
```

Tells quiki to serve the wiki by the name of `[name]`. If no wikis are enabled,
quiki will not start.

### server.wiki.[name].host

```
@server.wiki.mywiki.host: some.host.com;
```

_Optional_. Host on which to serve the wiki by the name of `[name]`.

quiki can serve any number of wikis, so long as their
[roots](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#root)
do not collide. Specifying a host allows wikis to have same roots, since quiki
can respect the Host header.

### server.wiki.[name].config

```
@server.wiki.mywiki.config: /home/www/wikis/mywiki/wiki.conf;
```

_Optional_. Absolute path to the configuration file for the wiki by the name of
`[name]`.

This is only technically optional when [`server.dir.wiki`](#serverdirwiki) is
set. When you do not specify the configuration file path explicitly, it is
assumed to be at `[server.dir.wiki]/[wiki name]/wiki.conf`.

If this is specified, your wiki does not necessarily have to be within the
[`server.dir.wiki`](#serverdirwiki) directory because quiki can use the
[`dir.wiki`](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#root)
directive to find the wiki.

### server.wiki.[name].password

```
@server.wiki.mywiki.password: secret;
```

_Optional_. Password for read access to the wiki.

wikiserver normally uses a password for read authentication, but since quiki
communicates with it via standard I/O, this is bypassed. quiki does not need
a password for each wiki unless the wikiserver runs independently of quiki and
is reached via a UNIX socket.

You may still need to
set a password though for other things which connect to the wikiserver via
sockets (such as [adminifier](https://github.com/cooper/adminifier)).

## wiki configuration

quiki reads the wiki configuration files associated with each enabled wiki.
quiki supports these wiki options, all of which are _optional_:

### name

```
@name: My Wiki;
```

Wiki option
[`name`](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#name).

quiki uses this in the `<title>` tag on most pages and possibly other places.

### template

```
@template: default;
```

Wiki extended option
[`template`](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#template).

Specifies the template to be used on the wiki. This may be an absolute path to
the template or a template name. If only a name is given, the directories in
[`server.dir.template`](#serverdirtemplate) will be searched.

If you do not specify a template at all, the
[default template](templates/default) will be assumed.

### main_page

```
@main_page: some_page;
```

Wiki extended option
[`main_page`](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#main_page).

Name of the main page. This should not be the page's title but rather a
filename, relative to [`dir.page`](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#dir).
The `.page` extension is not necessary.

### navigation

```
@navigation: {
    Main page: /page/welcome;
    Rules: /page/rules;
};
```

Wiki extended option
[`navigation`](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#navigation).

Map of navigation items. Keys are the displayed text; values are the URL. The
URLs are relative to the current page (i.e., they are used unchanged as the
`href` attribute).

Currently quiki only supports top-level navigation items.
