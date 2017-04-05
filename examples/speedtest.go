package main

import (
	"fmt"
	"math/rand"
	"time"

	gtp ".."
)

const RUNS = 1000
const DEPTH = 50

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {

	starttime := time.Now()

	startboard := gtp.NewBoard(19, 0)

	for r := 0; r < RUNS; r++ {
		board := startboard.Copy()
		for d := 0; d < DEPTH; d++ {
			x, y := genmove(gtp.BLACK, board)
			board.PlayMove(gtp.BLACK, x, y)
		}
	}

	fmt.Printf("%v\n", time.Now().Sub(starttime))
}

func genmove(colour int, board *gtp.Board) (int, int) {
	all_possible := board.AllLegalMoves(colour)
	i := rand.Intn(len(all_possible))
	return all_possible[i].X, all_possible[i].Y
}
