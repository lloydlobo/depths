package floor

import (
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

var (
	floorTileLargeModel rl.Model
)

type Floor struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
}

func NewFloor(pos, size rl.Vector3) Floor {
	return Floor{
		Position:    pos,
		Size:        size,
		BoundingBox: common.GetBoundingBoxPositionSizeV(pos, size),
	}
}

// InitFloor creates a new floor, loads a mesh, assigns material parameters and texture to a model.
// NOTE: A basic plane shape can be generated instead of being loaded from a model file
// func InitFloor(floor *Floor) {
// 	pos := rl.NewVector3(0., 0., 0.)
// 	const scale = 2.0
// 	size := rl.Vector3Multiply(rl.NewVector3(16., 0.001, 9.), rl.NewVector3(scale, 1., scale))
//
// 	// Set global var in internal/game.go for bounds, vertex information
// 	*floor = NewFloor(pos, size)
// }

func SetupFloorModel() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	floorTileLargeModel = common.ModelDungeonKit.OBJ.Floor // Floor,FloorDetail
	rl.SetMaterialTexture(floorTileLargeModel.Materials, rl.MapDiffuse, common.ModelDungeonKit.OBJ.Colormap)
}

func (fl Floor) Draw() {
	for x := float32(fl.BoundingBox.Min.X) - 1/2; x < float32(fl.BoundingBox.Max.X)+1; x += 1 {
		for z := float32(fl.BoundingBox.Min.Z) - 1/2; z < float32(fl.BoundingBox.Max.Z)+1; z += 1 {
			position := rl.Vector3{X: x, Y: (fl.BoundingBox.Max.Y - fl.BoundingBox.Min.Y) / 2, Z: z}
			rl.DrawModel(floorTileLargeModel, position, 1.0, rl.White)
		}
	}

	if false { // DEBUG
		rl.DrawBoundingBox(fl.BoundingBox, rl.DarkGray)
		common.DrawXYZOrbitV(rl.Vector3Zero(), 2.)
		common.DrawWorldXYZAxis()
	}
}
