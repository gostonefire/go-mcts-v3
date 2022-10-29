package main

import (
	"bufio"
	"fmt"
	"github.com/petestonefire/go-mcts-v3/internal/mcts"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/db"
	"github.com/petestonefire/go-mcts-v3/internal/othello"
	"github.com/petestonefire/go-mcts-v3/internal/tictactoe"
	"os"
	"strconv"
	"strings"
)

// getPlayOptions - Gets input from the executor
func getPlayOptions() (gameId int, size uint8, name string, err error) {
	var input string
	var s int
	size = 4

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Game [0 - TicTacToe, 1 - Othello]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.TrimSpace(input); input != "" {
		s, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("Error, malformed number give: %s\n", err)
			return
		}
		gameId = s
	}

	fmt.Print("Size [4]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.TrimSpace(input); input != "" {
		s, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("Error, malformed number given: %s\n", err)
			return
		}
		size = uint8(s)
	}

	name = fmt.Sprintf("nodetree%dx%d-%d", size, size, gameId)

	return
}

// main - Main function
func main() {
	fmt.Println("MCTS Play")
	var game mcts.BoardGame
	var passAllowed bool
	var playerA, playerB string

	gameId, size, name, err := getPlayOptions()
	if err != nil {
		return
	}

	if gameId == 0 {
		playerA = "X"
		playerB = "Y"
		game = tictactoe.NewTicTacToe(size, playerA, playerB)
	} else if gameId == 1 {
		playerA = "B"
		playerB = "W"
		game, err = othello.NewOthello(size, playerA, playerB)
		if err != nil {
			return
		}
		passAllowed = true
	} else {
		fmt.Println("No game corresponding to given game number")
		return
	}

	initialState, _ := game.GetState()
	nodeDB, err := db.NewPlayNodeTree(name, playerA, playerB, initialState)
	if err != nil {
		panic("Error while open/create file based node database")
	}
	defer func(ActionsFile *os.File) { _ = ActionsFile.Close() }(nodeDB.ActionsFile)
	defer func(NodeFile *os.File) { _ = NodeFile.Close() }(nodeDB.NodeFile)

	tree := mcts.NewPlayTree(game, nodeDB, fmt.Sprintf("%s.state", name))

	err = tree.ResetPlayPlayerB()
	if err != nil {
		fmt.Println("Unable to reset game")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Ready to play!")
	game.PrintBoard()

	var result mcts.MoveResult
	for {
		fmt.Print("Column [A,B...]: ")
		moveX, _ := reader.ReadString('\n')
		moveX = strings.ToUpper(strings.TrimSpace(moveX))
		x := strings.Index("ABCDEFGH", moveX)

		if moveX == "" && passAllowed {
			result, err = tree.PlayExploitPlayer(0, 0, true)
			if err != nil {
				fmt.Printf("Error while making move: %s\n", err)
				return
			}
		} else {
			fmt.Print("Row [1,2...]: ")
			moveY, _ := reader.ReadString('\n')
			y, _ := strconv.Atoi(strings.TrimSpace(moveY))

			result, err = tree.PlayExploitPlayer(uint8(x), uint8(y)-1, false)
			if err != nil {
				fmt.Printf("Error while making move: %s\n", err)
				return
			}
		}

		game.PrintBoard()

		if result.IsDone {
			if result.Winner != "" {
				fmt.Printf("Winner is %s\n", result.Winner)
			} else {
				fmt.Printf("Game is a draw\n")
			}
			return
		}
	}
}
