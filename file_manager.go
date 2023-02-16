package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type FileManager struct {
	Scenario string
	Generation int
}

func NewFileManager(scenario string, generation int) *FileManager {
	fm := &FileManager{scenario, generation}

	path := fmt.Sprintf("scenario/%s/gen_%d/scripts", scenario, generation)
	if err := os.MkdirAll(path, 0755); err != nil {
		logger.Fatalf("Failed to create directory %s: %v", path, err)
	}

	return fm
}

func (fm *FileManager) LoadScript(id int) *ScriptNode {
	path := fmt.Sprintf("scenario/%s/gen_%d/scripts/%d.l", fm.Scenario, fm.Generation, id)
	source, err := os.ReadFile(path)
	if err != nil {
		logger.Fatalf("Couldn't read script %s: %v", path, err)
	}

	return ParseScript(string(source))
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
	path := fmt.Sprintf("scenario/%s/gen_%d/cells", fm.Scenario, fm.Generation)
	buf := make([]byte, 0, match.Arena.Width * match.Arena.Height * MAX_BYTES_PER_CELL)

	for _, cell := range match.Arena.Cells {
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

	row := fmt.Sprintf("%d,%d,%d,%d,%d,%d\n", match.Id, match.ScriptA, match.ScriptB, match.ScoreA, match.ScoreB, match.Tick)
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
