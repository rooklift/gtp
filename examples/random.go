package main

import (
	"math/rand"
	"../gtp"
)

func main() {
	gtp.StartGTP(genmove, "Random", "1.0")
}

func genmove(colour int, board *gtp.Board) string {

	all_possible := board.AllLegalMoves(colour)

	if len(all_possible) == 0 {
		return "pass"
	}

	i := rand.Intn(len(all_possible))

	s := gtp.StringFromXY(all_possible[i].X, all_possible[i].Y, board.Size)

	return s
}
