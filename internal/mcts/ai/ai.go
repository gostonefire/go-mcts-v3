package ai

import (
	"container/list"
	"fmt"
	"os"
)

// AI - Represents a structure for managing with statistics and models in the AI domain
type AI struct {
	AIFile             *os.File
	AIOverflowFile     *os.File
	highValueThreshold float64
	lowValueThreshold  float64
	visitsThreshold    uint64
	stateLength        int
	maxListLengths     int
	highList           *list.List
	lowList            *list.List
	bulkSize           int
}

// item - Represents an item in one of the buffer lists
type item struct {
	player string
	state  string
	value  float64
}

// NewAI - Creates a new AI instance
func NewAI(
	dbName string,
	highValueThreshold,
	lowValueThreshold float64,
	visitsThreshold uint64,
	stateLength,
	maxListLengths,
	bulkSize int,
	newDB bool,
) (
	ai *AI,
	err error,
) {

	dFile := fmt.Sprintf("%s-aiDB.txt", dbName)
	oFile := fmt.Sprintf("%s-aiDBOverflow.txt", dbName)

	// Clear old db file(s) if indicated
	if newDB {
		if err = removeExistingFiles([]string{dFile, oFile}); err != nil {
			fmt.Println("Error while trying to remove existing node files")
			return
		}
	}

	// Open or create the AI DB files
	df, err := os.OpenFile(dFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error while open or create %s, %s\n", dFile, err)
		return
	}
	of, err := os.OpenFile(oFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error while open or create %s, %s\n", dFile, err)
		return
	}

	ai = &AI{
		AIFile:             df,
		AIOverflowFile:     of,
		highValueThreshold: highValueThreshold,
		lowValueThreshold:  lowValueThreshold,
		visitsThreshold:    visitsThreshold,
		stateLength:        stateLength,
		maxListLengths:     maxListLengths,
		highList:           list.New(),
		lowList:            list.New(),
		bulkSize:           bulkSize / 2, // Divide by 2 since we have two lists to draw the bulk from
	}

	return
}

// RecordStateStatistics - Records statistics for a player/state pair in the AI domain
func (A *AI) RecordStateStatistics(
	player,
	state string,
	oldVisits,
	visits uint64,
	oldPoints,
	points float64,
) (
	err error,
) {

	if visits < A.visitsThreshold {
		return
	}

	var oldValue, value float64
	value = points / float64(visits)
	if oldVisits > 0 {
		oldValue = oldPoints / float64(oldVisits)
	}

	if oldVisits < A.visitsThreshold {
		if value >= A.highValueThreshold {
			err = A.manageStateStatistics(player, state, value, true)
		} else if value <= A.lowValueThreshold {
			err = A.manageStateStatistics(player, state, value, false)
		}
	} else {
		if oldValue < A.highValueThreshold && value >= A.highValueThreshold {
			err = A.manageStateStatistics(player, state, value, true)
		} else if oldValue > A.lowValueThreshold && value <= A.lowValueThreshold {
			err = A.manageStateStatistics(player, state, value, false)
		}
	}

	return
}

// manageStateStatistics - Manages the fifo queues holding statistics until it is time to send away a bulk
func (A *AI) manageStateStatistics(player, state string, value float64, highValue bool) (err error) {
	// Ensure the state length is kept, important when pushing state values to the AI model
	diff := A.stateLength - len(state)
	if diff > 0 {
		state = fmt.Sprintf("%0*d%s", diff, 0, state)
	}

	statItem := item{player: player, state: state, value: value}

	// Update the buffers with the new item and ensure buffers are kept within size limits
	if highValue {
		if A.highList.Len() >= A.maxListLengths {
			listItem := A.highList.Back()
			err = A.writeOverflow(listItem.Value.(item), true)
			if err != nil {
				return
			}
			A.highList.Remove(listItem)
		}

		A.highList.PushFront(statItem)
	} else {
		if A.lowList.Len() >= A.maxListLengths {
			listItem := A.lowList.Back()
			err = A.writeOverflow(listItem.Value.(item), false)
			if err != nil {
				return
			}
			A.lowList.Remove(listItem)
		}

		A.lowList.PushFront(statItem)
	}

	// Time to send a bulk of items out.
	if A.highList.Len() >= A.bulkSize && A.lowList.Len() >= A.bulkSize {
		err = A.writeBulk(A.bulkSize)
		if err != nil {
			return
		}
	}

	return
}

// WriteAndCloseBuffers - Supposed to be run before closing down application and ensures that whatever is still left
// in buffers gets written to either AI DB or overflow DB
func (A *AI) WriteAndCloseBuffers() (err error) {
	lenHighList := A.highList.Len()
	lenLowList := A.lowList.Len()

	if lenHighList <= lenLowList {
		// High list contains fewer elements, write out according high list
		err = A.writeBulk(lenHighList)
		if err != nil {
			return
		}

		// Empty low list to overflow
		for e := A.lowList.Front(); e != nil; e = e.Next() {
			err = A.writeOverflow(e.Value.(item), false)
			if err != nil {
				return
			}
		}
	} else {
		// Low list contains fewer elements, write out according low list
		err = A.writeBulk(lenLowList)
		if err != nil {
			return
		}

		// Empty high list to overflow
		for e := A.highList.Front(); e != nil; e = e.Next() {
			err = A.writeOverflow(e.Value.(item), true)
			if err != nil {
				return
			}
		}
	}

	return
}

// writeBulk - Writes a chunk of data of the size of a bulk to whatever means for learning is decided.
// Currently, it is implemented to just go to file.
func (A *AI) writeBulk(bulkSize int) (err error) {
	var statItem item
	var listItem *list.Element
	for i := 0; i < bulkSize; i++ {
		listItem = A.highList.Back()
		statItem = listItem.Value.(item)
		_, err = fmt.Fprintf(A.AIFile, "%s,%s,%.2f,1\n", statItem.player, statItem.state, statItem.value)
		if err != nil {
			fmt.Printf("Error while writing AI bulk to file: %s", err)
			return
		}
		A.highList.Remove(listItem)

		listItem = A.lowList.Back()
		statItem = listItem.Value.(item)
		_, err = fmt.Fprintf(A.AIFile, "%s,%s,%.2f,0\n", statItem.player, statItem.state, statItem.value)
		if err != nil {
			fmt.Printf("Error while writing AI bulk to file: %s", err)
			return
		}
		A.lowList.Remove(listItem)
	}

	return
}

// writeOverflow - Writes overflow data to file
func (A *AI) writeOverflow(statItem item, highValue bool) (err error) {
	var highLabel int
	if highValue {
		highLabel = 1
	}
	_, err = fmt.Fprintf(A.AIOverflowFile, "%s,%s,%.2f,%d\n", statItem.player, statItem.state, statItem.value, highLabel)
	if err != nil {
		fmt.Printf("Error while writing AI overflow to file: %s", err)
		return
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
