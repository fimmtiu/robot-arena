package main

import "image/color"

type Heatmap struct {
	Writer ImageWriter
	Filename string
	Generation *Generation
	Color color.RGBA
}

const (
	MovesMap ResultType = iota
	ShotsMap
	KillsMap
	WaitsMap
	NumberOfHeatmapTypes
)

const HEATMAP_PIXELS_PER_CELL = 8 // FIXME: shrink to 6 once it's working
const TRANSPARENCY = 25 // out of 255; 90% transparent

func NewHeatmap(name string, gen *Generation, colour color.RGBA) *Heatmap {
	writer := NewImageWriter(name, HEATMAP_PIXELS_PER_CELL)
	writer.StartImage(gen.Arena)
	return &Heatmap{writer, writer.CurrentImage.Filename,	gen, colour}
}

func (hm *Heatmap) AddEvent(x, y int) {
	hm.Writer.CurrentImage.DrawCell(x, y, hm.Color)
}

func (hm *Heatmap) Write() {
	hm.Writer.FinishImage()
}

// TODO: Don't generate a heatmap if the destination file already exists.
func GenerateHeatmaps(gen *Generation) []*Heatmap {
	var results [NumberOfHeatmapTypes]*Heatmap
	results[MovesMap] = NewHeatmap("moves", gen, color.RGBA{0, 213, 255, TRANSPARENCY}) // teal
	results[ShotsMap] = NewHeatmap("shots", gen, color.RGBA{255, 136, 0, TRANSPARENCY}) // orange
	results[KillsMap] = NewHeatmap("kills", gen, color.RGBA{255, 0, 0, TRANSPARENCY})   // red
	results[WaitsMap] = NewHeatmap("waits", gen, color.RGBA{0, 255, 0, TRANSPARENCY})   // green

	gen.FileManager.EachCell(func (x, y, moves, shots, kills, waits int) {
		for i := 0; i < moves; i++ {
			results[MovesMap].AddEvent(x, y)
		}
		// for i := 0; i < shots; i++ {          // FIXME: Uncomment once it's working
		// 	results[ShotsMap].AddEvent(x, y)
		// }
		// for i := 0; i < kills; i++ {
		// 	results[KillsMap].AddEvent(x, y)
		// }
		// for i := 0; i < waits; i++ {
		// 	results[WaitsMap].AddEvent(x, y)
		// }
	})

	for _, hm := range results {
		hm.Write()
	}
	return results[:]
}
