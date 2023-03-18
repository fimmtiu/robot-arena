package main

import (
	"fmt"
	"io"
	"os"
)

type ResultsViewer struct {
	Scenario string
	Output io.Writer
	GenerationCount int
}

func NewResultsViewer(scenario string) *ResultsViewer {
	return &ResultsViewer{scenario, nil, CurrentHighestGeneration(scenario)}
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
		gen := NewGeneration(rv.Scenario, genId)
		rv.GenerateHeatmaps(gen)
		// rv.WriteHeatmaps(gen)
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
		gen := NewGeneration(rv.Scenario, genId)
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
	io.WriteString(rv.Output, "\n</table>\n")
}

func (rv *ResultsViewer) GenerateHeatmaps(gen *Generation) {
	// FIXME: huge mess, not ready yet
}

func (rv *ResultsViewer) WriteFooter() {
	io.WriteString(rv.Output, `
</body>
</html>
`)
}
