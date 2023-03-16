package main

import (
	"image"
	_ "image/png"
	"os"
)

type Team int
const (
	TeamA Team = iota
	TeamB
)

type CellType uint8
const (
	OpenCell CellType = iota
	WallCell
	SpawnCell
	GoalCell
)

type Cell struct {
	X int
	Y int
	Type CellType
	Team Team
	VisibleCells map[*Cell]bool

	// Statistics for building histogram maps of activity in the arena.
	Moves uint
	Shots uint
	Kills uint
	Waits uint
}

// I've divided this behaviour in case we want to introduce blocks that are shootable but not walkable in the future
// (open pits?), or if we want glass walls that you can see through but not shoot through, etc.
func (c *Cell) BotsCanPass() bool {
	return c.Type != WallCell && c.Type != GoalCell
}

func (c *Cell) ShotsCanPass() bool {
	return c.Type != WallCell && c.Type != GoalCell
}

func (c *Cell) BlocksVision() bool {
	return c.Type == WallCell
}

func (c *Cell) VisibleFrom(c2 *Cell) bool {
	_, found := c.VisibleCells[c2]
	return found
}

type Arena struct {
	Width int
	Height int
	Cells []Cell
	Spawns [2][]*Cell
	Goals [2]*Cell
}

func LoadArena(filename string) (a *Arena) {
	f, err := os.Open(filename)
	if err != nil {
		logger.Fatalf("Couldn't open '%s': %v", filename, err)
	}
	img, format, err := image.Decode(f)
	if err != nil {
		logger.Fatalf("Couldn't parse %s file '%s': %v", format, filename, err)
	}

	return NewArena(img)
}

func NewArena(img image.Image) *Arena {
	a := &Arena{
		Width: img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
		Cells: make([]Cell, img.Bounds().Dx() * img.Bounds().Dy()),
		Spawns: [2][]*Cell{make([]*Cell, 0, BOTS_PER_TEAM), make([]*Cell, 0, BOTS_PER_TEAM)},
		Goals: [2]*Cell{},
	}

	// Translate image pixels to cells
	for x := 0; x < a.Width; x++ {
		for y := 0; y < a.Height; y++ {
			cell := &a.Cells[x * a.Height + y]
			cell.X = x
			cell.Y = y

			red, green, blue, _ := img.At(x, y).RGBA()
			if red == 0 && green == 0 && blue == 0 {  // Black pixels indicate walls.
				cell.Type = WallCell
				} else if red == 65535 && green == 65535 && blue == 65535 { // White pixels indicate open space.
				cell.Type = OpenCell
			} else if red == 65535 { // Red pixels indicate spawn points.
				cell.Type = SpawnCell
				cell.Team = intToTeam(green)
				a.Spawns[cell.Team] = append(a.Spawns[cell.Team], cell)
			} else if green == 65535 { // Green pixels indicate the goal points for each team.
				cell.Type = GoalCell
				cell.Team = intToTeam(red)
				a.Goals[cell.Team] = cell
			} else {
				logger.Fatalf("Unknown color in image at (%d, %d): %02x, %02x, %02x", x, y, red, green, blue)
			}
		}
	}

	a.verifyValidArena()
	a.calculateVisibility()
	return a
}

func intToTeam(color uint32) Team {
	if color == 0 {
		return TeamA
	}
	return TeamB
}

// Pre-calculate visibility for all cells. It's slow (n^2), but it only happens once on startup.
// There are a lot of optimizations we could do here if we need to.
func (a *Arena) calculateVisibility() {
	visibleCells := 0

	for i := 0; i < len(a.Cells); i++ {
		cell := &a.Cells[i]
		cell.VisibleCells = make(map[*Cell]bool, 0)

		for j := 0; j < len(a.Cells); j++ {
			otherCell := &a.Cells[j]
			BresenhamLine(cell.X, cell.Y, otherCell.X, otherCell.Y, func (x, y int) bool {
				c := &a.Cells[x * a.Height + y]
				if c.BlocksVision() {
					return false
				} else {
					visibleCells++
					cell.VisibleCells[c] = true
					return true
				}
			})
		}
	}
	logger.Printf("Created %d cell visibility links for %d cells.", visibleCells, len(a.Cells))
}

func (a *Arena) CanSee(src *Cell, dest *Cell) bool {
	return src.VisibleFrom(dest)
}

// Manhattan distance between two points. Working with ints keeps things simple.
func (a *Arena) Distance(src *Cell, dest *Cell) int {
	return intAbs(src.X - dest.X) + intAbs(src.Y - dest.Y)
}

func (a *Arena) Reset() {
	for i := 0; i < len(a.Cells); i++ {
		a.Cells[i].Moves = 0
		a.Cells[i].Shots = 0
		a.Cells[i].Kills = 0
		a.Cells[i].Waits = 0
	}
}

// Can a unit in cell `c` move in direction `dir`, or is it blocked by a wall? Returns the destination cell if it's a
// valid move, or the current cell if the move is blocked by a wall or map border.
// (We also need to test for the presence of another unit in the destination space, but Match does that.)
func (a *Arena) DestinationCellAfterMove(c *Cell, dir Direction) *Cell {
	switch dir {
	case North:
		if c.Y > 0 && a.Cells[c.X * a.Height + c.Y - 1].BotsCanPass() {
			return &a.Cells[c.X * a.Height + c.Y - 1]
		}
	case South:
		if c.Y < a.Height - 1 && a.Cells[c.X * a.Height + c.Y + 1].BotsCanPass() {
			return &a.Cells[c.X * a.Height + c.Y + 1]
		}
	case East:
		if c.X < a.Width - 1 && a.Cells[(c.X + 1) * a.Height + c.Y].BotsCanPass() {
			return &a.Cells[(c.X + 1) * a.Height + c.Y]
		}
	case West:
		if c.X > 0 && a.Cells[(c.X - 1) * a.Height + c.Y].BotsCanPass() {
			return &a.Cells[(c.X - 1) * a.Height + c.Y]
		}
	}
	return c
}

// Verify that the map has five spawns and one goal for each team.
func (a *Arena) verifyValidArena() {
	var a_spawns, b_spawns, a_goals, b_goals int

	for i := 0; i < len(a.Cells); i++ {
		cell := &a.Cells[i]
		if cell.Type == SpawnCell && cell.Team == TeamA {
			a_spawns++
		} else if cell.Type == SpawnCell && cell.Team == TeamB {
			b_spawns++
		} else if cell.Type == GoalCell && cell.Team == TeamA {
			a_goals++
		} else if cell.Type == GoalCell && cell.Team == TeamB {
			b_goals++
		}
	}

	if a_spawns != BOTS_PER_TEAM || b_spawns != BOTS_PER_TEAM || a_goals != 1 || b_goals != 1 {
		logger.Fatalf("Bogus map! %d A spawns, %d B spawns, %d A goals, %d B goals.", a_spawns, b_spawns, a_goals, b_goals)
	}
}
