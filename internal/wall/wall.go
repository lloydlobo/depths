package wall

import (
	"example/depths/internal/common"
	"fmt"
	"path/filepath"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	wallModel       rl.Model
	wallCornerModel rl.Model
)

func InitWall() {}

func SetupWallModel(room common.RoomType) {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	switch room {
	case common.OpenWorldRoom:
		wallModel = common.ModelDungeonKit.OBJ.WallOpening
		rl.SetMaterialTexture(wallModel.Materials, rl.MapDiffuse, common.ModelDungeonKit.OBJ.Colormap)
		wallCornerModel = common.ModelDungeonKit.OBJ.Wall
		rl.SetMaterialTexture(wallCornerModel.Materials, rl.MapDiffuse, common.ModelDungeonKit.OBJ.Colormap)
	case common.DrillRoom:
		// PERF: Load once in common
		dir := filepath.Join("res", "kenney_prototype-kit", "Models")
		texture := rl.LoadTexture(filepath.Join(dir, "OBJ format", "Textures", "colormap.png"))
		wallModel = rl.LoadModel(filepath.Join(dir, "OBJ format", "wall.obj"))
		rl.SetMaterialTexture(wallModel.Materials, rl.MapDiffuse, texture)
		wallCornerModel = wallModel // not necessary as walls fill corner space.. still init it
		rl.SetMaterialTexture(wallCornerModel.Materials, rl.MapDiffuse, texture)
	default:
		panic(fmt.Sprintf("unexpected common.RoomType: %#v", room))
	}
}

// Use walls to avoid infinite-map generation
func DrawBatch(room common.RoomType, pos, size, scale rl.Vector3) {
	var (
		tint         = rl.White
		rotationAxis = common.YAxis
	)

	var (
		wallLen   float32
		wallthick float32
		wallboty  float32
	)

	// Draw walls along each X & Z axis
	switch room {
	case common.OpenWorldRoom:
		wallLen = float32(1.)
		wallthick = float32(1. / 2.)
		wallboty = size.Y
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
	case common.DrillRoom:
		// NOTE: Should just use floor.BoundingBox
		// NOTE: wallBox := rl.GetModelBoundingBox(wallModel) // For precision
		wallLen = float32(1.) * (scale.Z / 4.)
		wallboty = size.Y
		wallthick = float32(1./2.) * (scale.Z / 4.) // HACKY: works somehow when i want to make walls thick
		lo := -float32(size.X/2.) + wallLen + wallthick/2.
		hi := float32(size.X/2.) - wallLen/4. - wallthick
		for i := lo; i <= hi; i += 1 { // .. without sacrificing player vs floor bounds collision checks
			pos1 := rl.NewVector3(pos.X-i, pos.Y+wallboty, pos.Z-size.Z/2-wallthick) // back left->right plane (+-X -Z)
			pos2 := rl.NewVector3(pos.X+i, pos.Y+wallboty, pos.Z+size.Z/2+wallthick) // front left->right plane (+-X +Z)
			rl.DrawModelEx(wallModel, pos1, rotationAxis, 90, scale, tint)
			rl.DrawModelEx(wallModel, pos2, rotationAxis, 360-90, scale, tint)
		}
		wallLen = float32(1.) * (scale.X / 4.)
		wallthick = float32(1./2.) * (scale.X / 4.) // HACKY: works somehow when i want to make walls thick
		lo = -float32(size.Z/2.) + wallLen + wallthick/2.
		hi = float32(size.Z/2.) - wallLen/4. - wallthick
		for i := lo; i <= hi; i += 1 { // .. without sacrificing player vs floor bounds collision checks
			pos1 := rl.NewVector3(pos.X-size.X/2-wallthick, pos.Y+wallboty, pos.Z+i) // left back->front plane (-X +-Z)
			pos2 := rl.NewVector3(pos.X+size.X/2+wallthick, pos.Y+wallboty, pos.Z+i) // right back->front plane (+X +-Z)
			rl.DrawModelEx(wallModel, pos1, rotationAxis, 0, scale, tint)
			rl.DrawModelEx(wallModel, pos2, rotationAxis, 180, scale, tint)
		}
	default:
		panic(fmt.Sprintf("unexpected common.RoomType: %#v", room))
	}

	// Draw 4 wall corners
	switch room {
	case common.OpenWorldRoom:
		bottomLeft := rl.NewVector3(pos.X-size.X/2-wallthick, pos.Y+wallboty, pos.Z+size.Z/2+wallthick)
		bottomRight := rl.NewVector3(pos.X+size.X/2+wallthick, pos.Y+wallboty, pos.Z+size.Z/2+wallthick)
		topRight := rl.NewVector3(pos.X+size.X/2+wallthick, pos.Y+wallboty, pos.Z-size.Z/2-wallthick)
		topLeft := rl.NewVector3(pos.X-size.X/2-wallthick, pos.Y+wallboty, pos.Z-size.Z/2-wallthick)
		rl.DrawModelEx(wallCornerModel, topRight, rotationAxis, 0, scale, tint)
		rl.DrawModelEx(wallCornerModel, topLeft, rotationAxis, 90, scale, tint)
		rl.DrawModelEx(wallCornerModel, bottomLeft, rotationAxis, 180, scale, tint)
		rl.DrawModelEx(wallCornerModel, bottomRight, rotationAxis, 270, scale, tint)
	case common.DrillRoom:
		break // walls fill corner
	default:
		panic(fmt.Sprintf("unexpected common.RoomType: %#v", room))
	}
}
