// metrics.go
package main

import (
	"math"
)

// Modularity вычисляет модулярность разбиения
// Q = (1/2m) * Σ(a_ij - (k_i * k_j / 2m)) * δ(c_i, c_j)
// где:
// - a_ij = 1 если есть ребро между i и j
// - k_i = степень узла i
// - m = количество рёбер
// - δ(c_i, c_j) = 1 если узлы в одной коммьюнити
func Modularity(g *Graph, partition map[int]int) float64 {
	m := float64(g.NumEdges())
	if m == 0 {
		return 0
	}

	Q := 0.0

	for u := range g.Nodes {
		for v := range g.Nodes {
			if partition[u] == partition[v] {
				ku := float64(len(g.Edges[u]))
				kv := float64(len(g.Edges[v]))

				aij := 0.0
				if g.Edges[u][v] {
					aij = 1.0
				}

				Q += aij - (ku*kv)/(2*m)
			}
		}
	}

	return Q / (2 * m)
}

// SilhouetteCoefficient вычисляет коэффициент силуэта
// Мера того, насколько хорошо узел подходит к своей коммьюнити
// Значения от -1 (плохо) до 1 (хорошо)
func SilhouetteCoefficient(g *Graph, partition map[int]int) float64 {
	// Построить коммьюнити
	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	if len(comms) == 1 {
		// Только одна коммьюнити
		return 0
	}

	totalSilhouette := 0.0
	count := 0

	// Для каждого узла вычислить силуэт
	for node, nodeComm := range partition {
		nodes := comms[nodeComm]

		// a(i) = среднее расстояние внутри своей коммьюнити
		// расстояние = 0, если есть ребро; 1, если нет
		a := 0.0
		if len(nodes) > 1 {
			for _, other := range nodes {
				if node == other {
					continue
				}
				if g.HasEdge(node, other) {
					a += 0.0 // связаны — близко
				} else {
					a += 1.0 // не связаны — далеко
				}
			}
			a /= float64(len(nodes) - 1)
		}

		// b(i) = минимальное среднее расстояние до другой коммьюнити
		b := math.MaxFloat64
		for otherComm, otherNodes := range comms {
			if otherComm == nodeComm {
				continue
			}
			dist := 0.0
			for _, other := range otherNodes {
				if g.HasEdge(node, other) {
					dist += 0.0
				} else {
					dist += 1.0
				}
			}
			if len(otherNodes) > 0 {
				dist /= float64(len(otherNodes))
			}
			if dist < b {
				b = dist
			}
		}
		if b == math.MaxFloat64 {
			// Нет соседних коммьюнити или нет рёбер — считаем максимально далеко
			b = 1.0
		}

		// s(i) = (b(i) - a(i)) / max(a(i), b(i))
		denom := math.Max(a, b)
		if denom > 0 {
			s := (b - a) / denom
			totalSilhouette += s
			count++
		} else if a == 0 && b == 0 {
			// изолированный узел без связей
			count++
		}
	}

	if count > 0 {
		return totalSilhouette / float64(count)
	}
	return 0
}

// NumCommunities возвращает количество сообществ
func NumCommunities(partition map[int]int) int {
	comms := make(map[int]bool)
	for _, comm := range partition {
		comms[comm] = true
	}
	return len(comms)
}

// PrintCommunities выводит в консоль информацию о сообществах
func PrintCommunities(partition map[int]int) {
	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	for comm, nodes := range comms {
		println("Community", comm, ":", len(nodes), "nodes")
	}
}
