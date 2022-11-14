package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gostonefire/filehashmap"
	"io"
	"math"
	"os"
)

// NodeTree - Struct representing the file based database for a node tree
type NodeTree struct {
	ActionsFile *os.File
	NodeMap     *filehashmap.FileHashMap
	playerA     string
	playerB     string
}

// MCNode - Monte Carlo tree search node
type MCNode struct {
	Assigned       bool
	IsEnd          bool
	State          string
	Player         string
	Actions        []Action
	ActionsAddress uint64
}

// Action - Convenient structure for an action
type Action struct {
	Visits         uint64
	Points         uint64
	X              uint8
	Y              uint8
	Pass           bool
	ActionNode     MCNode
	ActionIndex    uint64
	ActionsAddress uint64
	ActionNodeKey  []byte
	// ActionNodeAddress uint64
}

// nodeState - Type to hold a state in the state registry
type nodeState struct {
	stateCodeHigh uint64
	stateCodeLow  uint64
	playerA       bool
}

// NewNodeTree - Creates a new NodeTree either using a new file or existing
func NewNodeTree(nodeTreeName, playerA, playerB, initialState string, uniqueStates int64, newTree bool) (nodeTree *NodeTree, err error) {
	mFile := fmt.Sprintf("%s-map.bin", nodeTreeName)
	oFile := fmt.Sprintf("%s-ovfl.bin", nodeTreeName)
	aFile := fmt.Sprintf("%s-actions.bin", nodeTreeName)

	_, err1 := os.Stat(mFile)
	_, err2 := os.Stat(oFile)
	_, err3 := os.Stat(aFile)
	if err1 != nil || err2 != nil || err3 != nil {
		newTree = true
	}

	if newTree {
		if err = removeExistingFiles([]string{mFile, oFile, aFile}); err != nil {
			fmt.Println("Error while trying to remove existing node files")
			return
		}
	}

	// Open or create the action file and hash map files
	af, err := os.OpenFile(aFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error while open or create %s, %s\n", aFile, err)
		return nil, err
	}

	var fhm *filehashmap.FileHashMap
	if newTree {
		fhm, _, err = filehashmap.NewFileHashMap(nodeTreeName, uniqueStates, 17, 9, nil)
		if err != nil {
			fmt.Printf("Error while creating FileHashMap, %s\n", err)
			return nil, err
		}
		err = fhm.CreateNewFiles()
		if err != nil {
			fmt.Printf("Error while creating FileHashMap files, %s\n", err)
			return nil, err
		}
	} else {
		fhm, _, err = filehashmap.NewFromExistingFiles(nodeTreeName, nil)
		if err != nil {
			fmt.Printf("Error while opening FileHashMap, %s\n", err)
			return nil, err
		}
	}

	nt := NodeTree{
		ActionsFile: af,
		NodeMap:     fhm,
		playerA:     playerA,
		playerB:     playerB,
	}

	// Add the top node if we are creating a new node tree
	if newTree {
		action := Action{
			X:    math.MaxUint8,
			Y:    math.MaxUint8,
			Pass: false,
		}
		_, err = nt.addAction(playerA, action, initialState)
		if err != nil {
			fmt.Printf("Error while creating first action/node in the new node tree\n")
			return
		}
	}

	nodeTree = &nt
	return
}

// NewPlayNodeTree - Creates a new NodeTree either using a new file or existing
func NewPlayNodeTree(nodeTreeName, playerA, playerB, initialState string) (nodeTree *NodeTree, err error) {
	mFile := fmt.Sprintf("%s-map.bin", nodeTreeName)
	oFile := fmt.Sprintf("%s-ovfl.bin", nodeTreeName)
	aFile := fmt.Sprintf("%s-actions.bin", nodeTreeName)

	_, err1 := os.Stat(mFile)
	_, err2 := os.Stat(oFile)
	_, err3 := os.Stat(aFile)
	if err1 != nil || err2 != nil || err3 != nil {
		return NewNodeTree(nodeTreeName, playerA, playerB, initialState, 10, true)
	}

	// Open the node files
	af, err := os.OpenFile(aFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Error while open node tree for play: %s\n", err)
		return nil, err
	}

	fhm, _, err := filehashmap.NewFromExistingFiles(nodeTreeName, nil)
	if err != nil {
		fmt.Printf("Error while opening FileHashMap, %s\n", err)
		return nil, err
	}

	nt := NodeTree{
		ActionsFile: af,
		NodeMap:     fhm,
		playerA:     playerA,
		playerB:     playerB,
	}

	nodeTree = &nt
	return
}

