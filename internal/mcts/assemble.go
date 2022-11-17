package mcts

import (
	"fmt"
	"github.com/gostonefire/go-mcts-v3/internal/conf"
	"github.com/gostonefire/go-mcts-v3/internal/mcts/ai"
	"github.com/gostonefire/go-mcts-v3/internal/mcts/db"
	"github.com/gostonefire/go-mcts-v3/internal/othello"
	"github.com/gostonefire/go-mcts-v3/internal/tictactoe"
	"github.com/gostonefire/go-mcts-v3/internal/verticalfourinarow"
)

// AssembleForLearning - Assembles all parts necessary for learning mode
func AssembleForLearning() (
	tree *Tree,
	deferFunc func(),
	err error,
) {

	var game BoardGame
	var playerA, playerB string

	// Get options from console
	gameId, size, maxRounds, uniqueStates, forceNew, name, err := conf.GetLearnOptions()
	if err != nil {
		return
	}

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
	} else if gameId == 2 {
		playerA = "B"
		playerB = "W"
		game = verticalfourinarow.NewVerticalFIR(playerA, playerB)
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
	nodeDB, err := db.NewNodeTree(name, playerA, playerB, initialState, uniqueStates, forceNew)
	if err != nil {
		fmt.Println("Error while creating file based node database")
		err = fmt.Errorf("error while creating file based node database")
		return
	}
	deferFunc = func() {
		_ = nodeDB.ActionsFile.Close()
		nodeDB.NodeMap.CloseFiles()
	}

	// Create AI management assets
	aiMgmt, err := ai.NewAI(name, conf.AIHighValueThreshold, conf.AILowValueThreshold, conf.AIVisitsThreshold, int(size*size), forceNew)
	if err != nil {
		fmt.Println("Error while creating AI management assets")
		err = fmt.Errorf("error while creating AI management assets")
		return
	}

	// Create the mcts tree instance
	tree = NewTree(game, nodeDB, aiMgmt, maxRounds, fmt.Sprintf("%s.state", name), forceNew)

	return
}

// AssembleForPlay - Assembles all parts necessary for play mode
func AssembleForPlay() (
	tree *Tree,
	passAllowed bool,
	deferFunc func(),
	err error,
) {

	var game BoardGame
	var playerA, playerB string

	// Get options from console
	gameId, size, name, err := conf.GetPlayOptions()
	if err != nil {
		return
	}

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
		nodeDB.NodeMap.CloseFiles()
	}

	// Create AI management assets
	aiMgmt, err := ai.NewAI(name, conf.AIHighValueThreshold, conf.AILowValueThreshold, conf.AIVisitsThreshold, int(size*size), false)
	if err != nil {
		fmt.Println("Error while creating AI management assets")
		err = fmt.Errorf("error while creating AI management assets")
		return
	}

	tree = NewPlayTree(game, nodeDB, aiMgmt, fmt.Sprintf("%s.state", name))

	return
}
