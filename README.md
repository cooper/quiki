# quiki

a standalone web server for [wikifier](https://github.com/cooper/wikifier)

* [install](#install)
* [configure](#configure)
  * [server\.http\.port](#serverhttpport)
  * [server\.http\.bind](#serverhttpbind)
  * [server\.dir\.template](#serverdirtemplate)
  * [server\.wiki\.[name]\.quiki](#serverwikinamequiki)
  * [server\.wiki\.[name]\.template](#serverwikinametemplate)
* [run](#run)

## install

```
go get github.com/cooper/quiki
```

## configure

quiki uses
[the same configuration](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#wikifierserver-options)
as the wikiserver. In addition to the existing wikiserver options, quiki adds these:

### server.http.port

__Required__. Port to run the HTTP server on.

### server.http.bind

_Optional_. Host to bind to. Defaults to all available hosts.

### server.dir.template

__Required__. Absolute path to the template directory.

If you are using a template packaged with quiki, do something like this:
```
@gopath: /home/me/go;
@server.dir.template: [@gopath]/src/github.com/cooper/quiki/templates;
```

### server.wiki.[name].quiki

__Required__. Boolean option which enables quiki on the wiki by the name of
`[name]`.

quiki can serve any number of wikis, so long as their
[roots](https://github.com/cooper/wikifier/blob/master/doc/configuration.md#root)
do not collide. Since quiki shares a configuration with the wikiserver, this
option tells quiki which wikis it should serve. If no wikis are enabled, quiki
will not start.

### server.wiki.[name].template

_Optional_. Specifies the template to be used on the wiki by the name of
`[name]`. This is relative to [`server.dir.template`](#serverwikinametemplate).

If you do not specify, the [default template](templates/default) will be
assumed.

## run

```
quiki /path/to/wikiserver.conf
```
