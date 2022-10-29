package othello

import (
	"fmt"
)

type column []string
type coords [2]int
type legit struct {
	move  coords
	flips []coords
}

// Othello - Represents the board game with the same name
type Othello struct {
	board            []column
	playerA          string
	playerB          string
	legitPlayerMoves map[string][]legit
	check            [8][2]int
	playerInTurn     string
	size             int
	done             bool
}

// NewOthello - Returns a new instance of the game
func NewOthello(size uint8, playerA string, playerB string) (*Othello, error) {
	var sizeOk bool
	allowed := [3]uint8{4, 6, 8}
	for i := 0; i < len(allowed); i++ {
		if size == allowed[i] {
			sizeOk = true
			break
		}
	}

	if !sizeOk {
		fmt.Printf("Error, size not allowed: %d\n", size)
		return nil, fmt.Errorf("error, size not allowed: %d", size)
	}

	t := Othello{
		size:             int(size),
		playerA:          playerA,
		playerB:          playerB,
		legitPlayerMoves: make(map[string][]legit),
	}
	t.check = [8][2]int{{-1, 1}, {0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}}
	t.Reset()

	return &t, nil
}

// Reset - Resets the game to be prepared for a new game
func (O *Othello) Reset() {
	board := make([]column, O.size)
	for r := 0; r < O.size; r++ {
		c := make(column, O.size)
		board[r] = c
	}

	for x := 0; x < O.size; x++ {
		for y := 0; y < O.size; y++ {
			board[x][y] = " "
		}
	}

	// Set board with initial markers stated by the game of Othello
	for _, c := range []coords{{O.size/2 - 1, O.size/2 - 1}, {O.size / 2, O.size / 2}} {
		board[c[0]][c[1]] = O.playerA
	}
	for _, c := range []coords{{O.size / 2, O.size/2 - 1}, {O.size/2 - 1, O.size / 2}} {
		board[c[0]][c[1]] = O.playerB
	}

	O.board = board
	O.playerInTurn = O.playerA
	O.done = false

	// Update legit moves for the player in turn (i.e. playerA after a reset)
	O.legitPlayerMoves[O.playerA] = O.evaluateLegitMoves(O.playerB, O.playerA)
}

func (O *Othello) Move(x uint8, y uint8, pass bool) (isDone bool, winner string, err error) {
	if pass && len(O.legitPlayerMoves[O.playerInTurn]) > 0 {
		fmt.Println("Illegal to pass while having legit moves to chose among")
		err = fmt.Errorf("error, illegal to pass while having legit moves to chose among")
		return
	}
	if !pass {
		var moveLegit bool
		for _, l := range O.legitPlayerMoves[O.playerInTurn] {
			if l.move[0] == int(x) && l.move[1] == int(y) {
				O.board[x][y] = O.playerInTurn
				for _, f := range l.flips {
					O.board[f[0]][f[1]] = O.playerInTurn
				}
				moveLegit = true
			}
		}
		if !moveLegit {
			fmt.Println("Proposed move is not legit")
			err = fmt.Errorf("error, proposed move is not legit")
			return
		}
	}

	// Switch players
	if O.playerInTurn == O.playerA {
		O.playerInTurn = O.playerB
	} else {
		O.playerInTurn = O.playerA
	}

	// Evaluate game after move
	O.legitPlayerMoves[O.playerA] = O.evaluateLegitMoves(O.playerB, O.playerA)
	O.legitPlayerMoves[O.playerB] = O.evaluateLegitMoves(O.playerA, O.playerB)
	winner = O.evaluateGame()
	isDone = O.done

	return
}

// AvailableActions - Returns available actions given who is the player in turn
func (O *Othello) AvailableActions() (legit [][2]uint8, pass bool) {
	// Get legit moves for whoever is the player in turn
	set := make(map[coords]bool)
	for _, c := range O.legitPlayerMoves[O.playerInTurn] {
		set[c.move] = true
	}

	i := 0
	legit = make([][2]uint8, len(set))
	for c := range set {
		legit[i] = [2]uint8{uint8(c[0]), uint8(c[1])}
		i++
	}

	// If there were no legit moves, then the player has to pass
	if len(legit) == 0 {
		pass = true
	}

	return
}

// GetPlayers - Returns the two players of the game in start order
func (O *Othello) GetPlayers() [2]string {
	return [2]string{O.playerA, O.playerB}
}

// SetPlayers - Sets the players of the game
func (O *Othello) SetPlayers(players [2]string) {
	O.playerA = players[0]
	O.playerB = players[1]
}

// GetState - Gets the state of the game as a base3 number formatted as a string and the player in turn
func (O *Othello) GetState() (string, string) {
	buf := make([]byte, O.size*O.size)
	i := 0
	for r := 0; r < O.size; r++ {
		for c := 0; c < O.size; c++ {
			switch O.board[r][c] {
			case O.playerA:
				buf[i] = '1'
			case O.playerB:
				buf[i] = '2'
			default:
				buf[i] = '0'
			}
			i++
		}
	}

	return string(buf), O.playerInTurn
}

