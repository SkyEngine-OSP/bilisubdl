# bilisubdl

bilisubdl is a command line tool for downloading subtitles from bilibili.tv. It supports downloading subtitles from both anime and episode ids.

## Usage

```bash
bilisubdl [command] [flags] [arguments]
```

## Commands

* `dl`: Download subtitle from ID.
* `search`: Search anime.
* `timeline`: Show timeline (sun|mon|tue|wed|thu|fri|sat).
* `list`: List episode, section and language.

## Examples

```bash
# Download subtitle from anime id 37738 with language en

$ bilisubdl dl 1049041 -l en

# Download subtitle from episode id 2075361 with language en

$ bilisubdl dl 2075361 -l en --dlepisode

# Search anime with keyword "one piece".

$ bilisubdl search "one piece"

# List available subtitle language of anime id 37738

$ bilisubdl list 37738 -L

# Show today timeline

$ bilisubdl timeline

# Show timeline on Sunday

$ bilisubdl timeline mon
```

## Installing

The `bilisubdl` command on Windows using [Scoop](https://scoop.sh/)

```txt
scoop install https://raw.githubusercontent.com/K0ng2/scoop-bucket/main/bucket/bilisubdl.json
```
