package db

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// numberBase - Number base to use when converting from nodeState of a board to uint64 and back
const numberBase string = "012"

// Node key offsets
const nodeKeyLength int = 17
const stateHighOffset uint64 = 0
const stateLowOffset uint64 = 8
const playerOffset uint64 = 16

// Node value offsets
const nodeValueLength int = 9
const isEndOffset uint64 = 0
const actionsOffset uint64 = 1

/*
	Assigned
	State         8 bytes
	Player        1 byte
	Actions
	ActionNodeAddress
	aIndexAddress 8 bytes

*/

// Total length of one action in file
const actionLength int = 36

// Action offsets

// Offsets in relation to first byte of an Action in file
// Note: code depends on visitsOffset and pointsOffset being first and in that order.
const visitsOffset uint64 = 0        // Needs to be first - 8 bytes
const pointsOffset uint64 = 8        // Needs to be second - 8 bytes
const actionXOffset uint64 = 16      // 1 byte
const actionYOffset uint64 = 17      // 1 byte
const actionPassOffset uint64 = 18   // 1 byte
const childNodeKeyOffset uint64 = 19 // 17 bytes

/*
	Visits         8 bytes
	Points         8 bytes
	X              1 byte
	Y              1 byte
	Pass           1 byte
	IsDone         1 byte
	ActionNode 8 byte
	ActionNodeAddress
*/

// nodeToBuffer - Converts an MCNode to a byte buffer (value only)
func nodeToBuffer(mcNode MCNode) (value []byte) {
	// Create value byte buffer
	value = make([]byte, nodeValueLength)

	// IsEnd in one byte
	if mcNode.IsEnd {
		value[isEndOffset] = 1
	}

	// Actions index address in file in 8 bytes
	binary.LittleEndian.PutUint64(value[actionsOffset:], mcNode.ActionsAddress)

	return
}

// bufferToNode - Converts a byte buffer to a Node
func bufferToNode(key []byte, value []byte, playerTrue, playerFalse string) (mcNode MCNode) {
	// State High in 8 bytes
	stateCodeHigh := binary.LittleEndian.Uint64(key[stateHighOffset:])

	// State Low in 8 bytes
	stateCodeLow := binary.LittleEndian.Uint64(key[stateLowOffset:])

	state := stateCodesToState(stateCodeHigh, stateCodeLow)

	// Player in one byte
	var player string
	if key[playerOffset] == 1 {
		player = playerTrue
	} else {
		player = playerFalse
	}

	// IsEnd in one byte
	isEnd := value[isEndOffset] == 1

	// Actions (file pointer to) in 8 bytes
	actions := binary.LittleEndian.Uint64(value[actionsOffset:])

	mcNode = MCNode{
		Assigned:       true,
		IsEnd:          isEnd,
		State:          state,
		Player:         player,
		ActionsAddress: actions,
	}

	return
}

// actionToBuffer - Converts an Action to a byte buffer
func actionToBuffer(action Action, buf []byte) {
	// Create byte buffer
	//buf := make([]byte, actionLength)

	// Convert all node data to bytes in byte buffer
	// Visits in 8 bytes
	binary.LittleEndian.PutUint64(buf[visitsOffset:], action.Visits)

	// Points in 8 bytes
	binary.LittleEndian.PutUint64(buf[pointsOffset:], action.Points)

	// Action X in one byte
	buf[actionXOffset] = action.X

	// Action Y in one byte
	buf[actionYOffset] = action.Y

	// Action Pass in one byte
	if action.Pass {
		buf[actionPassOffset] = 1
	}

	// Resulting child node key in file in 17 bytes
	for i := uint64(0); i < uint64(nodeKeyLength); i++ {
		buf[childNodeKeyOffset+i] = action.ActionNodeKey[i]
	}
}

// bufferToAction - Converts a byte buffer to an Action
func bufferToAction(buf []byte) Action {
	// Visits in 8 bytes
	visits := binary.LittleEndian.Uint64(buf[visitsOffset:])

	// Points in 8 bytes
	points := binary.LittleEndian.Uint64(buf[pointsOffset:])

	// Action X in one byte
	actionX := buf[actionXOffset]

	// Action Y in one byte
	actionY := buf[actionYOffset]

	// Action Pass in one byte
	actionPass := buf[actionPassOffset] == 1

	// Resulting child node key in 17 bytes
	actionNodeKey := buf[childNodeKeyOffset : childNodeKeyOffset+uint64(nodeKeyLength)]

	return Action{
		Visits:        visits,
		Points:        points,
		X:             actionX,
		Y:             actionY,
		Pass:          actionPass,
		ActionNodeKey: actionNodeKey,
	}
}

// baseEncode - Encodes decimal to new base
func baseEncode(nb uint64, buf *bytes.Buffer) {
	l := uint64(len(numberBase))
	if nb/l != 0 {
		baseEncode(nb/l, buf)
	}
	buf.WriteByte(numberBase[nb%l])
}

// baseDecode - Decodes base to decimal
func baseDecode(enc string) uint64 {
	var nb uint64
	lbase := len(numberBase)
	le := len(enc)
	for i := 0; i < le; i++ {
		mult := 1
		for j := 0; j < le-i-1; j++ {
			mult *= lbase
		}
		nb += uint64(strings.IndexByte(numberBase, enc[i]) * mult)
	}
	return nb
}

// stateToStateCodes - Converts a state given as a string (max 64 digits in base 3) to state codes in two
// unsigned 64-bit integers
func stateToStateCodes(state string) (stateCodeHigh, stateCodeLow uint64) {
	diff := 64 - len(state)
	if diff > 0 {
		state = fmt.Sprintf("%0*d%s", diff, 0, state)
	}

	return baseDecode(state[0:32]), baseDecode(state[32:])
}

// stateCodesToState - Converts state codes to a state string
func stateCodesToState(stateCodeHigh, stateCodeLow uint64) string {
	buffer := bytes.Buffer{}
	baseEncode(stateCodeHigh, &buffer)
	stateHigh := buffer.String()

	buffer.Reset()
	baseEncode(stateCodeLow, &buffer)
	stateLow := buffer.String()

	diff := 32 - len(stateLow)
	if diff > 0 {
		stateLow = fmt.Sprintf("%0*d%s", diff, 0, stateLow)
	}

	var stateBuilder strings.Builder
	stateBuilder.WriteString(stateHigh)
	stateBuilder.WriteString(stateLow)
	state := stateBuilder.String()

	state = strings.TrimLeft(state, "0")
	return state
}

// nodeStateToBuffer - Converts a nodeState struct to buffer
func nodeStateToBuffer(nodeState nodeState) (buf []byte) {
	buf = make([]byte, nodeKeyLength)
	// State High in 8 bytes
	binary.LittleEndian.PutUint64(buf[stateHighOffset:], nodeState.stateCodeHigh)

	// State High in 8 bytes
	binary.LittleEndian.PutUint64(buf[stateLowOffset:], nodeState.stateCodeLow)

	// Player in one byte
	if nodeState.playerA {
		buf[playerOffset] = 1
	}

	return
}
