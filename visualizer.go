package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"time"
)

type Visualizer interface {
	Init(state *GameState)     // Called when the game state is initialized
	Update(action Action)      // Called once per action to tell the visualizer to record the current state
	NoChange()                 // Time advanced, but the game state wasn't changed in any way
	TickComplete()             // Called at the end of a tick, once all robots have taken a turn
	Finish()                   // Cleans up and writes whatever output is required
}

// Generates a PNG image of the current state of the arena.
type ImageWriter struct {
	Dir string
	Prefix string
	NextFileIndex int

	blackSquare image.Image
	whiteSquare image.Image
	greenSquare image.Image
	redSquare image.Image
	blueSquare image.Image
	lightRedSquare image.Image
	lightBlueSquare image.Image
	lightGreenSquare image.Image
	laserSquare image.Image
}

type NullVisualizer struct {
	// Doesn't do shit.
}

// Uses ImageWriter to generate a bunch of images, one per tick, then stitches them together into an animated GIF.
type GifVisualizer struct {
	State *GameState
	img ImageWriter
}

// Uses ImageWriter to generate a bunch of images, one per action, then stitches them together into a movie.
type Mp4Visualizer struct {
	State *GameState
	img ImageWriter
}

// Each grid cell in the arena will be PIXELS_PER_CELL pixels wide in the output images.
const PIXELS_PER_CELL = 16

func NewImageWriter(prefix string) ImageWriter {
	dir := fmt.Sprintf("/tmp/robot-arena-%d-%d", os.Getpid(), time.Now().UnixNano())
	if err := os.Mkdir(dir, 0755); err != nil {
		logger.Fatalf("Could not create temporary directory %s: %v", dir, err)
	}

	laserWidth := PIXELS_PER_CELL / 8
	if laserWidth < 1 {
		laserWidth = 1
	}

	return ImageWriter{
		dir, prefix, 0,
		makeSolidSquare(PIXELS_PER_CELL, 0, 0, 0), makeSolidSquare(PIXELS_PER_CELL, 255, 255, 255),
		makeSolidSquare(PIXELS_PER_CELL, 0, 255, 0), makeSolidSquare(PIXELS_PER_CELL, 255, 0, 0),
		makeSolidSquare(PIXELS_PER_CELL, 0, 0, 255), makeSolidSquare(PIXELS_PER_CELL, 255, 100, 100),
		makeSolidSquare(PIXELS_PER_CELL, 100, 100, 255), makeSolidSquare(PIXELS_PER_CELL, 120, 255, 120),
		makeSolidSquare(laserWidth, 64, 255, 64),
	}
}

// The names have to be lexicographically sorted so that they're assembled in the right order.
func (img *ImageWriter) NextFileName() string {
	path := fmt.Sprintf("%s/%s_%011d.png", img.Dir, img.Prefix, img.NextFileIndex)
	img.NextFileIndex++
	return path
}

// Returns a glob path that matches all the images generated by this ImageWriter.
func (img *ImageWriter) WildCard() string {
	return fmt.Sprintf("%s/%s_*.png", img.Dir, img.Prefix)
}

