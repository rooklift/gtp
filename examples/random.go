package main

import (
	"math/rand"
	"time"

	gtp ".."
	sgf "github.com/rooklift/sgf"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	gtp.StartGTP(genmove, "Random", "1.0")
}

func genmove(board *sgf.Board, colour sgf.Colour) string {
	all_possible := gtp.AllLegalMoves(board, colour)	// Returns moves in SGF format, e.g. ["dd", "dg" ...]
	if len(all_possible) == 0 {
		return "resign"
	}
	i := rand.Intn(len(all_possible))
	return all_possible[i]
}
