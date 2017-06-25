# quiki

quiki is a fully-featured wiki suite and standalone web server that is
completely file-based. instead of storing content in a database, each page is
represented by a text file written in a clean and productive
[source language](https://github.com/cooper/wikifier/blob/master/doc/language.md).

the underlying [wikifier](https://github.com/cooper/wikifier) engine offers
image generation, category management,
[templates](https://github.com/cooper/wikifier/blob/master/doc/models.md),
[markdown integration](https://github.com/cooper/wikifier/blob/master/doc/markdown.md),
git-based revision tracking, and much more.
[adminifier](https://github.com/cooper/adminifier), a sister project, is a wiki
administrative panel featuring a web-based editor.

* [install](#install)
* [configure](#configure)
* [run](#run)

## install

you need [perl](http://perl.org), [go](http://golang.org), and (preferably)
[cpanm](https://metacpan.org/pod/App::cpanminus):
```sh
apt-get install perl golang # or similar
curl -L https://cpanmin.us | perl - --sudo App::cpanminus
```

install wikifier dependencies:
```sh
apt-get install libgd-dev # or similar
cpanm GD Git::Wrapper HTTP::Date HTML::Strip HTML::Entities JSON::XS URI::Escape
```

install [wikifier](https://github.com/cooper/wikifier):
```sh
git clone https://github.com/cooper/wikifier.git
```

install quiki:
```sh
go get github.com/cooper/quiki
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
