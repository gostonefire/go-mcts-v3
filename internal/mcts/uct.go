package mcts

import (
	"fmt"
	"math"
)

// uct - Returns the upper confidence bound for trees given a node
func uct(parentVisits, nodeVisits, nodePoints uint64) (float64, error) {
	if nodeVisits == 0 {
		return 0, fmt.Errorf("node has no Visits, would result in division by zero")
	}

	w := float64(nodePoints) / 2 // Half since we are using integers for points and thus gives 1 instead of 0.5 for draw
	n := float64(nodeVisits)
	N := float64(parentVisits)

	// uctValue := nodePoints/nodeVisits + math.Sqrt2*math.Sqrt(math.Log(parentVisits)/nodeVisits)
	uctValue := w/n + 10*math.Sqrt(math.Log(N)/n)
	return uctValue, nil
}
