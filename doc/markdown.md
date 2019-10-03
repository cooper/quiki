# Markdown

In addition to wiki source files, wikifier can serve content from documents
written in Markdown.

Markdown files are stored in a dedicated directory, determined by the
[`dir.md`](configuration.md#dir) directive. As with page files, the wikiserver
monitors the markdown file directory and generates them immediately upon
changes.

Rather than simply injecting the Markdown content as HTML, wikifier parses and
translates each Markdown file to the wikifier source language. These generated
wiki source files are then served just like any other page file.

wikifier uses the [cmark](https://github.com/jgm/cmark) Markdown parser, which
implements the [CommonMark](http://commonmark.org) specification. wikifier also
supports [GFM](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown)
tables via a custom extension.

## Setup

First, you need [cmark](https://github.com/jgm/cmark). Follow that link and
look at the README for details on how to install it, but here is an example
for a systemwide installation:

```bash
apt-get install cmake
git clone https://github.com/jgm/cmark
cd cmark
mkdir build
cd build
cmake ..
make
make test
make install
```

Now install the [CommonMark](https://metacpan.org/pod/CommonMark) module from
CPAN:

```bash
cpanm CommonMark
```

If the module from CPAN fails to install, this is likely because it was unable
to locate cmark on the system.