// removeExistingFiles - Removes any existing node related files if present
func removeExistingFiles(files []string) error {
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			err = os.Remove(file)
			if err != nil {
				fmt.Printf("Error while removing file %s: %s\n", file, err)
				return err
			}
		}
	}
	return nil
}

// AttachActionNodes - Attaches actions structure in the childrens file and updates the node identified with state
// accordingly. To each action a child node is created (or identified if already present) and attached.
// It returns the created children in a slice of Action.
func (N *NodeTree) AttachActionNodes(
	parentState,
	childPlayer string,
	actions []Action,
	actionResultStates []string,
) (
	attachedActions []Action,
	actionsAddress uint64,
	nReused int64,
	err error,
) {
	// Convert states to base3 and create a state key
	stateCodeHigh, stateCodeLow := stateToStateCodes(parentState)
	parentStateKey := nodeStateToBuffer(nodeState{
		stateCodeHigh: stateCodeHigh,
		stateCodeLow:  stateCodeLow,
		playerA:       childPlayer != N.playerA,
	})

	nActions := len(actions)
	attachedActions = make([]Action, nActions)

	parentValue, err := N.NodeMap.Get(parentStateKey)
	if errors.Is(err, filehashmap.NoRecordFound{}) {
		fmt.Printf("Error, no such parentState in node registry: %s\n", parentState)
		err = fmt.Errorf("error, no such parentState in node registry: %s", parentState)
		return
	} else if err != nil {
		return
	}

	buf := make([]byte, 1+nActions*actionLength)

	var reusedNode bool
	var actionNodeKey []byte
	var resultingChild MCNode
	buf[0] = uint8(nActions)
	for i := 0; i < nActions; i++ {
		o := 1 + i*actionLength

		resultingChild, actionNodeKey, reusedNode, err = N.addNode(actionResultStates[i], childPlayer)
		if err != nil {
			return
		}

		actions[i].ActionNode = resultingChild
		actions[i].ActionNodeKey = actionNodeKey
		actionToBuffer(actions[i], buf[o:])

		if reusedNode {
			nReused++
		}
	}

	actionsAddress, err = writeBufferToFile(N.ActionsFile, 0, io.SeekEnd, buf)
	if err != nil {
		return
	}

	binary.LittleEndian.PutUint64(parentValue[actionsOffset:], actionsAddress)
	err = N.NodeMap.Set(parentStateKey, parentValue)
	if err != nil {
		return
	}

	for i := 0; i < nActions; i++ {
		actions[i].ActionIndex = uint64(i)
		actions[i].ActionsAddress = actionsAddress
	}

	attachedActions = actions

	return
}

