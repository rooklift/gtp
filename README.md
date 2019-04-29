# gtp
Go Text Protocol helper library for engines in Golang

* The app registers its genmove() function with the library.
* The genmove() function receives a Board and a Colour.
* The genmove() function returns an SGF-formatted move, e.g. "dd"
* Or it can return "pass" or "resign"
