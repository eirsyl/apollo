package orchestrator

func getNodeWeights(nodes map[string]*Node) (map[string]float64, float64) {
	totalWeight := 0.0
	weights := map[string]float64{}

	for _, node := range nodes {
		weights[node.ID] = 1
		totalWeight += 1
	}

	return weights, totalWeight
}
