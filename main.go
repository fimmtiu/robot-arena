package main

import (
	"log"
	"os"
	"strings"

	"github.com/pkg/profile"
)

const BOTS_PER_TEAM = 5
const MAX_TICKS_PER_GAME = 200 // I'll crank this up to 2,000 after I'm done testing.

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

	scenario := os.Args[1]
	genCount := strToInt(os.Args[2])
	defer profile.Start(profile.ProfilePath(".")).Stop()

	for i := 0; i < genCount; i++ {
		gen := NewHighestGeneration(scenario)
		gen.Initialize(arena, getVisualizer(gen))
		logger.Printf("Running generation %d...", gen.Id)
		gen.Run()
	}

	logger.Printf("Done!")
}

func getVisualizer(gen *Generation) Visualizer {
	switch strings.ToLower(os.Getenv("VIS")) {
	case "gif":
		return NewGifVisualizer(gen.FileManager)
	case "mp4":
		return NewMp4Visualizer(gen.FileManager)
	default:
		return NewNullVisualizer()
	}
}
