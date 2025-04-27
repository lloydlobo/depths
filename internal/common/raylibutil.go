package common

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func GetBoundingBoxFromPositionSizeV(pos, size rl.Vector3) rl.BoundingBox {
	return rl.NewBoundingBox(
		rl.NewVector3(pos.X-size.X/2, pos.Y-size.Y/2, pos.Z-size.Z/2),
		rl.NewVector3(pos.X+size.X/2, pos.Y+size.Y/2, pos.Z+size.Z/2))
}

func GetCollisionPointBox(point rl.Vector3, box rl.BoundingBox) bool {
	return point.X >= box.Min.X && point.X <= box.Max.X &&
		point.Y >= box.Min.Y && point.Y <= box.Max.Y &&
		point.Z >= box.Min.Z && point.Z <= box.Max.Z
}
