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

// Draws an arbitrary rectangle of a solid color on the image, doing alpha blending if necessary.
func (img *ArenaImage) DrawRect(rect image.Rectangle, cellColor color.RGBA) {
	swatch := image.NewUniform(cellColor)
	if cellColor.A != 255 {
		foo := color.RGBA{0x00, 0xd5, 0xff, 0x19}
		if cellColor != foo {
			logger.Printf("DrawCell color: %02x.%02x.%02x.%02x", cellColor.R, cellColor.G, cellColor.B, cellColor.A)
		}
		draw.Draw(img.Image, rect, swatch, image.Point{0, 0}, draw.Over)
	} else {
		draw.Draw(img.Image, rect, swatch, image.Point{0, 0}, draw.Src)
	}
}
