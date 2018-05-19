package orchestrator

func getNodeWeights(nodes map[string]*Node) (map[string]float64, float64) {
	totalWeight := 0.0
	weights := map[string]float64{}

	for _, node := range nodes {
		weights[node.ID] = 1
		totalWeight++
	}

	return weights, totalWeight
}

func getReshardTable(source *Node, numSlots int) ([]int, error) {
	// Pick slots from source with most slots
	slots, err := source.MySelf.allSlots()
	if err != nil {
		return nil, err
	}

	if len(slots) < numSlots {
		numSlots = len(slots)
	}

	return slots[0:numSlots], nil
}