// SetState - Sets the game according given state and player in turn
func (O *Othello) SetState(state, playerInTurn string) (isDone bool, winner string) {
	// Fix length of state by left padding with zeros
	diff := int(O.size*O.size) - len(state)
	if diff > 0 {
		state = fmt.Sprintf("%0*d%s", diff, 0, state)
	}

	i := 0
	O.playerInTurn = playerInTurn
	O.done = false
	for r := 0; r < O.size; r++ {
		for c := 0; c < O.size; c++ {
			switch state[i] {
			case '1':
				O.board[r][c] = O.playerA
			case '2':
				O.board[r][c] = O.playerB
			default:
				O.board[r][c] = " "
			}
			i++
		}
	}
	// Evaluate the game
	O.legitPlayerMoves[O.playerA] = O.evaluateLegitMoves(O.playerB, O.playerA)
	O.legitPlayerMoves[O.playerB] = O.evaluateLegitMoves(O.playerA, O.playerB)
	winner = O.evaluateGame()
	isDone = O.done

	return

}

// PrintBoard - Prints out the game board
func (O *Othello) PrintBoard() {
	columns := "   A B C D E F G H"
	fmt.Println("")
	for r := int(O.size) - 1; r >= 0; r-- {
		fmt.Printf("%d ", r+1)
		for c := 0; c < int(O.size); c++ {
			fmt.Printf("|%s", O.board[c][r])
		}
		fmt.Print("|\n")
	}
	fmt.Printf("%s\n", columns[0:4+2*(O.size-1)])
	fmt.Println("")
}

// getPlayerCoords - Returns coordinates for every brick the given player has on the board
func (O *Othello) getPlayerCoords(player string) (playerCoords []coords) {
	playerCoords = make([]coords, 0)
	for r := int(O.size) - 1; r >= 0; r-- {
		for c := 0; c < int(O.size); c++ {
			if O.board[c][r] == player {
				playerCoords = append(playerCoords, coords{c, r})
			}
		}
	}
	return
}

// evaluateLegitMoves - Returns all legit moves for a player
func (O *Othello) evaluateLegitMoves(opponent, player string) []legit {
	legitMoves := make([]legit, 0)
	opponentCoords := O.getPlayerCoords(opponent)
	for _, coord := range opponentCoords {
		col := coord[0]
		row := coord[1]
		for _, ch := range O.check {
			flips := []coords{{col, row}}
			c := col + ch[0]
			r := row + ch[1]
			if c < 0 || c >= O.size || r < 0 || r >= O.size || O.board[c][r] != " " {
				continue
			}

			// Vertical line
			var valid bool
			if ch[0] == 0 {
				for i := row - ch[1]; i >= 0 && i < O.size; i -= ch[1] {
					if O.board[col][i] == player {
						valid = true
						break
					}
					flips = append(flips, coords{col, i})
				}
				if valid {
					legitMoves = append(legitMoves, legit{
						move:  coords{c, r},
						flips: flips,
					})
				}
				continue
			}

			// Horizontal line
			if ch[1] == 0 {
				for i := col - ch[0]; i >= 0 && i < O.size; i -= ch[0] {
					if O.board[i][row] == player {
						valid = true
						break
					}
					flips = append(flips, coords{i, row})
				}
				if valid {
					legitMoves = append(legitMoves, legit{
						move:  coords{c, r},
						flips: flips,
					})
				}
				continue
			}

			// Diagonal line
			i := col - ch[0]
			j := row - ch[1]
			for i >= 0 && i < O.size && j >= 0 && j < O.size {
				if O.board[i][j] == player {
					valid = true
					break
				}
				flips = append(flips, coords{i, j})
				i -= ch[0]
				j -= ch[1]
			}
			if valid {
				legitMoves = append(legitMoves, legit{
					move:  coords{c, r},
					flips: flips,
				})
			}
		}
	}

	return legitMoves
}

// evaluateGame - Evaluates whether the game is finished in which case done is set to true.
// It returns the winner if any, so a combination of the done flag and whether there was a winner gives
// whether it was a draw or not
func (O *Othello) evaluateGame() (winner string) {
	bricksA, bricksB := O.getLeaderBoard()

	if len(O.legitPlayerMoves[O.playerA]) == 0 && len(O.legitPlayerMoves[O.playerB]) == 0 || bricksA+bricksB == O.size*O.size {
		O.done = true
		if bricksA > bricksB {
			winner = O.playerA
		} else if bricksB > bricksA {
			winner = O.playerB
		}
	}

	return
}

// getLeaderBoard - Returns number of bricks for each player
func (O *Othello) getLeaderBoard() (bricksA, bricksB int) {
	for r := int(O.size) - 1; r >= 0; r-- {
		for c := 0; c < int(O.size); c++ {
			if O.board[c][r] == O.playerA {
				bricksA++
			} else if O.board[c][r] == O.playerB {
				bricksB++
			}
		}
	}

	return
}
