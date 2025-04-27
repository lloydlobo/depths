package projectile

import (
	"example/depths/internal/util/mathutil"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MaxProjectiles                  = int32(64)
	MaxTimeLeft                     = float32(.8)  // Should decrement time to compare with by dt (i.e. rl.GetFrameTime())
	MaxProjectileFireRateTimerLimit = float32(0.1) //  Reduce this to increase fire rate.
)

type ProjectileSOA struct { // size=1352 (0x548)
	Position           [MaxProjectiles]rl.Vector3
	AngleXZPlaneDegree [MaxProjectiles]float32 // Direction [MaxProjectiles]rl.Vector3
	TimeLeft           [MaxProjectiles]float32
	IsActive           [MaxProjectiles]bool

	CircularBufIndex int32
	FireRateTimer    float32
}

// See https://github.com/lloydlobo/tinycreatures/blob/210c4a44ed62fbb08b5f003872e046c99e288bb9/src/main.lua#L2522C3-L2529C61
func (ps *ProjectileSOA) Reset() {
	for i := range MaxProjectiles {
		ps.Position[i] = rl.Vector3{}
		ps.AngleXZPlaneDegree[i] = 0
		ps.TimeLeft[i] = 0.
		ps.IsActive[i] = false
	}
	ps.CircularBufIndex = 0
	ps.FireRateTimer = 0
}

// See https://github.com/lloydlobo/tinycreatures/blob/210c4a44ed62fbb08b5f003872e046c99e288bb9/src/main.lua#L341C1-L356C4
func (ps *ProjectileSOA) Emit(position rl.Vector3, rotationDegree float32) {
	ps.Position[ps.CircularBufIndex] = position
	ps.Position[ps.CircularBufIndex] = position
	ps.TimeLeft[ps.CircularBufIndex] = MaxTimeLeft
	ps.IsActive[ps.CircularBufIndex] = true
	ps.AngleXZPlaneDegree[ps.CircularBufIndex] = rotationDegree

	// Increment index: (ring like data structure / circular reusable buffer)
	ps.CircularBufIndex = (ps.CircularBufIndex + 1) % MaxProjectiles

	// Reset to default timer: (assume cooldown)
	ps.FireRateTimer = MaxProjectileFireRateTimerLimit
}

// Example
func FireEntityProjectile(ps *ProjectileSOA, entityPos, entitySize rl.Vector3, rotationDegree float32) {
	initialOrigin := rl.Vector3{
		X: entityPos.X + mathutil.CosF(rotationDegree*rl.Deg2rad)*entitySize.X/2,
		Y: entityPos.Y + entitySize.Y/2,
		Z: entityPos.Z + mathutil.SinF(rotationDegree*rl.Deg2rad)*entitySize.Z/2,
	}
	ps.Emit(initialOrigin, rotationDegree)
}
