package storage

import (
	"example/depths/internal/block"
	"example/depths/internal/floor"
	"example/depths/internal/player"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GameEntityData struct {
	LevelID int32 `json:"levelID"`

	Camera        rl.Camera3D `json:"camera"`
	FinishScreen  int         `json:"finishScreen"`
	FramesCounter int32       `json:"framesCounter"`

	// FIXME: Floor should be in additional GDT (SINCE FLOOR DIMENSIONS CHANGES BASED ON SCREEN)
	XFloor                 floor.Floor   `json:"xFloor"`
	XPlayer                player.Player `json:"xPlayer"`
	HasPlayerLeftDrillBase bool          `json:"hasPlayerLeftDrillBase"`
}

type GameAdditionalData struct {
	LevelID int32

	Blocks []block.Block `json:"blocks"`
}

type GameLogicData struct {
	LevelID int32

	Money      int32 `json:"money"`
	Experience int32 `json:"experience"`
	HitScore   int32 `json:"hitScore"`
	HitCount   int32 `json:"hitCount"`
}

type GameDataType int

const (
	EntityGDT GameDataType = iota
	LogicGDT
	AdditionalGDT
)

var (
	GameDataTypeToStringMap = map[GameDataType]string{
		EntityGDT:     "entity",
		LogicGDT:      "logic",
		AdditionalGDT: "additional",
	}
)
