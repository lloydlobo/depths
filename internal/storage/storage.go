package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type GameStorageLevelJSON struct {
	Version string `json:"version"`
	LevelID int32  `json:"levelID"`

	Data map[string]any `json:"data"`

	// ....
}

// TODO: Do not overwrite existing hiscore if current is less
//
//	It should be handled by game logic that loads level and applies/overwrites
//	g.Level struct with hiscore
func SaveStorageLevel(l GameStorageLevelJSON) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	saveDir := filepath.Join(cwd, "storage")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", saveDir, err)
	}
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(l.LevelID))+".json")
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("create %q: %w", name, err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(l); err != nil {
		return fmt.Errorf("encode level: %w", err)
	}
	return nil
}

// Extended
// filetag is added as a version suffix. e.g. filetag="blocks" => level_1_blocks.json
func SaveStorageLevelEx(l GameStorageLevelJSON, filetag string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	saveDir := filepath.Join(cwd, "storage")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", saveDir, err)
	}
	if len(filetag) == 0 {
		filetag = ""
	} else {
		if filetag[0] != '_' {
			filetag = "_" + filetag
		}
	}
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(l.LevelID))+filetag+".json")
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("create %q: %w", name, err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(l); err != nil {
		return fmt.Errorf("encode level: %w", err)
	}
	return nil
}

func LoadStorageLevel(ID int32) (*GameStorageLevelJSON, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	saveDir := filepath.Join(cwd, "storage")
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(ID))+".json")
	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", name, err)
	}
	var l GameStorageLevelJSON
	dec := json.NewDecoder(f)
	if err := dec.Decode(&l); err != nil {
		return nil, fmt.Errorf("encode level: %w", err)
	}

	return &l, nil
}
