package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type FileManager struct {
	Scenario string
	Generation int
	ScriptIds []int
}

var scriptIdRegexp = regexp.MustCompile(`/(\d+).l$`)
var generationRegexp = regexp.MustCompile(`/gen_(\d+)$`)

func NewFileManager(scenario string, generation int) *FileManager {
	fm := &FileManager{scenario, generation, make([]int, SCRIPTS_PER_GENERATION)}

	path := fmt.Sprintf("scenario/%s/gen_%d/scripts", scenario, generation)
	if err := os.MkdirAll(path, 0755); err != nil {
		logger.Fatalf("Failed to create directory %s: %v", path, err)
	}

	pattern := fmt.Sprintf("scenario/%s/gen_%d/scripts/*", fm.Scenario, fm.Generation)
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		logger.Fatalf("Can't glob %s: %v", pattern, err)
	}

	for _, filename := range filenames {
		submatches := scriptIdRegexp.FindStringSubmatch(filename)
		if len(submatches) != 2 {
			logger.Fatalf("Unparseable name in scripts directory: %v", filename)
		}
		number, err := strconv.Atoi(submatches[1])
		if err != nil {
			logger.Fatalf("Can't convert name to number: %v, %v", filename, err)
		}

		fm.ScriptIds = append(fm.ScriptIds, number)
	}
	sort.Ints(fm.ScriptIds)

	return fm
}

func currentHighestGeneration(scenario string) int {
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
		number, err := strconv.Atoi(submatches[0])
		if err != nil {
			logger.Fatalf("Can't convert name to number: %v, %v", dirname, err)
		}
		logger.Printf("    dirname '%s' => %d", dirname, number)

		generations = append(generations, number)
	}

	sort.Ints(generations)
	logger.Printf("generations int: %v", generations)
	return generations[len(generations)-1]
}

func (fm *FileManager) NewScriptFile() *os.File {
	highestId := fm.ScriptIds[len(fm.ScriptIds)-1]
	highestId++

	path := fmt.Sprintf("scenario/%s/gen_%d/scripts/%d.l", fm.Scenario, fm.Generation, highestId)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		logger.Fatalf("Can't open new script file %v: %v", path, err)
	}

	fm.ScriptIds = append(fm.ScriptIds, highestId)
	return f
}

func (fm *FileManager) LoadScript(state *GameState, id int) Script {
	path := fmt.Sprintf("scenario/%s/gen_%d/scripts/%d.l", fm.Scenario, fm.Generation, id)
	source, err := os.ReadFile(path)
	if err != nil {
		logger.Fatalf("Couldn't read script %s: %v", path, err)
	}

	return Script{ParseScript(string(source)), state}
}

// X,Y (1 byte each), then moves, shots, kills, waits at 4 bytes each. Actual size will be packed smaller.
const MAX_BYTES_PER_CELL = 2 + 4 + 4 + 4 + 4

func (fm *FileManager) RecordResults(match *Match) error {
	if err := fm.writeCellStatistics(match); err != nil {
		return err
	}

	if err := fm.writeMatchOutcome(match); err != nil {
		return err
	}

	return nil
}

func (fm *FileManager) writeCellStatistics(match *Match) error {
	arena := match.State.Arena
	path := fmt.Sprintf("scenario/%s/gen_%d/cells", fm.Scenario, fm.Generation)
	buf := make([]byte, 0, arena.Width * arena.Height * MAX_BYTES_PER_CELL)

	for _, cell := range arena.Cells {
		if cell.Moves == 0 && cell.Shots == 0 && cell.Kills == 0 && cell.Waits == 0 {
			continue
		}
		buf = binary.AppendUvarint(buf, uint64(cell.X))
		buf = binary.AppendUvarint(buf, uint64(cell.Y))
		buf = binary.AppendUvarint(buf, uint64(cell.Moves))
		buf = binary.AppendUvarint(buf, uint64(cell.Shots))
		buf = binary.AppendUvarint(buf, uint64(cell.Kills))
		buf = binary.AppendUvarint(buf, uint64(cell.Waits))
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		logger.Fatalf("Couldn't open %s for appending: %v", path, err)
	}
	written, err := file.Write(buf)
	if err != nil {
		logger.Fatalf("Couldn't write %d bytes to %s: %v", len(buf), path, err)
	}
	if written < len(buf) {
		logger.Fatalf("Only wrote %d of %d bytes to %s!", written, len(buf), path)
	}

	logger.Printf("Wrote %d bytes to cell statistics file.", written)
	return nil
}

func (fm *FileManager) writeMatchOutcome(match *Match) error {
	path := fmt.Sprintf("scenario/%s/gen_%d/results.csv", fm.Scenario, fm.Generation)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	stat, err := file.Stat()
	if (err != nil && errors.Is(err, fs.ErrNotExist)) || stat.Size() == 0 {
		file.WriteString("matchId,scriptA,scriptB,scoreA,scoreB,ticks\n")
	} else if err != nil {
		logger.Fatalf("Can't stat %s: %v", path, err)
	}

	row := fmt.Sprintf("%d,%d,%d,%d,%d,%d\n", match.Id, match.ScriptA, match.ScriptB,
											match.Scores[TeamA], match.Scores[TeamB], match.State.Tick)
	written, err := file.WriteString(row)
	if err != nil {
		logger.Fatalf("Couldn't write %d characters to %s: %v", len(row), path, err)
	}
	if written < len(row) {
		logger.Fatalf("Only wrote %d of %d characters to %s!", written, len(row), path)
	}

	logger.Printf("Wrote %d characters to results CSV.", written)
	return nil
}
