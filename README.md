# Posterr CLI

A CLI tool to update Plex posters to the best ones available on the internet!

## Usage

### Compare

This command will compare the posters suggested by MetaDB against the local and generate a HTML file with all posters
that would be updated by the `update` command. 

```
posterr compare --plex-base-url=<plex-base-url> --plex-token=<plex-token> [flags]
```

### Preview

This command will open a specific movie's poster (as suggested by MetaDB) in a new browser window.

```
posterr preview <imdb-id> [flags]
```

### Update

This command will update all local posters with the posters suggested by MetaDB after analyzing all available posters
from various sources.

```
posterr update --plex-base-url=<plex-base-url> --plex-token=<plex-token> [flags]
```

## MetaDB

This tool uses a custom API hosted by MetaDB to determine which poster to use for each movie. There are a few compelling
reasons why an API is used instead of adding the same functionality directly to the CLI tool:

- Performance
  - The API downloads and analyzes all posters available from its sources to determine which poster to recommend. Once
    this process is done, the result will be reused for all subsequent requests. Moving this functionality to the CLI
    tool would make it incredibly slow since a movie with a lot of available posters easily can take up towards 30
    seconds to process.
- Portability
  - The API uses OpenCV to compare and analyze all available posters. Moving this functionality to the CLI tool would
    require the user to install OpenCV manually themselves, making the CLI tool less portable. 
- Usability
  - The API uses various 3rd-party APIs as sources for its posters, most which require an API key. Moving this
    functionality to the CLI tool would require the user to register and supply API keys for all these sources instead.

The source code for the API is not publicly available at this time, but will be in the near future.

### Limitations

The API currently only support posters in the original language of the movie as well as in English. The CLI tool will,
for the time being, skip any library or movie with its metadata language not set to English because of this.