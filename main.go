package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/pkg/profile"
)

const BOTS_PER_TEAM = 5
const MAX_TICKS_PER_GAME = 200 // I'll crank this up higher after I'm done testing.

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", log.Ldate | log.Ltime)
	InitScript()
}

func main() {
	logger = log.New(os.Stdout, "", log.Ldate | log.Ltime)
	InitScript()

	arena := LoadArena("arena.png")
	logger.Printf("Loaded %dx%d arena.", arena.Width, arena.Height)
	if os.Getenv("PROF") != "" {
		defer profile.Start(profile.ProfilePath(".")).Stop()
	}

	action := os.Args[1]
	scenario := os.Args[2]

	switch action {
	case "run":
		genCount := strToInt(os.Args[3])

		for i := 0; i < genCount; i++ {
			gen := NewHighestGeneration(scenario, arena)
			gen.Initialize(NewNullVisualizer())
			logger.Printf("Running generation %d...", gen.Id)
			gen.Run()
		}
		NewResultsViewer(scenario, arena).GenerateResults()

	case "view":
		genId := strToInt(os.Args[3])
		matchId := strToInt(os.Args[4])

		gen := NewGeneration(scenario, genId, arena)
		vis := NewMp4Visualizer(gen.FileManager)
		gen.Initialize(vis)
		scriptA, scriptB := gen.FileManager.FindScriptIds(matchId)
		match := NewMatch(gen, matchId, scriptA, scriptB)
		match.Run()
		logger.Printf("Match %d: script %d: %d points, script %d: %d points", matchId, scriptA, match.Scores[TeamA], scriptB, match.Scores[TeamB])
		cmd := exec.Command("open", vis.OutputFile())
		err := cmd.Run()
		if err != nil {
			logger.Fatalf("Failed to run 'open': %v", err)
		}

	case "results":
		NewResultsViewer(scenario, arena).GenerateResults()

	default:
		logger.Fatalf("Unknown command: %s", action)
	}

	logger.Printf("Done!")
}
