package gtp

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	EMPTY = 0
	BLACK = 1
	WHITE = 2
)

type Point struct {
	X int
	Y int
}

type Board struct {
	State [][]int
	Ko Point
	Size int
	Komi float64
	NextPlayer int
}

func NewBoard(size int, komi float64) *Board {
	var board Board

	board.State = make([][]int, size)
	for i := range(board.State) {
		board.State[i] = make([]int, size)
	}

	board.Ko.X = -1
	board.Ko.Y = -1
	board.Size = size
	board.Komi = komi
	board.NextPlayer = BLACK

	return &board
}

func (b *Board) Clear() {
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			b.State[x][y] = EMPTY
		}
	}
	b.Ko.X = -1
	b.Ko.Y = -1
	b.NextPlayer = BLACK
}

func (b *Board) Copy() *Board {
	newboard := NewBoard(b.Size, b.Komi)
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			newboard.State[x][y] = b.State[x][y]
		}
	}
	newboard.Ko.X = b.Ko.X
	newboard.Ko.Y = b.Ko.Y
	newboard.NextPlayer = b.NextPlayer
	return newboard
}

func (b *Board) Dump() {
    for y := 0; y < b.Size; y++ {
        for x := 0; x < b.Size; x++ {
            c := '.'
            if b.Ko.X == x && b.Ko.Y == y {
            	c = '^'
            }
            if b.State[x][y] == BLACK {
                c = '*'
            } else if b.State[x][y] == WHITE {
                c = 'O'
            }
            fmt.Printf("%c ", c)
        }
        fmt.Printf("\n")
    }
}

func (b *Board) PlayMove(colour, x int, y int) error {

	if colour != BLACK && colour != WHITE {
		return fmt.Errorf("colour neither black nor white")
	}

	var opponent_colour int

	if colour == BLACK {
		opponent_colour = WHITE
	} else {
		opponent_colour = BLACK
	}

	if x < 0 || x >= b.Size || y < 0 || y >= b.Size {
		return fmt.Errorf("coordinate off board")
	}

	if b.State[x][y] != EMPTY {
		return fmt.Errorf("coordinate not empty")
	}

	// Disallow playing on the ko square...

	if colour == b.NextPlayer && b.Ko.X == x && b.Ko.Y == y {
		// b.State[x][y] = EMPTY								// Not needed as long as we make the move after ko check
		return fmt.Errorf("illegal ko recapture")
	}

	// NOTE: WE ACTUALLY MAKE THE MOVE HERE...

	b.State[x][y] = colour

	// Normal captures...

	last_point_captured := Point{-1, -1}						// If we captured exactly 1 stone, this will record it

	stones_destroyed := 0
	adj_points := b.AdjacentPoints(x, y)

	for _, point := range(adj_points) {
		if b.State[point.X][point.Y] == opponent_colour {
			if b.GroupHasLiberties(point.X, point.Y) == false {
				stones_destroyed += b.DestroyGroup(point.X, point.Y)
				last_point_captured = Point{point.X, point.Y}
			}
		}
	}

	// Disallow moves with no liberties (obviously after captures have been done)...

	if b.GroupHasLiberties(x, y) == false {
		b.State[x][y] = EMPTY
		return fmt.Errorf("move is suicidal")
	}

	// A square is a ko square if:
	//    - It was the site of the only stone captured this turn
	//    - The capturing stone has no friendly neighbours
	//    - The capturing stone has one liberty

	b.Ko.X = -1
	b.Ko.Y = -1

	if stones_destroyed == 1 {

		// Provisonally set the ko square to be the captured square...

		b.Ko.X = last_point_captured.X
		b.Ko.Y = last_point_captured.Y

		// But unset it if the capturing stone has any friendly neighbours or > 1 liberty

		liberties := 0
		friend_flag := false

		for _, point := range(adj_points) {
			if b.State[point.X][point.Y] == EMPTY {
				liberties += 1
			}
			if b.State[point.X][point.Y] == colour {
				friend_flag = true
				break
			}
		}

		if friend_flag || liberties > 1 {
			b.Ko.X = -1
			b.Ko.Y = -1
		}
	}

	// Set colour of next player...

	if colour == BLACK {
		b.NextPlayer = WHITE
	} else {
		b.NextPlayer = BLACK
	}

	return nil
}

