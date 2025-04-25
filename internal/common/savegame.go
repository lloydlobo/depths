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
