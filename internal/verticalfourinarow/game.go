package verticalfourinarow

import "fmt"

type column []string

// VerticalFIR - Represents the board game vertical four in a row.
// It has 7 columns and 6 rows and the objective is to get 4 markers in a row (vertical, horizontal or diagonal).
// Player can only select a column to drop their marker in, no choice of row and no possibility to pass
type VerticalFIR struct {
	board        []column
	playerA      string
	playerB      string
	playerInTurn string
	columns      int
	rows         int
	rounds       int
	done         bool
}

// NewVerticalFIR - Returns a new instance of the game
func NewVerticalFIR(playerA string, playerB string) *VerticalFIR {
	t := VerticalFIR{columns: 7, rows: 6, playerA: playerA, playerB: playerB}
	t.Reset()

	return &t
}

// Reset - Resets the game to be prepared for a new game
func (V *VerticalFIR) Reset() {
	board := make([]column, V.columns)
	for r := 0; r < V.columns; r++ {
		c := make(column, V.rows)
		board[r] = c
	}

	for x := 0; x < V.columns; x++ {
		for y := 0; y < V.rows; y++ {
			board[x][y] = " "
		}
	}

	V.board = board
	V.rounds = 0
	V.playerInTurn = V.playerA
	V.done = false
}

// Move - Makes a move in the game and evaluates the board for draw, win or continue play.
// It returns whether game is done, winner (empty string if a draw) and error.
func (V *VerticalFIR) Move(x uint8, y uint8, pass bool) (bool, string, error) {
	c := x
	var r int
	for r = 0; r < V.rows; r++ {
		if V.board[c][r] == " " {
			break
		}
	}
	if r == V.rows {
		return false, "", fmt.Errorf("illegal move, spot already occupied")
	}

	V.board[c][r] = V.playerInTurn
	V.rounds++
	draw := V.evaluateGame()

	// Check if the game is over, a draw results in 0 points and a win gives 1 point
	if draw {
		return true, "", nil
	} else if V.done {
		return true, V.playerInTurn, nil
	}

	// Switch playerInTurn who is next in turn
	if V.playerInTurn == V.playerA {
		V.playerInTurn = V.playerB
	} else {
		V.playerInTurn = V.playerA
	}

	return false, "", nil
}

// AvailableActions - Returns available actions, i.e. free spots on the board. The y and pass flag isn't relevant in
// the game of Vertical four in a row and will always be 0 respective false.
func (V *VerticalFIR) AvailableActions() ([][2]uint8, bool) {
	actions := make([][2]uint8, 0, V.columns)

	for x := 0; x < V.columns; x++ {
		if V.board[x][V.rows-1] == " " {
			actions = append(actions, [2]uint8{uint8(x), 0})
		}
	}

	return actions, false
}

