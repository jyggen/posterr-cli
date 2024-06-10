# Posterr CLI

A CLI tool for managing Plex posters.

## Usage

### Compare

```
posterr compare <plex-base-url> <plex-token> [flags]

Arguments:
  <plex-base-url>
  <plex-token>

Flags:
  -h, --help                                 Show context-sensitive help.
      --cache-base-path="$XDG_CACHE_HOME"
      --threads=10
      --version                              Show version number.

      --http-timeout=10s
      --output-file="-"                      Defaults to stdout.

```

### Update

```
posterr update <plex-base-url> <plex-token> [flags]

Arguments:
  <plex-base-url>
  <plex-token>

Flags:
  -h, --help                                 Show context-sensitive help.
      --cache-base-path="$XDG_CACHE_HOME"
      --threads=10
      --version                              Show version number.

      --http-timeout=10s
      --force
```