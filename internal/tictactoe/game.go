package tictactoe

import "fmt"

type column []string

// TicTacToe - Represents the board game with the same name
type TicTacToe struct {
	board        []column
	playerA      string
	playerB      string
	playerInTurn string
	size         uint8
	rounds       int
	done         bool
}

// NewTicTacToe - Returns a new instance of the game
func NewTicTacToe(size uint8, playerA string, playerB string) *TicTacToe {
	t := TicTacToe{size: size, playerA: playerA, playerB: playerB}
	t.Reset()

	return &t
}

// Reset - Resets the game to be prepared for a new game
func (T *TicTacToe) Reset() {
	board := make([]column, T.size)
	for r := uint8(0); r < T.size; r++ {
		c := make(column, T.size)
		board[r] = c
	}

	for x := uint8(0); x < T.size; x++ {
		for y := uint8(0); y < T.size; y++ {
			board[x][y] = " "
		}
	}

	T.board = board
	T.rounds = 0
	T.playerInTurn = T.playerA
	T.done = false
}

// Move - Makes a move in the game and evaluates the board for draw, win or continue play.
// It returns whether game is done, winner (empty string if a draw) and error.
func (T *TicTacToe) Move(x uint8, y uint8, pass bool) (bool, string, error) {
	if T.board[x][y] != " " {
		return false, "", fmt.Errorf("illegal move, spot already occupied")
	}

	T.board[x][y] = T.playerInTurn
	T.rounds++
	draw := T.evaluateGame()

	// Check if the game is over, a draw results in 0 points and a win gives 1 point
	if draw {
		return true, "", nil
	} else if T.done {
		return true, T.playerInTurn, nil
	}

	// Switch playerInTurn who is next in turn
	if T.playerInTurn == T.playerA {
		T.playerInTurn = T.playerB
	} else {
		T.playerInTurn = T.playerA
	}

	return false, "", nil
}

// AvailableActions - Returns available actions, i.e. free spots on the board. The pass flag isn't relevant in
// the game of TicTacToe and will always be false.
func (T *TicTacToe) AvailableActions() ([][2]uint8, bool) {
	actions := make([][2]uint8, 0, T.size*T.size)

	for x := uint8(0); x < T.size; x++ {
		for y := uint8(0); y < T.size; y++ {
			if T.board[x][y] == " " {
				actions = append(actions, [2]uint8{x, y})
			}
		}
	}

	return actions, false
}

// evaluateGame - Evaluates whether the game is finished in which case done is set to true.
// It returns true if the game is a draw, otherwise the winner is who is denoted in playerInTurn.
func (T *TicTacToe) evaluateGame() bool {
	// Check columns and rows
	for a := uint8(0); a < T.size; a++ {
		winCol := true
		winRow := true
		if T.board[a][0] == " " {
			winCol = false
		}
		if T.board[0][a] == " " {
			winRow = false
		}

		for b := uint8(1); b < T.size && (winCol || winRow); b++ {
			if T.board[a][0] != T.board[a][b] {
				winCol = false
			}
			if T.board[0][a] != T.board[b][a] {
				winRow = false
			}
		}
		if winCol || winRow {
			T.done = true
			return false
		}

	}

	// Check diagonals
	winD1 := true
	winD2 := true
	if T.board[0][0] == " " {
		winD1 = false
	}
	if T.board[0][T.size-1] == " " {
		winD2 = false
	}

	for a := uint8(1); a < T.size && (winD1 || winD2); a++ {
		if T.board[0][0] != T.board[a][a] {
			winD1 = false
		}
		if T.board[0][T.size-1] != T.board[a][T.size-1-a] {
			winD2 = false
		}
	}
	if winD1 || winD2 {
		T.done = true
		return false
	}

	// Check if game is a draw (i.e. no winner and board is full)
	return T.rounds == int(T.size)*int(T.size)
}

// GetPlayers - Returns the two players of the game in start order
func (T *TicTacToe) GetPlayers() [2]string {
	return [2]string{T.playerA, T.playerB}
}

// SetPlayers - Sets the players of the game
func (T *TicTacToe) SetPlayers(players [2]string) {
	T.playerA = players[0]
	T.playerB = players[1]
}

// GetState - Gets the state of the game as a base3 number formatted as a string and the player in turn
func (T *TicTacToe) GetState() (string, string) {
	buf := make([]byte, T.size*T.size)
	i := 0
	for r := uint8(0); r < T.size; r++ {
		for c := uint8(0); c < T.size; c++ {
			switch T.board[r][c] {
			case T.playerA:
				buf[i] = '1'
			case T.playerB:
				buf[i] = '2'
			default:
				buf[i] = '0'
			}
			i++
		}
	}

	return string(buf), T.playerInTurn
}

// SetState - Sets the game according given state and player in turn
func (T *TicTacToe) SetState(state, playerInTurn string) (bool, string) {
	// Fix length of state by left padding with zeros
	diff := int(T.size*T.size) - len(state)
	if diff > 0 {
		state = fmt.Sprintf("%0*d%s", diff, 0, state)
	}

	i := 0
	T.rounds = 0
	T.playerInTurn = playerInTurn
	T.done = false
	for r := uint8(0); r < T.size; r++ {
		for c := uint8(0); c < T.size; c++ {
			switch state[i] {
			case '1':
				T.board[r][c] = T.playerA
				T.rounds++
			case '2':
				T.board[r][c] = T.playerB
				T.rounds++
			default:
				T.board[r][c] = " "
			}
			i++
		}
	}
	// Check if the game is over, a draw results in 0 points and a win gives 1 point
	if draw := T.evaluateGame(); draw {
		return true, ""
	} else if T.done {
		return true, T.playerInTurn
	}

	return false, ""
}

// PrintBoard - Prints out the game board
func (T *TicTacToe) PrintBoard() {
	columns := "   A B C D E F G H"
	fmt.Println("")
	for r := int(T.size) - 1; r >= 0; r-- {
		fmt.Printf("%d ", r+1)
		for c := 0; c < int(T.size); c++ {
			fmt.Printf("|%s", T.board[c][r])
		}
		fmt.Print("|\n")
	}
	fmt.Printf("%s\n", columns[0:4+2*(T.size-1)])
	fmt.Println("")
}
