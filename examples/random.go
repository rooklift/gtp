package main

import (
	"math/rand"
	"time"

	gtp ".."
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	gtp.StartGTP(genmove, "Random", "1.0")
}

func genmove(colour int, board *gtp.Board) string {
	all_possible := board.AllLegalMoves(colour)
	if len(all_possible) == 0 {
		return "resign"
	}
	i := rand.Intn(len(all_possible))
	s := board.StringFromPoint(all_possible[i])
	return s
}
