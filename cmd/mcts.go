package main

import (
	"fmt"
	"github.com/petestonefire/go-mcts-v3/internal/conf"
	"github.com/petestonefire/go-mcts-v3/internal/mcts"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/db"
)

// main - Main function
func main() {
	fmt.Println("MCTS Learn")

	var actions []db.Action
	var isEnd bool
	var winner string

	// Get options from console
	gameId, size, maxRounds, forceNew, name, err := conf.GetLearnOptions()
	if err != nil {
		return
	}

	// Assemble all parts that conforms to an MCTS tree in learning mode
	tree, deferFunc, err := mcts.AssembleForLearning(gameId, size, maxRounds, forceNew, name)
	defer deferFunc()
	if err != nil {
		return
	}

	// Main learning loop
	for {
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
