// metrics.go
package main

import (
	"fmt"
	"math"
)

// ComputeModularity вычисляет модулярность разбиения
// Q = (1/2m) * Σ(a_ij - (k_i * k_j / 2m)) * δ(c_i, c_j)
func ComputeModularity(g *Graph, partition map[int]int) float64 {
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
func SilhouetteCoefficient(g *Graph, partition map[int]int) float64 {
	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	if len(comms) == 1 {
		return 0
	}

	totalSilhouette := 0.0
	count := 0

	for node, nodeComm := range partition {
		nodes := comms[nodeComm]

		a := 0.0
		if len(nodes) > 1 {
			for _, other := range nodes {
				if node == other {
					continue
				}

				if g.HasEdge(node, other) {
					a += 0.0
				} else {
					a += 1.0
				}
			}
			a /= float64(len(nodes) - 1)
		}

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
			b = 1.0
		}

		denom := math.Max(a, b)
		if denom > 0 {
			s := (b - a) / denom
			totalSilhouette += s
			count++
		} else if a == 0 && b == 0 {
			count++
		}
	}

	if count > 0 {
		return totalSilhouette / float64(count)
	}

	return 0
}

// PrintCommunities печатает разбиение узлов по сообществам
func PrintCommunities(partition map[int]int) {
	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	for comm, nodes := range comms {
		fmt.Printf("C%d: ", comm)
		for i, v := range nodes {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(v)
		}
		fmt.Println()
	}
	fmt.Println()
}
