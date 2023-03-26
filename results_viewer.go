package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type ResultsViewer struct {
	Scenario string
	Arena *Arena
	Output io.Writer
	GenerationCount int
}

const SCORES_PER_GENERATION = 10

func NewResultsViewer(scenario string, arena *Arena) *ResultsViewer {
	return &ResultsViewer{scenario, arena, nil, CurrentHighestGeneration(scenario)}
}

func (rv *ResultsViewer) GenerateResults() {
	logger.Printf("Generating results...")
	path := fmt.Sprintf("scenario/%s/results.html", rv.Scenario)
	file, err := os.Create(path)
	if err != nil {
		logger.Fatalf("Can't open %s for writing: %v", path, err)
	}
	rv.Output = file
	defer file.Close()

	rv.WriteHeader()
	rv.WriteSummary()
	for genId := 1; genId <= rv.GenerationCount; genId++ {
		gen := NewGeneration(rv.Scenario, genId, rv.Arena)

		rv.WriteBestScores(gen)
		heatmaps := GenerateHeatmaps(gen)
		rv.WriteHeatmaps(heatmaps)
	}
	rv.WriteFooter()
	logger.Printf("Results are at %s", path)
}

func (rv *ResultsViewer) WriteHeader() {
	io.WriteString(rv.Output, fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<title>Robot Arena Results: %s</title>
		</head>

		<body>
			<h1>Robot Arena Results: %s</h1>
	`, rv.Scenario, rv.Scenario))
}

func (rv *ResultsViewer) WriteSummary() {
	io.WriteString(rv.Output, `
		<h3>Summary</h3>
		<table>
		<tr>
			<th>Generation</th>
			<th>Successful runs</th>
			<th>Average script size</th>
		</tr>
	`)

	for genId := 1; genId <= rv.GenerationCount; genId++ {
		gen := NewGeneration(rv.Scenario, genId, rv.Arena)
		successes := 0
		gen.FileManager.EachResultRow(func (_, _, _, scoreA, scoreB, _ int) {
			if scoreA > 0 || scoreB > 0 {
				successes++
			}
		})

		io.WriteString(rv.Output, fmt.Sprintf(`
			<tr>
				<td>%d</td>
				<td>%d</td>
				<td>%d</td>
			</tr>
		`, genId, successes, gen.FileManager.AverageScriptSize()))
	}
	io.WriteString(rv.Output, `
		</table>
	`)
}

func (rv *ResultsViewer) WriteBestScores(gen *Generation) {
	io.WriteString(rv.Output, `
		<table>
			<tr>
				<th>Script ID</th>
				<th>Score</th>
			</tr>
	`)

	scores := gen.BestScores()
	for i := 0; i < SCORES_PER_GENERATION; i++ {
		io.WriteString(rv.Output, fmt.Sprintf(`
			<tr>
				<td><a href="gen_%d/scripts/%d.l">%d</td>
				<td>%.3f</td>
			</tr>
		`, gen.Id, scores[i].Id, scores[i].Id, scores[i].Score))
	}

	io.WriteString(rv.Output, `
		</table>
	`)
}

func (rv *ResultsViewer) WriteHeatmaps(heatmaps []*Heatmap) {
	io.WriteString(rv.Output, `
		<table>
		  <tr>
			  <th>Moves</th>
				<th>Waits</th>
			</tr>
			<tr>
	`)

	rv.WriteHeatmap(heatmaps[MovesMap])
	rv.WriteHeatmap(heatmaps[WaitsMap])

	io.WriteString(rv.Output, `
	    </tr>
		  <tr>
	`)

	rv.WriteHeatmap(heatmaps[ShotsMap])
	rv.WriteHeatmap(heatmaps[KillsMap])

	io.WriteString(rv.Output, `
			</tr>
  	</table>
	`)
}

func (rv *ResultsViewer) WriteHeatmap(heatmap *Heatmap) {
	relative_path := fmt.Sprintf("gen_%d/%s.png", heatmap.Generation.Id, heatmap.Writer.Prefix)
	destination := fmt.Sprintf("scenario/%s/%s", rv.Scenario, relative_path)
	logger.Printf("Copying heatmap from '%s' to '%s'", heatmap.Filename, destination)
	cmd := exec.Command("cp", heatmap.Filename, destination)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'cp': %v", err)
	}

	io.WriteString(rv.Output, fmt.Sprintf(`
		<td><img src="%s"></td>
	`, relative_path))
}

func (rv *ResultsViewer) WriteFooter() {
	io.WriteString(rv.Output, `
		</body>
		</html>
	`)
}
