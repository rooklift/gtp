# gtp
Go Text Protocol helper for engines in Golang.

Requires [rooklift/sgf](https://github.com/rooklift/sgf) library.

* The app registers its `genmove()` function with `gtp.StartGTP()`
* The `genmove()` function receives an `sgf.Board` and an `sgf.Colour`
* The `genmove()` function returns an *SGF-formatted* move, e.g. `dd`
* Or it can return `pass` or `resign`
* Other GTP commands are handled automatically by the library.

For an example of a fully functioning random-move generator constructed with this library, see [random.go](https://github.com/rooklift/gtp/blob/master/examples/random.go) inside the `examples` folder.

# Technical notes

Behind the scenes, the library is generating an SGF tree of the game, and using the `sgf` library for legality checking and undo functionality. The custom GTP command `savesgf` can be used to dump that as a file.
