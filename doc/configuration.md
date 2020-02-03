# Configuration

This document describes all of the available configuration options. The options
are categorized by the lowest-level quiki interface at which they are used.
Some are required for the generation of a single page, others for the operation
of a full wiki, and others yet for the operation of a webserver or frontend.

## Configuration files

The primary method of configuration is to define options in a configuration
file. All quiki configuration files are written in the quiki language:

    @name:      MyWiki;             /* assign a string option */
    @dir.page:  [@dir.wiki]/pages;  /* string option with embeded variable */
    @page.enable.cache;             /* enable a boolean option */
    -@page.enable.title;            /* disable a boolean option */

If you are using **quiki webserver**, you must have a dedicated configuration
file for the webserver. This tells it where to listen and where to find the
wikis you have configured on the server. This is typically called `quiki.conf`,
and is required as the first argument to the `quiki` executable.

**Every wiki** also requires its own configuration file, usually called
`wiki.conf` at the root level of the wiki directory.
When using webserver, the path of each wiki's configuration file is defined in
the server configuration (`quiki.conf`) by
[`server.wiki.[name].config`](#serverwikinameconfig).

Each wiki can optionally have a **private configuration** file. This is where
the administrative credentials and any other sensitive data can be stored.
This file is located elsewhere than the wiki directory so as to void any
possibility of it being served to HTTP clients.
When using webserver, this is defined by
[`server.wiki.[name].private`](#serverwikinameprivate).

## wikifier options

These options are available to the wikifier engine, the lowest level API 
for rendering pages.

### name

Name of the wiki.

__Default__: *Wiki*

### host

| Option        | Description   | Default          |
| -----         | -----         | -----            |
| `host.wiki`   | Wiki host     | None (all hosts) |

Hostname for the wiki.

It may be overridden in the server configuration by
[`server.wiki.[name].host`](#serverwikinamehost). If specified in neither
place, the wiki is accessible from all available hosts.

### root

| Option        | Description   | Default        |
| -----         | -----         | -----          |
| `root.wiki`   | Wiki root     | None (i.e. /)  |
| `root.page`   | Page root     | */page*        |
| `root.image`  | Image root    | */images*      |
| `root.file`   | File root     | None           |

HTTP roots. These are relative to the server HTTP root, NOT the wiki root.
They are used for link targets and image URLs; they will never be used to
locate content on the filesystem. Do not include trailing slashes.

It may be useful to use `root.wiki` within the definitions of the rest:

    @root.wiki:     /mywiki;
    @root.page:     [@root.wiki]/page;
    @root.image:    [@root.wiki]/images;

If you specify `root.file`, the entire wiki directory (as specified by
[`dir.wiki`](#dir)) will be indexed by the web server at this path. Note
that this will likely expose your wiki configuration.

### dir

| Option            | Description                                           |
| -----             | -----                                                 |
| `dir.wiki`        | Wiki root directory (usually inferred)                |
| `dir.page`        | Page files stored here                                |
| `dir.image`       | Image originals stored here                           |
| `dir.model`       | Model files stored here                               |
| `dir.cache`       | Generated content and metadata stored here            |

Directories on the filesystem. It is strongly recommended that they are
absolute paths; otherwise they will be dictated by whichever directory the
program is started from. 

Best practice to achieve this is to reference `dir.wiki` within each. With
webserver, `dir.wiki` is predefined as long as your wiki exists within
[`server.dir.wiki`](#serverdirwiki).

    @dir.page:      [@dir.wiki]/pages;
    @dir.image:     [@dir.wiki]/images;
    @dir.model:     [@dir.wiki]/models;
    @dir.cache:     [@dir.wiki]/cache;

### external

External wiki information.

You can configure any number of external wikis, each referred to by a shorthand
identifier `[wiki_id]` consisting of word-like characters.

| Option                        | Description | Default
| -----                         | -----       | -----
| `external.[wiki_id].name`     | External wiki name, displayed in link tooltips |
| `external.[wiki_id].root`     | External wiki page root |
| `external.[wiki_id].type`     | External wiki type | *quiki*

Accepted values for `type`
* *quiki* (this is default)
* *mediawiki*
* *none* (URI encode only)

The default configuration includes the `wp` identifier for the
[English Wikipedia](http://en.wikipedia.org):

    @external.wp: {
        name: Wikipedia;  /* appears in tooltip */
        root: http://en.wikipedia.org/wiki;
        type: mediawiki;
    };

From the page source, this looks like:

    [[ Cats | wp: Cat ]] are a type of [[ wp: animal ]].

### page.enable.title

If enabled, the first section's heading defaults to the title of the
page, and all other headings on the page are sized down one.

You may want to disable this at the wiki level if your wiki content is
embedded within a template that has its own place for the page title:

    -@page.enable.title;

__Default__: Enabled

### image.size_method

The method which quiki should use to scale images.

This is here for purposes of documentation only; there's probably no reason
to ever stray away from the default for whichever quiki interface is used.

**Accepted values**
* _javascript_ - JavaScript-injected image sizing
* _server_ - server-side image sizing using [`image.sizer`](#imagesizer) and
  [`image.calc`](#imagecalc) (recommended)

__Default__ (webserver): _server_

__Default__ (low-level): _javascript_

### image.calc

A function reference that calculates a missing dimension of an image.
This is utilized only when [`image.size_method`](#imagesize_method) is _server_.

This is here for purposes of documentation only and can only be configured
using quiki's wikifier engine API directly.

__Default__: (webserver) built-in function

### image.sizer

A function reference that returns the URL to a sized version of an image. After
using [`image.calc`](#imagecalc) to find the dimensions of an image,
`image.sizer` is called to generate a URL for the image at those dimensions.

This is here for purposes of documentation only and can only be configured
using quiki's wikifier engine API directly.

__Default__: (webserver) built-in function

## Wiki public options

These options are available to the wiki website interface.

### image.type

The desired file type for generated images.

When configured, all resulting images will be in this format, regardless of
their original format. Unless you have a specific reason to do this, omit
this option to preserve original image formats.

**Accepted values**
* _png_ - larger, lossless compression
* _jpeg_ - smaller, lossy compression

__Default__: none (preserve original format)

### image.quality

The desired quality of generated images with lossy compression. This is only
utilized if [`image.type`](#imagetype) is set to *jpeg*.

__Default__: *100*

### image.retina

Image scales to enable for displays with a greater than 1:1 pixel ratio.

For instance, to support both @2x and @3x scaling:

    @image.retina: 2, 3;

__Default__: *2, 3*

### page.enable.cache

Enable caching of generated pages.

For best performance, page caching is enabled by default, and quiki will
only generate pages if the source file has been modified since the cache
file was last written.

__Default__: Enabled

### cat.per_page

Maximum number of pages to display on a single category posts page.

__Default__: _5_

### cat.[name].main

Set the main page for the category by the name of `[name]`.
This means that it will be displayed before all other categories, regardless
of their creation dates. The value of this option is the page filename.

You can also mark a page as the main page of a category from within the page
source itself, like so:

    @category.some_cat.main;

If multiple pages are marked as the main page of a category, the one with
the most recent creation time is preferred. If this option is provided,
however, the page specified by it will always take precedence.

__Default__: none (show newest first)

### cat.[name].title

Sets the human-readable title for the category by the name of `[name]`.

__Default__: none

### var.*

Global wiki variable space. Variables defined in this space will be
available throughout the wiki. However they may be overwritten on a
particular page.

Example (in config):

    @var.site.url: http://mywiki.example.com;
    @var.site.display_name: MyWiki;

Example (on main page):

    Welcome to [@site.display_name]!

## Wiki extended options

These options are not used by the wiki API directly but are standardized 
here so that there is consistency amongst various frontends such as webserver.

### main_page

Name of the main page.

This should not be the page's title but rather a
filename, relative to [`dir.page`](#dir). The `.page` extension is not
necessary.

```
@main_page: Welcome Page; /* normalized to welcome_page.page */
```

### main_redirect

If enabled, the wiki root redirects to the main page rather than just
rendering it at the root location.

```
@main_redirect;
```

### error_page

Name of the error page.

This should not be the page's title but rather a
filename, relative to [`dir.page`](#dir). The `.page` extension is not
necessary.

```
@error_page: Error; /* normalized to error.page */
```

### navigation

Navigation items.

Keys are unformatted text to be displayed, with spaces permitted.
Values are URLs, relative to the current page (NOT the wiki root).

It is good practice to refer to [`root`](#root):

```
@navigation: {
    Main page:  [@root.page]/welcome_page;
    Rules:      [@root.page]/rules;
};
```

You can nest maps to create sublists at any level, but the number of levels
supported depends on the frontend or template being used by the wiki.

```
@navigation: {
    Home: /page/welcome;
    About: {
        Company info: /page/our_company;
        Facebook: http://facebook.com/our.company;
    };
    Contact: /contact;
};
```

### template

Name or path of the template used by the wiki.

```
@template: default;
```

__Default__ (webserver): *default*

### logo

Filename for the wiki logo, relative to [`dir.image`](#dir).

Frontends like webserver automatically generate the logo in whatever dimensions
are needed and display it where appropriate.

```
@logo: logo.png;
```

## Wiki private options

Private wiki options may be in a separation configuration file.
This is where administrator credentials are stored.

### admin.[username].name

Real name for the administrator `[username]`.

This is used to for revision tracking to attribute page creation,
image upload, etc. to the user. Depending on the frontend and/or template,
it may be displayed to the public as the author or maintainer of a page or
the owner of some file.

### admin.[username].email

Email address of the administrator `[username]`.

Used for revision tracking.

### admin.[username].crypt

Ttype of encryption used for the password of administrator `[username]`.

**Accepted values**
* _none_ (plain text)
* _sha1_
* _sha256_
* _sha512_

__Default__: *sha1*

### admin.[username].password

Password of the administrator `[username]`.

It must be encrypted in
the crypt set by [`admin.[username].crypt`](#adminusernamecrypt).

## webserver options

These options are respected by the quiki webserver.

### server.dir.wiki

Path to some directory where wikis are stored.

Your wikis do not all have to be in the same directory, so this is optional.
However, if you make use of this, quiki can infer [`dir.wiki`](#dir) and
[`server.wiki.[name].config`](#serverwikinameconfig) for each wiki.


### server.dir.adminifier

Path to adminifier resources.

### server.enable.pregeneration

If enabled, webserver pre-generates all pages and images upon start.

__Requires__: [`page.enable.cache`](#pageenablecache)

__Default__: Enabled

### server.enable.monitor

If enabled, webserver uses operating system facilities to monitor wiki
directories and pre-generate content when the source files are changed.

__Requires__: [`page.enable.cache`](#pageenablecache)

__Default__: Enabled

### server.http.port

__Required__. Port for HTTP server to listen on.

```
server.http.port: 8080;
```

You can also say `unix` to listen on a UNIX socket.

### server.http.bind

_Optional_. Host to bind to.

```
@server.http.bind: 127.0.0.1;
```

Or, you can listen on a UNIX socket:
```
@server.http.port:  unix;
@server.http.bind:  /var/run/quiki.sock;
```

__Default__: none (bind to all available hosts)

### server.dir.template

_Optional_. Template search paths.

```
@server.dir.template: /home/www/wiki-templates;
```

This is a comma-separated list of paths to look for templates when they are
specified by name in the wiki configuration. If you're running multiple wikis
that share a template, or if you are using the default template, this optional
is useful. Otherwise, you can just specify the absolute path to each wiki's
template in the [template](#template) directive.

### server.wiki.[name].enable

Enable the wiki by the name of `[name]`.

Any wikis configured that do not have this option present are skipped.
Any number of wikis can be configured on a single server using this.

### server.wiki.[name].config

Ppath to the configuration file for the wiki by the name of `[name]`.

This is required for each wiki unless [`server.dir.wiki`](#serverdirwiki) is
set and the wiki is located there.

__Default__: [`server.dir.wiki`](#serverdirwiki)`/[name]/wiki.conf`

### server.wiki.[name].private

The path to the PRIVATE configuration file for the wiki by the name of
`[name]`.

This is where administrative credentials are stored. Be sure that the
private configuration is not inside the HTTP server root.