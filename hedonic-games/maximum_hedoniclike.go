// maximum_hedoniclike.go
package main

import (
	"math"
	"math/rand"
)

type MLModel struct {
	G       *Graph
	Pout    float64
	Pin     float64
	Alpha   float64
	Beta    float64
	TargetK int // желаемое число сообществ (-1 = не ограничено)
}

// NewMLModel создаёт модель без ограничения на число кластеров
func NewMLModel(g *Graph, alpha, beta float64) *MLModel {
	return &MLModel{
		G:       g,
		Alpha:   alpha,
		Beta:    beta,
		TargetK: -1,
	}
}

// NewMLModelWithTargetK создаёт модель с желаемым числом сообществ
func NewMLModelWithTargetK(g *Graph, alpha, beta float64, targetK int) *MLModel {
	return &MLModel{
		G:       g,
		Alpha:   alpha,
		Beta:    beta,
		TargetK: targetK,
	}
}

func (ml *MLModel) ComputeLikelihood(partition map[int]int) float64 {
	ml.computeOptimalProbs(partition)

	if ml.Pin <= 0 || ml.Pin >= 1 || ml.Pout <= 0 || ml.Pout >= 1 {
		return math.Inf(-1)
	}

	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	mk := make(map[int]int)
	for comm, nodes := range comms {
		count := 0
		for i, u := range nodes {
			for _, v := range nodes[i+1:] {
				if ml.G.HasEdge(u, v) {
					count++
				}
			}
		}
		mk[comm] = count
	}

	ll := 0.0

	for _, m := range mk {
		if m > 0 {
			ll += float64(m) * math.Log(ml.Pin)
		}
	}

	for comm, nodes := range comms {
		maxEdges := len(nodes) * (len(nodes) - 1) / 2
		missingEdges := maxEdges - mk[comm]

		if missingEdges > 0 {
			ll += float64(missingEdges) * math.Log(1-ml.Pin)
		}
	}

	totalM := ml.G.NumEdges()
	sumMk := 0

	for _, m := range mk {
		sumMk += m
	}

	edgesBetween := totalM - sumMk

	if edgesBetween > 0 {
		ll += float64(edgesBetween) * math.Log(ml.Pout)
	}

	n := float64(ml.G.NumNodes())
	sumNk2 := 0.0

	for _, nodes := range comms {
		nk := float64(len(nodes))
		sumNk2 += nk * nk
	}

	maxEdgesBetween := int(n*n/2 - sumNk2/2)
	missingEdgesBetween := maxEdgesBetween - edgesBetween

	if missingEdgesBetween > 0 {
		ll += float64(missingEdgesBetween) * math.Log(1-ml.Pout)
	}

	return ll
}

func (ml *MLModel) ComputeObjectiveFunction(partition map[int]int) float64 {
	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	var totalInnerEdges int
	var sumNk2 float64

	for _, nodes := range comms {
		innerEdges := 0
		for i, u := range nodes {
			for _, v := range nodes[i+1:] {
				if ml.G.HasEdge(u, v) {
					innerEdges++
				}
			}
		}
		totalInnerEdges += innerEdges
		nk := float64(len(nodes))
		sumNk2 += nk * nk
	}

	P := float64(totalInnerEdges) - 0.5*sumNk2*ml.Alpha
	return P
}

func (ml *MLModel) computeOptimalProbs(partition map[int]int) {
	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	var totalMk int
	var sumNk2 float64

	for _, nodes := range comms {
		mk := 0
		for i, u := range nodes {
			for _, v := range nodes[i+1:] {
				if ml.G.HasEdge(u, v) {
					mk++
				}
			}
		}
		totalMk += mk
		nk := float64(len(nodes))
		sumNk2 += nk * nk
	}

	n := float64(ml.G.NumNodes())
	m := float64(ml.G.NumEdges())

	denomPin := sumNk2 - n

	if denomPin > 0 {
		ml.Pin = 2 * float64(totalMk) / denomPin
	} else {
		ml.Pin = 0.5
	}

	denomPout := n*n - sumNk2

	if denomPout > 0 {
		ml.Pout = 2 * (m - float64(totalMk)) / denomPout
	} else {
		ml.Pout = 0.5
	}

	if ml.Pin < 0.001 {
		ml.Pin = 0.001
	}

	if ml.Pin > 0.999 {
		ml.Pin = 0.999
	}

	if ml.Pout < 0.001 {
		ml.Pout = 0.001
	}

	if ml.Pout > 0.999 {
		ml.Pout = 0.999
	}
}

type GibbsSamplingResult struct {
	BestPartition    map[int]int
	BestObjective    float64
	ObjectiveHistory []float64
	OptimalAlpha     float64
	OptimalPin       float64
	OptimalPout      float64
	NumCommunities   int
	ConvergedAt      int
	TotalIterations  int
}

