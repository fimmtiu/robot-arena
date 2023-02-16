package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"time"
)

type Visualizer interface {
	Init(match *Match)         // Called when the game state is initialized
	Update()                   // Called once per action to tell the visualizer to record the current state
	Finish(outputPath string)  // Writes the entire story to a file
}

type GifVisualizer struct {
	Dir string
	Match *Match
	NextFileIndex int

	blackSquare image.Image
	whiteSquare image.Image
	greenSquare image.Image
}

const GIF_SCALE = 32
const NUM_COLORS = 256 // FIXME: Look into how to reduce this without the palettizing breaking everything.
// (I think we'd have to create a custom palette and pass it in.)

func NewGifVisualizer() *GifVisualizer {
	dir := fmt.Sprintf("/tmp/robot-arena-%d-%d", os.Getpid(), time.Now().UnixNano())
	if err := os.Mkdir(dir, 0755); err != nil {
		logger.Fatalf("Could not create temporary directory %s: %v", dir, err)
	}

	return &GifVisualizer{dir, nil, 0, makeSolidSquare(0, 0, 0), makeSolidSquare(255, 255, 255), makeSolidSquare(0, 255, 0)}
}

func (vis *GifVisualizer) Init(match *Match) {
	vis.Match = match
	logger.Printf("White: %v", vis.whiteSquare)
}

func (vis *GifVisualizer) Update() {
	frame := image.NewRGBA(image.Rect(0, 0, vis.Match.Arena.Width * GIF_SCALE, vis.Match.Arena.Height * GIF_SCALE))
	swatch := image.Rect(0, 0, GIF_SCALE, GIF_SCALE)

	// Draw the map
	for x := 0; x < vis.Match.Arena.Width; x++ {
		for y := 0; y < vis.Match.Arena.Height; y++ {
			cell := vis.Match.Arena.Cells[x * vis.Match.Arena.Height + y]
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
	// FIXME make gif

	// if err := os.RemoveAll(vis.Dir); err != nil {
	// 	logger.Fatalf("Could not destroy temporary directory %s: %v", vis.Dir, err)
	// }
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
