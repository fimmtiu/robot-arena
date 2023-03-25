package main

import (
	"fmt"
	"os"
	"os/exec"
)

type Visualizer interface {
	Init(state *GameState)     // Called when the game state is initialized
	Update(action Action)      // Called once per action to tell the visualizer to record the current state
	NoChange()                 // Time advanced, but the game state wasn't changed in any way
	TickComplete()             // Called at the end of a tick, once all robots have taken a turn
	Finish()                   // Cleans up and writes whatever output is required
}

type NullVisualizer struct {
	// Doesn't do shit.
}

// Uses ImageWriter to generate a bunch of images, one per tick, then stitches them together into an animated GIF.
type GifVisualizer struct {
	FileManager *FileManager
	State *GameState
	img ImageWriter
}

// Uses ImageWriter to generate a bunch of images, one per action, then stitches them together into a movie.
type Mp4Visualizer struct {
	FileManager *FileManager
	State *GameState
	img ImageWriter
}

// Each grid cell in the arena will be PIXELS_PER_CELL pixels wide in the output images.
const DEFAULT_PIXELS_PER_CELL = 16

func NewNullVisualizer() *NullVisualizer {
	return &NullVisualizer{}
}

func (vis *NullVisualizer) Init(state *GameState) {}
func (vis *NullVisualizer) Update(action Action) {}
func (vis *NullVisualizer) NoChange() {}
func (vis *NullVisualizer) TickComplete() {}
func (vis *NullVisualizer) Finish() {}

func NewGifVisualizer(fm *FileManager) *GifVisualizer {
	return &GifVisualizer{fm, nil, NewImageWriter("tick", DEFAULT_PIXELS_PER_CELL)}
}

func (vis *GifVisualizer) Init(state *GameState) {
	vis.State = state
	// Before the first action of the game, write an image showing the initial game state.
	vis.img.WriteImage(state, nil)
}

func (vis *GifVisualizer) Update(action Action) {
	// To save space, the GIF visualizer doesn't bother showing the effects of individual actions. It only
	// generates one image per tick rather than one image per robot move.
}

func (vis *GifVisualizer) NoChange() {
	// No-op.
	// Because we only show one frame per tick, we don't care about keeping the intervals between actions consistent.
}

func (vis *GifVisualizer) TickComplete() {
	vis.img.WriteImage(vis.State, nil)
}

func (vis *GifVisualizer) Finish() {
	convertCommand := fmt.Sprintf("convert -delay 20 -loop 0 %s %s/game.gif", vis.img.WildCard(), vis.FileManager.GenerationDir())
	cmd := exec.Command("/bin/sh", "-c", convertCommand)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'convert': %v", err)
	}
	logger.Printf("Created GIF at %s/game.gif", vis.FileManager.GenerationDir())
}

func NewMp4Visualizer(fm *FileManager) *Mp4Visualizer {
	return &Mp4Visualizer{fm, nil, NewImageWriter("frame", DEFAULT_PIXELS_PER_CELL)}
}

func (vis *Mp4Visualizer) Init(state *GameState) {
	vis.State = state
	// Before the first action of the game, write an image showing the initial game state.
	vis.img.WriteImage(state, nil)
}

func (vis *Mp4Visualizer) Update(action Action) {
	vis.img.WriteImage(vis.State, &action)
}

// We write images on the turns of dead robots so that the speed of the visualization stays consistent, instead of
// speeding up as fewer robots are alive, Space Invaders-style.
func (vis *Mp4Visualizer) NoChange() {
	vis.img.WriteImage(vis.State, nil)
}

func (vis *Mp4Visualizer) TickComplete() {
	// No-op. We've already written images for all the individual actions.
}

func (vis *Mp4Visualizer) Finish() {
	ffmpegCommand := fmt.Sprintf("ffmpeg -y -framerate 30 -pattern_type glob -i '%s' -c:v libx264 -pix_fmt yuv420p %s", vis.img.WildCard(), vis.OutputFile())
	cmd := exec.Command("/bin/sh", "-c", ffmpegCommand)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'ffmpeg': %v", err)
	}
	logger.Printf("Created MP4 at %s/game.mp4", vis.FileManager.GenerationDir())

	if err := os.RemoveAll(vis.img.Dir); err != nil {
		logger.Fatalf("Could not destroy temporary directory %s: %v", vis.img.Dir, err)
	}
}

func (vis *Mp4Visualizer) OutputFile() string {
	return fmt.Sprintf("%s/game.mp4", vis.FileManager.GenerationDir())
}
