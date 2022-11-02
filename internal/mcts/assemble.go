package mcts

import (
	"fmt"
	"github.com/petestonefire/go-mcts-v3/internal/conf"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/ai"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/db"
	"github.com/petestonefire/go-mcts-v3/internal/othello"
	"github.com/petestonefire/go-mcts-v3/internal/tictactoe"
)

// AssembleForLearning - Assembles all parts necessary for learning mode
func AssembleForLearning(
	gameId int,
	size uint8,
	maxRounds float64,
	forceNew bool,
	name string,
) (
	tree *Tree,
	deferFunc func(),
	err error,
) {

	var game BoardGame
	var playerA, playerB string

	deferFunc = func() {}

	// Create the game instance
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
		err = fmt.Errorf("error, no game corresponding to given game number")
		return
	}
	initialState, _ := game.GetState()

	// Create the node tree db instance
	nodeDB, err := db.NewNodeTree(name, playerA, playerB, initialState, forceNew)
	if err != nil {
		fmt.Println("Error while creating file based node database")
		err = fmt.Errorf("error while creating file based node database")
		return
	}
	deferFunc = func() {
		_ = nodeDB.ActionsFile.Close()
		_ = nodeDB.NodeFile.Close()
	}

	// Create AI management assets
	aiMgmt, err := ai.NewAI(name, conf.AIHighValueThreshold, conf.AILowValueThreshold, conf.AIVisitsThreshold, int(size*size), conf.AIMaxListLengths, conf.AIBulkSize, forceNew)
	if err != nil {
		fmt.Println("Error while creating AI management assets")
		err = fmt.Errorf("error while creating AI management assets")
		return
	}
	deferFunc = func() {
		_ = nodeDB.ActionsFile.Close()
		_ = nodeDB.NodeFile.Close()
		_ = aiMgmt.AIFile.Close()
	}

	// Create the mcts tree instance
	tree = NewTree(game, nodeDB, aiMgmt, maxRounds, fmt.Sprintf("%s.state", name), forceNew)

	return
}

// AssembleForPlay - Assembles all parts necessary for play mode
func AssembleForPlay(
	gameId int,
	size uint8,
	name string,
) (
	tree *Tree,
	passAllowed bool,
	deferFunc func(),
	err error,
) {

	var game BoardGame
	var playerA, playerB string

	deferFunc = func() {}

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
		err = fmt.Errorf("error, no game corresponding to given game number")
		return
	}

	initialState, _ := game.GetState()
	nodeDB, err := db.NewPlayNodeTree(name, playerA, playerB, initialState)
	if err != nil {
		fmt.Println("Error while open/create file based node database")
		err = fmt.Errorf("error while open/create file based node database")
		return
	}
	deferFunc = func() {
		_ = nodeDB.ActionsFile.Close()
		_ = nodeDB.NodeFile.Close()
	}

	// Create AI management assets
	aiMgmt, err := ai.NewAI(name, conf.AIHighValueThreshold, conf.AILowValueThreshold, conf.AIVisitsThreshold, int(size*size), conf.AIMaxListLengths, conf.AIBulkSize, false)
	if err != nil {
		fmt.Println("Error while creating AI management assets")
		err = fmt.Errorf("error while creating AI management assets")
		return
	}
	deferFunc = func() {
		_ = nodeDB.ActionsFile.Close()
		_ = nodeDB.NodeFile.Close()
		_ = aiMgmt.AIFile.Close()
	}

	tree = NewPlayTree(game, nodeDB, aiMgmt, fmt.Sprintf("%s.state", name))

	return
}
