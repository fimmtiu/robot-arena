package main

// A basic implementation of Bresenham's line drawing algorithm. Calls the user-provided callback on each coordinate
// pair that's part of the line. This will be used for tons of stuff, from calculating visibility to drawing actual
// lines on an image.
//
// The callback gets (x, y) coords of each grid cell on the line. If it returns false, we stop executing immediately and
// return false. Otherwise, we return true at the end of the line.
func BresenhamLine(x0, y0, x1, y1 int, callback func(x, y int) bool) bool {
	var dx, dy, sx, sy, error int

	dx = intAbs(x1 - x0)
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	dy = intAbs(y1 - y0)
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	error = dx + dy

	for {
		if callback(x0, y0) == false {
			return false
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * error
		if e2 >= dy {
			if x0 == x1 {
				break
			}
			error += dy
			x0 += sx
		}
		if e2 <= dx {
			if y0 == y1 {
				break
			}
			error += dx
			y0 += sy
		}
	}

	return true
}

func intAbs(n int) int {
	if n < 0 {
		return -n
	} else {
		return n
	}
}

// Each team considers "north" to be the direction of the enemy's goal, and "south" to be the direction of its own side.
// This, plus the horizontal and vertical symmetry of the map, allows a script to run identically regardless of whether
// it's controlling Team A or Team B.
func relativeToActualDirection(relative Direction, team Team) Direction {
	if team == TeamA {
		switch relative {
		case North: return East
		case South: return West
		case East:  return South
		case West:  return North
		}
	} else {
		switch relative {
		case North: return West
		case South: return East
		case East:  return North
		case West:  return South
		}
	}
	logger.Fatalf("Weird direction %d for team %d!", relative, team)
	return North
}
