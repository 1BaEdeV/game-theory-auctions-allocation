package main

import (
	"fmt"
	"time"
)

func main() {
	teachers, err := LoadAMteachers("../ds/relations_graph.json")
	if err != nil {
		panic(err)
	}
	G, _, _ := teachers.ToGraph()
	fmt.Printf("Graph loaded: %d nodes, %d edges\n", G.NumNodes(), G.NumEdges())

	var allHGResults []Result
	var allMLResults []Result

	// ============ HG: перебор разных γ ============
	gammaValues := []float64{0.5, 1.0, 1.5, 2.0, 2.5}

	fmt.Println("\n🎮 HEDONIC GAMES: Testing different gamma values...")
	for i, gamma := range gammaValues {
		start := time.Now()

		game := NewHedonicGame(G, gamma)
		partition := game.BetterResponseDynamics()

		elapsed := time.Since(start).Seconds()

		mod := Modularity(G, partition)
		sil := SilhouetteCoefficient(G, partition)
		numComm := NumCommunities(partition)

		result := Result{
			Graph:       "Teachers",
			Nodes:       G.NumNodes(),
			Edges:       G.NumEdges(),
			Gamma:       gamma,
			Iterations:  0,
			Modularity:  mod,
			Silhouette:  sil,
			Communities: numComm,
			Duration:    fmt.Sprintf("%.4f", elapsed),
		}

		allHGResults = append(allHGResults, result)

		// JSON для каждого γ
		jsonFile := fmt.Sprintf("results/graph_hg_gamma_%.1f.json", gamma)
		if err := ExportJSON(G, partition, mod, sil, jsonFile); err != nil {
			panic(err)
		}

		PrintCommunities(partition)
		fmt.Printf("[%d/%d] HG γ=%.1f: Communities=%d, Modularity=%.4f, Silhouette=%.4f, Time=%.4f\n",
			i+1, len(gammaValues), gamma, numComm, mod, sil, elapsed)
	}

	// ============ ML: перебор разных α ============
	alphaValues := []float64{0.02, 0.1, 0.2, 0.5, 1.0}

	fmt.Println("\n🧬 MAXIMUM LIKELIHOOD: Testing different alpha values...")
	for i, alpha := range alphaValues {
		start := time.Now()

		mlResult := MaximumLikelihoodImproved(
			G,
			[]float64{alpha},
			100,
			[]float64{5.0},
			50,
			3,
			50,
		)

		elapsed := time.Since(start).Seconds()

		mod := Modularity(G, mlResult.BestPartition)
		sil := SilhouetteCoefficient(G, mlResult.BestPartition)

		result := Result{
			Graph:       "Teachers",
			Nodes:       G.NumNodes(),
			Edges:       G.NumEdges(),
			Gamma:       alpha,
			Iterations:  mlResult.TotalIterations,
			Modularity:  mod,
			Silhouette:  sil,
			Communities: mlResult.NumCommunities,
			Duration:    fmt.Sprintf("%.4f", elapsed),
		}

		allMLResults = append(allMLResults, result)

		// JSON для каждого α
		jsonFile := fmt.Sprintf("results/graph_ml_alpha_%.2f.json", alpha)
		if err := ExportJSON(G, mlResult.BestPartition, mod, sil, jsonFile); err != nil {
			panic(err)
		}

		PrintCommunities(mlResult.BestPartition)
		fmt.Printf("[%d/%d] ML α=%.2f: Communities=%d, Modularity=%.4f, Silhouette=%.4f, Time=%.4f\n",
			i+1, len(alphaValues), alpha, mlResult.NumCommunities, mod, sil, elapsed)
	}

	// ============ СОХРАНЕНИЕ CSV (APPEND) ============
	fmt.Println("\n💾 Saving results...")

	// Если файлы существуют, добавляем в конец, иначе создаём с заголовком
	if err := AppendResultsCSV(allHGResults, "results/hedonic_experiments.csv"); err != nil {
		panic(err)
	}
	if err := AppendResultsCSV(allMLResults, "results/ml_experiments.csv"); err != nil {
		panic(err)
	}

	// итоговый CSV со ВСЕМИ результатами
	allResults := append(allHGResults, allMLResults...)
	if err := AppendResultsCSV(allResults, "results/all_experiments.csv"); err != nil {
		panic(err)
	}

	fmt.Println("\n✅ All results saved:")
	fmt.Println("   - results/graph_hg_gamma_*.json (5 файлов)")
	fmt.Println("   - results/graph_ml_alpha_*.json (5 файлов)")
	fmt.Println("   - results/hedonic_experiments.csv")
	fmt.Println("   - results/ml_experiments.csv")
	fmt.Println("   - results/all_experiments.csv")
}

// ===== ВСПОМОГАТЕЛЬНОЕ =====

func NewHedonicGame(g *Graph, gamma float64) *HedonicGameState {
	state := &HedonicGameState{
		G:          g,
		Gamma:      gamma,
		Partition:  initializeRandomPartition(g, 5),
		Iterations: 0,
	}
	return state
}

type HedonicGameState struct {
	G          *Graph
	Gamma      float64
	Partition  map[int]int
	Iterations int
}

func (hg *HedonicGameState) BetterResponseDynamics() map[int]int {
	maxIter := 10000
	for hg.Iterations = 0; hg.Iterations < maxIter; hg.Iterations++ {
		changed := false
		for _, node := range hg.G.GetNodeList() {
			oldComm := hg.Partition[node]
			bestComm := hg.GetBestCommunity(node)
			if oldComm != bestComm {
				hg.Partition[node] = bestComm
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return hg.Partition
}

func (hg *HedonicGameState) GetBestCommunity(node int) int {
	bestComm := hg.Partition[node]
	bestUtil := hg.ComputeUtility(node, bestComm)

	for neighbor := range hg.G.Edges[node] {
		comm := hg.Partition[neighbor]
		util := hg.ComputeUtility(node, comm)
		if util > bestUtil {
			bestUtil = util
			bestComm = comm
		}
	}

	newComm := 0
	for _, comm := range hg.Partition {
		if comm >= newComm {
			newComm = comm + 1
		}
	}
	util := hg.ComputeUtility(node, newComm)
	if util > bestUtil {
		bestComm = newComm
	}

	return bestComm
}

func (hg *HedonicGameState) ComputeUtility(node, community int) float64 {
	friends := 0.0
	strangers := 0.0

	for neighbor := range hg.G.Edges[node] {
		if hg.Partition[neighbor] == community {
			friends += 1.0
		} else {
			strangers += 1.0
		}
	}

	return friends - hg.Gamma*strangers
}

func NumCommunities(partition map[int]int) int {
	comms := make(map[int]bool)
	for _, comm := range partition {
		comms[comm] = true
	}
	return len(comms)
}
