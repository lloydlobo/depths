package common

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// DrawXYZOrbitV draws perpendicular 3D circles to all 3 (x y z) axis.
func DrawXYZOrbitV(pos rl.Vector3, radius float32) {
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, 1, 0), 90, XAxisColor)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(1, 0, 0), 90, YAxisColor)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, -1, 0), 0, ZAxisColor)
}

// DrawWorldXYZAxis draws all 3 (x y z) axis intersecting at (0,0,0).
func DrawWorldXYZAxis() {
	rl.DrawLine3D(rl.NewVector3(500, 0, 0), rl.NewVector3(-500, 0, 0), XAxisColor)
	rl.DrawLine3D(rl.NewVector3(0, 500, 0), rl.NewVector3(0, -500, 0), YAxisColor)
	rl.DrawLine3D(rl.NewVector3(0, 0, 500), rl.NewVector3(0, 0, -500), ZAxisColor)
}
