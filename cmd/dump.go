package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// numberBase - Number base to use when converting from nodeState of a board to uint64 and back
const numberBase string = "012"

// Total length of one node in file
const nodeLength int = 17

// Node offsets
const stateOffset uint64 = 0
const playerOffset uint64 = 8
const actionsOffset uint64 = 9

/*
	Assigned
	State         8 bytes
	Player        1 byte
	Actions
	ActionNodeAddress
	aIndexAddress 8 bytes

*/

// Total length of one action in file
const actionLength int = 28

// Action offsets

// Offsets in relation to first byte of an Action in file
// Note: code depends on visitsOffset and pointsOffset being first and in that order.
const visitsOffset uint64 = 0 // Needs to be first
const pointsOffset uint64 = 8 // Needs to be second
const actionXOffset uint64 = 16
const actionYOffset uint64 = 17
const actionPassOffset uint64 = 18
const isDoneOffset uint64 = 19
const childNodeOffset uint64 = 20

func openFiles(nodeTreeName string) (nodeFile, actionsFile *os.File, err error) {
	nFile := fmt.Sprintf("%s-nodes.bin", nodeTreeName)
	aFile := fmt.Sprintf("%s-actions.bin", nodeTreeName)

	// Open or create the node files
	nodeFile, err = os.OpenFile(nFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Error while open %s, %s\n", nFile, err)
		return
	}
	actionsFile, err = os.OpenFile(aFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Error while open %s, %s\n", aFile, err)
		return
	}

	return
}

// baseEncode - Encodes decimal to new base
func baseEncode(nb uint64, buf *bytes.Buffer) {
	l := uint64(len(numberBase))
	if nb/l != 0 {
		baseEncode(nb/l, buf)
	}
	buf.WriteByte(numberBase[nb%l])
}

// bufferToNodeString - Converts a byte buffer to a Node string
func bufferToNodeString(buf []byte, playerTrue, playerFalse string) (node string, actionsAddress uint64) {
	// State in 8 bytes
	buffer := bytes.Buffer{}
	baseEncode(binary.LittleEndian.Uint64(buf[stateOffset:]), &buffer)
	state := buffer.String()
	diff := 9 - len(state)
	if diff > 0 {
		state = fmt.Sprintf("%0*d%s", diff, 0, state)
	}

	// Player in one byte
	var player string
	if buf[playerOffset] == 1 {
		player = playerTrue
	} else {
		player = playerFalse
	}

	// Actions (file pointer to) in 8 bytes
	actionsAddress = binary.LittleEndian.Uint64(buf[actionsOffset:])

	var aString string
	if actionsAddress == math.MaxUint64 {
		aString = "*"
	} else {
		aString = fmt.Sprintf("%d", actionsAddress)
	}

	node = fmt.Sprintf("%s|%s|%s", state, player, aString)
	return
}

// bufferToActionsString - Converts a buffer to an actions string
func bufferToActionsString(buf []byte) (action string) {
	// Visits in 8 bytes
	visits := binary.LittleEndian.Uint64(buf[visitsOffset:])

	// Points in 8 bytes
	points := float64(binary.LittleEndian.Uint64(buf[pointsOffset:])) / 2

	// Action X in one byte
	actionX := buf[actionXOffset]

	// Action Y in one byte
	actionY := buf[actionYOffset]

	// Action Pass in one byte
	// actionPass := buf[actionPassOffset] == 1

	// IsDone in one byte
	isDone := buf[isDoneOffset] == 1

	// Resulting child node address in file in 8 bytes
	nAddress := binary.LittleEndian.Uint64(buf[childNodeOffset:])

	action = fmt.Sprintf("|%d|%0.1f|%d|%d|%v| -> |%d|", visits, points, actionX, actionY, isDone, nAddress)

	return
}

func actionsAndNodeAddresses(filePtr *os.File, actionIndexAddress uint64) (string, error) {
	if actionIndexAddress != math.MaxUint64 {
		_, err := filePtr.Seek(int64(actionIndexAddress), io.SeekStart)
		if err != nil {
			fmt.Printf("Error while seeking for actions record: %s\n", err)
			return "", err
		}
		buf := make([]byte, 1)
		_, err = filePtr.Read(buf)
		if err != nil {
			fmt.Printf("Error while reading number of actions in record: %s\n", err)
			return "", err
		}
		nRelatives := int(buf[0])
		buf = make([]byte, nRelatives*actionLength)
		_, err = filePtr.Read(buf)
		if err != nil {
			fmt.Printf("Error while reading actions record: %s\n", err)
			return "", err
		}
		actions := make([]string, nRelatives)
		for i := 0; i < nRelatives; i++ {
			action := bufferToActionsString(buf[i*actionLength:])
			actions[i] = fmt.Sprintf("(%s)", action)
		}
		return strings.Join(actions, ","), nil
	}

	return "*", nil
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Must submit tree name")
		return
	}

	nf, af, err := openFiles(args[0])
	if err != nil {
		return
	}
	defer func(af *os.File) { err = af.Close() }(af)
	defer func(nf *os.File) { err = nf.Close() }(nf)

	var node, actionsString string
	var actionsAddress uint64
	a := 0
	buf := make([]byte, nodeLength)
	for {
		if _, err = nf.Read(buf); err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				fmt.Printf("Error while reading from node file: %s\n", err)
				return
			}
		}
		node, actionsAddress = bufferToNodeString(buf, "X", "Y")
		actionsString, err = actionsAndNodeAddresses(af, actionsAddress)
		if err != nil {
			return
		}
		fmt.Printf("%d|%s| -> %s\n", a, node, actionsString)
		a += nodeLength
	}

	fmt.Println("\nActions:")

	a = 0
	_, err = af.Seek(0, io.SeekStart)
	if err != nil {
		fmt.Printf("Error while seeking start in actionsString file: %s", err)
		return
	}
	for {
		buf = make([]byte, 1)
		if _, err = af.Read(buf); err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				fmt.Printf("Error while reading from actionsString file: %s\n", err)
				return
			}
		}
		nActions := int(buf[0])
		buf = make([]byte, nActions*actionLength)
		if _, err = af.Read(buf); err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				fmt.Printf("Error while reading from actionsString file: %s\n", err)
				return
			}
		}

		actions := make([]string, nActions)
		for i := 0; i < nActions; i++ {

			actions[i] = fmt.Sprintf("(%s)", bufferToActionsString(buf[i*actionLength:]))
		}
		fmt.Printf("%d -> %s\n", a, strings.Join(actions, ", "))

		a += 1 + nActions*actionLength
	}

}
