# Markdown

In addition to pages written in the [quiki source language](language.md),
quiki can serve content rendered from
[Markdown](https://en.wikipedia.org/wiki/Markdown).

Markdown files are stored alongside `.page` files within the wiki page directory.
They are identified by the `.md` file extension.

Rather than simply injecting the Markdown content as HTML, quiki translates
Markdown files to the quiki source language on the fly. Although this
generated page source is never stored or exposed to where you would see it, 
translation ensures that pages rendered from Markdown are formatted exactly as
would be those written directly in the quiki source language.

quiki uses [blackfriday](https://github.com/russross/blackfriday/tree/v2)
with [extensions](https://github.com/russross/blackfriday/tree/v2#extensions)
enabled to closely resemble
[GitHub Flavored Markdown](https://guides.github.com/features/mastering-markdown/).