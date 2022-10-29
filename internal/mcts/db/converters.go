package db

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// numberBase - Number base to use when converting from nodeState of a board to uint64 and back
const numberBase string = "012"

// Total length of one node in file
const nodeLength int = 18

// Node offsets
const stateOffset uint64 = 0
const playerOffset uint64 = 8
const isEndOffset uint64 = 9
const actionsOffset uint64 = 10

/*
	Assigned
	State         8 bytes
	Player        1 byte
	Actions
	ActionNodeAddress
	aIndexAddress 8 bytes

*/

// Total length of one action in file
const actionLength int = 27

// Action offsets

// Offsets in relation to first byte of an Action in file
// Note: code depends on visitsOffset and pointsOffset being first and in that order.
const visitsOffset uint64 = 0 // Needs to be first
const pointsOffset uint64 = 8 // Needs to be second
const actionXOffset uint64 = 16
const actionYOffset uint64 = 17
const actionPassOffset uint64 = 18
const childNodeOffset uint64 = 19

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

// nodeToBuffer - Converts an MCNode to a byte buffer
func nodeToBuffer(mcNode MCNode, playerTrue string) []byte {
	// Create byte buffer
	buf := make([]byte, nodeLength)

	// State in 8 bytes
	binary.LittleEndian.PutUint64(buf[stateOffset:], baseDecode(mcNode.State))

	// Player in one byte
	if mcNode.Player == playerTrue {
		buf[playerOffset] = 1
	}

	// IsEnd in one byte
	if mcNode.IsEnd {
		buf[isEndOffset] = 1
	}

	// Actions index address in file in 8 bytes
	binary.LittleEndian.PutUint64(buf[actionsOffset:], mcNode.ActionsAddress)

	return buf
}

// bufferToNode - Converts a byte buffer to a Node
func bufferToNode(buf []byte, playerTrue, playerFalse string, nodeAddress uint64) MCNode {
	// State in 8 bytes
	buffer := bytes.Buffer{}
	baseEncode(binary.LittleEndian.Uint64(buf[stateOffset:]), &buffer)
	state := buffer.String()

	// Player in one byte
	var player string
	if buf[playerOffset]&1 == 1 {
		player = playerTrue
	} else {
		player = playerFalse
	}

	// IsEnd in one byte
	isEnd := buf[isEndOffset] == 1

	// Actions (file pointer to) in 8 bytes
	actions := binary.LittleEndian.Uint64(buf[actionsOffset:])

	return MCNode{
		Assigned:       true,
		IsEnd:          isEnd,
		State:          state,
		Player:         player,
		nAddress:       nodeAddress,
		ActionsAddress: actions,
	}
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

	// Resulting child node address in file in 8 bytes
	binary.LittleEndian.PutUint64(buf[childNodeOffset:], action.ActionNodeAddress)
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

	// Resulting child node address in file in 8 bytes
	nAddress := binary.LittleEndian.Uint64(buf[childNodeOffset:])

	return Action{
		Visits:            visits,
		Points:            points,
		X:                 actionX,
		Y:                 actionY,
		Pass:              actionPass,
		ActionNodeAddress: nAddress,
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
