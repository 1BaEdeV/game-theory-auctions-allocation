// Модуль для сохранения результатов в JSON и CSV форматы

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Result представляет результаты одного эксперимента
type Result struct {
	Graph       string
	Nodes       int
	Edges       int
	Gamma       float64
	Iterations  int
	Modularity  float64
	Silhouette  float64
	Communities int
	Duration    string
}

// JSONNode представляет узел в JSON экспорте
type JSONNode struct {
	ID        int `json:"id"`
	Community int `json:"community"`
}

// JSONEdge представляет ребро в JSON экспорте
type JSONEdge struct {
	Source int `json:"source"`
	Target int `json:"target"`
}

// JSONMetadata содержит метаданные в JSON экспорте
type JSONMetadata struct {
	NumNodes       int     `json:"numNodes"`
	NumEdges       int     `json:"numEdges"`
	NumCommunities int     `json:"numCommunities"`
	Modularity     float64 `json:"modularity"`
	Silhouette     float64 `json:"silhouette"`
}

// JSONData полная структура JSON экспорта
type JSONData struct {
	Nodes    []JSONNode   `json:"nodes"`
	Edges    []JSONEdge   `json:"edges"`
	Metadata JSONMetadata `json:"metadata"`
}

// ExportJSON экспортирует результаты в JSON формат
// Используется для дальнейшей визуализации и анализа
func ExportJSON(
	g *Graph,
	partition map[int]int,
	modularity, silhouette float64,
	filename string,
) error {
	// Создать директорию если её нет
	os.MkdirAll("results", 0755)

	// ============ ПОДГОТОВИТЬ УЗЛЫ ============
	// Преобразовать узлы в JSON формат
	nodes := make([]JSONNode, 0, g.NumNodes())
	for _, node := range g.GetNodeList() {
		nodes = append(nodes, JSONNode{
			ID:        node,
			Community: partition[node],
		})
	}
	// Преобразовать рёбра в JSON формат
	// Важно: избегаем дублирования (ребро (u,v) = ребро (v,u))
	edges := make([]JSONEdge, 0)
	for u := range g.Nodes {
		for v := range g.Edges[u] {
			if u < v { // только если u < v, чтобы не дублировать
				edges = append(edges, JSONEdge{
					Source: u,
					Target: v,
				})
			}
		}
	}
	metadata := JSONMetadata{
		NumNodes:       g.NumNodes(),
		NumEdges:       g.NumEdges(),
		NumCommunities: NumCommunities(partition),
		Modularity:     modularity,
		Silhouette:     silhouette,
	}
	data := JSONData{
		Nodes:    nodes,
		Edges:    edges,
		Metadata: metadata,
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("не могу создать файл %s: %w", filename, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // красивое форматирование (2 пробела)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("ошибка при записи JSON: %w", err)
	}

	fmt.Printf("JSON экспортирован: %s (%d узлов, %d рёбер)\n", filename, g.NumNodes(), g.NumEdges())
	return nil
}

// ExportCSV экспортирует результаты экспериментов в CSV таблицу
func ExportCSV(results []Result, filename string) error {
	// Создать директорию если её нет
	os.MkdirAll("results", 0755)

	// Открыть файл для записи
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("❌ не могу создать файл %s: %w", filename, err)
	}
	defer file.Close()

	// Создать CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"Graph",
		"Nodes",
		"Edges",
		"Gamma",
		"Iterations",
		"Modularity",
		"Silhouette",
		"Communities",
		"Duration",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("❌ ошибка при записи заголовка: %w", err)
	}
	for _, r := range results {
		row := []string{
			r.Graph,
			strconv.Itoa(r.Nodes),
			strconv.Itoa(r.Edges),
			fmt.Sprintf("%.1f", r.Gamma),
			strconv.Itoa(r.Iterations),
			fmt.Sprintf("%.4f", r.Modularity),
			fmt.Sprintf("%.4f", r.Silhouette),
			strconv.Itoa(r.Communities),
			r.Duration,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("ошибка при записи строки: %w", err)
		}
	}

	fmt.Printf("CSV экспортирован: %s (%d экспериментов)\n", filename, len(results))
	return nil
}

// PrintResults выводит результаты в красивом формате в консоль
func PrintResults(results []Result) {
	if len(results) == 0 {
		fmt.Println("Нет результатов для вывода")
		return
	}

	// Вывести заголовок
	fmt.Printf("%-18s | %-7s | %-7s | %-7s | %-11s | %-12s | %-12s | %-12s\n",
		"Graph", "Nodes", "Edges", "Gamma", "Iterations", "Modularity", "Silhouette", "Communities")

	// Вывести данные
	for _, r := range results {
		fmt.Printf("%-18s | %7d | %7d | %7.1f | %11d | %12.4f | %12.4f | %12d\n",
			r.Graph,
			r.Nodes,
			r.Edges,
			r.Gamma,
			r.Iterations,
			r.Modularity,
			r.Silhouette,
			r.Communities,
		)
	}
}

// PrintResultsDetailed выводит подробную информацию о каждом результате
func PrintResultsDetailed(results []Result) {
	for i, r := range results {
		fmt.Printf("\n РЕЗУЛЬТАТ %d\n", i+1)
		fmt.Println("─────────────────────────────")
		fmt.Printf("  Граф:           %s\n", r.Graph)
		fmt.Printf("  Узлов:          %d\n", r.Nodes)
		fmt.Printf("  Рёбер:          %d\n", r.Edges)
		fmt.Printf("  Параметр γ:     %.1f\n", r.Gamma)
		fmt.Printf("  Итерации:       %d\n", r.Iterations)
		fmt.Printf("  Модулярность:   %.4f\n", r.Modularity)
		fmt.Printf("  Силуэт:         %.4f\n", r.Silhouette)
		fmt.Printf("  Сообществ:      %d\n", r.Communities)
		fmt.Printf("  Время:          %s\n", r.Duration)
	}
	fmt.Println()
}
