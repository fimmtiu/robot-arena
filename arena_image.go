package main

import (
	"image"
	"image/color"
	"image/draw"
)


type ColorIndex uint8
const (
	ColorBlack ColorIndex = iota
	ColorWhite
	ColorRed
	ColorGreen
	ColorBlue
	ColorOrange
	ColorLightRed
	ColorLightBlue
	NumberOfColors
)

// Source data for a PNG image of the current state of the arena.
type ArenaImage struct {
	Name string
	Width int
	Height int
	PixelsPerCell int
	Image *image.RGBA
}

var colorSwatches = make(map[int]image.Image, NumberOfColors)

func NewArenaImage(name string, width, height, pixelsPerCell int) *ArenaImage {
	return &ArenaImage{
		name, width, height, pixelsPerCell,
		image.NewRGBA(image.Rect(0, 0, width * pixelsPerCell, height * pixelsPerCell)),
	}
}

// Fills in a single cell of the arena grid.
func (img *ArenaImage) DrawCell(x, y int, cellColor color.RGBA) {
	rect := image.Rect(x * img.PixelsPerCell, y * img.PixelsPerCell, (x + 1) * img.PixelsPerCell, (y + 1) * img.PixelsPerCell)
	img.DrawRect(rect, cellColor)
}

// Draws an arbitrary rectangle of swatch size or smaller on the image.
func (img *ArenaImage) DrawRect(rect image.Rectangle, cellColor color.RGBA) {
	swatch := img.GetColorSwatch(cellColor)
	if cellColor.A == 255 {
		draw.Draw(img.Image, rect, swatch, swatch.Bounds().Min, draw.Src)
	} else {
		draw.Draw(img.Image, rect, swatch, swatch.Bounds().Min, draw.Over)
	}
}

// We generate and cache solid color swatches so that we can copy them over in big rectangles instead of laboriously
// filling in each pixel on each frame.
func (img *ArenaImage) GetColorSwatch(c color.RGBA) image.Image {
	key := (int(c.R) << 24) | (int(c.G) << 16) | (int(c.B) << 8) | int(c.A)
	swatch, found := colorSwatches[key]
	if !found {
		rgba := image.NewRGBA(image.Rect(0, 0, img.PixelsPerCell, img.PixelsPerCell))
		for x := 0; x < img.PixelsPerCell; x++ {
			for y := 0; y < img.PixelsPerCell; y++ {
				rgba.Set(x, y, c)
			}
		}
		swatch = rgba
		colorSwatches[key] = rgba
	}
	return swatch
}
