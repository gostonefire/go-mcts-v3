package mcts

import (
	"fmt"
	"github.com/petestonefire/go-mcts-v3/internal/conf"
	"github.com/petestonefire/go-mcts-v3/internal/mcts/db"
	"math/rand"
	"sort"
)

// Select - Traverses the tree to find node to explore or exploit.
// It returns all traversed node up to and including the leaf node.
func (T *Tree) Select() (actions []db.Action, err error) {
	action, err := T.NodeDB.GetTopAction()
	if err != nil {
		return
	}

	// Print statistics
	if T.Rounds > 0 && int64(T.Rounds)%10000 == 0 {
		T.printStatistics(false)
	}

	// Check if we have reached max number of finished rounds
	if T.Rounds >= T.MaxRounds {
		fmt.Println("\nMax rounds reached")
		T.printStatistics(true)

		return
	}

	actions = []db.Action{action}

	if action.ActionNode.Actions == nil {
		return
	}

	for {
		var uctValue float64
		var selected int
		var maxUCT float64

		if rand.Float32() < conf.RandomRoundThreshold {
			selected = rand.Intn(len(action.ActionNode.Actions))
		} else {
			for i, a := range action.ActionNode.Actions {
				// Nodes without Visits shall always be chosen ahead already visited nodes
				if a.Visits == 0 {
					selected = i
					break
				}

				// Get UCT from child
				uctValue, err = uct(action.Visits, a.Visits, a.Points)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}

				// Register new selected if child has better UCT value
				if uctValue > maxUCT {
					selected = i
					maxUCT = uctValue
				}
			}
		}

		action = action.ActionNode.Actions[selected]

		action.ActionNode, err = T.NodeDB.GetNode(action.ActionNodeAddress)
		if err != nil {
			return
		}

		actions = append(actions, action)

		// If selected node does not have any Children (leafs) then we are at the finally selected node from the tree
		if action.ActionNode.Actions == nil {
			return
		}
	}
}

// Expand - expands one leaf with new unvisited nodes.
// It returns one random child out of the created.
func (T *Tree) Expand(actions []db.Action) (resultActions []db.Action, err error) {
	// Get available gameActions from the associated game, a nil indicates no available gameActions and the game branch is ended
	lastAction := len(actions) - 1
	action := actions[lastAction]
	_, _ = T.Game.SetState(action.ActionNode.State, action.ActionNode.Player)
	gameActions := T.availableGameActions()
	if gameActions == nil {
		return
	}

	// Make some preparation and decide who is player for the expanded nodes
	var player string
	var nReused int64
	var actionsAddress uint64
	nActions := len(gameActions)
	newActions := make([]db.Action, nActions)
	states := make([]string, nActions)
	if action.ActionNode.Player == T.PlayerA {
		player = T.PlayerB
	} else {
		player = T.PlayerA
	}

	for n := 0; n < nActions; n++ {
		_, _, err = T.Game.Move(gameActions[n].X, gameActions[n].Y, gameActions[n].Pass)
		if err != nil {
			return
		}
		state, _ := T.Game.GetState()
		_, _ = T.Game.SetState(action.ActionNode.State, action.ActionNode.Player)

		newActions[n].X = gameActions[n].X
		newActions[n].Y = gameActions[n].Y
		newActions[n].Pass = gameActions[n].Pass
		states[n] = state
	}

	newActions, actionsAddress, nReused, err = T.NodeDB.AttachActionNodes(action.ActionNode.State, player, newActions, states)
	if err != nil {
		return
	}
	actions[lastAction].ActionNode.Actions = newActions
	actions[lastAction].ActionNode.ActionsAddress = actionsAddress

	newNodes := int64(len(newActions)) - nReused
	T.NNodes += newNodes
	T.NReusedNodes += nReused
	T.NUnexpandedNodes += newNodes - 1 // Removing one since we now have expanded one node

	// Pick one random action out of the created ones
	resultActions = append(actions, newActions[rand.Intn(nActions)])

	// Update depth stats
	if newNodes > 0 {
		depth := len(resultActions)
		_, ok := T.DepthStats[depth]
		if ok {
			T.DepthStats[depth] += newNodes
		} else {
			T.DepthStats[depth] = newNodes
		}
	}
	return
}

