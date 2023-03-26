package main

import "image/color"

type Heatmap struct {
	Writer ImageWriter
	Cells map[int]int // FIXME not used any more, remove this
	Color color.RGBA
}

const (
	MovesMap ResultType = iota
	ShotsMap
	KillsMap
	WaitsMap
	NumberOfHeatmapTypes
)

const HEATMAP_PIXELS_PER_CELL = 8
const TRANSPARENCY = 255 / 10

func NewHeatmap(name string, arena *Arena, colour color.RGBA) *Heatmap {
	writer := NewImageWriter(name, HEATMAP_PIXELS_PER_CELL)
	writer.StartImage(arena)
	return &Heatmap{writer,	make(map[int]int), colour}
}

func (hm *Heatmap) AddEvent(x, y int) {
	hm.Writer.CurrentImage.DrawCell(x, y, hm.Color)
}

func (hm *Heatmap) Write() {
	hm.Writer.FinishImage()
}

func GenerateHeatmaps(gen *Generation) []*Heatmap {
	var results [NumberOfHeatmapTypes]*Heatmap
	results[MovesMap] = NewHeatmap("moves", gen.Arena, color.RGBA{0, 213, 255, TRANSPARENCY}) // teal
	results[ShotsMap] = NewHeatmap("shots", gen.Arena, color.RGBA{255, 136, 0, TRANSPARENCY}) // orange
	results[KillsMap] = NewHeatmap("kills", gen.Arena, color.RGBA{255, 0, 0, TRANSPARENCY})   // red
	results[WaitsMap] = NewHeatmap("waits", gen.Arena, color.RGBA{0, 255, 0, TRANSPARENCY})   // green

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
