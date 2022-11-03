package ai

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// AI - Represents a structure for managing with statistics and models in the AI domain
type AI struct {
	AIFileName         string
	highValueThreshold float32
	lowValueThreshold  float32
	visitsThreshold    uint64
	stateLength        int
	ofInterest         map[string]bool
}

// NewAI - Creates a new AI instance
func NewAI(
	dbName string,
	highValueThreshold,
	lowValueThreshold float32,
	visitsThreshold uint64,
	stateLength int,
	newDB bool,
) (
	ai *AI,
	err error,
) {
	dFilePrefix := fmt.Sprintf("%s-aiDB-", dbName)

	// Get existing db files and record new file index
	var dirEntries []os.DirEntry
	dirEntries, err = os.ReadDir(".")
	if err != nil {
		fmt.Printf("Error listing directory content, %s\n", err)
		return
	}

	dbFiles := make([]string, 0)
	lenPrefix := len(dFilePrefix)
	maxIdx := -1
	for _, e := range dirEntries {
		if strings.HasPrefix(e.Name(), dFilePrefix) {
			dbFiles = append(dbFiles, e.Name())
			idx, _ := strconv.Atoi(e.Name()[lenPrefix : lenPrefix+1])
			if idx > maxIdx {
				maxIdx = idx
			}
		}
	}
	newIdx := maxIdx + 1

	// Clear old db file(s) if indicated
	if newDB {
		if err = removeExistingFiles(dbFiles); err != nil {
			fmt.Println("Error while trying to remove existing AI DB files")
			return
		}
		newIdx = 0
	}

	// New file to fill
	dFile := fmt.Sprintf("%s%d.txt", dFilePrefix, newIdx)

	// Test to open or create the new AI file and then close immediately
	df, err := os.OpenFile(dFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error while open or create %s, %s\n", dFile, err)
		return
	}
	err = df.Close()
	if err != nil {
		fmt.Printf("Error while closing %s, %s\n", dFile, err)
		return
	}

	ai = &AI{
		AIFileName:         dFile,
		highValueThreshold: highValueThreshold,
		lowValueThreshold:  lowValueThreshold,
		visitsThreshold:    visitsThreshold,
		stateLength:        stateLength,
		ofInterest:         make(map[string]bool),
	}

	return
}

// RecordStateStatistics - Records statistics for a player/state pair in the AI domain
func (A *AI) RecordStateStatistics(player, state string, visits uint64, points float64) {

	if visits < A.visitsThreshold {
		return
	}

	state = fmt.Sprintf("%s%s", state, player)

	value := float32(points / float64(visits))

	if value >= A.highValueThreshold {
		A.ofInterest[state] = true
	} else if value <= A.lowValueThreshold {
		A.ofInterest[state] = false
	} else {
		delete(A.ofInterest, state)
	}
}

// WriteAndCloseBuffers - Supposed to be run before closing down application and ensures that whatever is still left
// in buffers gets written to either AI DB or overflow DB
func (A *AI) WriteAndCloseBuffers() (err error) {
	// Open AI file
	df, err := os.OpenFile(A.AIFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error while open or create %s, %s\n", A.AIFileName, err)
		return
	}
	defer func(df *os.File) { _ = df.Close() }(df)

	var kLen int
	var player, state string
	for k, v := range A.ofInterest {
		kLen = len(k)
		player = k[kLen-1:]
		state = k[:kLen-1]

		diff := A.stateLength - len(state)
		if diff > 0 {
			state = fmt.Sprintf("%0*d%s", diff, 0, state)
		}

		if v {
			_, err = fmt.Fprintf(df, "%s,%s,1\n", player, state)
		} else {
			_, err = fmt.Fprintf(df, "%s,%s,0\n", player, state)
		}
		if err != nil {
			fmt.Printf("Error while writing state values to file, %s\n", err)
		}
	}
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
