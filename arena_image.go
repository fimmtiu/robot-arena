package main

import (
	"image"
	"image/color"
	"image/draw"
)

// Source data for a PNG image of the current state of the arena.
type ArenaImage struct {
	Filename string
	Width int
	Height int
	PixelsPerCell int
	Image *image.RGBA
}

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

// Draws an arbitrary rectangle of a solid color on the image. Doesn't respect opacity.
func (img *ArenaImage) DrawRect(rect image.Rectangle, cellColor color.RGBA) {
	swatch := image.NewUniform(cellColor)
	draw.Draw(img.Image, rect, swatch, image.Point{0, 0}, draw.Src)
}
