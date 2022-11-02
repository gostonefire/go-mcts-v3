package main

import (
	"context"
	"fmt"
	"github.com/petestonefire/go-mcts-v3/internal/mcts"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/db"
	"os"
	"os/signal"
)

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

// main - Main function
func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(exitCodeInterrupt)
	}()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitCodeErr)
	}
}

// run - Runs the application in learning mode
func run(ctx context.Context) (err error) {
	fmt.Println("MCTS Learn")

	var complete bool

	// Assemble all parts that conforms to an MCTS tree in learning mode
	tree, deferFunc, err := mcts.AssembleForLearning()
	defer deferFunc()
	if err != nil {
		return
	}

	// Main learning loop
Loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Received interrupt signal, saving and stopping learn loop")
			break Loop
		default:
			// do a piece of work
			complete, err = learnIteration(tree)
			if err != nil {
				return
			}
			if complete {
				break Loop
			}
		}
	}

	fmt.Printf("\nNumber of nodes in final tree: %d\n", tree.NNodes)

	err = tree.SaveState()
	if err != nil {
		err = fmt.Errorf("error while saving state")
		return
	}
	err = tree.WriteAndCloseAIBuffers()
	if err != nil {
		err = fmt.Errorf("error while writing leftovers from AI buffers to file")
		return
	}

	return
}

// learnIteration - One learning iteration according Monte Carlo Tree Search
func learnIteration(tree *mcts.Tree) (complete bool, err error) {
	var actions []db.Action
	var isEnd bool
	var winner string

	// Execute an MCTS Select to find node to exploit or explore
	actions, err = tree.Select()
	if err != nil {
		err = fmt.Errorf("error performing Select")
		return
	}
	if actions == nil {
		// Learning complete
		complete = true
		return
	}

	// Play the game up to and including te selected node
	isEnd, winner = tree.PlayAction(actions[len(actions)-1])

	if !isEnd {
		// Execute an MCTS Expand to add new nodes to explore, one of the new nodes is randomly chosen and returned
		actions, err = tree.Expand(actions)
		if err != nil || actions == nil {
			err = fmt.Errorf("error while expanding node tree")
			return
		}

		// Play the expanded node in the game
		isEnd, winner = tree.PlayAction(actions[len(actions)-1])

		if !isEnd {
			// Simulate the game to an end using any simulation policy
			winner, err = tree.Simulate()
			if err != nil {
				err = fmt.Errorf("error performing Simulate")
				return
			}
		}
	}

	// Update the IsDone flag in the tree
	if isEnd {
		err = tree.SetNodeIsEnd(actions[len(actions)-1])
		if err != nil {
			err = fmt.Errorf("error while updating IsEnd flag in nodes")
			return
		}
	}

	// Update statistics in the game tree
	err = tree.BackPropagation(actions, winner)
	if err != nil {
		err = fmt.Errorf("error while performing back propagation")
		return
	}

	return
}
