package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"os/exec"
	"time"
)

type Visualizer interface {
	Init(state *GameState)         // Called when the game state is initialized
	Update(action Action)      // Called once per action to tell the visualizer to record the current state
	Finish(outputPath string)  // Writes the entire story to a file
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
}

const GIF_SCALE = 16
const NUM_COLORS = 256 // FIXME: Look into how to reduce this without the palettizing breaking everything.
// (I think we'd have to create a custom palette and pass it in.)

func NewGifVisualizer() *GifVisualizer {
	dir := fmt.Sprintf("/tmp/robot-arena-%d-%d", os.Getpid(), time.Now().UnixNano())
	if err := os.Mkdir(dir, 0755); err != nil {
		logger.Fatalf("Could not create temporary directory %s: %v", dir, err)
	}

	return &GifVisualizer{
		dir, nil, 0,
		makeSolidSquare(0, 0, 0), makeSolidSquare(255, 255, 255),	makeSolidSquare(0, 255, 0), makeSolidSquare(255, 0, 0),
		makeSolidSquare(0, 0, 255), makeSolidSquare(255, 100, 100),	makeSolidSquare(100, 100, 255),
	}
}

func (vis *GifVisualizer) Init(state *GameState) {
	vis.State = state
}

func (vis *GifVisualizer) Update(action Action) {
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
			case Wall:
				draw.Draw(frame, rect, vis.blackSquare, swatch.Min, draw.Src)
			case Spawn, Open:
				draw.Draw(frame, rect, vis.whiteSquare, swatch.Min, draw.Src)
			case Goal:
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

	path := fmt.Sprintf("%s/frame_%d.gif", vis.Dir, vis.NextFileIndex)
	vis.NextFileIndex++

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logger.Fatalf("Can't open file %s: %v", path, err)
	}
	gif.Encode(f, frame, &gif.Options{NumColors: NUM_COLORS})
	f.Close()
}

func (vis *GifVisualizer) Finish(outputPath string) {
	convertCommand := fmt.Sprintf("convert -delay 100 -loop 0 %s/*.gif %s", vis.Dir, outputPath)
	cmd := exec.Command("/bin/sh", "-c", convertCommand)
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("Failed to run 'convert': %v", err)
	}

	if err := os.RemoveAll(vis.Dir); err != nil {
		logger.Fatalf("Could not destroy temporary directory %s: %v", vis.Dir, err)
	}
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
