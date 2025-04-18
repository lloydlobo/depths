// Package common provides assets and resources initialized once (if possible)
// before the game is run.
package common

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	Font struct {
		Primary   rl.Font
		Secondary rl.Font
	}
	Music struct {
		Theme   rl.Music
		Ambient rl.Music
	}
	FX struct {
		Coin rl.Sound
	}
	Shader struct {
		PBR rl.Shader
	}
)

func GetBoundingBoxFromPositionSizeV(pos, size rl.Vector3) rl.BoundingBox {
	return rl.NewBoundingBox(
		rl.NewVector3(pos.X-size.X/2, pos.Y-size.Y/2, pos.Z-size.Z/2),
		rl.NewVector3(pos.X+size.X/2, pos.Y+size.Y/2, pos.Z+size.Z/2),
	)
}
