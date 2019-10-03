# quiki

quiki is a fully-featured wiki suite and standalone web server that is
completely file-based. instead of storing content in a database, each page
is represented by a text file written in the clean and productive
[quiki source language](doc/language.md).

it sports caching, image generation, category management,
[templates](doc/models.md),
[markdown integration](doc/markdown.md),
git-based revision tracking, and much more.
[adminifier](https://github.com/cooper/adminifier), a sister project, is an
administrative panel featuring a web-based editor.

* [install](#install)
* [configure](#configure)
* [run](#run)

## install

```sh
go get github.com/cooper/quiki
```

## configure

quiki ships with a working example configuration.  
there is also a detailed [configuration spec](doc/configuration.md).

```sh
cp quiki.conf.example quiki.conf
nano -w quiki.conf
```

## run

```sh
quiki quiki.conf    # ($GOPATH/bin/quiki if PATH not configured for go)
```
