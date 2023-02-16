package main

import (
	"log"
	"os"
)

const BOTS_PER_TEAM = 5

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
}
