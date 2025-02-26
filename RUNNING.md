# Running

The `quiki` executable can be used for several different functions.

Please note: it is intended to be run as a non-super user.

## Cheat Sheet

#### First Run Wizard

```
quiki -w        # Run wizard with defaults of *:8080 and data directory at ~/quiki 
quiki -w -bind=1.2.3.4 -port=9091 -host=sites.example.com   # custom server params
quiki -w -config=/etc/quiki.conf -wikis-dir=/var/www/wikis  # custom paths
```

#### Run Webserver

```
quiki                                   # using server config at ~/quiki/quiki.conf
quiki -config=/path/to/quiki.conf       # using config in another location
quiki -host=sites.example.com           # override default http host
quiki -bind=1.2.3.4 -port=8090          # override host and port in config
```

#### Generate Pages to STDOUT

HTML Output
```
quiki /path/to/my_page.page             # generate a standalone page, any .page or .md file
quiki -wiki=/path/to/wiki my_page       # generate a page within a wiki
quiki -i                                # interactive mode - read from STDIN
```

JSON Output
```
quiki -json /path/to/my_page.page       # generate a standalone page, output JSON
quiki -json -wiki=/path/to/wiki my_page # generate a page within a wiki, output JSON
```

#### Wiki Operations

```
quiki -wiki=/path/to/wiki               # pregenerate all pages in a wiki
quiki -create-wiki=/path/to/wiki        # create a new wiki at the path
```

#### Server Operations

Use with `-config=/path/to/quiki.conf`, otherwise assumes `~/quiki/quiki.conf`.

```
quiki -enable-wiki=shortcode            # start serving a wiki within the server wikis directory
quiki -wiki=/path/to/wiki -enable-wiki=shortcode  # ...outside the server wikis directory
quiki -disable-wiki=shortcode           # stop serving a wiki
quiki -import-wiki=/path/to/wiki        # import a wiki to the server wikis directory
quiki -reload                           # reload the server configuration

# these commands can be combined, e.g.
quiki -import-wiki=/path/to/wiki -enable-wiki=shortcode -reload
```

# Wizard

To set up quiki webserver for the first time, run the setup wizard with
```
quiki -w
```

### Alternate bind, port, and default host
The server by default binds to all available hosts on port 8080 and
serves sites that do not specify a host on all HTTP hosts. If you need
to do something different, you can specify them to the wizard like:
```
quiki -w -bind=1.2.3.4 -port=9091 -host=sites.example.com
```

### Alternate config and sites locations
By default, the wizard creates a directory at `~/quiki` to store your quiki
configuration and sites. It will write the configuration file to `~/quiki/quiki.conf` and
create a sites root directory at `~/quiki/wikis`.

If you want to store the config and sites elsewhere, you can specify like so:
```
quiki -w -config=/etc/quiki.conf -wikis-dir=/var/www/wikis
```

See the [configuration spec](doc/configuration.md) for all options.

### Admin account setup
When running with `-w`, after the configuration file is written, the server
begins listening. You should navigate your browser to the adminifier setup page,
e.g. `http://localhost:8080/admin/create-user`. Here, complete the form to set
up your first admin user.

The form requests an authentication token, which is printed to stdout to copy
and paste. You can also find it designated as `@adminifier.token` in the webserver
config, usually at `~/quiki/quiki.conf`.

It also allows you to specify a location to store your wiki sites. This will be
prepopulated with the recommended location which is the `wikis` dir relative
to the config location, typically `~/quiki/wikis`.

Upon successful completion of this form, you are logged in and ready to create
your first wiki!

# Webserver

To run the quiki webserver if you've already set it up, simply run
```
quiki
```
This will assume the webserver config path of `~/quiki/quiki.conf`.

### Alternate config location
To run with a different config, use
```
quiki -config=/path/to/quiki.conf
```

See the [configuration spec](doc/configuration.md) for all options.

### Override bind, port, and default host
The server by default binds to all available hosts on port 8080 and
serves sites that do not specify a host on all HTTP hosts. If you need
to do something different, you can specify them to the wizard like:

If you specify any of the these via flags, they override the
following options in the server configuration:
* **-bind** overrides `server.http.bind`
* **-port** overrides `server.http.port`
* **-host** overrides `server.http.host`

```
quiki -bind=1.2.3.4 -port=9091 -host=sites.example.com
```

# Standalone page

Standalone page mode allows you to render a single quiki page file with
no wiki context.

```
quiki path/to/my.page
```

# Standalone wiki

Standalone wiki mode allows you to perform wiki operations without
running the webserver.

### Render page within a wiki

To render a page within a wiki, use
```
quiki -wiki=/path/to/wiki some_page_name
```
This will, for example, render the page at
`/path/to/wiki/pages/some_page_name.page` in the context of the specified
wiki. The HTML is written to STDOUT, and then the program exits.

### Pregeneration

To pregenerate all the pages in a wiki, run without a page argument:
```
quiki -wiki=/path/to/wiki
```

# Interactive Mode

Interactive mode reads quiki source from STDIN until reaching EOF (i.e. Ctrl-D).
Then, it prints the HTML output to STDOUT and exits. 

To run quiki in interactive mode, use
```
quiki -i
```