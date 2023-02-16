package main

import (
	"log"
	"os"
)

const BOTS_PER_TEAM = 5

var logger *log.Logger
var fileManager *FileManager

func init() {
	logger = log.New(os.Stdout, "", log.Ldate | log.Ltime)
	InitScript()
}

func main() {
	logger = log.New(os.Stdout, "", log.Ldate | log.Ltime)
	InitScript()

	// FIXME: read command line arguments, including scenario name
	fileManager = NewFileManager("monkey", 1)

	arena := LoadArena("arena.png")
	logger.Printf("Loaded %dx%d arena.", arena.Width, arena.Height)

	vis := NewGifVisualizer()

	currentMatch = NewMatch(arena, vis, 1, 1, 1)
	vis.Init(currentMatch)
	currentMatch.RunTick()
}
