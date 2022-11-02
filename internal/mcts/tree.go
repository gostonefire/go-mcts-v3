package mcts

import (
	"bufio"
	"fmt"
	"github.com/petestonefire/go-mcts-v3/internal/conf"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/db"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type BoardGame interface {
	Reset()                                                 // Resets game to starting position
	Move(x uint8, y uint8, pass bool) (bool, string, error) // Return: IsDone, Winner (empty string is a draw) and error
	AvailableActions() ([][2]uint8, bool)                   // Returns: [X,Y] coordinates and Pass (with empty slice)
	GetPlayers() [2]string                                  // Player names in start order
	SetPlayers(players [2]string)                           // Player names in start order
	GetState() (string, string)                             // Return: State, Player in turn
	SetState(state, playerInTurn string) (bool, string)     // Sets the game in a specific state, return as Move
	PrintBoard()                                            // Prints the board on console
}

type NodeDB interface {
	AttachActionNodes(parentState, childPlayer string, actions []db.Action, actionResultStates []string) (attachedActions []db.Action, actionsAddress uint64, nReused int64, err error)
	GetTopAction() (action db.Action, err error)
	GetNode(nodeAddress uint64) (mcNode db.MCNode, err error)
	UpdateActionStatistics(actionsAddress uint64, actionIndex uint64, newVisits, newPoints uint64) error
	SetNodeIsEnd(nodeAddress uint64) (err error)
}

type AI interface {
	RecordStateStatistics(player, state string, oldVisits, visits uint64, oldPoints, points float64) (err error)
}

// Action - Convenient structure for an action
type Action struct {
	X    uint8
	Y    uint8
	Pass bool
}

// Tree - Structure representing an MCTS tree
type Tree struct {
	Game             BoardGame
	NodeDB           NodeDB
	AI               AI
	PlayerA          string
	PlayerB          string
	AtNode           db.MCNode
	NNodes           int64
	NReusedNodes     int64
	NUnexpandedNodes int64
	DepthStats       map[int]int64
	Rounds           float64
	MaxRounds        float64
	OverlearnRounds  float64
	OverlearnFactor  float64
	StateFilename    string
}

// NewTree - Returns a new tree with a single node at the top
func NewTree(game BoardGame, nodeDb NodeDB, aiMgmt AI, maxRounds float64, stateFilename string, forceNew bool) *Tree {
	// Seed random generator
	time.Now().UnixNano()
	rand.Seed(time.Now().UnixNano())

	// Get players from the game
	players := game.GetPlayers()

	// Create a tree
	tree := Tree{
		Game:             game,
		NodeDB:           nodeDb,
		AI:               aiMgmt,
		PlayerA:          players[0],
		PlayerB:          players[1],
		NNodes:           1,
		NUnexpandedNodes: 1,
		DepthStats:       make(map[int]int64),
		MaxRounds:        maxRounds,
		OverlearnFactor:  conf.OverlearnFactor,
		StateFilename:    stateFilename,
	}

	tree.DepthStats[1] = 1

	// Remove state file if needed
	if _, err := os.Stat(stateFilename); err == nil {
		if forceNew {
			err = os.Remove(stateFilename)
			if err != nil {
				fmt.Printf("Error while removing old file: %s\n", err)
				return nil
			}
		} else {
			delete(tree.DepthStats, 1)
			err = tree.ReadAndSetState()
			if err != nil {
				return nil
			}
		}
	}

	return &tree
}

// NewPlayTree - Returns a new tree with a single node at the top
func NewPlayTree(game BoardGame, nodeDb NodeDB, aiMgmt AI, stateFilename string) *Tree {
	// Seed random generator
	time.Now().UnixNano()
	rand.Seed(time.Now().UnixNano())

	// Get players from the game
	players := game.GetPlayers()

	// Create a tree
	tree := Tree{
		Game:             game,
		NodeDB:           nodeDb,
		AI:               aiMgmt,
		PlayerA:          players[0],
		PlayerB:          players[1],
		NNodes:           1,
		NUnexpandedNodes: 1,
		OverlearnFactor:  conf.OverlearnFactor,
		StateFilename:    stateFilename,
	}

	err := tree.ReadAndSetState()
	if err != nil {
		return nil
	}

	return &tree
}

// SaveState - Saves the current state of the tree for use if we want to continue to learn later
func (T *Tree) SaveState() error {
	f, err := os.OpenFile(T.StateFilename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error while open or create %s, %s\n", T.StateFilename, err)
		return err
	}
	defer func(f *os.File) { err = f.Close() }(f)

	_, err = fmt.Fprintf(
		f,
		"%s,%s,%d,%.0f,%.0f,%d,%d\n",
		T.PlayerA,
		T.PlayerB,
		T.NNodes,
		T.Rounds,
		T.OverlearnRounds,
		T.NReusedNodes,
		T.NUnexpandedNodes,
	)

	if err != nil {
		fmt.Printf("Error while writing state to file")
		return err
	}

	return nil
}

func (T *Tree) ReadAndSetState() error {
	// If there is no state file we just have to go with defaults
	if _, err := os.Stat(T.StateFilename); err != nil {
		return nil
	}

	f, err := os.OpenFile(T.StateFilename, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Error while open %s, %s\n", T.StateFilename, err)
		return err
	}
	defer func(f *os.File) { err = f.Close() }(f)

	r := bufio.NewReader(f)
	buf, _, err := r.ReadLine()
	state := string(buf)
	states := strings.Split(state, ",")

	playerA := states[0]
	playerB := states[1]
	nNodes, err := strconv.Atoi(states[2])
	if err != nil {
		fmt.Printf("Error while reading number of nodes: %s", err)
		return err
	}
	rounds, err := strconv.Atoi(states[3])
	if err != nil {
		fmt.Printf("Error while reading rounds: %s", err)
		return err
	}
	overlearnRounds, err := strconv.Atoi(states[4])
	if err != nil {
		fmt.Printf("Error while reading overlearnRounds: %s", err)
		return err
	}
	nReusedNodes, err := strconv.Atoi(states[5])
	if err != nil {
		fmt.Printf("Error while reading number of reused nodes: %s", err)
		return err
	}
	nUnexpandedNodes, err := strconv.Atoi(states[6])
	if err != nil {
		fmt.Printf("Error while reading number of unexpanded nodes: %s", err)
		return err
	}

	T.PlayerA = playerA
	T.PlayerB = playerB
	T.NNodes = int64(nNodes)
	T.Rounds = float64(rounds)
	T.MaxRounds += T.Rounds
	T.OverlearnRounds = float64(overlearnRounds)
	T.NReusedNodes = int64(nReusedNodes)
	T.NUnexpandedNodes = int64(nUnexpandedNodes)

	T.Game.SetPlayers([2]string{playerA, playerB})

	return nil
}
