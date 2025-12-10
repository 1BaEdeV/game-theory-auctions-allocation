// hedonic.go
package main

// HedonicGame представляет гедоническую игру на графе
type HedonicGame struct {
	G          *Graph
	Partition  map[int]int // node → community_id
	Gamma      float64     // параметр игры (вес штрафа за чужаков)
	Iterations int         // количество проведённых итераций
}

// NewHedonicGame создаёт новую игру
// Каждый узел начинает в своей собственной коммьюнити
func NewHedonicGame(g *Graph, gamma float64) *HedonicGame {
	partition := make(map[int]int)
	for node := range g.Nodes {
		partition[node] = node // каждый узел - своя коммьюнити
	}

	return &HedonicGame{
		G:          g,
		Partition:  partition,
		Gamma:      gamma,
		Iterations: 0,
	}
}

// ComputeUtility вычисляет утилиту узла в данной коммьюнити
// u_i(C) = (количество друзей в C) - gamma * (количество чужаков в C)
func (hg *HedonicGame) ComputeUtility(node, community int) float64 {
	friends := 0.0
	strangers := 0.0

	for neighbor := range hg.G.Edges[node] {
		if hg.Partition[neighbor] == community {
			friends += 1.0
		} else if hg.Partition[neighbor] != community {
			strangers += 1.0
		}
	}

	return friends - hg.Gamma*strangers
}

// BestCommunity находит сообщество, которое максимизирует утилиту узла
func (hg *HedonicGame) BestCommunity(node int) int {
	// текущая коммьюнити узла
	currentComm := hg.Partition[node]
	bestComm := currentComm
	bestUtil := hg.ComputeUtility(node, currentComm)

	// проверить все коммьюнити соседей
	checkedComms := make(map[int]bool)
	for neighbor := range hg.G.Edges[node] {
		comm := hg.Partition[neighbor]
		if !checkedComms[comm] {
			checkedComms[comm] = true
			util := hg.ComputeUtility(node, comm)
			if util > bestUtil {
				bestUtil = util
				bestComm = comm
			}
		}
	}

	// проверить свою коммьюнити (создание новой коммьюнити)
	soloUtil := hg.ComputeUtility(node, node)
	if soloUtil > bestUtil {
		bestComm = node
	}

	return bestComm
}

// BetterResponseDynamics запускает алгоритм поиска Нэш-равновесия
// Каждый узел по очереди выбирает сообщество, которое максимизирует его утилиту
// Алгоритм продолжается до сходимости (когда никто не хочет менять коммьюнити)
func (hg *HedonicGame) BetterResponseDynamics() map[int]int {
	maxIterations := 1000

	for iter := 0; iter < maxIterations; iter++ {
		changed := false

		// каждый узел выбирает свою лучшую коммьюнити
		for _, node := range hg.G.GetNodeList() {
			oldComm := hg.Partition[node]
			newComm := hg.BestCommunity(node)

			if oldComm != newComm {
				hg.Partition[node] = newComm
				changed = true
			}
		}

		hg.Iterations = iter + 1

		// если никто не менялся, достигли равновесия
		if !changed {
			break
		}
	}

	return hg.Partition
}

// GetCommunityStructure возвращает разбиение узлов по сообществам
func (hg *HedonicGame) GetCommunityStructure() map[int][]int {
	comms := make(map[int][]int)
	for node, comm := range hg.Partition {
		comms[comm] = append(comms[comm], node)
	}
	return comms
}
