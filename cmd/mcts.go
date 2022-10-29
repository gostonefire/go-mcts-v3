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

// getLearnOptions - Gets input from the executor
func getLearnOptions() (gameId int, size uint8, maxRounds float64, forceNew bool, name string, err error) {
	var input string
	var s, m int
	size = 4
	maxRounds = 1000000

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
			fmt.Printf("Error, malformed number give: %s\n", err)
			return
		}
		size = uint8(s)
	}

	fmt.Print("Max learning rounds [1000000]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.TrimSpace(input); input != "" {
		m, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("Error, malformed number give: %s\n", err)
			return
		}
		maxRounds = float64(m)
	}

	fmt.Print("Force new tree [false]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.ToUpper(strings.TrimSpace(input)); input == "TRUE" {
		forceNew = true
	}

	name = fmt.Sprintf("nodetree%dx%d-%d", size, size, gameId)

	return
}

// main - Main function
func main() {
	fmt.Println("MCTS Learn")
	var game mcts.BoardGame
	var playerA, playerB string

	gameId, size, maxRounds, forceNew, name, err := getLearnOptions()
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
	} else {
		fmt.Println("No game corresponding to given game number")
		return
	}

	initialState, _ := game.GetState()
	nodeDB, err := db.NewNodeTree(name, playerA, playerB, initialState, forceNew)
	if err != nil {
		panic("Error while creating file based node database")
	}
	defer func(ActionsFile *os.File) { _ = ActionsFile.Close() }(nodeDB.ActionsFile)
	defer func(NodeFile *os.File) { _ = NodeFile.Close() }(nodeDB.NodeFile)

	tree := mcts.NewTree(game, nodeDB, maxRounds, fmt.Sprintf("%s.state", name), forceNew)

	var actions []db.Action
	var isEnd bool
	var winner string

	for {
		// Reset game to a starting position
		game.Reset()

		// Execute an MCTS Select to find node to exploit or explore
		actions, err = tree.Select()
		if err != nil {
			panic("Error performing Select")
		}
		if actions == nil {
			// Learning complete
			break
		}

		// Play the game up to and including te selected node
		isEnd, winner = tree.PlayAction(actions[len(actions)-1])

		if !isEnd {
			// Execute an MCTS Expand to add new nodes to explore, one of the new nodes is randomly chosen and returned
			actions, err = tree.Expand(actions)
			if err != nil || actions == nil {
				panic("Error while expanding node tree")
			}

			// Play the expanded node in the game
			isEnd, winner = tree.PlayAction(actions[len(actions)-1])

			if !isEnd {
				// Simulate the game to an end using any simulation policy
				winner, err = tree.Simulate()
				if err != nil {
					panic("Error performing Simulate")
				}
			}
		}

		// Update the IsDone flag in the tree
		if isEnd {
			err = tree.SetNodeIsEnd(actions[len(actions)-1])
			if err != nil {
				panic("Error while updating IsEnd flag in nodes")
			}
		}

		// Update statistics in the game tree
		err = tree.BackPropagation(actions, winner)
		if err != nil {
			panic("Error while performing back propagation")
		}
	}
	fmt.Printf("\nNumber of nodes in final tree: %d\n", tree.NNodes)
	err = tree.SaveState()
	if err != nil {
		fmt.Println("Error while saving state")
	}
}
