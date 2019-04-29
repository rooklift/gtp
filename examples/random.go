package main

import (
	"math/rand"
	"time"

	gtp ".."
	sgf "github.com/fohristiwhirl/sgf"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	gtp.StartGTP(genmove, "Random", "1.0")
}

func genmove(board *sgf.Board, colour sgf.Colour) string {
	all_possible := gtp.AllLegalMoves(board, colour)
	if len(all_possible) == 0 {
		return "resign"
	}
	i := rand.Intn(len(all_possible))
	return all_possible[i]
}
