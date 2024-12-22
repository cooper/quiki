# quiki

[quiki](https://quiki.app) is a wiki suite and standalone web server that is
completely file-based. instead of storing content in a database, each page is
represented by a text file written in the clean and productive
[quiki source language](doc/language.md) or [markdown](doc/markdown.md).

it sports caching, image generation, category management, [templates](doc/models.md),
git-based revision tracking, and more. while it is meant to be easily maintainable
from the command line, you may optionally enable the web-based editor.

* [install](#install)
* [configure](#configure)
* [run](#run)

## install

```sh
go get github.com/cooper/quiki
```

## configure

create a `quiki.conf` configuration file based on the
[provided example](quiki.conf.example) and place it somewhere readable by the user
that will run quiki.

see the [configuration spec](doc/configuration.md) for all options.

## run

```sh
quiki quiki.conf
```