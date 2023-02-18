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
	TickComplete()             // Called at the end of a tick, once all robots have taken a turn
	Finish()                   // Cleans up and writes whatever output is required
}

type GifVisualizer struct {
	Dir string
	State *GameState
	NextFileIndex int

	blackSquare image.Image
	whiteSquare image.Image
	greenSquare image.Image
	redSquare image.Image
	blueSquare image.Image
	lightRedSquare image.Image
	lightBlueSquare image.Image
	lightGreenSquare image.Image
}

const GIF_SCALE = 16

func NewGifVisualizer() *GifVisualizer {
	dir := fmt.Sprintf("/tmp/robot-arena-%d-%d", os.Getpid(), time.Now().UnixNano())
	if err := os.Mkdir(dir, 0755); err != nil {
		logger.Fatalf("Could not create temporary directory %s: %v", dir, err)
	}

	return &GifVisualizer{
		dir, nil, 0,
		makeSolidSquare(0, 0, 0), makeSolidSquare(255, 255, 255),	makeSolidSquare(0, 255, 0), makeSolidSquare(255, 0, 0),
		makeSolidSquare(0, 0, 255), makeSolidSquare(255, 100, 100),	makeSolidSquare(100, 100, 255),	makeSolidSquare(120, 255, 120),
	}
}

func (vis *GifVisualizer) Init(state *GameState) {
	vis.State = state
}

func (vis *GifVisualizer) Update(action Action) {
	// If this is the first action of the game, write a tick image showing the initial game state.
	if vis.NextFileIndex == 0 {
		path := fmt.Sprintf("%s/tick_%05d.png", vis.Dir, 0)
		vis.writePng(Action{Type: ActionWait}, path)
	}

	// The names have to be lexicographically sorted so that they're assembled in the right order.
	path := fmt.Sprintf("%s/frame_%010d.png", vis.Dir, vis.NextFileIndex)
	vis.NextFileIndex++
	vis.writePng(action, path)
}

func (vis *GifVisualizer) TickComplete() {
	path := fmt.Sprintf("%s/tick_%05d.png", vis.Dir, vis.State.Tick + 1)
	vis.writePng(Action{Type: ActionWait}, path)
}

func (vis *GifVisualizer) Finish() {
	convertCommand := fmt.Sprintf("convert -delay 20 -loop 0 %s/tick_*.png %s/game.gif", vis.Dir, fileManager.GenerationDir())
	cmd := exec.Command("/bin/sh", "-c", convertCommand)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'convert': %v", err)
	}
	logger.Printf("Created GIF at %s/game.gif", fileManager.GenerationDir())

	ffmpegCommand := fmt.Sprintf("ffmpeg -y -framerate 30 -pattern_type glob -i '%s/frame_*.png' -c:v libx264 -pix_fmt yuv420p %s/game.mp4", vis.Dir, fileManager.GenerationDir())
	cmd = exec.Command("/bin/sh", "-c", ffmpegCommand)
	err = cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'ffmpeg': %v", err)
	}
	logger.Printf("Created MP4 at %s/game.mp4", fileManager.GenerationDir())

	if err := os.RemoveAll(vis.Dir); err != nil {
		logger.Fatalf("Could not destroy temporary directory %s: %v", vis.Dir, err)
	}
}

func (vis *GifVisualizer) writePng(action Action, outfile string) {
	width := vis.State.Arena.Width
	height := vis.State.Arena.Height
	frame := image.NewRGBA(image.Rect(0, 0, width * GIF_SCALE, height * GIF_SCALE))
	swatch := image.Rect(0, 0, GIF_SCALE, GIF_SCALE)

	// Draw the map
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			cell := vis.State.Arena.Cells[x * height + y]
			rect := image.Rect(x * GIF_SCALE, y * GIF_SCALE, (x + 1) * GIF_SCALE, (y + 1) * GIF_SCALE)
			switch cell.Type {
			case WallCell:
				draw.Draw(frame, rect, vis.blackSquare, swatch.Min, draw.Src)
			case SpawnCell, OpenCell:
				draw.Draw(frame, rect, vis.whiteSquare, swatch.Min, draw.Src)
			case GoalCell:
				draw.Draw(frame, rect, vis.greenSquare, swatch.Min, draw.Src)
			}
		}
	}

	// Draw the bots
	for _, bot := range vis.State.Bots {
		rect := image.Rect(bot.Position.X * GIF_SCALE, bot.Position.Y * GIF_SCALE,
											(bot.Position.X + 1) * GIF_SCALE, (bot.Position.Y + 1) * GIF_SCALE)
		if bot.Team == TeamA {
			if bot.Alive {
				draw.Draw(frame, rect, vis.redSquare, swatch.Min, draw.Src)
			} else {
				draw.Draw(frame, rect, vis.lightRedSquare, swatch.Min, draw.Src)
			}
		} else {
			if bot.Alive {
				draw.Draw(frame, rect, vis.blueSquare, swatch.Min, draw.Src)
			} else {
				draw.Draw(frame, rect, vis.lightBlueSquare, swatch.Min, draw.Src)
			}
		}
	}

	// Draw the lasers in a nice bright green.
	if action.Type == ActionShoot {
		BresenhamLine(vis.State.CurrentBot.Position.X, vis.State.CurrentBot.Position.Y, action.Target.X, action.Target.Y, func (x, y int) bool {
			scaledX := vis.State.CurrentBot.Position.X * GIF_SCALE + GIF_SCALE / 2
			scaledY := vis.State.CurrentBot.Position.Y * GIF_SCALE + GIF_SCALE / 2
			frame.Set(scaledX, scaledY, color.RGBA{64, 255, 64, 255})
			return true
		})
	}

	f, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logger.Fatalf("Can't open file %s: %v", outfile, err)
	}
	png.Encode(f, frame)
	f.Close()
}

// We pre-generate some solid color swatches so that we can copy them over in big rectangles instead of laboriously
// filling in each pixel on each frame.
func makeSolidSquare(r, g, b uint8) image.Image {
	square := image.NewRGBA(image.Rect(0, 0, GIF_SCALE, GIF_SCALE))
	for x := 0; x < GIF_SCALE; x++ {
		for y := 0; y < GIF_SCALE; y++ {
			square.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return square
}
