package main

import "image/color"

type Heatmap struct {
	Writer ImageWriter
	Filename string
	Generation *Generation
	Color color.RGBA
	eventCounts []int
}

const (
	MovesMap ResultType = iota
	ShotsMap
	KillsMap
	WaitsMap
	NumberOfHeatmapTypes
)

const HEATMAP_PIXELS_PER_CELL = 6
const MIN_OPACITY = 38 // out of 255; about 15%

func NewHeatmap(name string, gen *Generation, colour color.RGBA) *Heatmap {
	writer := NewImageWriter(name, HEATMAP_PIXELS_PER_CELL)
	writer.StartImage(gen.Arena)
	return &Heatmap{writer, writer.CurrentImage.Filename,	gen, colour, make([]int, gen.Arena.Width * gen.Arena.Height)}
}

func (hm *Heatmap) AddEvent(x, y int) {
	hm.eventCounts[x * hm.Generation.Arena.Height + y]++
}

func (hm *Heatmap) Write() {
	maxEvents := 0
	for _, count := range hm.eventCounts {
		if count > maxEvents {
			maxEvents = count
		}
	}

	for i, count := range hm.eventCounts {
		x, y := i / hm.Generation.Arena.Height, i % hm.Generation.Arena.Height
		if count > 0 && maxEvents > 0 {
			alpha := uint8(255 * float32(count) / float32(maxEvents))
			if alpha < MIN_OPACITY {
				alpha = MIN_OPACITY
			}
			hm.Writer.CurrentImage.DrawCell(x, y, color.RGBA{
				lightenColor(hm.Color.R, alpha),
				lightenColor(hm.Color.G, alpha),
				lightenColor(hm.Color.B, alpha),
				255,
			})
		}
	}
	hm.Writer.FinishImage()
}

// Washes out a color based on the alpha (opacity) value. We want to avoid doing proper transparency because it's
// noticeably more time-consuming than just filling in a solid pixel, and we never need to write to the same pixel more
// than once.
func lightenColor(c, alpha uint8) uint8 {
	distance := 255.0 - float32(c)
	ratio := (float32(255 - alpha) / 255.0)
	return uint8(float32(c) + distance * ratio)
}

// TODO: Don't generate a heatmap if the destination file already exists.
func GenerateHeatmaps(gen *Generation) []*Heatmap {
	var results [NumberOfHeatmapTypes]*Heatmap
	results[MovesMap] = NewHeatmap("moves", gen, color.RGBA{0, 213, 255, 255}) // teal
	results[ShotsMap] = NewHeatmap("shots", gen, color.RGBA{255, 136, 0, 255}) // orange
	results[KillsMap] = NewHeatmap("kills", gen, color.RGBA{255, 0, 0, 255})   // red
	results[WaitsMap] = NewHeatmap("waits", gen, color.RGBA{0, 255, 0, 255})   // green

	gen.FileManager.EachCell(func (x, y, moves, shots, kills, waits int) {
		for i := 0; i < moves; i++ {
			results[MovesMap].AddEvent(x, y)
		}
		for i := 0; i < shots; i++ {
			results[ShotsMap].AddEvent(x, y)
		}
		for i := 0; i < kills; i++ {
			results[KillsMap].AddEvent(x, y)
		}
		for i := 0; i < waits; i++ {
			results[WaitsMap].AddEvent(x, y)
		}
	})

	for _, hm := range results {
		hm.Write()
	}
	return results[:]
}
