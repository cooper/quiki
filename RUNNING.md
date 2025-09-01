# Running

## Choose a binary

There are a few options to install and interact with quiki:

- **`quiki`** (38MB) - the main executable includes the fully featured webserver
  and admin panel as a standalone app, and can also be used for all CLI capabilities.
- **`quiki-wiki`** (28MB) - provides CLI interface for wiki operations without server.
- **`wikifier`** (15MB) - tiny binary to interact with the underlying quiki
  language parser, providing page rendering only.

The quiki executables have zero dependencies. All resources are embedded.

### Install

```sh
go install github.com/cooper/quiki                  # full suite
go install github.com/cooper/quiki/cli/quiki-wiki   # wiki cli
go install github.com/cooper/quiki/cli/wikifier     # standalone renderer
```

Please note: quiki is intended to be run as a non-super user.

All examples in documentation reference the `quiki` binary, but the syntax for the
other binaries is the same.

## About the quiki dir

quiki uses a unified directory to store all server-level data:
- configuration file (`quiki.conf`)
- sites that it serves (`wikis/` subdirectory)
- server state like pid file, authenticator db

By default, this directory is `~/quiki`. In the docker image, this becomes `/quiki`,
which should be mapped to a volume or directory on the host for persistence.

For atypical configurations, can specify a different directory with the `-dir` flag.

## Cheat Sheet

#### First Run Wizard

```
quiki -w                                # run wizard with defaults (*:8080, ~/quiki)
quiki -w -bind=1.2.3.4 -port=9091 -host=sites.example.com   # custom server params
quiki -w -dir=/data/quiki                # custom quiki directory
```

#### Run Webserver

```
quiki                                   # using default directory ~/quiki
quiki -dir=/path/to/quiki               # using custom data directory
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

Use with `-dir=/path/to/quiki-data` if your quiki dir is not `~/quiki`.

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

### Alternate quiki directory
By default, the wizard creates a directory at `~/quiki` to store your quiki
configuration, user database, sites, and other server data. It will write the 
configuration file to `~/quiki/quiki.conf` and create a sites root directory 
at `~/quiki/wikis`.

If you want to store data in a different location, you can specify:
```
quiki -w -dir=/data/quiki
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
to the quiki directory, typically `~/quiki/wikis`.

Upon successful completion of this form, you are logged in and ready to create
your first wiki!

# Webserver

To run the quiki webserver if you've already set it up, simply run
```
quiki
```
This will use the default quiki directory at `~/quiki`.

### Alternate quiki directory
To run with a different quiki directory, use
```
quiki -dir=/path/to/quiki
```

### Backward compatibility
The legacy `-config` flag is still supported for now but deprecated:
```
quiki -config=/path/to/quiki.conf       # deprecated, use -dir instead
```

Please update your service/autostart files accordingly, or moved to a containerized
approach using the example [Docker compose file](./docker-compose.yml).

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