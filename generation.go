// That's "generation" as in a cohort of a particular age, not as in "the act of generating stuff".

package main

import (
	"math/rand"
	"sort"
)

type Generation struct {
	Id int
	Previous *Generation
	FileManager *FileManager
	scriptEditor *ScriptEditor
	Arena *Arena
	Visualizer Visualizer
	matchups [][2]int   // A list of [scriptA, scriptB] pairs.
}

const SCRIPTS_PER_GENERATION = 10000
const MATCHES_PER_SCRIPT = 6

const KEEP_PERCENT = 0.20
const RANDOM_PERCENT = 0.35
const MUTATE_PERCENT = 0.30
const SPLICE_PERCENT = 0.35

func NewHighestGeneration(scenario string) *Generation {
	highestGeneration := CurrentHighestGeneration(scenario)
	return NewGeneration(scenario, highestGeneration + 1)
}

func NewGeneration(scenario string, id int) *Generation {
	var previous *Generation = nil
	if id > 1 {
		previous = &Generation{id - 1, nil, NewFileManager(scenario, id - 1), NewScriptEditor(), nil, nil, [][2]int{}}
	}

	fileManager := NewFileManager(scenario, id)
	return &Generation{id, previous, fileManager, NewScriptEditor(), nil, nil, [][2]int{}}
}

func (g *Generation) Initialize(arena *Arena, vis Visualizer) {
	g.Arena = arena
	g.Visualizer = vis

	// Ensure that we have a minimum number of scripts in the scripts folder.
	g.FileManager.ReadScriptIds()
	if len(g.FileManager.ScriptIds) < SCRIPTS_PER_GENERATION {
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

	g.FileManager.ReadScriptIds()
	g.calculateMatchups(g.FileManager.ScriptIds, MATCHES_PER_SCRIPT)
}

// Populate some record-keeping data structures that we use to track which scripts will play each other.
// Note that, depending on the number of scripts and the value of matchesPerScript, some scripts might play more than
// matchesPerScript matches — it's more of a minimum than a limit.
func (g *Generation) calculateMatchups(scriptIds []int, matchesPerScript int) {
	randomScriptIds := make([]int, 0, len(scriptIds))
	randomScriptIds = append(randomScriptIds, scriptIds...)
	rand.Shuffle(len(randomScriptIds), func(i, j int) {
		randomScriptIds[i], randomScriptIds[j] = randomScriptIds[j], randomScriptIds[i]
	})

	matchCounts := make(map[int]int, len(randomScriptIds))

	findAvailableOpponent := func(id, startAt, maxMatches int) bool {
		for i := 0; i < len(randomScriptIds); i++ {
			index := (startAt + i) % len(randomScriptIds)
			opponentId := randomScriptIds[index]

			if matchCounts[opponentId] < maxMatches && id != opponentId {
				matchCounts[id]++
				matchCounts[opponentId]++
				g.matchups = append(g.matchups, [2]int{id, opponentId})
				return true
			}
		}
		return false
	}

	for i, id := range randomScriptIds {
		for n := 1; matchCounts[id] < matchesPerScript; n++ {
			if !findAvailableOpponent(id, i + n, matchesPerScript) {
				findAvailableOpponent(id, i + n, matchesPerScript + 1)
			}
		}
	}
}

func (g *Generation) CopyScriptFromPreviousGen(scriptId int) {
	code := g.Previous.FileManager.ScriptCode(scriptId)
	g.FileManager.WriteNewScript(code)
}

func (g *Generation) MakeNewRandomScript() {
	code := g.scriptEditor.RandomScript(MIN_EXPRS_PER_SCRIPT)
	g.FileManager.WriteNewScript(code)
}

func (g *Generation) MutateScript(scriptId int) {
	code := g.scriptEditor.MutateScript(g.Previous.FileManager.ScriptCode(scriptId))
	g.FileManager.WriteNewScript(code)
}

func (g *Generation) SpliceScripts(scriptA, scriptB int) {
	code := g.scriptEditor.SpliceScripts(g.Previous.FileManager.ScriptCode(scriptA), g.Previous.FileManager.ScriptCode(scriptB))
	g.FileManager.WriteNewScript(code)
}

type ScriptScore struct {
	Id int
	Score float64
	Sum int
	Count int
}

// Returns the IDs of the top-scoring KEEP_PERCENT scripts.
func (g *Generation) BestScores() []int {
	scores := make([]ScriptScore, len(g.FileManager.ScriptIds))

	g.FileManager.EachResultRow(func (matchId, scriptA, scriptB, scoreA, scoreB, ticks int) {
		scores[scriptA - 1].Id = scriptA
		scores[scriptA - 1].Sum += scoreA
		scores[scriptA - 1].Count++
		scores[scriptB - 1].Id = scriptB
		scores[scriptB - 1].Sum += scoreB
		scores[scriptB - 1].Count++
	})

	for i := range scores {
		scores[i].Score = float64(scores[i].Sum) / float64(scores[i].Count)
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score   // Sorts in reverse order so the best scripts are first
	})

	elements_to_keep := int(float64(len(scores)) * KEEP_PERCENT)
	ids := make([]int, elements_to_keep)
	for i := 0; i < elements_to_keep; i++ { // FIXME: this can just be a copy() now
		ids[i] = scores[i].Id
	}
	return ids
}

func (g *Generation) Run() {
	for matchId := 0; ; matchId++ {
		done, scriptA, scriptB := g.pickTwoScripts()
		if done {
			break
		}
		match := NewMatch(g, matchId, scriptA, scriptB)
		match.Run()

		logger.Printf("Match %d: script %d: %d points, script %d: %d points", matchId, scriptA, match.Scores[TeamA], scriptB, match.Scores[TeamB])
		g.FileManager.writeCellStatistics(match)
		g.FileManager.writeMatchOutcome(match)
	}
}

// Find two scripts that haven't yet played each other and return their IDs.
func (g *Generation) pickTwoScripts() (bool, int, int) {
	if len(g.matchups) > 0 {
		matchup := g.matchups[0]
		g.matchups = g.matchups[1:]
		return false, matchup[0], matchup[1]
	} else {
		return true, -1, -1
	}
}