func (b *Board) GroupHasLiberties(x int, y int) bool {

	if x < 0 || y < 0 || x >= b.Size || y >= b.Size {
		panic("GroupHasLiberties() called with illegal x,y")
	}

	checked_stones := make(map[Point]bool)
	return b.__group_has_liberties(x, y, checked_stones)
}

func (b *Board) __group_has_liberties(x int, y int, checked_stones map[Point]bool) bool {

	checked_stones[Point{x, y}] = true

	adj_points := b.AdjacentPoints(x, y)

	for _, adj := range(adj_points) {
		if b.State[adj.X][adj.Y] == EMPTY {
			return true
		}
	}

	for _, adj := range(adj_points) {
		if b.State[adj.X][adj.Y] == b.State[x][y] {
			if checked_stones[Point{adj.X, adj.Y}] == false {
				if b.__group_has_liberties(adj.X, adj.Y, checked_stones) {
					return true
				}
			}
		}
	}

	return false
}

func (b *Board) DestroyGroup(x int, y int) int {

	if x < 0 || y < 0 || x >= b.Size || y >= b.Size {
		panic("DestroyGroup() called with illegal x,y")
	}

	stones_destroyed := 1
	colour := b.State[x][y]
	b.State[x][y] = EMPTY

	for _, adj := range(b.AdjacentPoints(x, y)) {
		if b.State[adj.X][adj.Y] == colour {
			stones_destroyed += b.DestroyGroup(adj.X, adj.Y)
		}
	}

	return stones_destroyed
}

func (b *Board) Pass(colour int) error {

	if colour != BLACK && colour != WHITE {
		return fmt.Errorf("colour neither black nor white")
	}

	b.Ko.X = -1
	b.Ko.Y = -1
	if colour == BLACK {
		b.NextPlayer = WHITE
	} else {
		b.NextPlayer = BLACK
	}

	return nil
}

func (b *Board) NewFromMove(colour int, x int, y int) (*Board, error) {
	newboard := b.Copy()
	err := newboard.PlayMove(colour, x, y)
	if err != nil {
		return nil, err
	}
	return newboard, nil
}

func (b *Board) NewFromPass(colour int) (*Board, error) {
	newboard := b.Copy()
	err := newboard.Pass(colour)
	if err != nil {
		return nil, err
	}
	return newboard, nil
}

func (b *Board) AllLegalMoves(colour int) []Point {

	if colour != BLACK && colour != WHITE {
		return nil
	}

	var all_possible []Point

	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {

			if b.State[x][y] != EMPTY {
				continue
			}

			_, err := b.NewFromMove(colour, x, y)		// This is very crude

			if err != nil {
				continue
			}

			all_possible = append(all_possible, Point{x, y})
		}
	}

	return all_possible
}

func (b *Board) AdjacentPoints(x int, y int) []Point {

	var points []Point

	i := x - 1
	j := y

	if i >= 0 && i < b.Size && j >= 0 && j < b.Size {
		points = append(points, Point{i, j})
	}

	i = x + 1
	j = y

	if i >= 0 && i < b.Size && j >= 0 && j < b.Size {
		points = append(points, Point{i, j})
	}

	i = x
	j = y - 1

	if i >= 0 && i < b.Size && j >= 0 && j < b.Size {
		points = append(points, Point{i, j})
	}

	i = x
	j = y + 1

	if i >= 0 && i < b.Size && j >= 0 && j < b.Size {
		points = append(points, Point{i, j})
	}

	return points
}

func (b *Board) StringFromXY(x, y int) string {
	letter := 'A' + x
	if letter >= 'I' {
		letter += 1
	}
	number := b.Size - y
	return fmt.Sprintf("%c%d", letter, number)
}

func (b *Board) StringFromPoint(p Point) string {
	return b.StringFromXY(p.X, p.Y)
}

func (b *Board) XYFromString(s string) (int, int, error) {

	if len(s) < 2 {
		return -1, -1, fmt.Errorf("coordinate string too short")
	}

	letter := strings.ToLower(s)[0]

	if letter < 'a' || letter > 'z' {
		return -1, -1, fmt.Errorf("letter part of coordinate not in range a-z")
	}

	if letter == 'i' {
		return -1, -1, fmt.Errorf("letter i not permitted")
	}

	x := int((letter - 'a'))
	if letter > 'i' {
		x -= 1
	}

	tmp, err := strconv.Atoi(s[1:])
	if err != nil {
		return -1, -1, fmt.Errorf("couldn't parse number part of coordinate")
	}
	y := (b.Size - tmp)

	if x >= b.Size || y >= b.Size || x < 0 || y < 0 {
		return -1, -1, fmt.Errorf("coordinate off board")
	}

	return x, y, nil
}