func MaximumLikelihoodImproved(
	g *Graph,
	alphaValues []float64,
	numIterations int,
	betaValues []float64,
	numBetaIterations int,
	numInitializations int,
	burnInIterations int,
) *GibbsSamplingResult {
	bestPartition := make(map[int]int)
	bestObjective := math.Inf(-1)
	bestAlpha := 0.0
	bestPin := 0.5
	bestPout := 0.5
	bestConvergedAt := 0
	bestTotalIterations := 0
	objectiveHistory := make([]float64, 0)

	for _, alpha := range alphaValues {
		for init := 0; init < numInitializations; init++ {
			partition := initializeRandomPartition(g, rand.Intn(5)+3)
			ml := NewMLModel(g, alpha, 0.0)

			currentBest := math.Inf(-1)
			currentBestPartition := copyPartition(partition)
			convergedAt := 0

			for _, beta := range betaValues {
				ml.Beta = beta

				for iter := 0; iter < burnInIterations; iter++ {
					for _, node := range g.GetNodeList() {
						partition[node] = selectNewCommunityImproved(ml, node, partition)
					}
				}

				changeCount := 0

				for iter := 0; iter < numBetaIterations; iter++ {
					for _, node := range g.GetNodeList() {
						partition[node] = selectNewCommunityImproved(ml, node, partition)
					}

					objective := ml.ComputeObjectiveFunction(partition)
					objectiveHistory = append(objectiveHistory, objective)

					if objective > currentBest {
						currentBest = objective
						currentBestPartition = copyPartition(partition)
						changeCount = 0
						convergedAt = iter
					} else {
						changeCount++
					}

					if changeCount > 5 && iter > 10 {
						break
					}
				}
			}

			if currentBest > bestObjective {
				bestObjective = currentBest
				bestPartition = copyPartition(currentBestPartition)
				bestAlpha = alpha
				bestConvergedAt = convergedAt
				bestTotalIterations = len(objectiveHistory)

				ml.ComputeLikelihood(bestPartition)
				bestPin = ml.Pin
				bestPout = ml.Pout
			}
		}
	}

	return &GibbsSamplingResult{
		BestPartition:    bestPartition,
		BestObjective:    bestObjective,
		ObjectiveHistory: objectiveHistory,
		OptimalAlpha:     bestAlpha,
		OptimalPin:       bestPin,
		OptimalPout:      bestPout,
		NumCommunities:   NumCommunities(bestPartition),
		ConvergedAt:      bestConvergedAt,
		TotalIterations:  bestTotalIterations,
	}
}

// Оптимизация с учетом целевого числа кластеров
func selectNewCommunityImprovedWithTarget(ml *MLModel, node int, partition map[int]int) int {
	oldComm := partition[node]
	commsToTry := make(map[int]bool)
	commsToTry[oldComm] = true

	for neighbor := range ml.G.Edges[node] {
		commsToTry[partition[neighbor]] = true
	}

	// Вероятность создать новую коммьюнити
	currentK := NumCommunities(partition)
	canCreateNew := ml.TargetK < 0 || currentK < ml.TargetK

	if canCreateNew && rand.Float64() < 0.15 {
		newComm := rand.Intn(currentK + 3)
		commsToTry[newComm] = true
	}

	energies := make(map[int]float64)
	var maxEnergy float64 = math.Inf(-1)

	for comm := range commsToTry {
		partition[node] = comm
		energy := ml.ComputeObjectiveFunction(partition)
		energies[comm] = energy

		if energy > maxEnergy {
			maxEnergy = energy
		}
	}

	partition[node] = oldComm

	probabilities := make(map[int]float64)
	var sumProb float64

	for comm, energy := range energies {
		expVal := math.Exp(ml.Beta * (energy - maxEnergy))
		probabilities[comm] = expVal
		sumProb += expVal
	}

	if sumProb > 0 {
		for comm := range probabilities {
			probabilities[comm] /= sumProb
		}
	}

	r := rand.Float64()
	cumulative := 0.0

	for comm, prob := range probabilities {
		cumulative += prob
		if r < cumulative {
			return comm
		}
	}

	return oldComm
}

func selectNewCommunityImproved(ml *MLModel, node int, partition map[int]int) int {
	oldComm := partition[node]
	commsToTry := make(map[int]bool)
	commsToTry[oldComm] = true

	for neighbor := range ml.G.Edges[node] {
		commsToTry[partition[neighbor]] = true
	}

	if rand.Float64() < 0.15 {
		newComm := rand.Intn(NumCommunities(partition) + 3)
		commsToTry[newComm] = true
	}

	energies := make(map[int]float64)
	var maxEnergy float64 = math.Inf(-1)

	for comm := range commsToTry {
		partition[node] = comm
		energy := ml.ComputeObjectiveFunction(partition)
		energies[comm] = energy

		if energy > maxEnergy {
			maxEnergy = energy
		}
	}

	partition[node] = oldComm

	probabilities := make(map[int]float64)
	var sumProb float64

	for comm, energy := range energies {
		expVal := math.Exp(ml.Beta * (energy - maxEnergy))
		probabilities[comm] = expVal
		sumProb += expVal
	}

	if sumProb > 0 {
		for comm := range probabilities {
			probabilities[comm] /= sumProb
		}
	}

	r := rand.Float64()
	cumulative := 0.0

	for comm, prob := range probabilities {
		cumulative += prob
		if r < cumulative {
			return comm
		}
	}

	return oldComm
}

func initializeRandomPartition(g *Graph, numComms int) map[int]int {
	partition := make(map[int]int)
	for node := range g.Nodes {
		partition[node] = rand.Intn(numComms)
	}
	return partition
}

func copyPartition(partition map[int]int) map[int]int {
	res := make(map[int]int, len(partition))
	for k, v := range partition {
		res[k] = v
	}
	return res
}
