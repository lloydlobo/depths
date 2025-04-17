// Package common provides assets and resources initialized once (if possible)
// before the game is run.
package common

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	Font  struct{ Primary, Secondary rl.Font }
	Music struct{ Theme, Ambient rl.Music }
	FX   struct{ Coin rl.Sound }
)
