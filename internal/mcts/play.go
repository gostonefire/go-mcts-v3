package mcts

import (
	"fmt"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/db"
	"math/rand"
)

// MoveResult - A structure holding information about the result after a play move
type MoveResult struct {
	IsDone           bool
	Winner           string
	OpponentMoveX    uint8
	OpponentMoveY    uint8
	OpponentMovePass bool
}

// ResetPlayPlayerA - Resets the tree for use in a new play against the tree as an opponent where the human act as
// Player A, i.e. the Player that makes the first move
func (T *Tree) ResetPlayPlayerA() error {
	var err error
	T.Game.Reset()
	T.AtNode, err = T.NodeDB.GetNode(0)
	if err != nil {
		return err
	}

	return nil
}

// ResetPlayPlayerB - Resets the tree for use in a new play against the tree as an opponent where the human act as
// Player B, i.e. the Player that makes the second move
func (T *Tree) ResetPlayPlayerB() error {
	var err error
	T.Game.Reset()
	T.AtNode, err = T.NodeDB.GetNode(0)
	if err != nil {
		return err
	}

	_, _ = T.opponentMove()

	return nil
}

// PlayExploitPlayer - Plays a move as Player and the model uses strict tree evaluation,
// i.e. it never explores it only exploits on statistics learned during learn phase.
// If game tree is not fully explored the opponent will play either by policy if such exist or
// randomly (which of course in itself is a policy)
func (T *Tree) PlayExploitPlayer(x uint8, y uint8, pass bool) (MoveResult, error) {
	var err error

	// Check if proposed move is valid given the current state of the game
	var isValidMove bool
	actions := T.availableGameActions()
	if actions == nil {
		return MoveResult{}, fmt.Errorf("no available actions, game is already over")
	}

	for _, action := range actions {
		if action.X == x && action.Y == y && action.Pass == pass {
			isValidMove = true
			break
		}
	}

	if !isValidMove {
		return MoveResult{}, fmt.Errorf("not a valid move")
	}

	// Check if we are still in an explored part of the game tree and update state (AtNode) accordingly
	if T.AtNode.Assigned && T.AtNode.Actions != nil {
		for _, a := range T.AtNode.Actions {
			if a.X == x && a.Y == y && a.Pass == pass {
				T.AtNode, err = T.NodeDB.GetNode(a.ActionNodeAddress)
				if err != nil {
					return MoveResult{}, err
				}
				break
			}
		}
	}

	// Make move in game
	isDone, winner, _ := T.Game.Move(x, y, pass)

	// Did the move result in end of game
	if isDone {
		return MoveResult{
			IsDone:           true,
			Winner:           winner,
			OpponentMoveX:    0,
			OpponentMoveY:    0,
			OpponentMovePass: false,
		}, nil
	}

	return T.opponentMove()

	/*
		// Find best move for opponent
		var action Action
		if T.AtNode.Actions != nil {
			var selected int
			var maxScore float64

			for i, a := range T.AtNode.Actions {
				if a.Visits == 0 {
					continue
				}

				score := float64(a.Points) / 2 / float64(a.Visits)
				if score > maxScore {
					selected = i
					maxScore = score
				}
			}

			action = Action{
				X:    T.AtNode.Actions[selected].X,
				Y:    T.AtNode.Actions[selected].Y,
				Pass: T.AtNode.Actions[selected].Pass,
			}

			T.AtNode, err = T.NodeDB.GetNode(T.AtNode.Actions[selected].ActionNodeAddress)
			if err != nil {
				return MoveResult{}, err
			}

		} else {
			actions = T.availableGameActions()
			if actions == nil {
				return MoveResult{}, fmt.Errorf("no available actions, should not be possible")
			}

			T.AtNode = db.MCNode{}
			action = actions[rand.Intn(len(actions))]
		}

		isDone, winner, _ = T.Game.Move(action.X, action.Y, action.Pass)

		return MoveResult{
			IsDone:           isDone,
			Winner:           winner,
			OpponentMoveX:    action.X,
			OpponentMoveY:    action.Y,
			OpponentMovePass: action.Pass,
		}, nil
	*/
}

func (T *Tree) opponentMove() (MoveResult, error) {
	// Find best move for opponent
	var action Action
	if T.AtNode.Actions != nil {
		var selected int
		var maxScore float64

		for i, a := range T.AtNode.Actions {
			if a.Visits == 0 {
				continue
			}

			score := float64(a.Points) / 2 / float64(a.Visits)
			if score > maxScore {
				selected = i
				maxScore = score
			}
		}

		action = Action{
			X:    T.AtNode.Actions[selected].X,
			Y:    T.AtNode.Actions[selected].Y,
			Pass: T.AtNode.Actions[selected].Pass,
		}

		var err error
		T.AtNode, err = T.NodeDB.GetNode(T.AtNode.Actions[selected].ActionNodeAddress)
		if err != nil {
			return MoveResult{}, err
		}

	} else {
		actions := T.availableGameActions()
		if actions == nil {
			return MoveResult{}, fmt.Errorf("no available actions, should not be possible")
		}

		T.AtNode = db.MCNode{}
		action = actions[rand.Intn(len(actions))]
	}

	isDone, winner, _ := T.Game.Move(action.X, action.Y, action.Pass)

	return MoveResult{
		IsDone:           isDone,
		Winner:           winner,
		OpponentMoveX:    action.X,
		OpponentMoveY:    action.Y,
		OpponentMovePass: action.Pass,
	}, nil

}

func (T *Tree) PrintBoard() {
	T.Game.PrintBoard()
}
