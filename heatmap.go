package main

import "image/color"

type Heatmap struct {
	Writer ImageWriter
	Cells map[int]int
	Color color.RGBA
	Gen *Generation // FIXME not used any more, remove this
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

func NewHeatmap(name string, colour color.RGBA, gen *Generation) *Heatmap {
	return &Heatmap{
		NewImageWriter(name, HEATMAP_PIXELS_PER_CELL),
		make(map[int]int), colour, gen,
	}
}

func (hm *Heatmap) AddEvent(x, y int) {
	hm.Writer.CurrentImage.DrawCell(x, y, hm.Color)
}

func GenerateHeatmaps(gen *Generation) []*Heatmap {
	var results [NumberOfHeatmapTypes]*Heatmap
	results[MovesMap] = NewHeatmap("moves", color.RGBA{0, 213, 255, TRANSPARENCY}, gen) // teal
	results[ShotsMap] = NewHeatmap("shots", color.RGBA{255, 136, 0, TRANSPARENCY}, gen) // orange
	results[KillsMap] = NewHeatmap("kills", color.RGBA{255, 0, 0, TRANSPARENCY}, gen)   // red
	results[WaitsMap] = NewHeatmap("waits", color.RGBA{0, 255, 0, TRANSPARENCY}, gen)   // green

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

	return results[:]
}
