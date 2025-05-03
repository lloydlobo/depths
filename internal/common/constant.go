package common

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	Phi            = math.Phi   // 1.61803
	InvPhi         = 1 / Phi    // 0.618034
	OneMinusInvPhi = 1 - InvPhi // 0.381966
	TwoMinusInvPhi = 2 - InvPhi // 1.38197
)

var (
	Vector3One       = rl.Vector3One()
	Vector3Zero      = rl.Vector3Zero()
	Vector2One       = rl.Vector2One()
	Vector2Zero      = rl.Vector2Zero()
	Vector3OneLength = rl.Vector3Length(Vector3One)
	Vector2OneLength = rl.Vector2Length(Vector2One)
)

var (
	XAxisColor = rl.Fade(rl.Red, .2)
	YAxisColor = rl.Fade(rl.Green, .2)
	ZAxisColor = rl.Fade(rl.Blue, .2)
)

var (
	XAxis = rl.NewVector3(1, 0, 0)
	YAxis = rl.NewVector3(0, 1, 0)
	ZAxis = rl.NewVector3(0, 0, 1)
)

// RoomType enumerates different type of rooms based on screen/scenes to switch use cases.
type RoomType uint8

const (
	OpenWorldRoom RoomType = iota
	DrillRoom
)

const (
	FPS = 60
)
