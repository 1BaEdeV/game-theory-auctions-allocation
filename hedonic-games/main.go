// main.go
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	os.MkdirAll("results", 0755)

	teachers, err := LoadAMteachers("../ds/relations_graph.json")
	if err != nil {
		panic(err)
	}

	g, _, idToName := teachers.ToGraph()
	results := make([]ExperimentResult, 0)

	alphaValues := []float64{0.1, 0.3, 0.5, 0.7, 0.9}

	for _, alpha := range alphaValues {
		start := time.Now()

		hg := NewHedonicGameWithTargetK(*g, alpha, 21)
		partition := hg.FindNashStablePartition_WithPotential(1000, false)
		potential := hg.ComputePotential_Formula71()
		modularity := ComputeModularity(g, partition)

		elapsed := time.Since(start).Seconds()

		result := NewExperimentResult(
			"Hedonic_Game",
			"Hedonic",
			alpha,
			g,
			partition,
			potential,
			modularity,
			hg.Iterations,
			hg.Iterations,
			elapsed,
		)
		results = append(results, result)

		filename := fmt.Sprintf("results/partition_hedonic_alpha_%.1f.json", alpha)
		if err := ExportPartitionToJSON(g, partition, idToName, filename); err != nil {
			fmt.Printf("export error: %v\n", err)
		}
	}

	alphaMLValues := []float64{0.2, 0.5, 0.8}
	betaValues := []float64{0.1, 0.5, 1.0}

	for _, alpha := range alphaMLValues {
		for _, beta := range betaValues {
			start := time.Now()

			ml := NewMLModel(g, alpha, beta)
			partition := initializeRandomPartition(g, 5)

			for iter := 0; iter < 50; iter++ {
				for _, node := range g.GetNodeList() {
					partition[node] = selectNewCommunityImproved(ml, node, partition)
				}
			}

			objective := ml.ComputeObjectiveFunction(partition)
			modularity := ComputeModularity(g, partition)
			elapsed := time.Since(start).Seconds()

			result := NewExperimentResult(
				"Maximum_Likelihood",
				fmt.Sprintf("ML_beta_%.1f", beta),
				alpha,
				g,
				partition,
				objective,
				modularity,
				50,
				50,
				elapsed,
			)
			results = append(results, result)

			filename := fmt.Sprintf("results/partition_ml_alpha_%.1f_beta_%.1f.json", alpha, beta)
			if err := ExportPartitionToJSON(g, partition, idToName, filename); err != nil {
				fmt.Printf("export error: %v\n", err)
			}
		}
	}

	if err := SaveResultsToCSV(results, "results/experiments.csv"); err != nil {
		fmt.Printf("CSV error: %v\n", err)
	}
}
