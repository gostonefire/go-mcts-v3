package main

import (
	"fmt"
	"github.com/gostonefire/go-mcts-v3/internal/othello"
)

func main() {
	game, err := othello.NewOthello(4, "B", "W")
	if err != nil {
		return
	}

	game.PrintBoard()

	// Black
	isDone, winner, err := game.Move(1, 3, false)
	game.PrintBoard()
	actions, pass := game.AvailableActions()
	fmt.Println(actions, pass)

	// White
	isDone, winner, err = game.Move(0, 3, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// Black
	isDone, winner, err = game.Move(0, 2, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// White
	isDone, winner, err = game.Move(0, 1, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// Black
	isDone, winner, err = game.Move(0, 0, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// White
	isDone, winner, err = game.Move(2, 3, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// Black
	isDone, winner, err = game.Move(3, 1, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// White
	isDone, winner, err = game.Move(3, 0, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// Black
	isDone, winner, err = game.Move(3, 3, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// White
	isDone, winner, err = game.Move(3, 2, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// Black
	isDone, winner, err = game.Move(0, 0, true)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	// White
	isDone, winner, err = game.Move(2, 0, false)
	game.PrintBoard()
	actions, pass = game.AvailableActions()
	fmt.Println(actions, pass)

	fmt.Println(isDone, winner)

}
