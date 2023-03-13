package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
)

// A basic implementation of Bresenham's line drawing algorithm. Calls the user-provided callback on each coordinate
// pair that's part of the line. This will be used for tons of stuff, from calculating visibility to drawing actual
// lines on an image.
//
// The callback gets (x, y) coords of each grid cell on the line. If it returns false, we stop executing immediately and
// return false. Otherwise, we return true at the end of the line.
func BresenhamLine(x0, y0, x1, y1 int, callback func(x, y int) bool) bool {
	var delta_x, delta_y, sign_x, sign_y int
	flip := false

	delta_x = intAbs(x1 - x0)
	if x0 < x1 {
		sign_x = 1
	} else {
		sign_x = -1
	}
	delta_y = intAbs(y1 - y0)
	if y0 < y1 {
		sign_y = 1
	} else {
		sign_y = -1
	}

	if delta_x < delta_y {
		temp := delta_x
		delta_x = delta_y
		delta_y = temp
		flip = true
	}

	a := 2 * delta_y
	e := a - delta_x
	b := a - 2 * delta_x

	if callback(x0, y0) == false {
		return false
	}
	for i := 1; i < delta_x + 1; i++ {
		if e < 0 {
			if flip {
				y0 += sign_y
			} else {
				x0 += sign_x
			}
			e += a
		} else {
			x0 += sign_x
			y0 += sign_y
			e += b
		}
		if callback(x0, y0) == false {
			return false
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
// This, plus the symmetry of the map, allows a script to run identically regardless of whether it's controlling Team A
// or Team B.
func relativeToAbsoluteDirection(relative Direction, team Team) Direction {
	relative = Direction(intAbs(int(relative)))
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

func strToInt(s string) int {
	number, err := strconv.Atoi(s)
	if err != nil {
		logger.Fatalf("Can't convert string to number: \"%s\", %v", s, err)
	}
	return number
}

func CurrentHighestGeneration(scenario string) int {
	generations := []int{}
	pattern := fmt.Sprintf("scenario/%s/gen_*", scenario)
	dirnames, err := filepath.Glob(pattern)
	if err != nil {
		logger.Fatalf("Can't glob %s: %v", pattern, err)
	}

	for _, dirname := range dirnames {
		submatches := scriptIdRegexp.FindStringSubmatch(dirname)
		if len(submatches) != 2 {
			logger.Fatalf("Unparseable name in scenario directory: %v", dirname)
		}
		number := strToInt(submatches[0])
		logger.Printf("    dirname '%s' => %d", dirname, number)

		generations = append(generations, number)
	}

	sort.Ints(generations)
	logger.Printf("generations int: %v", generations)
	return generations[len(generations)-1]
}