// addNode - Adds a node to the tree.
// It returns the newly created node and any error
func (N *NodeTree) addNode(state, player string) (mcNode MCNode, stateKey []byte, reusedNode bool, err error) {

	// Convert states to base3 and create a nodeState instance
	// stateCode := nodeState{playerA: player == N.playerA}
	// stateCode.stateCodeHigh, stateCode.stateCodeLow = stateToStateCodes(state)

	// Convert states to base3 and create a state key
	stateCodeHigh, stateCodeLow := stateToStateCodes(state)
	stateKey = nodeStateToBuffer(nodeState{
		stateCodeHigh: stateCodeHigh,
		stateCodeLow:  stateCodeLow,
		playerA:       player == N.playerA,
	})

	nodeValue, err := N.NodeMap.Get(stateKey)
	if errors.Is(err, filehashmap.NoRecordFound{}) {
		mcNode = MCNode{
			Assigned:       true,
			State:          state,
			Player:         player,
			ActionsAddress: math.MaxUint64,
		}
		nodeValue = nodeToBuffer(mcNode)
		err = N.NodeMap.Set(stateKey, nodeValue)
		if err != nil {
			return
		}

	} else if err != nil {
		return
	} else {
		mcNode = bufferToNode(stateKey, nodeValue, N.playerA, N.playerB)
		reusedNode = true
	}

	return
}

// GetTopAction - Returns the top action from the tree
func (N *NodeTree) GetTopAction() (action Action, err error) {
	actions, err := N.getActionsByAddress(0)
	if err != nil {
		return
	}
	if len(actions) == 0 {
		fmt.Println("Error while retrieving action, got zero length slice from file")
		err = fmt.Errorf("got zero length slice from file")
		return
	}
	actions[0].ActionIndex = 0
	actions[0].ActionsAddress = 0

	// Get nodes for associated actions
	for i := 0; i < len(actions); i++ {
		actions[i].ActionNode, err = N.getNodeByNodeKey(actions[i].ActionNodeKey)
		if err != nil {
			return
		}

		actions[i].ActionNode.Actions, err = N.getActionsByAddress(actions[i].ActionNode.ActionsAddress)
		if err != nil {
			return
		}
	}

	action = actions[0]
	return
}

// GetNode - Retrieves a node given its node key
func (N *NodeTree) GetNode(nodeKey []byte) (mcNode MCNode, err error) {
	// Get node data from file
	mcNode, err = N.getNodeByNodeKey(nodeKey)
	if err != nil {
		return
	}

	// Get associated actions from file
	mcNode.Actions, err = N.getActionsByAddress(mcNode.ActionsAddress)
	if err != nil {
		return
	}

	/*
		// Get nodes for associated actions
		for i := 0; i < len(mcNode.Actions); i++ {
			mcNode.Actions[i].ActionNode, err = N.getNodeByAddress(mcNode.Actions[i].ActionNodeAddress)
			if err != nil {
				return
			}

			mcNode.Actions[i].ActionNode.Actions, err = N.getActionsByAddress(mcNode.Actions[i].ActionNode.ActionsAddress)
			if err != nil {
				return
			}
		}
	*/

	return
}

// getNodeByAddress - Retrieves a node given its node key.
func (N *NodeTree) getNodeByNodeKey(nodeKey []byte) (mcNode MCNode, err error) {
	// Get node data from file
	value, err := N.NodeMap.Get(nodeKey)
	if err != nil {
		return
	}

	// Convert data in buffer to a node
	mcNode = bufferToNode(nodeKey, value, N.playerA, N.playerB)

	return
}

// UpdateActionStatistics - Updates visits and points for an action
func (N *NodeTree) UpdateActionStatistics(actionsAddress, actionIndex, newVisits, newPoints uint64) error {
	// Check for a valid actionsAddress
	if actionsAddress == math.MaxUint64 {
		fmt.Println("Error, unassigned actions address provided")
		return fmt.Errorf("unassigned actions address provided")
	}

	buf := make([]byte, 16)

	// Visits in 8 bytes
	binary.LittleEndian.PutUint64(buf, newVisits)

	// Points in 8 bytes
	binary.LittleEndian.PutUint64(buf[pointsOffset:], newPoints)

	fileAddress := actionsAddress + 1 + uint64(actionLength)*actionIndex
	_, err := writeBufferToFile(N.ActionsFile, fileAddress, io.SeekStart, buf)
	if err != nil {
		fmt.Printf("Error while writing updates statistics to action in file\n")
		return err
	}

	return nil
}

