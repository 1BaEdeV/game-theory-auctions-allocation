// export.go
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ExperimentResult struct {
	TestName      string
	Algorithm     string
	Parameter     float64
	NumNodes      int
	NumEdges      int
	Communities   int
	Potential     float64
	Modularity    float64
	Iterations    int
	ConvergedAt   int
	ExecutionTime float64
	Timestamp     string
}

type PartitionJSON struct {
	Directed   bool        `json:"directed"`
	Multigraph bool        `json:"multigraph"`
	Graph      interface{} `json:"graph"`
	Nodes      []NodeJSON  `json:"nodes"`
	Links      []LinkJSON  `json:"links"`
}

type NodeJSON struct {
	ID        string `json:"id"`
	Community int    `json:"community"`
}

type LinkJSON struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

func ExportPartitionToJSON(g *Graph, partition map[int]int, idToName map[int]string, filename string) error {
	nodes := make([]NodeJSON, 0, len(partition))
	for nodeID, commID := range partition {
		name := idToName[nodeID]
		nodes = append(nodes, NodeJSON{
			ID:        name,
			Community: commID,
		})
	}

	links := make([]LinkJSON, 0, g.NumEdges())
	visited := make(map[[2]int]bool)

	for u := 0; u < g.NumNodes(); u++ {
		for v := range g.Edges[u] {
			if u < v {
				u1, u2 := u, v
				if u > v {
					u1, u2 = v, u
				}
				key := [2]int{u1, u2}
				if !visited[key] {
					links = append(links, LinkJSON{
						Source: idToName[u],
						Target: idToName[v],
					})
					visited[key] = true
				}
			}
		}
	}

	pj := PartitionJSON{
		Directed:   false,
		Multigraph: false,
		Graph:      map[string]interface{}{},
		Nodes:      nodes,
		Links:      links,
	}

	data, err := json.MarshalIndent(pj, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON error: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	return nil
}

func SaveResultsToCSV(results []ExperimentResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create error: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"TestName",
		"Algorithm",
		"Parameter",
		"NumNodes",
		"NumEdges",
		"Communities",
		"Potential",
		"Modularity",
		"Iterations",
		"ConvergedAt",
		"ExecutionTime",
		"Timestamp",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("header error: %w", err)
	}

	for _, r := range results {
		row := []string{
			r.TestName,
			r.Algorithm,
			fmt.Sprintf("%.6f", r.Parameter),
			fmt.Sprintf("%d", r.NumNodes),
			fmt.Sprintf("%d", r.NumEdges),
			fmt.Sprintf("%d", r.Communities),
			fmt.Sprintf("%.6f", r.Potential),
			fmt.Sprintf("%.6f", r.Modularity),
			fmt.Sprintf("%d", r.Iterations),
			fmt.Sprintf("%d", r.ConvergedAt),
			fmt.Sprintf("%.4f", r.ExecutionTime),
			r.Timestamp,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}

	return nil
}

func NewExperimentResult(
	testName string,
	algorithm string,
	parameter float64,
	g *Graph,
	partition map[int]int,
	potential float64,
	modularity float64,
	iterations int,
	convergedAt int,
	executionTime float64,
) ExperimentResult {
	comms := make(map[int][]int)
	for node, comm := range partition {
		comms[comm] = append(comms[comm], node)
	}

	return ExperimentResult{
		TestName:      testName,
		Algorithm:     algorithm,
		Parameter:     parameter,
		NumNodes:      g.NumNodes(),
		NumEdges:      g.NumEdges(),
		Communities:   len(comms),
		Potential:     potential,
		Modularity:    modularity,
		Iterations:    iterations,
		ConvergedAt:   convergedAt,
		ExecutionTime: executionTime,
		Timestamp:     time.Now().Format(time.RFC3339),
	}
}
