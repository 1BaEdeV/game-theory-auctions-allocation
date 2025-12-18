package main

type HedonicGame struct {
	G          Graph
	Partition  map[int]int
	Alpha      float64
	Beta       float64 // Параметр модулярности
	Iterations int
}

func NewHedonicGame(g Graph, alpha float64) *HedonicGame {
	partition := make(map[int]int)
	for i := 0; i < g.NumNodes(); i++ {
		partition[i] = i // начальное разбиение - каждый в своем комьюнити в одиночку
	}
	return &HedonicGame{
		G:          g,
		Partition:  partition,
		Alpha:      alpha,
		Iterations: 0,
	}

}

// GetCommunityStructure возвращает структуру коммьюнити
func (hg *HedonicGame) GetCommunityStructure() map[int][]int {
	comms := make(map[int][]int)
	for node, comm := range hg.Partition {
		comms[comm] = append(comms[comm], node)
	}
	return comms
}

// GetNumberOfCommunities возвращает количество кластеров
func (hg *HedonicGame) GetNumberOfCommunities() int {
	return len(hg.GetCommunityStructure())
}

func NumCommunities(partition map[int]int) int {
	comms := make(map[int]bool)
	for _, comm := range partition {
		comms[comm] = true
	}
	return len(comms)
}

// ========== Формула (7.1) со стр. 183 ==========
// P(Π) = Σ_k [m(S_k) - n(S_k)(n(S_k)-1)α/2]
// где m(S_k) - число ребер в кластере k
//      n(S_k) - число узлов в кластере k

// ComputePotential_Formula71 вычисляет потенциал по формуле (7.1)
func (hg *HedonicGame) ComputePotential_Formula71() float64 {
	communities := hg.GetCommunityStructure()
	P := 0.0

	for _, nodes := range communities {
		m_sk := 0.0 // Количество ребер в кластере
		n_sk := float64(len(nodes))

		// Подсчитаем ребра внутри кластера
		for i := 0; i < len(nodes); i++ {
			for j := i + 1; j < len(nodes); j++ {
				u := nodes[i]
				v := nodes[j]
				if hg.G.HasEdge(u, v) {
					m_sk += 1.0
				}
			}
		}

		// P(Π) += m(S_k) - n(S_k)(n(S_k)-1)α/2
		contribution := m_sk - (n_sk * (n_sk - 1) * hg.Alpha / 2.0)
		P += contribution
	}

	return P
}

// ========== Формула (7.2) со стр. 184 ==========
// P(Π) = Σ_k Σ_{i,j ∈ S_k, i≠j} (A_ij - d_i*d_j/(2m))

// ComputePotential_Formula72 вычисляет потенциал по формуле (7.2) - модулярность
func (hg *HedonicGame) ComputePotential_Formula72() float64 {
	communities := hg.GetCommunityStructure()
	m := float64(hg.G.NumEdges())
	if m == 0 {
		return 0
	}

	P := 0.0

	for _, nodes := range communities {
		for i := 0; i < len(nodes); i++ {
			for j := i + 1; j < len(nodes); j++ {
				u := nodes[i]
				v := nodes[j]

				A_ij := 0.0
				if hg.G.HasEdge(u, v) {
					A_ij = 1.0
				}

				d_u := float64(len(hg.G.Edges[u]))
				d_v := float64(len(hg.G.Edges[v]))

				// P += A_ij - d_u*d_v/(2m)
				P += A_ij - (d_u * d_v / (2 * m))
			}
		}
	}

	return P
}

// ComputePotentialCurrent вычисляет текущий потенциал
func (hg *HedonicGame) ComputePotentialCurrent(useModularity bool) float64 {
	if useModularity {
		return hg.ComputePotential_Formula72()
	}
	return hg.ComputePotential_Formula71()
}

// ComputeUtility_BetterResponse вычисляет полезность для динамики наилучших ответов (стр. 178)
func (hg *HedonicGame) ComputeUtility_BetterResponse(node, community int) float64 {
	friends := 0.0
	strangers := 0.0

	for neighbor := range hg.G.Edges[node] {
		if hg.Partition[neighbor] == community {
			friends += 1.0
		} else if hg.Partition[neighbor] != community {
			strangers += 1.0
		}
	}

	return friends - hg.Alpha*strangers
}

// FindNashStablePartition_WithPotential находит Нэш-стабильное разбиение
func (hg *HedonicGame) FindNashStablePartition_WithPotential(maxIterations int, useModularity bool) map[int]int {
	for iter := 0; iter < maxIterations; iter++ {
		changed := false
		nodes := hg.G.GetNodeList()

		for _, node := range nodes {
			oldComm := hg.Partition[node]

			// Попробуем переместить узел в каждую коммьюнити и измерим изменение потенциала
			bestComm := oldComm
			bestPotential := hg.ComputePotentialCurrent(useModularity)

			// Получаем все соседние коммьюнити
			neighborComms := make(map[int]bool)
			for neighbor := range hg.G.Edges[node] {
				neighborComms[hg.Partition[neighbor]] = true
			}

			// Пробуем каждую соседнюю коммьюнити
			for comm := range neighborComms {
				hg.Partition[node] = comm
				newPotential := hg.ComputePotentialCurrent(useModularity)

				if newPotential > bestPotential {
					bestPotential = newPotential
					bestComm = comm
				}
			}

			// Пробуем solo
			hg.Partition[node] = node
			newPotential := hg.ComputePotentialCurrent(useModularity)
			if newPotential > bestPotential {
				bestComm = node
			}

			// Устанавливаем лучшую коммьюнити
			if bestComm != oldComm {
				hg.Partition[node] = bestComm
				changed = true
			} else {
				hg.Partition[node] = oldComm
			}
		}

		hg.Iterations = iter + 1

		if !changed {
			break
		}
	}

	return hg.Partition
}

// IsNashStable проверяет, является ли разбиение Нэш-стабильным
func (hg *HedonicGame) IsNashStable(useModularity bool) bool {
	currentPotential := hg.ComputePotentialCurrent(useModularity)
	nodes := hg.G.GetNodeList()

	for _, node := range nodes {
		oldComm := hg.Partition[node]

		// Пробуем переместить в другие коммьюнити
		neighborComms := make(map[int]bool)
		for neighbor := range hg.G.Edges[node] {
			neighborComms[hg.Partition[neighbor]] = true
		}

		for comm := range neighborComms {
			if comm == oldComm {
				continue
			}

			hg.Partition[node] = comm
			newPotential := hg.ComputePotentialCurrent(useModularity)
			hg.Partition[node] = oldComm

			if newPotential > currentPotential+1e-9 {
				return false
			}
		}
	}

	return true
}
