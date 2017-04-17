# quiki

quiki is a fully-featured wiki suite and standalone web server that is
completely file-based. instead of storing content in a database, each page i
represented by a text file written in a clean and productive source language.

the underlying [wikifier](https://github.com/cooper/wikifier) wiki engine offers
image generation, category management, templates, markdown integration,
git-based revision tracking, and much more.

* [install](#install)
* [configure](#configure)
* [run](#run)

## install

you need [perl](http://perl.org),
[cpanm](https://metacpan.org/pod/App::cpanminus), and [go](http://golang.org).

the wikifier engine is included as a submodule of this repository, so it does
not need to be installed manually.

install wikifier dependencies:
```
cpanm GD Git::Wrapper HTTP::Date HTML::Strip HTML::Entities JSON::XS URI::Escape
```

install quiki:
```
go get github.com/cooper/quiki
cd $GOPATH/src/github.com/cooper/quiki
git submodule update --init
```

## configure

copy and edit the example configuration:
```
cp quiki.conf.example quiki.conf
nano -w quiki.conf
```

there is also a detailed [configuration spec](doc/configuration.md).

## run

run quiki from your GOPATH:
```
$GOPATH/bin/quiki quiki.conf
```
