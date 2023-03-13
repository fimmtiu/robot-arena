package main

import (
	"log"
	"os"
	"strings"
)

const BOTS_PER_TEAM = 5
const MAX_TICKS_PER_GAME = 200 // I'll crank this up to 2,000 after I'm done testing.

var logger *log.Logger

// FIXME: Can we rearrange things so we don't need this global, where all file-related calls go through the current Generation?
var fileManager *FileManager

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

	for i := 0; i < genCount; i++ {
		// fileManager = NewFileManager(scenario, i)
		gen := NewHighestGeneration(scenario)
		gen.Initialize()
		logger.Printf("Running generation %d...", gen.Id)
		gen.Run(arena, getVisualizer())
	}

	logger.Printf("Done!")
}

func getVisualizer() Visualizer {
	switch strings.ToLower(os.Getenv("VIS")) {
	case "gif":
		return NewGifVisualizer()
	case "mp4":
		return NewMp4Visualizer()
	default:
		return NewNullVisualizer()
	}
}
