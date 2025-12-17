package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type TestResult struct {
	TestName        string
	GraphName       string
	Nodes           int
	Edges           int
	Algorithm       string
	Parameter1      float64
	Parameter2      float64
	Communities     int
	Modularity      float64
	Silhouette      float64
	ObjectiveValue  float64
	Iterations      int
	ConvergedAt     int
	ExecutionTime   float64
	Timestamp       string
}

func SaveResultsJSON(results []TestResult, filename string) error {
	os.MkdirAll("results", 0755)

	file, err := os.Create(fmt.Sprintf("results/%s.json", filename))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	fmt.Printf("✅ JSON saved: results/%s.json (%d records)\n", filename, len(results))
	return nil
}

func SaveResultsCSV(results []TestResult, filename string) error {
	os.MkdirAll("results", 0755)

	file, err := os.Create(fmt.Sprintf("results/%s.csv", filename))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"TestName",
		"GraphName",
		"Nodes",
		"Edges",
		"Algorithm",
		"Parameter1",
		"Parameter2",
		"Communities",
		"Modularity",
		"Silhouette",
		"ObjectiveValue",
		"Iterations",
		"ConvergedAt",
		"ExecutionTime_sec",
		"Timestamp",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	for _, r := range results {
		row := []string{
			r.TestName,
			r.GraphName,
			strconv.Itoa(r.Nodes),
			strconv.Itoa(r.Edges),
			r.Algorithm,
			fmt.Sprintf("%.6f", r.Parameter1),
			fmt.Sprintf("%.6f", r.Parameter2),
			strconv.Itoa(r.Communities),
			fmt.Sprintf("%.6f", r.Modularity),
			fmt.Sprintf("%.6f", r.Silhouette),
			fmt.Sprintf("%.6f", r.ObjectiveValue),
			strconv.Itoa(r.Iterations),
			strconv.Itoa(r.ConvergedAt),
			fmt.Sprintf("%.4f", r.ExecutionTime),
			r.Timestamp,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}

	fmt.Printf("✅ CSV saved: results/%s.csv (%d records)\n", filename, len(results))
	return nil
}

func SavePartitionJSON(partition map[int]int, filename string, metadata map[string]interface{}) error {
	os.MkdirAll("results", 0755)

	communities := make(map[int][]int)
	for node, comm := range partition {
		communities[comm] = append(communities[comm], node)
	}

	data := map[string]interface{}{
		"partition":    partition,
		"communities":  communities,
		"num_nodes":    len(partition),
		"num_communities": len(communities),
		"metadata":     metadata,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	file, err := os.Create(fmt.Sprintf("results/%s.json", filename))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	fmt.Printf("✅ Partition saved: results/%s.json\n", filename)
	return nil
}

func NewTestResult(
	testName string,
	graphName string,
	nodes int,
	edges int,
	algorithm string,
	param1 float64,
	param2 float64,
	communities int,
	modularity float64,
	silhouette float64,
	objectiveValue float64,
	iterations int,
	convergedAt int,
	executionTime float64,
) TestResult {
	return TestResult{
		TestName:       testName,
		GraphName:      graphName,
		Nodes:          nodes,
		Edges:          edges,
		Algorithm:      algorithm,
		Parameter1:     param1,
		Parameter2:     param2,
		Communities:    communities,
		Modularity:     modularity,
		Silhouette:     silhouette,
		ObjectiveValue: objectiveValue,
		Iterations:     iterations,
		ConvergedAt:    convergedAt,
		ExecutionTime:  executionTime,
		Timestamp:      time.Now().Format(time.RFC3339),
	}
}

func PrintTestResult(r TestResult) {
	fmt.Printf("\n📊 TEST RESULT: %s\n", r.TestName)
	fmt.Printf("   Graph: %s (%d nodes, %d edges)\n", r.GraphName, r.Nodes, r.Edges)
	fmt.Printf("   Algorithm: %s\n", r.Algorithm)
	if r.Algorithm == "HG" {
		fmt.Printf("   Gamma: %.4f\n", r.Parameter1)
	} else {
		fmt.Printf("   Alpha: %.4f, Beta: %.4f\n", r.Parameter1, r.Parameter2)
	}
	fmt.Printf("   Communities: %d\n", r.Communities)
	fmt.Printf("   Modularity: %.4f\n", r.Modularity)
	fmt.Printf("   Silhouette: %.4f\n", r.Silhouette)
	fmt.Printf("   Objective Value: %.4f\n", r.ObjectiveValue)
	fmt.Printf("   Iterations: %d (Converged at: %d)\n", r.Iterations, r.ConvergedAt)
	fmt.Printf("   Execution Time: %.4f sec\n", r.ExecutionTime)
}
