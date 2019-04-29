# gtp
Go Text Protocol helper for engines in Golang.

Requires [fohristiwhirl/sgf](https://github.com/fohristiwhirl/sgf) library.

* The app registers its genmove() function with gtp.StartGTP().
* The genmove() function receives an sgf.Board and an sgf.Colour.
* The genmove() function returns an *SGF-formatted* move, e.g. "dd"
* Or it can return "pass" or "resign"
