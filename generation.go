// That's "generation" as in a cohort of a particular age, not as in "the act of generating stuff".

package main

import (
	"math/rand"
	"sort"
)

type Generation struct {
	Id int
	Previous *Generation
	fileManager *FileManager
	scriptEditor *ScriptEditor
}

const SCRIPTS_PER_GENERATION = 1000

const KEEP_PERCENT = 0.20
const RANDOM_PERCENT = 0.35
const MUTATE_PERCENT = 0.30
const SPLICE_PERCENT = 0.35

func NewGeneration(scenario string, id int) *Generation {
	var previous *Generation = nil
	if id > 1 {
		previous = &Generation{id - 1, nil, NewFileManager(scenario, id - 1), NewScriptEditor()}
	}
	return &Generation{id, previous, NewFileManager(scenario, id), NewScriptEditor()}
}

// Ensure that we have a minimum number of scripts in the scripts folder.
func (g *Generation) Initialize() {
	if g.Previous == nil {
		for i := 0; i < SCRIPTS_PER_GENERATION; i++ {
			g.MakeNewRandomScript()
		}
	} else {
		best := g.Previous.BestScores()
		var count int

		for count = 0; count < len(best); count++ {
			g.CopyScriptFromPreviousGen(best[count])
			count++
		}
		for ; count < SCRIPTS_PER_GENERATION; count++ {
			n := rand.Float32()
			if n < RANDOM_PERCENT {
				g.MakeNewRandomScript()
			} else if n < RANDOM_PERCENT + MUTATE_PERCENT {
				g.MutateScript(best[rand.Intn(len(best))])
			} else {
				g.SpliceScripts(best[rand.Intn(len(best))], best[rand.Intn(len(best))])
			}
		}
	}
}

func (g *Generation) CopyScriptFromPreviousGen(scriptId int) {
	code := g.Previous.fileManager.ScriptCode(scriptId)

	file := g.fileManager.NewScriptFile()
	file.WriteString(code)
	file.Close()
}

func (g *Generation) MakeNewRandomScript() {
	code := g.scriptEditor.RandomScript()
	file := fileManager.NewScriptFile()
	file.WriteString(code)
	file.Close()
}

func (g *Generation) MutateScript(scriptId int) {
	code := g.scriptEditor.MutateScript(g.fileManager.ScriptCode(scriptId))
	file := fileManager.NewScriptFile()
	file.WriteString(code)
	file.Close()
}

func (g *Generation) SpliceScripts(scriptA, scriptB int) {
	code := g.scriptEditor.SpliceScripts(g.fileManager.ScriptCode(scriptA), g.fileManager.ScriptCode(scriptB))
	file := fileManager.NewScriptFile()
	file.WriteString(code)
	file.Close()
}

type ScriptScore struct {
	Id int
	Score float64
	Sum int
	Count int
}

// Returns the IDs of the top-scoring KEEP_PERCENT scripts.
func (g *Generation) BestScores() []int {
	scores := make([]ScriptScore, 0, len(g.fileManager.ScriptIds))

	fileManager.EachResultRow(func (matchId, scriptA, scriptB, scoreA, scoreB, ticks int) {
		scores[scriptA].Id = scriptA
		scores[scriptA].Sum += scoreA
		scores[scriptA].Count++
		scores[scriptB].Id = scriptB
		scores[scriptB].Sum += scoreB
		scores[scriptB].Count++
	})

	for i := range scores {
		scores[i].Score = float64(scores[i].Sum) / float64(scores[i].Count)
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score   // Sorts in reverse order so the best scripts are first
	})

	elements_to_keep := int(float64(len(scores)) * KEEP_PERCENT)
	ids := make([]int, 0, elements_to_keep)
	for i := 0; i < elements_to_keep; i++ {
		ids[i] = scores[i].Id
	}
	return ids
}

func (g *Generation) Run(arena *Arena) {

}

