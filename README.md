# Boorumux

Boorumux is a **work-in-progress** booru aggregator server written in Go that:

- aims to be dead simple and nicely written
- is easy to extend
- supports proxies right out of the box
- has better blacklisting than most, I think
- all while having no external dependencies

It was written to fulfill a desire that may have already been fulfilled
elsewhere, but I haven't tried any other aggregators than this one so I think
it's pretty good.

## Roadmap

Boorumux is currently a **work-in-progress** and isn't completely finished yet.
While most of the features planned are already implemented, they may not be
fully implemented nor work properly; generally, it is still perfectly usable.

- Configuration needs work on it to make it more user-friendly; currently it is
  just a JSON file that you write yourself with little documentation on how it
  is actually supposed to look.
- Parts of the UI need work
- Mobile support
- Make searching several boorus at once easier: the `-mux` part was left off to
  the side until everything worked okay, so it feels like it's a hack.
  I wasn't sure how best to implement something like this, I'm still not, but I
  have an idea.

## Supported APIs

- Danbooru (100% finished)
- Gelbooru (100% finished)

If something you want isn't here, contributions to add or improve existing booru
APIs are encouraged.

You can always
[complain](https://github.com/KushBlazingJudah/boorumux/issues/new) or
[check if somebody already did](https://github.com/KushBlazingJudah/boorumux/labels/booru%20request).

## Running

You can simply `go run ./cmd/boorumux` or `go build ./cmd/boorumux` to run or
build Boorumux respectively.

Boorumux depends on a few folders to work correctly:

- the `static` folder
- the `views` folder

In the future, these will optionally be bundled into the application at
compile-time to ease distribution.

Boorumux, by default, listens on `localhost:8080`.
Point your web browser there to use it.
