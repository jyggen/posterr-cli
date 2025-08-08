# Posterr CLI

A CLI tool to update your Plex posters to the best<sup>[*](#disclaimer)</sup> posters available on the internet!

## Usage

To learn more about Posterr CLI, its commands as well as available flags, the `--help` flag can be used.

### Compare

This command will compare your current Plex posters against the best posters available and generate an HTML file with
all posters that do not match. 

```
posterr compare --plex-base-url=<plex-base-url> --plex-token=<plex-token> [flags]
```

### Preview

This command will open the best poster available for the movie specified in a new browser window.

```
posterr preview <imdb-id> [flags]
```

### Update

This command will update any Plex poster that does not match the best poster available.

```
posterr update --plex-base-url=<plex-base-url> --plex-token=<plex-token> [flags]
```

## MetaDB

Posterr CLI uses an API hosted by MetaDB that suggests the best poster for each movie. There are a few compelling
reasons why an API is used instead of adding the API's functionality directly to the tool:

- Performance
  - In order to determine the best poster available, we need download, analyze and algorithmically rank all posters from
    all known sources. Embedding this functionality directly within the tool itself would slow it down significantly 
    since a single movie with a lot of available posters can easily take up towards 30 seconds to process. With the API
    doing the heavy lifting instead, there's a high probability that the movie's been processed and cached already, and
    that the best poster available can be suggested within milliseconds.
- Responsiveness
  - The algorithm which ranks posters is frequently tweaked and new sources of posters can appear (and disappear) at any
    given time. Having the functionality in an API makes Posterr CLI more responsive to changes, which allows users of
    the tool to benefit from gradual improvements without having to keep Posterr CLI itself up-to-date. 
- Usability
  - Posters are sourced from various APIs, services and other kind of sources, most of which require an API key or some
    other form of authentication. Instead of each user of Posterr CLI having to supply their own credentials for each
    source, they can piggyback off the API's credentials.

The source code of the API is not publicly available at this time, but the plan is to release it under an open-source
license in the future to enable self-hosting and community contributions.

### Privacy

In order to keep the usage of the API to a minimum, the communication between Poster CLI and the API is optimized and
heavily cached in Cloudflare's CDN as well as locally by the tool itself. However, the following information has a
chance of reaching MetaDB and the API:
- The IMDb ID of the movie we want to know the best available poster for.
- Your IP address (due to the nature of the Internet Protocol).
- The current version of Posterr CLI (included as part of the `User-Agent` header).

Although neither MetaDB nor the API deliberately stores any of this information, it will end up in access logs that are
kept for a short duration of time to ensure stability of the service as well as prevent abuse.

In addition, your IP address and the current version of Posterr CLI will, of course, also be visible to the Plex Media
Server instance specified, as well as any poster source the tool is prompted to download a poster from. In both cases,
Posterr CLI will cache these requests locally as well, when possible.

Your Plex token is only ever sent to the Plex Media Server instance you've specified and will **never** reach MetaDB,
the API or any of the poster sources.

### Limitations

The API will currently only suggest posters in either the original language of the movie or in English. Posterr CLI
will, because of this limitation, skip any Plex library or movie where the metadata language is not set to English.

## Disclaimer

"Best" in the movie poster universe is, as one would expect, highly subjective. However, user feedback suggests that the
posters set by Posterr CLI are _subjectively_ better than Plex's default posters in almost every case (or at least on
par with). If Posterr CLI sets an _objectively_ worse poster for one of your movies, feel free to open an issue - it
could be a bug! 