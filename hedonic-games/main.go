// main.go - для Zachary Karate Club
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	os.MkdirAll("results", 0755)

	// Загружаем Karate Club (34 узла, 78 рёбер)
	g := LoadKarateClub()
	idToName := generateNodeNames(g.NumNodes())

	results := make([]ExperimentResult, 0)

	// ======== ЭКСПЕРИМЕНТ 1: Гедонические игры с разными альфа ========
	alphaValues := []float64{0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.09, 0.1, 0.3, 0.5, 0.7, 0.9}

	for _, alpha := range alphaValues {
		start := time.Now()

		hg := NewHedonicGame(*g, alpha)
		partition := hg.FindNashStablePartition_WithPotential(1000, false)
		potential := hg.ComputePotential_Formula71()
		modularity := ComputeModularity(g, partition)

		elapsed := time.Since(start).Seconds()

		result := NewExperimentResult(
			"Hedonic_NoConstraint",
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

		filename := fmt.Sprintf("results/karate_hedonic_alpha_%.1f.json", alpha)
		if err := ExportPartitionToJSON(g, partition, idToName, filename); err != nil {
			fmt.Printf("export error: %v\n", err)
		}
	}

	// ======== ЭКСПЕРИМЕНТ 2: Гедонические игры с фиксированным K ========
	targetKValues := []int{2, 3, 4, 5, 6}

	for _, targetK := range targetKValues {
		start := time.Now()

		hg := NewHedonicGameWithTargetK(*g, 0.3, targetK) // используем alpha=0.3
		partition := hg.FindNashStablePartition_WithPotential(1000, false)
		potential := hg.ComputePotential_Formula71()
		modularity := ComputeModularity(g, partition)

		actualK := hg.GetNumberOfCommunities()
		elapsed := time.Since(start).Seconds()

		result := NewExperimentResult(
			fmt.Sprintf("Hedonic_K%d", targetK),
			"Hedonic_FixedK",
			float64(targetK),
			g,
			partition,
			potential,
			modularity,
			hg.Iterations,
			hg.Iterations,
			elapsed,
		)
		results = append(results, result)

		filename := fmt.Sprintf("results/karate_hedonic_k%d_actual%d.json", targetK, actualK)
		if err := ExportPartitionToJSON(g, partition, idToName, filename); err != nil {
			fmt.Printf("export error: %v\n", err)
		}
	}

	// ======== ЭКСПЕРИМЕНТ 3: Maximum Likelihood с разными параметрами ========
	alphaMLValues := []float64{0.2, 0.5, 0.8}
	betaValues := []float64{0.1, 0.5, 1.0}

	for _, alpha := range alphaMLValues {
		for _, beta := range betaValues {
			start := time.Now()

			ml := NewMLModel(g, alpha, beta)
			partition := initializeRandomPartition(g, 4)

			for iter := 0; iter < 100; iter++ {
				for _, node := range g.GetNodeList() {
					partition[node] = selectNewCommunityImproved(ml, node, partition)
				}
			}

			objective := ml.ComputeObjectiveFunction(partition)
			modularity := ComputeModularity(g, partition)
			elapsed := time.Since(start).Seconds()

			result := NewExperimentResult(
				"ML_NoConstraint",
				fmt.Sprintf("ML_beta_%.1f", beta),
				alpha,
				g,
				partition,
				objective,
				modularity,
				100,
				100,
				elapsed,
			)
			results = append(results, result)

			filename := fmt.Sprintf("results/karate_ml_alpha_%.1f_beta_%.1f.json", alpha, beta)
			if err := ExportPartitionToJSON(g, partition, idToName, filename); err != nil {
				fmt.Printf("export error: %v\n", err)
			}
		}
	}

	// ======== ЭКСПЕРИМЕНТ 4: ML с фиксированным K ========
	for _, targetK := range targetKValues {
		start := time.Now()

		ml := NewMLModelWithTargetK(g, 0.5, 1.0, targetK)
		partition := initializeRandomPartition(g, targetK)

		for iter := 0; iter < 100; iter++ {
			for _, node := range g.GetNodeList() {
				partition[node] = selectNewCommunityImprovedWithTarget(ml, node, partition)
			}
		}

		objective := ml.ComputeObjectiveFunction(partition)
		modularity := ComputeModularity(g, partition)
		actualK := NumCommunities(partition)
		elapsed := time.Since(start).Seconds()

		result := NewExperimentResult(
			fmt.Sprintf("ML_K%d", targetK),
			"ML_FixedK",
			float64(targetK),
			g,
			partition,
			objective,
			modularity,
			100,
			100,
			elapsed,
		)
		results = append(results, result)

		filename := fmt.Sprintf("results/karate_ml_k%d_actual%d.json", targetK, actualK)
		if err := ExportPartitionToJSON(g, partition, idToName, filename); err != nil {
			fmt.Printf("export error: %v\n", err)
		}
	}

	// Сохраняем все результаты
	if err := SaveResultsToCSV(results, "results/karate_experiments.csv"); err != nil {
		fmt.Printf("CSV error: %v\n", err)
	}
}

// generateNodeNames создаёт имена для узлов Karate Club (0-33)
func generateNodeNames(numNodes int) map[int]string {
	idToName := make(map[int]string)
	for i := 0; i < numNodes; i++ {
		idToName[i] = fmt.Sprintf("Node_%d", i)
	}
	return idToName
}