// SetNodeIsEnd - Marks a node as is end, i.e. there are no more actions to take from that node
func (N *NodeTree) SetNodeIsEnd(nodeKey []byte) (err error) {
	value, err := N.NodeMap.Get(nodeKey)
	if err != nil {
		fmt.Printf("Error while setting the IsEnd flag to a node in file: %s\n", err)
	}

	value[isEndOffset] = 1

	err = N.NodeMap.Set(nodeKey, value)
	if err != nil {
		fmt.Printf("Error while setting the IsEnd flag to a node in file: %s\n", err)
	}

	return
}

// getActionsByAddress - Retrieves all actions given a file position pointer
func (N *NodeTree) getActionsByAddress(actionsAddress uint64) (actions []Action, err error) {
	// Check for a valid actionsAddress, otherwise just return
	if actionsAddress == math.MaxUint64 {
		return
	}

	// Get number of connected actions in the index record
	var buf []byte
	buf, err = readFileToBuffer(N.ActionsFile, actionsAddress, io.SeekStart, 1)
	if err != nil {
		return
	}
	nActions := int(buf[0])

	// Get index record
	buf, err = readFileToBuffer(N.ActionsFile, 0, io.SeekCurrent, nActions*actionLength)
	if err != nil {
		return nil, err
	}

	actions = make([]Action, nActions)
	for i := 0; i < nActions; i++ {
		actions[i] = bufferToAction(buf[i*actionLength:])
		actions[i].ActionIndex = uint64(i)
		actions[i].ActionsAddress = actionsAddress
		if err != nil {
			return
		}
	}
	return
}

// addAction - Add one single action and create or identify its resulting node. It is not attached to any parent.
// It returns the created action and its actions address in the actions file
func (N *NodeTree) addAction(
	actionPlayer string,
	action Action,
	actionResultState string,
) (
	addedAction Action,
	err error,
) {
	buf := make([]byte, 1+actionLength)

	var resultingChild MCNode
	var childNodeKey []byte
	buf[0] = 1

	resultingChild, childNodeKey, _, err = N.addNode(actionResultState, actionPlayer)
	if err != nil {
		return
	}

	action.ActionNode = resultingChild
	action.ActionNodeKey = childNodeKey
	actionToBuffer(action, buf[1:])

	action.ActionsAddress, err = writeBufferToFile(N.ActionsFile, 0, io.SeekEnd, buf)
	if err != nil {
		return
	}

	action.ActionIndex = uint64(0)
	addedAction = action

	return
}

// readFileToBuffer - Reads n bytes from file and returns a buffer with the data
func readFileToBuffer(filePtr *os.File, offset uint64, whence, nBytes int) ([]byte, error) {
	// Seek read position
	_, err := filePtr.Seek(int64(offset), whence)
	if err != nil {
		fmt.Printf("Error while setting file pointer: %s\n", err)
		return nil, err
	}

	// Create a buffer and read file addresses for all children
	buf := make([]byte, nBytes)
	n, err := filePtr.Read(buf)
	if err != nil {
		fmt.Printf("Error while reading from file: %s\n", err)
		return nil, err
	}
	if n != nBytes {
		fmt.Printf("Error while reading from file: Expected %d bytes, got %d bytes\n", nBytes, n)
		return nil, fmt.Errorf("error while reading from file: Expected %d bytes, got %d bytes\n", nBytes, n)
	}

	return buf, nil
}

// writeBufferToFile - Writes a buffer to the node file
func writeBufferToFile(filePtr *os.File, offset uint64, whence int, buffer []byte) (uint64, error) {
	// Seek write position
	writePos, err := filePtr.Seek(int64(offset), whence)
	if err != nil {
		fmt.Printf("Error while setting file pointer: %s\n", err)
		return 0, err
	}

	// Write buffer to the file
	_, err = filePtr.Write(buffer)
	if err != nil {
		fmt.Printf("Error while writing buffer to file: %s", err)
		return 0, err
	}

	return uint64(writePos), nil
}
