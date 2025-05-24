package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SavedgameSlotDataType struct {
	Version          string    `json:"version"`
	SlotID           uint8     `json:"slotID"`
	ModifiedAt       time.Time `json:"modifiedAt"`
	CreatedAt        time.Time `json:"createdAt"`
	AllLevelIDS      []uint8   `json:"allLevelIDS"`
	UnlockedLevelIDS []uint8   `json:"unlockedLevelIDS"`
	CurrentLevelID   uint8     `json:"currentLevelID"`
}

func LoadSavegameSlot(slotID uint8) (*SavedgameSlotDataType, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	saveDir := filepath.Join(cwd, "storage", "savegame", "slot")
	fname := filepath.Join(saveDir, fmt.Sprintf("%d.json", slotID))

	buf, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	var sg *SavedgameSlotDataType

	if err := json.Unmarshal(buf, &sg); err != nil {
		return nil, err
	}

	return sg, nil
}

func SaveSavegameSlot(slotID uint8, data SavedgameSlotDataType) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	saveDir := filepath.Join(cwd, "storage", "savegame", "slot")

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", saveDir, err)
	}

	fname := filepath.Join(saveDir, fmt.Sprintf("%d.json", slotID))

	// name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(l.LevelID))+".json")
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("create %q: %w", fname, err)
	}

	b := Must(json.MarshalIndent(data, "", ""))
	_ = Must(f.Write(b))

	// enc := json.NewEncoder(f)
	// if err := enc.Encode(data); err != nil {
	// 	return fmt.Errorf("encode level: %w", err)
	// }

	return nil
}
