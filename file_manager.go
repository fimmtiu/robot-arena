package main

import (
	"bufio"
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
type CellProcessor func(x, y, moves, shots, kills, waits int)

var scriptIdRegexp = regexp.MustCompile(`/(\d+).l$`)
var generationRegexp = regexp.MustCompile(`/gen_(\d+)$`)

func NewFileManager(scenario string, generation int) *FileManager {
	fm := &FileManager{scenario, generation, make([]int, 0, SCRIPTS_PER_GENERATION)}

	if err := os.MkdirAll(fm.SimpleScriptsDir(), 0755); err != nil {
		logger.Fatalf("Failed to create directory %s: %v", fm.ScriptsDir(), err)
	}

	fm.ReadScriptIds()
	return fm
}

func (fm *FileManager) ReadScriptIds() {
	pattern := fm.ScriptsDir() + "/*.l"
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

func (fm *FileManager) SimpleScriptsDir() string {
	return fmt.Sprintf("scenario/%s/gen_%d/scripts/simple", fm.Scenario, fm.Generation)
}

func (fm *FileManager) WriteNewScript(code string) {
	highestId := 1
	if len(fm.ScriptIds) > 0 {
		highestId = fm.ScriptIds[len(fm.ScriptIds)-1]
		highestId++
	}
	fm.ScriptIds = append(fm.ScriptIds, highestId)

	path := fmt.Sprintf("%s/%d.l", fm.ScriptsDir(), highestId)
	fm.WriteFile(path, code)

	tree := ParseScript(code)
	tree2 := ParseScript(code)
	ts := tree.Size()
	t2s := tree2.Size()
	SimplifyTree(tree)
	if ts < t2s {
		logger.Printf("Shrunk script %d (%d - %d = %d)", highestId, t2s, ts, t2s - ts)
	}
	path = fmt.Sprintf("%s/%d.l", fm.SimpleScriptsDir(), highestId)
	fm.WriteFile(path, FormatScript(tree))
}

func (fm *FileManager) WriteFile(path string, contents string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		logger.Fatalf("Can't open new script file %v: %v", path, err)
	}
	f.WriteString(contents)
	f.Close()
}

func (fm *FileManager) ScriptCode(id int) string {
	path := fmt.Sprintf("%s/%d.l", fm.ScriptsDir(), id)
	source, err := os.ReadFile(path)
	if err != nil {
		logger.Fatalf("Couldn't read script %s: %v", path, err)
	}
	return string(source)
}

func (fm *FileManager) AverageScriptSize() int {
	sum := 0
	for _, id := range fm.ScriptIds {
		path := fmt.Sprintf("%s/%d.l", fm.ScriptsDir(), id)
		stat, err := os.Stat(path)
		if err != nil {
			logger.Fatalf("Couldn't stat script %s: %v", path, err)
		}
		sum += int(stat.Size())
	}

	return sum / len(fm.ScriptIds)
}

func (fm *FileManager) LoadScript(state *GameState, id int) Script {
	source := fm.ScriptCode(id)
	return Script{ParseScript(source), state}
}

// X,Y (1 byte each), then moves, shots, kills, waits at 4 bytes each. Actual size will be packed smaller.
const MAX_BYTES_PER_CELL = 2 + 4 + 4 + 4 + 4

func (fm *FileManager) WriteCellStatistics(arena *Arena) {
	path := fmt.Sprintf("scenario/%s/gen_%d/cells.csv", fm.Scenario, fm.Generation)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Fatalf("Couldn't open %s for writing: %v", path, err)
	}
	defer file.Close()

	file.WriteString("x,y,moves,shots,kills,waits\n")

	for _, cell := range arena.Cells {
		if cell.Moves > 0 || cell.Shots > 0 || cell.Kills > 0 || cell.Waits > 0 {
			file.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d\n", cell.X, cell.Y, cell.Moves, cell.Shots, cell.Kills, cell.Waits))
		}
	}
}

func (fm *FileManager) WriteMatchOutcome(match *Match) {
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
		strColumns := strings.Split(strings.TrimSpace(row), ",")
		columns := make([]int, len(strColumns))
		for i, str := range strColumns {
			columns[i] = strToInt(str)
		}
		callback(columns[0], columns[1], columns[2], columns[3], columns[4], columns[5])
	}

	file.Close()
}

func (fm *FileManager) EachCellRow(callback CellProcessor) {
	path := fmt.Sprintf("scenario/%s/gen_%d/cells.csv", fm.Scenario, fm.Generation)
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
		strColumns := strings.Split(strings.TrimSpace(row), ",")
		columns := make([]int, len(strColumns))
		for i, str := range strColumns {
			columns[i] = strToInt(str)
		}
		callback(columns[0], columns[1], columns[2], columns[3], columns[4], columns[5])
	}

	file.Close()
}

func (fm *FileManager) FindScriptIds(matchId int) (int, int) {
	scriptA, scriptB := -1, -1

	fm.EachResultRow(func (m, a, b, _, _, _ int) {
		if m == matchId {
			scriptA, scriptB = a, b
		}
	})

	if scriptA < 0 || scriptB < 0 {
		logger.Fatalf("Can't find a match in generation %d with id %d!", fm.Generation, matchId)
	}
	return scriptA, scriptB
}
