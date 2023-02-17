package main

import (
	"log"
	"os"
	"strconv"
)

const BOTS_PER_TEAM = 5
const MAX_TICKS_PER_GAME = 10

var logger *log.Logger
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
	genCount, err := strconv.Atoi(os.Args[2])
	if err != nil {
		logger.Fatalf("Invalid generation count: %s", os.Args[2])
	}

	// FIXME: What if we're not starting from zero? Adding new generations to an existing scenario?
	for i := 0; i < genCount; i++ {
		fileManager = NewFileManager(scenario, i)
		gen := NewGeneration(i)
		gen.Initialize()
		gen.Run(arena)
	}

	vis := NewGifVisualizer()
	currentMatch = NewMatch(arena, vis, 1, 1, 1)
	vis.Init(currentMatch.State)

	for i := 0; i < MAX_TICKS_PER_GAME; i++ {
		if currentMatch.RunTick() {
			break
		}
	}

	vis.Finish("/tmp/game.gif")
	logger.Printf("Done!")
}
