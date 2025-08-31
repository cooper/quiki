# Configuration

This document describes all of the available configuration options. The options
are categorized by the lowest-level quiki interface at which they are used.
Some are required for the generation of a single page, others for the operation
of a full wiki, and others yet for the operation of a webserver or frontend.

## Configuration files

The primary method of configuration is to define options in a configuration
file. All quiki configuration files are written in the quiki language:

    @name:      MyWiki;             /* assign a string option */
    @page.enable.cache;             /* enable a boolean option */
    -@page.enable.title;            /* disable a boolean option */

If you are using **quiki webserver**, you must have a dedicated configuration
file for the webserver. This tells it where to listen and where to find the
wikis you have configured on the server. This file is called `quiki.conf` and
is located in your quiki directory (default: `~/quiki/quiki.conf`).

**Every wiki** also requires its own configuration file called
`wiki.conf` at the root level of the wiki directory.

## wikifier options

These options are available to the wikifier engine, the lowest level API 
for rendering pages.

### name

Name of the wiki.

__Default__: *Wiki*

### host.wiki

Hostname for the wiki.

It may be overridden in the server configuration by
[`server.wiki.[name].host`](#serverwikinamehost).

If specified in neither place, the wiki is accessible from all available hosts.

__Default__: None (all hosts)

### dir.wiki

Path to the wiki.

With webserver, this can be omitted if either:

1. [`server.dir.wiki`](#serverdirwiki) is configured, and the wiki is located
   in that directory.
2. [`server.wiki.[name].dir`](#serverwikinamedir) is configured.

In all other cases, it is __Required__.

__Default__ (webserver): [`server.dir.wiki`](#serverdirwiki)`/[name]`

### root

| Option        | Description   | Default        |
| -----         | -----         | -----          |
| `root.wiki`   | Wiki root     | None (i.e. /)  |
| `root.page`   | Page root     | */page*        |
| `root.image`  | Image root    | */images*      |
| `root.file`   | File root     | None           |

_Optional_. HTTP roots. These are relative to the server HTTP root, NOT the
wiki root. They are used for link targets and image URLs; they will never be
used to locate content on the filesystem. Do not include trailing slashes.

It may be useful to use `root.wiki` within the definitions of the rest:

    @root.wiki:     /mywiki;
    @root.page:     [@root.wiki]/page;
    @root.image:    [@root.wiki]/images;

If you specify `root.file`, the entire wiki directory (as specified by
[`dir.wiki`](#dirwiki)) will be indexed by the web server at this path. Note
that this will likely expose your wiki configuration.

### root.ext

_Optional_. The full external prefix of the wiki.

It is used by adminifier to link to the wiki. It may also be used by frontends.
You should configure this if serving quiki through a reverse proxy. Otherwise, the
default should probably suffice. Do not include trailing slash.

```
@root.ext: https://mywiki.example.com;
```

__Default__ (webserver/adminifier): `http://<host.wiki || server.http.host>:<server.http.port>/<root.wiki>`

### external

_Optional_. External wiki information.

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

_Optional_. If enabled, the first section's heading defaults to the title of the
page, and all other headings on the page are sized down one.

You may want to disable this at the wiki level if your wiki content is
embedded within a template that has its own place for the page title:

    -@page.enable.title;

### page.code.lang

_Optional_. The default language to use for syntax highlighting of
[`code{}`](blocks.md#code) blocks.

See [this site](https://github.com/alecthomas/chroma#supported-languages) for
a list of supported languages.

### page.code.style

_Optional_. The default style to use for syntax highlighting of
[`code{}`](blocks.md#code) blocks.

See [this site](https://xyproto.github.io/splash/docs/index.html) for
a list of available styles.

__Default__: Enabled

### image.size_method

_Optional_. The method which quiki should use to scale images.

This is here for purposes of documentation only; there's probably no reason
to ever stray away from the default for whichever quiki interface is used.

**Accepted values**
* _javascript_ - JavaScript-injected image sizing
* _server_ - server-side image sizing using [`image.sizer`](#imagesizer) and
  [`image.calc`](#imagecalc) (recommended)

__Default__ (webserver): _server_

__Default__ (low-level): _javascript_

### image.calc

_Optional_. A function reference that calculates a missing dimension of an image.
This is utilized only when [`image.size_method`](#imagesize_method) is _server_.

This is here for purposes of documentation only and can only be configured
using quiki's wikifier engine API directly.

__Default__: (webserver) built-in function

### image.max_concurrent

_Optional_. Maximum number of concurrent image processing operations.

Limits how many images can be processed simultaneously to prevent system
overload and memory exhaustion. quiki also monitors available memory and throttles accordingly.

__Default__: 2

### image.processor

_Optional_. The image processing backend to use.

Controls which image processing engine is used for resizing and manipulation. By default, we try processors in order of performance:

**Accepted values**
* _auto_ - Automatic selection: libvips -> imagemagick -> pure go (recommended)
* _vips_ - Use libvips only (highest performance, 4-8x faster than ImageMagick)
* _imagemagick_ - Use ImageMagick only (high performance as compared to Go)
* _go_ - Use Pure Go only (most portable, worst performance, crashes with large images)

**Installation**
- _vips_: `brew install vips` or `apt-get install libvips-tools`
- _imagemagick_: `brew install imagemagick` or `apt-get install imagemagick`

__Example__: `@image.processor: vips;`

__Default__: auto

### image.quality

_Optional_. JPEG quality setting for libvips/ImageMagick processors (1-100).

Only used when `processor` is set to "vips", "imagemagick", or "auto" (when libvips/ImageMagick is available). Higher values mean better quality but larger file sizes to store and transfer.

__Example__: `@image.quality: 85;`

__Default__: 85

### image.max_memory_mb

_Optional_. Maximum memory usage per image in megabytes.

Prevents crashes from loading extremely large images by checking dimensions
before processing. An image requiring more memory will be rejected. Additionally,
quiki monitors memory usage and reduces concurrency when memory is low to prevent crashing.

__Default__: 256

### image.timeout_seconds

_Optional_. Maximum processing time per image operation in seconds.

Prevents hanging on slow image operations. Processing will be aborted if
it takes longer than this limit.

__Default__: 20

### image.pregen_thumbs

_Optional_. Comma-separated list of thumbnail sizes to pregenerate automatically.

This setting controls which thumbnail sizes are generated in the background
when images are first processed, improving performance for common sizes.

Each entry can be either:
* A single number (e.g., "250") - constrains the larger dimension to this size
* Exact dimensions (e.g., "400x300") - generates exactly this size

The default value of "250" ensures adminifier gallery previews load quickly.
Additional sizes can be added for custom interfaces or common usage patterns.

__Example__: `@image.pregen_thumbs: 250,150,400x300,800;`

__Default__: 250 (for adminifier thumbnail preloading)

### image.sizer

_Optional_. A function reference that returns the URL to a sized version of an image. After
using [`image.calc`](#imagecalc) to find the dimensions of an image,
`image.sizer` is called to generate a URL for the image at those dimensions.

This is here for purposes of documentation only and can only be configured
using quiki's wikifier engine API directly.

__Default__: (webserver) built-in function

## Wiki public options

These options are available to the wiki website interface.

### auth.enable

_Optional_. When enabled, the wiki has its own user management system and can create 
and manage users independently. When disabled, user management is prohibited, and all 
HTTP users are assumed to have read access to the wiki.

    @auth.enable;     /* enable wiki user management (default) */
    -@auth.enable;    /* disable user management, assume public read access */

__Default__: Enabled (but with public viewing / no login requirement)

### auth.require

_Optional_. When enabled, users must be logged in to view the wiki content. When disabled 
(default), the wiki is publicly accessible and can be viewed by anyone.

`auth.require` implies `auth.enable` also.

    @auth.require;     /* require authentication to view wiki */
    -@auth.require;    /* public wiki (default) */

__Default__: Disabled (public)

### auth.register

_Optional_. When enabled, allows new users to register accounts through the web. 
This option only has effect when `auth.enable` is also enabled.

Further, if you wish to restrict access to authenticated users,
see `auth.require`.

    @auth.register;    /* allow web registration */
    -@auth.register;   /* no web registration (default) */

__Default__: Disabled

### image.type

_Optional_. The desired file type for generated images.

When configured, all resulting images will be in this format, regardless of
their original format. Unless you have a specific reason to do this, omit
this option to preserve original image formats.

**Accepted values**
* _png_ - larger, lossless compression
* _jpeg_ - smaller, lossy compression

__Default__: none (preserve original format)

### image.quality

_Optional_. The desired quality of generated images with lossy compression. This is only
utilized if [`image.type`](#imagetype) is set to *jpeg*.

__Default__: *100*

### image.retina

_Optional_. Image scales to enable for displays with a greater than 1:1 pixel ratio.

For instance, to support both @2x and @3x scaling:

    @image.retina: 2, 3;

__Default__: *2, 3*

### image.arbitrary_sizes

_Optional_. When enabled, users can request any image size, even those not referenced 
in the wiki content. When disabled (default), users can only access image sizes that 
are explicitly referenced somewhere in the wiki content.

    @image.arbitrary_sizes;     /* enabled (not recommended for public wikis) */
    -@image.arbitrary_sizes;    /* disabled (recommended) */

__Default__: Disabled

### page.enable.cache

_Optional_. Enable caching of generated pages.

For best performance, page caching is enabled by default, and quiki will
only generate pages if the source file has been modified since the cache
file was last written.

__Default__: Enabled

### cat.per_page

_Optional_. Maximum number of pages to display on a single category posts page.

__Default__: _5_

### cat.[name].main

_Optional_. Set the main page for the category by the name of `[name]`.
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

_Optional_. Sets the human-readable title for the category by the name of `[name]`.

__Default__: none

### var.*

_Optional_. Global wiki variable space. Variables defined in this space will be
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

_Optional_. Name of the main page.

This should not be the page's title but rather a filename, relative to the
wiki page directory. The extension is not necessary.

```
@main_page: Welcome Page; /* normalized to welcome_page.page */
```

__Default__ (webserver): None

### main_redirect

_Optional_. If enabled, the wiki root redirects to the main page rather than just
rendering it at the root location.

```
@main_redirect;
```

__Default__ (webserver): Disabled

### error_page

_Optional_. Name of the error page.

This should not be the page's title but rather a filename, relative to the
wiki page directory. The extension is not necessary.

```
@error_page: Error; /* normalized to error.page */
```

__Default__ (webserver): None (show generic error text)

### navigation

_Optional_. Navigation items.

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

__Default__ (webserver): None

### template

_Optional_. Name or path of the template used by the wiki.

```
@template: default;
```

__Default__ (webserver): *default*

### logo

_Optional_. Filename for the wiki logo, relative to the wiki image directory.

Frontends like webserver automatically generate the logo in whatever dimensions
are needed and display it where appropriate.

```
@logo: logo.png;
```

## webserver options

These options are respected by the quiki webserver.

### server.name

_Optional_. The human-readable name of the server.

This is used for page titles and administrative interface headings. When not 
specified, "quiki" is used as the default server name.

    @server.name: My Server;

__Default__: *quiki*

### server.dir.resource

Path to quiki's `resources` directory.

    @repo: .;       /* where quiki is cloned */
    @server.dir.resource:   [@repo]/resources;

### server.dir.wiki

_Optional_. Path to some directory where wikis are stored.

Your wikis do not all have to be in the same directory, so this is optional.
However, if you make use of this, quiki can infer [`dir.wiki`](#dirwiki) for each
wiki, allowing you to omit the [`server.wiki.[name].dir`](#serverwikinamedir)
options.

### server.enable.pregeneration

_Optional_. If enabled, webserver pre-generates all pages and images upon start.

When disabled, the unified queue-only generation system is still used for all requests,
but startup background queuing is skipped.

__Requires__: [`page.enable.cache`](#pageenablecache)

__Default__: Enabled

### server.pregen.mode

_Optional_. Controls the performance characteristics of the pregeneration system.

Available modes:
- `default`: Balanced performance with intelligent worker scaling
- `fast`: High-performance mode with more workers and aggressive timeouts  
- `slow`: Resource-conservative mode with fewer workers and longer intervals

__Example__:
```
@server.pregen.mode: fast;
```

__Default__: `default`

## Advanced Pregeneration Options

The following options allow fine-tuning of individual pregeneration parameters, 
overriding the preset modes. All are optional.

### server.pregen.rate_limit

Time between background generation operations. Affects both pages and images (images are automatically 2x slower). Does not affect HTTP requests, only background pregeneration.

__Example__: `@server.pregen.rate_limit: 5ms;`

### server.pregen.page_priority

Buffer size for high-priority requests (HTTP requests). Larger = less blocking.

__Example__: `@server.pregen.page_priority: 1000;`

### server.pregen.page_background

Buffer size for background pregeneration. Larger = more content can be queued.

__Example__: `@server.pregen.page_background: 5000;`

### server.pregen.img_priority

Buffer size for high-priority image generation requests.

__Example__: `@server.pregen.img_priority: 200;`

### server.pregen.img_background

Buffer size for background image pregeneration.

__Example__: `@server.pregen.img_background: 500;`

### server.pregen.page_workers

Number of workers handling high-priority requests. More = better concurrency.

__Example__: `@server.pregen.page_workers: 8;`

### server.pregen.page_bg_workers

Number of workers handling background pregeneration. More = faster startup.

__Example__: `@server.pregen.page_bg_workers: 4;`

### server.pregen.img_workers

Number of workers handling high-priority image requests.

__Example__: `@server.pregen.img_workers: 6;`

__Default__: max(1, CPU_cores/4)

### server.pregen.img_bg_workers

Number of workers handling background image pregeneration.

__Example__: `@server.pregen.img_bg_workers: 2;`

__Default__: max(1, CPU_cores/8)

### server.pregen.timeout

Maximum time to wait for synchronous generation requests.

__Example__: `@server.pregen.timeout: 45s;`

### server.pregen.force

Force regeneration even if cache is fresh. Useful for debugging.

__Example__: `@server.pregen.force: true;`

### server.pregen.verbose

Enable detailed logging of pregeneration operations.

__Example__: `@server.pregen.verbose: true;`

### server.pregen.images

Enable pregeneration of image thumbnails.

__Example__: `@server.pregen.images: false;`

### server.pregen.cleanup

How often to clean up tracking maps. Set to 0 to disable.

__Example__: `@server.pregen.cleanup: 15m;`

### server.pregen.max_tracking

Maximum entries in tracking maps before forced cleanup.

__Example__: `@server.pregen.max_tracking: 50000;`

## Pregeneration Configuration Examples

### High-Performance Setup
```
@server.pregeneration.mode: fast;
@server.pregen.page_workers: 16;
@server.pregen.page_priority: 2000;
@server.pregeneration.rate_limit: 1ms;
@server.pregen.timeout: 15s;
```

### Resource-Conservative Setup  
```
@server.pregeneration.mode: slow;
@server.pregen.page_bg_workers: 1;
@server.pregeneration.rate_limit: 100ms;
@server.pregen.cleanup: 1h;
```

### Custom Balanced Setup
```
@server.pregeneration.mode: default;
@server.pregen.page_workers: 8;
@server.pregen.page_bg_workers: 3;
@server.pregen.img_workers: 4;
@server.pregen.verbose: true;
```

### server.enable.monitor

_Optional_. If enabled, webserver uses operating system facilities to monitor wiki
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
@server.http.port: 8080;
@server.http.bind: 127.0.0.1;
```

Or, you can listen on a UNIX socket:
```
@server.http.port:  unix;
@server.http.bind:  /var/run/quiki.sock;
```

__Default__: none (bind to all available hosts)

### server.http.host

_Optional_. The default host to serve wikis on if they don't specify `host.wiki`.

Note this helps inform the value of `root.ext` for wikis that do not specify an
external root in their configuration. It is recommended you configure it to the
public host of your server.

Should be a hostname only or IP only, e.g. `server.example.com`.

__Default__: None (serve wikis with no host configured on all hosts)

### server.domain

_Optional_. Cookie domain for session sharing across subdomains.

When specified, session cookies will be set with this domain, allowing login
sessions to be shared between the adminifier and wikis served on different
subdomains. For example, setting `@server.domain: .example.com;` allows
sessions to work across `admin.example.com` and `wiki.example.com`.

```
@server.domain: .example.com;
```

__Important__: It must start with a dot (.) for proper subdomain sharing.
__Default__: None (sessions limited to same host)

### server.dir.template

_Optional_. Template search paths.

```
@server.dir.template: /home/www/wiki-templates;
```

This is a comma-separated list of paths to look for templates when they are
specified by name in the wiki configuration. If you're running multiple wikis
that share a template, or if you are using the default template, this
is useful. Otherwise, you can just specify the absolute path to each wiki's
template in the [template](#template) directive.

### server.wiki.[name].enable

_Optional_. Enable the wiki with shortname `[name]`.

Any wikis configured that do not have this option present are skipped.
Any number of wikis can be configured on a single server using this.

### server.wiki.[name].dir

_Optional_. Path to the wiki with shortname `[name]`.

This can be omitted if [`server.dir.wiki`](#serverdirwiki) is set and
the wiki is located within that container. The shortname becomes the name of
its directory inside `server.dir.wiki`.

__Default__: [`server.dir.wiki`](#serverdirwiki)`/[name]`

### adminifier.enable

_Optional_. Enables the adminifier server administration panel.

__Default__: Disabled (but enabled in the example configuration)

### adminifier.host

_Optional_. Specifies an HTTP host to bind to for the adminifier web panel.

__Default__: None (i.e., adminifier is available on all hosts)

### adminifier.root

_Optional_. Specifies the HTTP root for the adminifier web panel.

If [`adminifier.host`](#adminifierhost) is specified and that hostname is
dedicated to adminifier, you can set `adminifier.root` empty to occupy the
entire host, for example:

    @adminifier.host: admin.mywiki.example.com;
    @adminifier.root: ;

__Default__: None (i.e., `/`)