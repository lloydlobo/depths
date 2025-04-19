package common

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	Phi            = math.Phi
	InvPhi         = 1 / Phi
	OneMinusInvPhi = 1 - InvPhi
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
