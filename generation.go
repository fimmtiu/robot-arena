// That's "generation" as in a cohort of a particular age, not as in "the act of generating stuff".

package main

import (
	"math/rand"
	"sort"

	"golang.org/x/exp/slices"
)

type Generation struct {
	Id int
	Previous *Generation
	fileManager *FileManager
	scriptEditor *ScriptEditor
	randomScriptIds []int         // A randomly shuffled array of script IDs in this generation
	matchesPlayed map[int][]int   // A map of (script ID) -> [IDs of scripts this script has already fought]
}

const SCRIPTS_PER_GENERATION = 1000
const MATCHES_PER_SCRIPT = 10

const KEEP_PERCENT = 0.20
const RANDOM_PERCENT = 0.35
const MUTATE_PERCENT = 0.30
const SPLICE_PERCENT = 0.35

func NewGeneration(scenario string, id int) *Generation {
	var previous *Generation = nil
	if id > 1 {
		previous = &Generation{id - 1, nil, NewFileManager(scenario, id - 1), NewScriptEditor(), []int{}, map[int][]int{}}
	}

	fileManager := NewFileManager(scenario, id)

	scriptIds := make([]int, len(fileManager.ScriptIds))
	copy(scriptIds, fileManager.ScriptIds)
	rand.Shuffle(len(scriptIds), func(i, j int) {
		scriptIds[i], scriptIds[j] = scriptIds[j], scriptIds[i]
	})

	matchesPlayed := make(map[int][]int, len(scriptIds))
	for _, id := range fileManager.ScriptIds {
		matchesPlayed[id] = make([]int, MATCHES_PER_SCRIPT)
	}

	return &Generation{id, previous, fileManager, NewScriptEditor(), scriptIds, matchesPlayed}
}

// Ensure that we have a minimum number of scripts in the scripts folder.
func (g *Generation) Initialize() {
	if g.Previous == nil {
		logger.Printf("Gen %d: Creating %d new random scripts", g.Id, SCRIPTS_PER_GENERATION)
		for i := 0; i < SCRIPTS_PER_GENERATION; i++ {
			g.MakeNewRandomScript()
		}
	} else {
		best := g.Previous.BestScores()
		var count int

		logger.Printf("Gen %d: Copying the %d best scripts from generation %d", g.Id, len(best), g.Previous.Id)
		for count = 0; count < len(best); count++ {
			g.CopyScriptFromPreviousGen(best[count])
			count++
		}
		logger.Printf("Gen %d: Mangling %d scripts", g.Id, SCRIPTS_PER_GENERATION - count)
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

func (g *Generation) Run(arena *Arena, vis Visualizer) {
	for matchId := 0; matchId < SCRIPTS_PER_GENERATION * MATCHES_PER_SCRIPT; matchId++ {
		scriptA, scriptB := g.pickTwoScripts()
		match := NewMatch(arena, vis, matchId, scriptA, scriptB)

		for {
			if match.RunTick() {   // Returns true when the match is over
				break
			}
		}

		logger.Printf("Match %d: script %d: %d points, script %d: %d points", matchId, scriptA, match.Scores[TeamA], scriptB, match.Scores[TeamB])
		g.fileManager.writeCellStatistics(match)
		g.fileManager.writeMatchOutcome(match)
	}
}

// Find two random scripts that haven't yet played each other and return their IDs.
func (g *Generation) pickTwoScripts() (int, int) {
	for id, played := range g.matchesPlayed {
		for _, opponentId := range g.randomScriptIds {
			if id != opponentId && !slices.Contains(played, opponentId) {
				g.matchesPlayed[id] = append(g.matchesPlayed[id], opponentId)
				g.matchesPlayed[opponentId] = append(g.matchesPlayed[opponentId], id)
				g.removeScriptIfFinished(id)
				g.removeScriptIfFinished(opponentId)

				return id, opponentId
			}
		}
	}

	logger.Fatalf("Couldn't find two valid scripts to fight!\nmatchesPlayed: %v\nrandomScriptIds:", g.matchesPlayed, g.randomScriptIds)
	return -1, -1
}

// If a script has already played MATCHES_PER_SCRIPT matches, remove it from our data structures.
func (g *Generation) removeScriptIfFinished(id int) {
	if len(g.matchesPlayed[id]) >= MATCHES_PER_SCRIPT {
		delete(g.matchesPlayed, id)
		index := slices.Index(g.randomScriptIds, id)
		slices.Delete(g.randomScriptIds, index, index + 1)
	}
}