func StartGTP(genmove func(colour int, board *Board) string, name string, version string) {

	board := NewBoard(19, 0.0)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.ToLower(line)        // Note this lowercase conversion
		tokens := strings.Fields(line)

		if len(tokens) == 0 {
			continue
		}

		var id int = -1

		if unicode.IsDigit(rune(tokens[0][0])) {
			var err error
			id, err = strconv.Atoi(tokens[0])
			if err != nil {
				fmt.Printf("? Couldn't parse ID\n\n")
				continue
			}
			tokens = tokens[1:]
		}

		if len(tokens) == 0 {
			continue            // This is GNU Go's behaviour when receiving just an ID
		}

		// So, by now, tokens is a list of the actual command; meanwhile id (if any) is saved
		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "name" {
			one_line_success(id, name)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "version" {
			one_line_success(id, version)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "protocol_version" {
			one_line_success(id, "2")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "komi" {
			if len(tokens) < 2 {
				one_line_failure(id, "no argument received for komi")
				continue
			}
			komi, err := strconv.ParseFloat(tokens[1], 64)
			if err != nil {
				one_line_failure(id, "couldn't parse komi float")
				continue
			}
			board.Komi = komi
			one_line_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "clear_board" {
			board.Clear()
			one_line_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "quit" {
			one_line_success(id, "")
			os.Exit(0)
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "showboard" {
			one_line_success(id, "")
			board.Dump()
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "boardsize" {
			if len(tokens) < 2 {
				one_line_failure(id, "no argument received for boardsize")
				continue
			}
			size, err := strconv.Atoi(tokens[1])
			if err != nil {
				one_line_failure(id, "couldn't parse boardsize int")
				continue
			}
			if size < 7 || size > 19 {
				one_line_failure(id, "boardsize not in range 7 - 19")
				continue
			}
			board = NewBoard(size, board.Komi)
			one_line_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "play" {

			if len(tokens) < 3 {
				one_line_failure(id, "insufficient arguments received for play")
				continue
			}
			if tokens[1] != "black" && tokens[1] != "b" && tokens[1] != "white" && tokens[1] != "w" {
				one_line_failure(id, "did not understand colour for play")
				continue
			}
			x, y, err := board.XYFromString(tokens[2])
			if err != nil {
				one_line_failure(id, err.Error())
				continue
			}

			var colour int
			if tokens[1][0] == 'w' {
				colour = WHITE
			} else {
				colour = BLACK
			}

			if tokens[2] == "pass" {
				board.Pass(colour)
				one_line_success(id, "")
				continue
			}

			err = board.PlayMove(colour, x, y)

			if err != nil {
				one_line_failure(id, err.Error())
				continue
			}

			one_line_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "genmove" {
			if len(tokens) < 2 {
				one_line_failure(id, "no argument received for genmove")
				continue
			}
			if tokens[1] != "black" && tokens[1] != "b" && tokens[1] != "white" && tokens[1] != "w" {
				one_line_failure(id, "did not understand colour for play")
				continue
			}

			var colour int
			if tokens[1][0] == 'w' {
				colour = WHITE
			} else {
				colour = BLACK
			}

			s := genmove(colour, board.Copy())		// Send the engine a copy, not the real thing

			if s == "pass" {
				board.Pass(colour)
			} else {
				x, y, err := board.XYFromString(s)
				if err != nil {
					one_line_failure(id, fmt.Sprintf("illegal move from engine: %s (%v)", s, err))
					continue
				}
				err = board.PlayMove(colour, x, y)
				if err != nil {
					one_line_failure(id, fmt.Sprintf("illegal move from engine: %s (%v)", s, err))
					continue
				}
			}

			one_line_success(id, s)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		one_line_failure(id, "unknown command")
	}

	return
}

func one_line_reply(id int, s string, shebang string) {
	fmt.Printf(shebang)
	if id != -1 {
		fmt.Printf("%d", id)
	}
	if s != "" {
		fmt.Printf(" %s\n\n", s)
	} else {
		fmt.Printf("\n\n")
	}
}

func one_line_success(id int, s string) {
	one_line_reply(id, s, "=")
}

func one_line_failure(id int, s string) {
	one_line_reply(id, s, "?")
}