// evaluateGame - Evaluates whether the game is finished in which case done is set to true.
// It returns true if the game is a draw, otherwise the winner is who is denoted in playerInTurn.
func (V *VerticalFIR) evaluateGame() bool {
	// Check columns
	for c := 0; c < V.columns; c++ {
		if V.board[c][0] == " " {
			continue
		}
		inRow := 1
		prev := V.board[c][0]

		for r := 1; r < V.rows; r++ {
			if V.board[c][r] == " " {
				break
			} else if V.board[c][r] == prev {
				inRow++
			} else {
				prev = V.board[c][r]
				inRow = 1
			}
			if inRow == 4 {
				V.done = true
				return false
			}
		}
	}

	// Check rows
	for r := 0; r < V.rows; r++ {
		var inRow int
		prev := V.board[0][r]
		if prev != " " {
			inRow = 1
		}

		for c := 1; c < V.columns; c++ {
			if V.board[c][r] == " " {
				prev = " "
				inRow = 0
			} else if V.board[c][r] == prev {
				inRow++
			} else {
				prev = V.board[c][r]
				inRow = 1
			}
			if inRow == 4 {
				V.done = true
				return false
			}
		}
	}

	// Check up diagonals
	start := [][2]int{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {2, 0}, {3, 0}}
	coordInc := func(a, b int) (int, int) { return a + 1, b + 1 }

	for _, s := range start {
		c, r := s[0], s[1]
		var inRow int
		prev := V.board[c][r]
		if prev != " " {
			inRow = 1
		}
		c, r = coordInc(c, r)

		for ; c < V.columns && r < V.rows; c, r = coordInc(c, r) {
			if V.board[c][r] == " " {
				prev = " "
				inRow = 0
			} else if V.board[c][r] == prev {
				inRow++
			} else {
				prev = V.board[c][r]
				inRow = 1
			}
			if inRow == 4 {
				V.done = true
				return false
			}
		}
	}

	// Check down diagonals
	start = [][2]int{{0, 3}, {0, 4}, {0, 5}, {1, 5}, {2, 5}, {3, 5}}
	coordInc = func(a, b int) (int, int) { return a + 1, b - 1 }

	for _, s := range start {
		c, r := s[0], s[1]
		var inRow int
		prev := V.board[c][r]
		if prev != " " {
			inRow = 1
		}
		c, r = coordInc(c, r)

		for ; c < V.columns && r >= 0; c, r = coordInc(c, r) {
			if V.board[c][r] == " " {
				prev = " "
				inRow = 0
			} else if V.board[c][r] == prev {
				inRow++
			} else {
				prev = V.board[c][r]
				inRow = 1
			}
			if inRow == 4 {
				V.done = true
				return false
			}
		}
	}

	// Check if game is a draw (i.e. no winner and board is full)
	return V.rounds == V.columns*V.rows
}

// GetPlayers - Returns the two players of the game in start order
func (V *VerticalFIR) GetPlayers() [2]string {
	return [2]string{V.playerA, V.playerB}
}

// SetPlayers - Sets the players of the game
func (V *VerticalFIR) SetPlayers(players [2]string) {
	V.playerA = players[0]
	V.playerB = players[1]
}

// GetState - Gets the state of the game as a base3 number formatted as a string and the player in turn
func (V *VerticalFIR) GetState() (string, string) {
	buf := make([]byte, V.columns*V.rows)
	i := 0
	for c := 0; c < V.columns; c++ {
		for r := 0; r < V.rows; r++ {
			switch V.board[c][r] {
			case V.playerA:
				buf[i] = '1'
			case V.playerB:
				buf[i] = '2'
			default:
				buf[i] = '0'
			}
			i++
		}
	}

	return string(buf), V.playerInTurn
}

// SetState - Sets the game according given state and player in turn
func (V *VerticalFIR) SetState(state, playerInTurn string) (bool, string) {
	// Fix length of state by left padding with zeros
	diff := V.columns*V.rows - len(state)
	if diff > 0 {
		state = fmt.Sprintf("%0*d%s", diff, 0, state)
	}

	i := 0
	V.rounds = 0
	V.playerInTurn = playerInTurn
	V.done = false
	for c := 0; c < V.columns; c++ {
		for r := 0; r < V.rows; r++ {
			switch state[i] {
			case '1':
				V.board[c][r] = V.playerA
				V.rounds++
			case '2':
				V.board[c][r] = V.playerB
				V.rounds++
			default:
				V.board[c][r] = " "
			}
			i++
		}
	}
	// Check if the game is over, a draw results in 0 points and a win gives 1 point
	if draw := V.evaluateGame(); draw {
		return true, ""
	} else if V.done {
		return true, V.playerInTurn
	}

	return false, ""
}

// PrintBoard - Prints out the game board
func (V *VerticalFIR) PrintBoard() {
	columns := "   A B C D E F G H"
	fmt.Println("")
	for r := int(V.rows) - 1; r >= 0; r-- {
		fmt.Printf("%d ", r+1)
		for c := 0; c < V.columns; c++ {
			fmt.Printf("|%s", V.board[c][r])
		}
		fmt.Print("|\n")
	}
	fmt.Printf("%s\n", columns[0:4+2*(V.columns-1)])
	fmt.Println("")
}