// Simulate - Plays a game to the end using simulation policy
func (T *Tree) Simulate() (string, error) {
	// Start play out simulation
	for {
		action := T.simulationPolicy()

		isDone, winner, err := T.Game.Move(action.X, action.Y, action.Pass)
		if err != nil {
			fmt.Printf("Error while making a move: %s\n", err)
			return "", err
		}
		if isDone {
			return winner, nil
		}
	}
}

// BackPropagation - Updates the tree with statistics after a simulation
func (T *Tree) BackPropagation(actions []db.Action, winner string) error {
	for i := len(actions) - 1; i >= 0; i-- {
		err := T.updateActionStatistics(actions[i], winner)
		if err != nil {
			return err
		}
	}

	T.Rounds++

	return nil
}

// PlayAction - Plays the game given the action.
// It returns whether the game is over, Winner (empty string is a draw) and error
func (T *Tree) PlayAction(action db.Action) (bool, string) {
	isDone, winner := T.Game.SetState(action.ActionNode.State, action.ActionNode.Player)

	return isDone, winner
}

// SetNodeIsEnd - Marks a node as is end, i.e. there are no more actions to take from that node
func (T *Tree) SetNodeIsEnd(action db.Action) (err error) {
	if !action.ActionNode.IsEnd {
		// Remove one from unexpanded nodes since this one is at the end and cannot be expanded
		T.NUnexpandedNodes--

		return T.NodeDB.SetNodeIsEnd(action.ActionNodeAddress)
	}

	return
}

// simulationPolicy - Gets next Action in a simulation and applies whatever policy determined suitable
func (T *Tree) simulationPolicy() Action {
	actions := T.availableGameActions()

	return actions[rand.Intn(len(actions))]
}

// availableGameActions - Returns available actions from the game in mcts Action format
func (T *Tree) availableGameActions() []Action {
	// Get available actions from game
	availableActions, pass := T.Game.AvailableActions()

	// Handle situation where the game responds with nil instead of empty slice
	nAvailableActions := 0
	if availableActions != nil {
		nAvailableActions = len(availableActions)
	}

	// No actions and not a Pass indicates that there are no actions to take att all, i.e. end of game
	if nAvailableActions == 0 && !pass {
		return nil
	}

	// If it is a Pass Action, just return the one Action with that information
	if pass {
		return []Action{
			{
				X:    0,
				Y:    0,
				Pass: true,
			},
		}
	}

	// At this point we know there are actions to take besides a Pass, so format them according mcts format
	actions := make([]Action, nAvailableActions)
	for n, a := range availableActions {
		actions[n] = Action{
			X: a[0],
			Y: a[1],
		}
	}

	return actions
}

// updateActionStatistics - Wrapper function over the NodeDB function with similar name, but this one adds the
// points and visits logic
func (T *Tree) updateActionStatistics(action db.Action, winner string) (err error) {
	newPoints := action.Points
	if winner == "" {
		// It's a draw
		newPoints += 1
	} else if action.ActionNode.Player == winner {
		newPoints += 2
	}

	newVisits := action.Visits + 1

	err = T.NodeDB.UpdateActionStatistics(action.ActionsAddress, action.ActionIndex, newVisits, newPoints)
	if err != nil {
		return
	}

	if T.Rounds > conf.AIWarmUpRounds {
		T.AI.RecordStateStatistics(action.ActionNode.Player, action.ActionNode.State, newVisits, float64(newPoints)/2)
	}

	return
}

func (T *Tree) printStatistics(finalPrint bool) {
	fmt.Printf(
		"%.0f rounds, %d unique nodes, %d reused nodes, %d unexpanded nodes\n",
		T.Rounds,
		T.NNodes,
		T.NReusedNodes,
		T.NUnexpandedNodes,
	)

	if finalPrint {
		i := 0
		depths := make([]int, len(T.DepthStats))
		for k := range T.DepthStats {
			depths[i] = k
			i++
		}
		sort.Ints(depths)
		for _, d := range depths {
			fmt.Printf("depth: %d [%d]\n", d, T.DepthStats[d])
		}

	}
}