func (img *ImageWriter) WriteImage(state *GameState, action *Action) {
	width := state.Arena.Width
	height := state.Arena.Height
	frame := image.NewRGBA(image.Rect(0, 0, width * PIXELS_PER_CELL, height * PIXELS_PER_CELL))
	swatch := image.Rect(0, 0, PIXELS_PER_CELL, PIXELS_PER_CELL)

	// Draw the map
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			cell := state.Arena.Cells[x * height + y]
			rect := image.Rect(x * PIXELS_PER_CELL, y * PIXELS_PER_CELL, (x + 1) * PIXELS_PER_CELL, (y + 1) * PIXELS_PER_CELL)
			switch cell.Type {
			case WallCell:
				draw.Draw(frame, rect, img.blackSquare, swatch.Min, draw.Src)
			case SpawnCell, OpenCell:
				draw.Draw(frame, rect, img.whiteSquare, swatch.Min, draw.Src)
			case GoalCell:
				draw.Draw(frame, rect, img.greenSquare, swatch.Min, draw.Src)
			}
		}
	}

	// Draw the bots. Dead bots show up as a light color.
	for _, bot := range state.Bots {
		rect := image.Rect(bot.Position.X * PIXELS_PER_CELL, bot.Position.Y * PIXELS_PER_CELL,
											(bot.Position.X + 1) * PIXELS_PER_CELL, (bot.Position.Y + 1) * PIXELS_PER_CELL)
		if bot.Team == TeamA {
			if bot.Alive {
				draw.Draw(frame, rect, img.redSquare, swatch.Min, draw.Src)
			} else {
				draw.Draw(frame, rect, img.lightRedSquare, swatch.Min, draw.Src)
			}
		} else {
			if bot.Alive {
				draw.Draw(frame, rect, img.blueSquare, swatch.Min, draw.Src)
			} else {
				draw.Draw(frame, rect, img.lightBlueSquare, swatch.Min, draw.Src)
			}
		}
	}

	// Draw the lasers in a nice bright green. (This is not terribly efficient. Lots of overdraw.)
	if action != nil && action.Type == ActionShoot {
		shooterX := state.CurrentBot.Position.X * PIXELS_PER_CELL + PIXELS_PER_CELL / 2
		shooterY := state.CurrentBot.Position.Y * PIXELS_PER_CELL + PIXELS_PER_CELL / 2
		targetX := action.Target.X * PIXELS_PER_CELL + PIXELS_PER_CELL / 2
		targetY := action.Target.Y * PIXELS_PER_CELL + PIXELS_PER_CELL / 2
		halfWidth := img.laserSquare.Bounds().Size().X / 2

		BresenhamLine(shooterX, shooterY, targetX, targetY, func (x, y int) bool {
			rect := image.Rect(x - halfWidth, y - halfWidth, x + halfWidth, y + halfWidth)
			draw.Draw(frame, rect, img.laserSquare, swatch.Min, draw.Src)
			return true
		})
	}

	path := img.NextFileName()
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logger.Fatalf("Can't open file %s: %v", path, err)
	}
	png.Encode(f, frame)
	f.Close()
}

// We pre-generate some solid color swatches so that we can copy them over in big rectangles instead of laboriously
// filling in each pixel on each frame.
func makeSolidSquare(sideLen int, r, g, b uint8) image.Image {
	square := image.NewRGBA(image.Rect(0, 0, sideLen, sideLen))
	for x := 0; x < sideLen; x++ {
		for y := 0; y < sideLen; y++ {
			square.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return square
}

func NewNullVisualizer() *NullVisualizer {
	return &NullVisualizer{}
}

func (vis *NullVisualizer) Init(state *GameState) {}
func (vis *NullVisualizer) Update(action Action) {}
func (vis *NullVisualizer) NoChange() {}
func (vis *NullVisualizer) TickComplete() {}
func (vis *NullVisualizer) Finish() {}

func NewGifVisualizer() *GifVisualizer {
	return &GifVisualizer{nil, NewImageWriter("tick")}
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
	convertCommand := fmt.Sprintf("convert -delay 20 -loop 0 %s %s/game.gif", vis.img.WildCard(), fileManager.GenerationDir())
	cmd := exec.Command("/bin/sh", "-c", convertCommand)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'convert': %v", err)
	}
	logger.Printf("Created GIF at %s/game.gif", fileManager.GenerationDir())
}

func NewMp4Visualizer() *Mp4Visualizer {
	return &Mp4Visualizer{nil, NewImageWriter("frame")}
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
// speeding up the fewer robots are alive, Space Invaders-style.
func (vis *Mp4Visualizer) NoChange() {
	vis.img.WriteImage(vis.State, nil)
}

func (vis *Mp4Visualizer) TickComplete() {
	// No-op. We've already written images for all the individual actions.
}

func (vis *Mp4Visualizer) Finish() {
	ffmpegCommand := fmt.Sprintf("ffmpeg -y -framerate 30 -pattern_type glob -i '%s' -c:v libx264 -pix_fmt yuv420p %s/game.mp4", vis.img.WildCard(), fileManager.GenerationDir())
	cmd := exec.Command("/bin/sh", "-c", ffmpegCommand)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'ffmpeg': %v", err)
	}
	logger.Printf("Created MP4 at %s/game.mp4", fileManager.GenerationDir())

	if err := os.RemoveAll(vis.img.Dir); err != nil {
		logger.Fatalf("Could not destroy temporary directory %s: %v", vis.img.Dir, err)
	}
}
