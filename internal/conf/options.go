package conf

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetLearnOptions - Gets input from the executor
func GetLearnOptions() (gameId int, size uint8, maxRounds float64, uniqueStates int64, forceNew bool, name string, err error) {
	var input string
	var s, m, u int
	size = 4
	maxRounds = 1000000
	uniqueStates = 100000000

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Game [0 - TicTacToe, 1 - Othello, 2 - V Four in a Row]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.TrimSpace(input); input != "" {
		s, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("Error, malformed number give: %s\n", err)
			return
		}
		gameId = s
	}

	if gameId != 2 {
		fmt.Print("Size [4]: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error while reading input from console: %s\n", err)
			return
		}
		if input = strings.TrimSpace(input); input != "" {
			s, err = strconv.Atoi(strings.TrimSpace(input))
			if err != nil {
				fmt.Printf("Error, malformed number give: %s\n", err)
				return
			}
			size = uint8(s)
		}
	}

	fmt.Print("Max learning rounds [1000000]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.TrimSpace(input); input != "" {
		m, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("Error, malformed number give: %s\n", err)
			return
		}
		maxRounds = float64(m)
	}

	fmt.Print("Estimated unique states [100000000]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.TrimSpace(input); input != "" {
		u, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("Error, malformed number give: %s\n", err)
			return
		}
		uniqueStates = int64(u)
	}

	fmt.Print("Force new tree [false]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.ToUpper(strings.TrimSpace(input)); input == "TRUE" {
		forceNew = true
	}

	name = fmt.Sprintf("nodetree%dx%d-%d", size, size, gameId)

	return
}

// GetPlayOptions - Gets input from the executor
func GetPlayOptions() (gameId int, size uint8, name string, err error) {
	var input string
	var s int
	size = 4

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Game [0 - TicTacToe, 1 - Othello, 2 - V Four in a Row]: ")
	input, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error while reading input from console: %s\n", err)
		return
	}
	if input = strings.TrimSpace(input); input != "" {
		s, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("Error, malformed number give: %s\n", err)
			return
		}
		gameId = s
	}

	if gameId != 2 {
		fmt.Print("Size [4]: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error while reading input from console: %s\n", err)
			return
		}
		if input = strings.TrimSpace(input); input != "" {
			s, err = strconv.Atoi(strings.TrimSpace(input))
			if err != nil {
				fmt.Printf("Error, malformed number given: %s\n", err)
				return
			}
			size = uint8(s)
		}
	}

	name = fmt.Sprintf("nodetree%dx%d-%d", size, size, gameId)

	return
}
