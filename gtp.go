package gtp

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fohristiwhirl/sgf"
)

var known_commands = []string{
	"boardsize", "clear_board", "genmove", "known_command", "komi", "list_commands",
	"name", "play", "protocol_version", "quit", "savesgf", "showboard", "undo", "version",
}

func StartGTP(genmove func(board *sgf.Board, colour sgf.Colour) string, name string, version string) {

	root := sgf.NewNode(nil)
	root.SetValue("SZ", "19")
	root.SetValue("KM", "0")
	node := root

	scanner := bufio.NewScanner(os.Stdin)
	var err error

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

		if tokens[0][0] >= '0' && tokens[0][0] <= '9' {
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
			print_success(id, name)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "version" {
			print_success(id, version)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "protocol_version" {
			print_success(id, "2")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "list_commands" {
			response := ""
			for _, command := range known_commands {
				response += command + "\n"
			}
			print_success(id, response)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "known_command" {
			if len(tokens) < 2 {
				print_failure(id, "no argument received for known_command")
				continue
			}
			response := "false"
			for _, command := range known_commands {
				if command == tokens[1] {
					response = "true"
					break
				}
			}
			print_success(id, response)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "komi" {
			if len(tokens) < 2 {
				print_failure(id, "no argument received for komi")
				continue
			}
			komi, err := strconv.ParseFloat(tokens[1], 64)
			if err != nil {
				print_failure(id, "couldn't parse komi float")
				continue
			}

			s := fmt.Sprintf("%f", komi)
			s = strings.TrimRight(s, "0")		// Clear all trailing zeroes e.g. "7.5000" --> "7.5"
			if strings.HasSuffix(s, ".") {		// But fix if it now ends with "."
				s += "0"
			}

			root.SetValue("KM", s)

			print_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "clear_board" {

			root.DeleteKey("HA")
			root.DeleteKey("AB")
			root.DeleteKey("AW")

			for _, child := range root.Children() {
				child.Detach()
			}

			node = root

			print_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "quit" {
			print_success(id, "")
			os.Exit(0)
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "showboard" {
			print_success(id, "Board:\n" + node.Board().String())
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "savesgf" {
			if len(tokens) < 2 {
				print_failure(id, "no argument received for savesgf")
				continue
			}

			err := node.Save(tokens[1])
			if err != nil {
				print_failure(id, err.Error())
				continue
			}

			print_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "boardsize" {
			if len(tokens) < 2 {
				print_failure(id, "no argument received for boardsize")
				continue
			}
			size, err := strconv.Atoi(tokens[1])
			if err != nil {
				print_failure(id, "couldn't parse boardsize int")
				continue
			}
			if size < 3 || size > 25 {
				print_failure(id, "boardsize not in range 3 - 25")
				continue
			}

			root.SetValue("SZ", strconv.Itoa(size))
			root.DeleteKey("HA")
			root.DeleteKey("AB")
			root.DeleteKey("AW")

			for _, child := range root.Children() {
				child.Detach()
			}

			node = root

			print_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "play" {
			if len(tokens) < 3 {
				print_failure(id, "insufficient arguments received for play")
				continue
			}
			if tokens[1] != "black" && tokens[1] != "b" && tokens[1] != "white" && tokens[1] != "w" {
				print_failure(id, "did not understand colour for play")
				continue
			}

			var colour sgf.Colour
			if tokens[1][0] == 'w' {
				colour = sgf.WHITE
			} else {
				colour = sgf.BLACK
			}

			if tokens[2] == "pass" {

				node = node.PassColour(colour)

			} else {

				p := sgf.ParseGTP(tokens[2], root.RootBoardSize())

				if p == "" {
					print_failure(id, "illegal move")
					continue
				}

				node, err = node.PlayColour(p, colour)
				if err != nil {
					print_failure(id, "illegal move")
					continue
				}
			}

			node.MakeMainLine()

			print_success(id, "")
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "genmove" {
			if len(tokens) < 2 {
				print_failure(id, "no argument received for genmove")
				continue
			}
			if tokens[1] != "black" && tokens[1] != "b" && tokens[1] != "white" && tokens[1] != "w" {
				print_failure(id, "did not understand colour for genmove")
				continue
			}

			var colour sgf.Colour
			if tokens[1][0] == 'w' {
				colour = sgf.WHITE
			} else {
				colour = sgf.BLACK
			}

			s := genmove(node.Board(), colour)

			if s == "pass" {

				node = node.PassColour(colour)

			} else if s == "resign" {

				// no need to do anything here

			} else {

				node, err = node.PlayColour(s, colour)
				if err != nil {
					print_failure(id, fmt.Sprintf("wanted to play illegal move: %q (%v)", s, err))
					continue
				}

				// Finally, convert the returned string to GTP format...

				s = GTP(s, root.RootBoardSize())
			}

			node.MakeMainLine()

			print_success(id, s)
			continue
		}

		// --------------------------------------------------------------------------------------------------

		if tokens[0] == "undo" {
			if node.Parent() == nil {
				print_failure(id, "cannot undo")
				continue
			} else {
				node = node.Parent()			// No need to delete children
				print_success(id, "")
				continue
			}
		}

		// --------------------------------------------------------------------------------------------------

		print_failure(id, "unknown command")
	}
}

func AllLegalMoves(board *sgf.Board, colour sgf.Colour) []string {		// Returns strings in format "dd" (SGF format)

	if colour != sgf.BLACK && colour != sgf.WHITE {
		return nil
	}

	if board == nil {
		return nil
	}

	var ret []string

	for x := 0; x < board.Size; x++ {

		Y_LOOP: for y := 0; y < board.Size; y++ {

			p := sgf.Point(x, y)

			if board.Ko == p && board.Player == colour {
				continue Y_LOOP
			}

			if board.Get(p) != sgf.EMPTY {
				continue Y_LOOP
			}

			for _, a := range sgf.AdjacentPoints(p, board.Size) {
				if board.Get(a) == sgf.EMPTY {
					ret = append(ret, p)				// Move is clearly legal since some of its neighbours are empty
					continue Y_LOOP
				}
			}

			// The move we are playing will have no liberties of its own.
			// Therefore, it will be legal iff it has a neighbour which:
			//
			//		- Is an enemy group with 1 liberty.
			//		- Is a friendly group with 2 or more liberties.

			for _, a := range sgf.AdjacentPoints(p, board.Size) {
				if board.Get(a) == colour.Opposite() {
					if len(board.Liberties(a)) == 1 {
						ret = append(ret, p)
						continue Y_LOOP
					}
				} else if board.Get(a) == colour {
					if len(board.Liberties(a)) >= 2 {
						ret = append(ret, p)
						continue Y_LOOP
					}
				} else {
					panic("wat")
				}
			}
		}
	}

	return ret
}

func print_reply(id int, s string, shebang string) {
	s = strings.TrimSpace(s)
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

func print_success(id int, s string) {
	print_reply(id, s, "=")
}

func print_failure(id int, s string) {
	print_reply(id, s, "?")
}

func GTP(p string, size int) string {

	x, y, onboard := sgf.ParsePoint(p, size)

	if onboard == false {
		return "offboard"
	}

	letter := 'A' + x
	if letter >= 'I' {
		letter += 1
	}
	number := size - y
	return fmt.Sprintf("%c%d", letter, number)
}
