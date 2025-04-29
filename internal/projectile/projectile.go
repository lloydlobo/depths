package projectile

import (
	"example/depths/internal/util/mathutil"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MaxProjectiles                  = int32(16)    // Cyclic buffer capacity
	MaxTimeLeft                     = float32(1.0) // 1 second in 60fps. Should decrement time to compare with by dt (i.e. rl.GetFrameTime())
	MaxProjectileFireRateTimerLimit = float32(0.5) // Reduce this to increase fire rate.
)

type ProjectileSOA struct { // size=344 (0x158)
	Position [MaxProjectiles]rl.Vector3
	Rotation [MaxProjectiles]float32 // (degrees) XZ plane
	TimeLeft [MaxProjectiles]float32
	IsActive [MaxProjectiles]bool

	CircularBufIndex int32
	FireRateTimer    float32 // Wait for MaxProjectileFireRateTimerLimit cooldown wait time before next emit. Update -= dt each frame
}

// See https://github.com/lloydlobo/tinycreatures/blob/210c4a44ed62fbb08b5f003872e046c99e288bb9/src/main.lua#L2522C3-L2529C61
func (ps *ProjectileSOA) Reset() {
	for i := range MaxProjectiles {
		ps.Position[i] = rl.Vector3{}
		ps.Rotation[i] = 0
		ps.TimeLeft[i] = 0.
		ps.IsActive[i] = false
	}
	ps.CircularBufIndex = 0
	ps.FireRateTimer = 0
}

// See https://github.com/lloydlobo/tinycreatures/blob/210c4a44ed62fbb08b5f003872e046c99e288bb9/src/main.lua#L341C1-L356C4
func (ps *ProjectileSOA) Emit(position rl.Vector3, rotationDegree float32) {
	ps.Position[ps.CircularBufIndex] = position
	ps.Rotation[ps.CircularBufIndex] = rotationDegree

	ps.TimeLeft[ps.CircularBufIndex] = MaxTimeLeft
	ps.IsActive[ps.CircularBufIndex] = true

	// Increment index: (ring like data structure / circular reusable buffer)
	ps.CircularBufIndex = (ps.CircularBufIndex + 1) % MaxProjectiles

	// Reset to default timer: (assume cooldown)
	ps.FireRateTimer = MaxProjectileFireRateTimerLimit
}

// Example
func FireEntityProjectile(ps *ProjectileSOA, entityPos, entitySize rl.Vector3, rotationDegree float32) (didFire bool) {
	if ps.FireRateTimer <= 0 {
		initialOrigin := rl.Vector3{
			X: entityPos.X + mathutil.CosF(rotationDegree*rl.Deg2rad)*entitySize.X/2,
			Y: entityPos.Y + entitySize.Y/2,
			Z: entityPos.Z + mathutil.SinF(rotationDegree*rl.Deg2rad)*entitySize.Z/2,
		}
		ps.Emit(initialOrigin, rotationDegree)
		return true
	}
	return false
}
