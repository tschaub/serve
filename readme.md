# Serve

Serve files via HTTP.

```
Usage: serve <dir>

Arguments:
  <dir>    Serve files from this directory.

Flags:
  -h, --help              Show context-sensitive help.
      --port=4000         Listen on this port.
      --[no-]cors         Include CORS support (on by default).
      --dot               Serve dot files (files prefixed with a '.').
      --explicit-index    Only serve index.html files if URL path includes it.
```

The `serve <dir>` command can be used to browse files in a directory via HTTP.  For example, `serve .` starts a server for browsing the files in the current working directory.  See below for more detail on the [usage](#usage).

## Installation

The `serve` program can be installed by downloading one of the archives from [the latest release](https://github.com/tschaub/serve/releases).

Extract the archive and place the `serve` executable somewhere on your path.  See a list of available commands by running `serve` in your terminal.

Homebrew users can install the `serve` program with [`brew`](https://brew.sh/):

```shell
brew update
brew install tschaub/tap/serve
```

## Usage

### `--port`

By default, files are served on port 4000 (e.g. `http://localhost:4000`).  To have the server listen on a different port, pass a different value to the `--port` argument (e.g. `serve --port 9000 .`).

### `--no-cors`

By default, files are served with [CORS headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS).  To turn off this behavior, use the `--no-cors` argument (e.g. `serve --no-cors .`).

### `--dot`

By default, files and directories starting with a `.` will not be listed or served.  To allow browsing `.`-prefixed files, use the `--dot` argument (e.g. `serve --dot .`).

### `--explicit-index`

By default, if a directory does not include an `index.html` file, a listing of files in the directory will be served.  If a directory does include an `index.html` file, that file will be served instead of the directory listing.

You can use the `--explicit-index` argument to make it so an existing `index.html` file is only served if the request URL ends with `/index.html`.  When this argument is used, requests that end in `/` will be served with a directory listing, and requests that end in `/index.html` will be served with the contents of the `index.html` file (or a 404 if that file doesn't exist).

| URL path                  | `serve .` (with defaults)         | `serve --expcit-index .`          |
| ------------------------- | --------------------------------- | --------------------------------- |
| `/path/to/dir/`           | existing `index.html` is served   | directory listing is served       |
| `/path/to/dir/index.html` | redirect to `/path/to/dir/`       | existing `index.html` is served   |
