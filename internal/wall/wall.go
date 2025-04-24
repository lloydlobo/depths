package wall

import (
	"example/depths/internal/common"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	wallModel       rl.Model
	wallCornerModel rl.Model
)

func InitWall() {}

func SetupWallModel() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	// wallModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall.obj"))
	wallModel = common.Model.OBJ.Wall
	wallModel = common.Model.OBJ.Stairs
	wallModel = common.Model.OBJ.WallHalf
	wallModel = common.Model.OBJ.WallNarrow
	wallModel = common.Model.OBJ.Dirt
	wallModel = common.Model.OBJ.WallOpening

	rl.SetMaterialTexture(wallModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)

	wallCornerModel = common.Model.OBJ.Wall
	wallCornerModel = common.Model.OBJ.Rocks
	rl.SetMaterialTexture(wallCornerModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
}

// Use walls to avoid infinite-map generation
func DrawBatch(pos, size, scale rl.Vector3) {
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

	rl.DrawModelEx(wallCornerModel, topRight, rotationAxis, 0, scale, tint)
	rl.DrawModelEx(wallCornerModel, topLeft, rotationAxis, 90, scale, tint)
	rl.DrawModelEx(wallCornerModel, bottomLeft, rotationAxis, 180, scale, tint)
	rl.DrawModelEx(wallCornerModel, bottomRight, rotationAxis, 270, scale, tint)
}
