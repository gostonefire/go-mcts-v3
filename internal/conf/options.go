package conf

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetLearnOptions - Gets input from the executor
func GetLearnOptions() (gameId int, size uint8, maxRounds float64, forceNew bool, name string, err error) {
	var input string
	var s, m int
	size = 4
	maxRounds = 1000000

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Game [0 - TicTacToe, 1 - Othello]: ")
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

	fmt.Print("Game [0 - TicTacToe, 1 - Othello]: ")
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

	name = fmt.Sprintf("nodetree%dx%d-%d", size, size, gameId)

	return
}
