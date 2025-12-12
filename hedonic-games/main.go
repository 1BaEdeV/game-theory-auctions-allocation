// main.go
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("🎮 HEDONIC GAMES: COMMUNITY DETECTION")

	// ============= ЗАГРУЗИТЬ ДАТАСЕТ УЧИТЕЛЕЙ =============
	fmt.Println("📚 LOADING TEACHERS DATASET")

	teachers, err := LoadAMteachers("../ds/relations_graph.json")
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	// Конвертировать в граф
	G_teachers, _, idToName := teachers.ToGraph()

	// Печать информации
	PrintGraphInfo(G_teachers, "Teachers Network", idToName)

	// ============= ЭКСПЕРИМЕНТЫ ПО ГАММА =============
	fmt.Println("🚀 RUNNING HEDONIC GAMES EXPERIMENTS")

	// Набор значений гамма
	gammas := []float64{0.1, 0.2, 0.3, 0.5, 0.7, 1.0}

	// Все результаты сюда
	var allResults []Result

	for _, gamma := range gammas {
		fmt.Printf("\n=== GAMMA = %.2f ===\n", gamma)

		// 10 запусков для данной гамма
		for run := 1; run <= 10; run++ {
			fmt.Printf("  ▶ Run %d/10\n", run)

			start := time.Now()
			game := NewHedonicGame(G_teachers, gamma)
			partition := game.BetterResponseDynamics()
			elapsed := time.Since(start)

			mod := Modularity(G_teachers, partition)
			sil := SilhouetteCoefficient(G_teachers, partition)
			comms := NumCommunities(partition)

			fmt.Printf("    Modularity:  %.4f\n", mod)
			fmt.Printf("    Silhouette:  %.4f\n", sil)
			fmt.Printf("    Communities: %d\n", comms)
			fmt.Printf("    Iterations:  %d\n", game.Iterations)
			fmt.Printf("    Time:        %v\n", elapsed)

			// Можно печатать сообщества только для первого запуска
			if run == 1 {
				fmt.Println("    First run communities:")
				PrintCommunities(partition)
			}

			// Собираем результат
			res := Result{
				Graph:       "Teachers Network",
				Nodes:       G_teachers.NumNodes(),
				Edges:       G_teachers.NumEdges(),
				Gamma:       gamma,
				Iterations:  game.Iterations,
				Modularity:  mod,
				Silhouette:  sil,
				Communities: comms,
				Duration:    elapsed.String(),
			}
			allResults = append(allResults, res)

			// Экспорт JSON для каждого случая (опционально)
			filenameJSON := fmt.Sprintf("results/teachers_gamma_%.2f_run_%02d.json", gamma, run)
			ExportJSON(G_teachers, partition, mod, sil, filenameJSON)
		}
	}

	// ============= ЭКСПОРТ СВОДНЫХ РЕЗУЛЬТАТОВ =============
	fmt.Println("\n💾 EXPORTING SUMMARY RESULTS")

	// Печать таблицы в консоль
	PrintResults(allResults)

	// Один общий CSV по всем гамма и всем 10 прогоном
	ExportCSV(allResults, "results/results_teachers_gamma_grid.csv")

	fmt.Println("✅ ALL COMPLETED SUCCESSFULLY!")
}
