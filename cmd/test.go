package main

import (
	"fmt"
	"github.com/gostonefire/go-mcts-v3/internal/verticalfourinarow"
)

func main() {
	game := verticalfourinarow.NewVerticalFIR("B", "W")

	game.PrintBoard()
	fmt.Println(game.GetState())

	moves := []uint8{2, 2, 2, 2, 2, 3, 3, 3, 3, 6, 4, 4, 4, 1, 5, 6, 5}

	var state, player string
	for _, m := range moves {
		isDone, winner, err := game.Move(m, 0, false)
		if err != nil {
			fmt.Printf("Got error: %s\n", err)
			break
		}
		game.PrintBoard()
		fmt.Println(game.GetState())
		if isDone {
			fmt.Printf("Winner: %s\n", winner)
			state, player = game.GetState()
			break
		}
	}

	game.Reset()
	fmt.Println(game.SetState(state, player))
}
