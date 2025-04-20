package gameplay

import (
	"example/depths/internal/common"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	wallModel       rl.Model
	wallCornerModel rl.Model
)

func InitWall() {
	wallModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall.obj"))
	rl.SetMaterialTexture(wallModel.Materials, rl.MapDiffuse, dungeonTexture)

	wallCornerModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall_corner.obj"))
	rl.SetMaterialTexture(wallCornerModel.Materials, rl.MapDiffuse, dungeonTexture)
}

// Draw all walls
func DrawWalls() {
	const (
		wallLen   = 4.
		wallthick = 1. / 2.
		wallboty  = (1. / 1.85)
	)
	var (
		rotationAxis = rl.Vector3{X: 0, Y: 1, Z: 0} // Y-axis
		scale        = common.Vector3One
		tint         = rl.White
	)
	// Draw walls along Z axis
	for i := -float32(floor.Size.Z/2) + wallLen/2; i < float32(floor.Size.Z/2); i += wallLen {
		pos1 := rl.NewVector3(floor.Position.X-floor.Size.X/2-wallthick, floor.Position.Y+wallboty, floor.Position.Z+i) // left back->front plane (-X +-Z)
		pos2 := rl.NewVector3(floor.Position.X+floor.Size.X/2+wallthick, floor.Position.Y+wallboty, floor.Position.Z+i) // right back->front plane (+X +-Z)
		rl.DrawModelEx(wallModel, pos1, rotationAxis, 90, scale, tint)
		rl.DrawModelEx(wallModel, pos2, rotationAxis, 90, scale, tint)
	}
	// Draw walls along X axis
	for i := -float32(floor.Size.X/2) + wallLen/2; i < float32(floor.Size.X/2); i += wallLen {
		pos1 := rl.NewVector3(floor.Position.X-i, floor.Position.Y+wallboty, floor.Position.Z-floor.Size.Z/2-wallthick) // back left->right plane (+-X -Z)
		pos2 := rl.NewVector3(floor.Position.X+i, floor.Position.Y+wallboty, floor.Position.Z+floor.Size.Z/2+wallthick) // front left->right plane (+-X +Z)
		rl.DrawModelEx(wallModel, pos1, rotationAxis, 180, scale, tint)
		rl.DrawModelEx(wallModel, pos2, rotationAxis, 180, scale, tint)
	}
	// Draw 4 wall corners
	bottomLeft := rl.NewVector3(floor.Position.X-floor.Size.X/2-wallthick, floor.Position.Y+wallboty, floor.Position.Z+floor.Size.Z/2+wallthick)
	bottomRight := rl.NewVector3(floor.Position.X+floor.Size.X/2+wallthick, floor.Position.Y+wallboty, floor.Position.Z+floor.Size.Z/2+wallthick)
	topRight := rl.NewVector3(floor.Position.X+floor.Size.X/2+wallthick, floor.Position.Y+wallboty, floor.Position.Z-floor.Size.Z/2-wallthick)
	topLeft := rl.NewVector3(floor.Position.X-floor.Size.X/2-wallthick, floor.Position.Y+wallboty, floor.Position.Z-floor.Size.Z/2-wallthick)
	rl.DrawModelEx(wallCornerModel, topRight, rotationAxis, 0, scale, tint)
	rl.DrawModelEx(wallCornerModel, topLeft, rotationAxis, 90, scale, tint)
	rl.DrawModelEx(wallCornerModel, bottomLeft, rotationAxis, 180, scale, tint)
	rl.DrawModelEx(wallCornerModel, bottomRight, rotationAxis, 270, scale, tint)
}
