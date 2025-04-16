package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type LevelPlaceholder struct {
	ID int32

	// ....
}

// TODO: Do not overwrite existing hiscore if current is less
//
//	It should be handled by game logic that loads level and applies/overwrites
//	g.Level struct with hiscore
func SaveStorageLevel(l LevelPlaceholder) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	saveDir := filepath.Join(cwd, "storage")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", saveDir, err)
	}
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(l.ID))+".json")
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

func LoadStorageLevel(ID int32) (*LevelPlaceholder, error) {
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
	var l LevelPlaceholder
	dec := json.NewDecoder(f)
	if err := dec.Decode(&l); err != nil {
		return nil, fmt.Errorf("encode level: %w", err)
	}

	return &l, nil
}
