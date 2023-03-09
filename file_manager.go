package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type FileManager struct {
	Scenario string
	Generation int
	ScriptIds []int
}

type ResultProcessor func(matchId, scriptA, scriptB, scoreA, scoreB, ticks int)

var scriptIdRegexp = regexp.MustCompile(`/(\d+).l$`)
var generationRegexp = regexp.MustCompile(`/gen_(\d+)$`)

func NewFileManager(scenario string, generation int) *FileManager {
	fm := &FileManager{scenario, generation, make([]int, SCRIPTS_PER_GENERATION)}

	if err := os.MkdirAll(fm.ScriptsDir(), 0755); err != nil {
		logger.Fatalf("Failed to create directory %s: %v", fm.ScriptsDir(), err)
	}

	pattern := fm.ScriptsDir() + "/*"
	filenames, err := filepath.Glob(pattern)
	if err != nil {
		logger.Fatalf("Can't glob %s: %v", pattern, err)
	}

	for _, filename := range filenames {
		submatches := scriptIdRegexp.FindStringSubmatch(filename)
		if len(submatches) != 2 {
			logger.Fatalf("Unparseable name in scripts directory: %v", filename)
		}
		fm.ScriptIds = append(fm.ScriptIds, strToInt(submatches[1]))
	}
	sort.Ints(fm.ScriptIds)

	return fm
}

func (fm *FileManager) GenerationDir() string {
	return fmt.Sprintf("scenario/%s/gen_%d", fm.Scenario, fm.Generation)
}

func (fm *FileManager) PreviousGenerationDir() string {
	if fm.Generation == 1 {
		logger.Fatal("Can't call PreviousGenerationDir when there's no previous generation!")
	}
	return fmt.Sprintf("scenario/%s/gen_%d", fm.Scenario, fm.Generation - 1)
}

func (fm *FileManager) ScriptsDir() string {
	return fmt.Sprintf("scenario/%s/gen_%d/scripts", fm.Scenario, fm.Generation)
}

func (fm *FileManager) currentHighestGeneration() int {
	generations := []int{}
	pattern := fmt.Sprintf("scenario/%s/gen_*", fm.Scenario)
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

func (fm *FileManager) NewScriptFile() *os.File {
	highestId := fm.ScriptIds[len(fm.ScriptIds)-1]
	highestId++

	path := fmt.Sprintf("%s/%d.l", fm.ScriptsDir(), highestId)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		logger.Fatalf("Can't open new script file %v: %v", path, err)
	}

	fm.ScriptIds = append(fm.ScriptIds, highestId)
	return f
}

func (fm *FileManager) ScriptCode(id int) string {
	path := fmt.Sprintf("%s/%d.l", fm.ScriptsDir(), id)
	source, err := os.ReadFile(path)
	if err != nil {
		logger.Fatalf("Couldn't read script %s: %v", path, err)
	}
	return string(source)
}

func (fm *FileManager) LoadScript(state *GameState, id int) Script {
	source := fm.ScriptCode(id)
	return Script{ParseScript(source), state}
}

// X,Y (1 byte each), then moves, shots, kills, waits at 4 bytes each. Actual size will be packed smaller.
const MAX_BYTES_PER_CELL = 2 + 4 + 4 + 4 + 4

func (fm *FileManager) RecordResults(match *Match) {
	fm.writeCellStatistics(match)
	fm.writeMatchOutcome(match)
}

func (fm *FileManager) writeCellStatistics(match *Match) {
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
}

func (fm *FileManager) writeMatchOutcome(match *Match) {
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
}

func (fm *FileManager) EachResultRow(callback ResultProcessor) {
	path := fmt.Sprintf("scenario/%s/gen_%d/results.csv", fm.Scenario, fm.Generation)
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		logger.Fatalf("Can't open %s: %v", path, err)
	}
	reader := bufio.NewReader(file)
	_, err = reader.ReadString('\n')
	if err != nil {
		logger.Fatalf("Can't read first line from %s: %v", path, err)
	}

	for {
		row, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Fatalf("Can't read line from %s: %v", path, err)
		}
		strColumns := strings.Split(row, ",")
		columns := make([]int, 0, len(strColumns))
		for i, str := range strColumns {
			columns[i] = strToInt(str)
		}
		callback(columns[0], columns[1], columns[2], columns[3], columns[4], columns[5])
	}

	file.Close()
}
