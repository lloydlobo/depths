package gameplay

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

// DrawXYZOrbitV draws perpendicular 3D circles to all 3 (x y z) axis.
func DrawXYZOrbitV(pos rl.Vector3, radius float32) {
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, 1, 0), 90, common.XAxisColor)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(1, 0, 0), 90, common.YAxisColor)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, -1, 0), 0, common.ZAxisColor)
}

// DrawWorldXYZAxis draws all 3 (x y z) axis intersecting at (0,0,0).
func DrawWorldXYZAxis() {
	rl.DrawLine3D(rl.NewVector3(500, 0, 0), rl.NewVector3(-500, 0, 0), common.XAxisColor)
	rl.DrawLine3D(rl.NewVector3(0, 500, 0), rl.NewVector3(0, -500, 0), common.YAxisColor)
	rl.DrawLine3D(rl.NewVector3(0, 0, 500), rl.NewVector3(0, 0, -500), common.ZAxisColor)
}
