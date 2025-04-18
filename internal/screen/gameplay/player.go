package gameplay

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
	Collisions  rl.Quaternion
}

func NewPlayer() Player {
	out := Player{
		Position:   camera.Target,
		Size:       rl.NewVector3(1, 2, 1),
		Collisions: rl.NewQuaternion(0, 0, 0, 0)}
	out.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(camera.Target.X-player.Size.X/2, camera.Target.Y-player.Size.Y/2, camera.Target.Z-player.Size.Z/2),
		rl.NewVector3(camera.Target.X+player.Size.X/2, camera.Target.Y+player.Size.Y/2, camera.Target.Z+player.Size.Z/2))
	return out
}

func (p *Player) Update() {
	// Project the player as the camera target
	p.Position = camera.Target

	p.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(p.Position.X-p.Size.X/2, p.Position.Y-p.Size.Y/2, p.Position.Z-p.Size.Z/2),
		rl.NewVector3(p.Position.X+p.Size.X/2, p.Position.Y+p.Size.Y/2, p.Position.Z+p.Size.Z/2))

	// Wall collisions
	if p.BoundingBox.Min.X <= floor.BoundingBox.Min.X {
		isPlayerWallCollision = true
		p.Collisions.X = -1
	}
	if p.BoundingBox.Max.X >= floor.BoundingBox.Max.X {
		isPlayerWallCollision = true
		p.Collisions.X = 1
	}
	if p.BoundingBox.Min.Z <= floor.BoundingBox.Min.Z {
		isPlayerWallCollision = true
		p.Collisions.Z = -1
	}
	if p.BoundingBox.Max.Z >= floor.BoundingBox.Max.Z {
		isPlayerWallCollision = true
		p.Collisions.Z = 1
	}

	// Floor collisions
	if p.BoundingBox.Min.Y <= floor.BoundingBox.Min.Y {
		p.Collisions.Y = 1 // Player head below floor
	}
	if p.BoundingBox.Max.Y >= floor.BoundingBox.Min.Y { // On floor
		p.Collisions.W = -1 // Allow walking freely
	}
}

func (p Player) Draw() {
	col := rl.Beige
	rl.DrawCapsule(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/4, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
		p.Size.X/2, 16, 16, col)
	rl.DrawCapsuleWires(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/4, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
		p.Size.X/2, 16, 16, col)
	rl.DrawCylinderWiresEx(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/2, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/2, 0)),
		p.Size.X/2, p.Size.X/2, 16, col)
	if isPlayerWallCollision {
		rl.DrawBoundingBox(p.BoundingBox, rl.Red)
	} else {
		rl.DrawBoundingBox(p.BoundingBox, rl.LightGray)
	}

	if p.Collisions.X != 0 {
		pos := p.Position
		pos.X += p.Collisions.X * p.Size.X / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), xCol)
	}
	if p.Collisions.Y != 0 {
		pos := p.Position
		pos.Y += p.Collisions.Y * p.Size.Y / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), yCol)
	}
	if p.Collisions.Z != 0 {
		pos := p.Position
		pos.Z += p.Collisions.Z * p.Size.Z / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), zCol)
	}
	if p.Collisions.W != 0 { // Floor
		pos := p.Position
		pos.Y += p.Collisions.W * p.Size.Y / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), yCol)
	}

	DrawXYZOrbitV(p.Position, 1.)
	draw_cube_texture_main()
}
