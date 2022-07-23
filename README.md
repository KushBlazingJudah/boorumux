# boorumux

Boorumux is a **work-in-progress** booru aggregator server written in Go that:

- aims to be dead simple and nicely written
- is extremely easy to extend
- supports proxies
- all while having no external dependencies

## Roadmap

Boorumux is currently a work-in-progress and isn't completely finished yet.
Here's a list of things planned, strikethrough is complete:

- ~~Proxy~~
- Page view
- Post view
- Aggregating several boorus at once
- Caching

## Supported APIs

- Danbooru (100% finished)

## Running

You can simply `go run ./cmd/boorumux` or `go build ./cmd/boorumux` to run or
build Boorumux respectively.

Boorumux depends on a few folders to work correctly:

- the `static` folder
- the `views` folder
