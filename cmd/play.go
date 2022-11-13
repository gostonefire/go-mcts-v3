package main

import (
	"bufio"
	"fmt"
	"github.com/gostonefire/go-mcts-v3/internal/mcts"
	"os"
	"strconv"
	"strings"
)

// main - Main function
func main() {
	fmt.Println("MCTS Play")

	// Assemble all parts that conforms to an MCTS tree in learning mode
	tree, passAllowed, deferFunc, err := mcts.AssembleForPlay()
	defer deferFunc()

	err = tree.ResetPlayPlayerB()
	if err != nil {
		fmt.Println("Unable to reset game")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Ready to play!")
	tree.PrintBoard()

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

		tree.PrintBoard()

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
