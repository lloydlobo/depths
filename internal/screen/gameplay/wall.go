package gameplay

import (
	"example/depths/internal/common"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	wallModel     rl.Model
	wallHalfModel rl.Model
)

func InitWall() {
	// wallModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall.obj"))
	wallModel = common.Model.OBJ.Wall
	wallModel = common.Model.OBJ.WallNarrow
	wallModel = common.Model.OBJ.WallOpening
	wallModel = common.Model.OBJ.WallHalf

	rl.SetMaterialTexture(wallModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)

	// wallCornerModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall_corner.obj"))
	wallHalfModel = common.Model.OBJ.Wall
	rl.SetMaterialTexture(wallHalfModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
}

func DrawWalls(pos, size, scale rl.Vector3) {
	wallLen := float32(1.)
	wallthick := float32(1. / 2.)
	wallboty := size.Y

	// Draw walls along each X & Z axis
	var (
		rotationAxis = rl.Vector3{X: 0, Y: 1, Z: 0} // Y-axis
		tint         = rl.White
	)
	for i := -float32(size.X/2) + wallLen/2; i < float32(size.X/2); i += 1 {
		pos1 := rl.NewVector3(pos.X-i, pos.Y+wallboty, pos.Z-size.Z/2-wallthick) // back left->right plane (+-X -Z)
		pos2 := rl.NewVector3(pos.X+i, pos.Y+wallboty, pos.Z+size.Z/2+wallthick) // front left->right plane (+-X +Z)
		rl.DrawModelEx(wallModel, pos1, rotationAxis, 180, scale, tint)
		rl.DrawModelEx(wallModel, pos2, rotationAxis, 0, scale, tint)
	}
	for i := -float32(size.Z/2) + wallLen/2; i < float32(size.Z/2); i += 1 {
		pos1 := rl.NewVector3(pos.X-size.X/2-wallthick, pos.Y+wallboty, pos.Z+i) // left back->front plane (-X +-Z)
		pos2 := rl.NewVector3(pos.X+size.X/2+wallthick, pos.Y+wallboty, pos.Z+i) // right back->front plane (+X +-Z)
		rl.DrawModelEx(wallModel, pos1, rotationAxis, -90, scale, tint)
		rl.DrawModelEx(wallModel, pos2, rotationAxis, 90, scale, tint)
	}

	// Draw 4 wall corners
	bottomLeft := rl.NewVector3(pos.X-size.X/2-wallthick, pos.Y+wallboty, pos.Z+size.Z/2+wallthick)
	bottomRight := rl.NewVector3(pos.X+size.X/2+wallthick, pos.Y+wallboty, pos.Z+size.Z/2+wallthick)
	topRight := rl.NewVector3(pos.X+size.X/2+wallthick, pos.Y+wallboty, pos.Z-size.Z/2-wallthick)
	topLeft := rl.NewVector3(pos.X-size.X/2-wallthick, pos.Y+wallboty, pos.Z-size.Z/2-wallthick)

	rl.DrawModelEx(wallHalfModel, topRight, rotationAxis, 0, scale, tint)
	rl.DrawModelEx(wallHalfModel, topLeft, rotationAxis, 90, scale, tint)
	rl.DrawModelEx(wallHalfModel, bottomLeft, rotationAxis, 180, scale, tint)
	rl.DrawModelEx(wallHalfModel, bottomRight, rotationAxis, 270, scale, tint)
}
