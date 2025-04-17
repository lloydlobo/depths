package gameplay

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenTitleText    = "GAMEPLAY SCREEN" // This should be temporary during prototype
	screenSubtitleText = "press enter or tap to jump to ending screen"
)

var (
	framesCounter int32
	finishScreen  int

	camera rl.Camera3D

	player Player

	floor Floor

	isWallCollision bool
)

var (
	xCol, yCol, zCol = rl.Fade(rl.Red, .3), rl.Fade(rl.Green, .3), rl.Fade(rl.Blue, .3)
)

type Player struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
	Collisions  rl.Quaternion
}

type Floor struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
}

func Init() {
	framesCounter = 0
	finishScreen = 0

	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 10., 10.),
		Target:     rl.NewVector3(0., 1+0.5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       60.0,
		Projection: rl.CameraPerspective,
	}

	player = Player{
		Position: camera.Target,
		Size:     rl.NewVector3(1, 2, 1),
		BoundingBox: rl.NewBoundingBox(
			rl.NewVector3(camera.Target.X-player.Size.X/2, camera.Target.Y-player.Size.Y/2, camera.Target.Z-player.Size.Z/2),
			rl.NewVector3(camera.Target.X+player.Size.X/2, camera.Target.Y+player.Size.Y/2, camera.Target.Z+player.Size.Z/2),
		),
		Collisions: rl.NewQuaternion(0, 0, 0, 0),
	}

	floor.Position = rl.Vector3Zero()
	floor.Size = rl.NewVector3(20, 1, 20)
	floor.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(-floor.Size.X/2, -floor.Size.Y/2, -floor.Size.Z/2),
		rl.NewVector3(floor.Size.X/2, floor.Size.Y/2, floor.Size.Z/2),
	)

	isWallCollision = false

	rl.DisableCursor() // for ThirdPersonPerspective
}

func Update() {
	dt := rl.GetFrameTime()
	_ = dt

	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) || rl.IsGestureDetected(rl.GestureDoubletap) {
		finishScreen = 1
		// rl.PlaySound(fxCoin)
	}

	// Save variables this frame
	oldCam := camera
	oldPlayer := player

	// Reset single frame flags/variables
	player.Collisions = rl.Quaternion{}
	isWallCollision = false

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)
	player.Update()

	if isWallCollision {
		player.Position = oldPlayer.Position
		player.BoundingBox = rl.NewBoundingBox(
			rl.NewVector3(player.Position.X-player.Size.X/2, player.Position.Y-player.Size.Y/2, player.Position.Z-player.Size.Z/2),
			rl.NewVector3(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2, player.Position.Z+player.Size.Z/2))

		camera.Target = oldCam.Target
		camera.Position = oldCam.Position
	}

	framesCounter++
}

func Draw() {
	// 3D World
	rl.BeginMode3D(camera)
	rl.ClearBackground(rl.RayWhite)

	// Floor
	rl.DrawCubeV(floor.Position, floor.Size, rl.White)
	rl.DrawBoundingBox(floor.BoundingBox, rl.LightGray)
	drawWorldXYZAxis()
	drawXYZOrbitV(rl.Vector3Zero(), 3.)

	// Player
	player.Draw()

	rl.EndMode3D()

	// 2D HUD
	fontThatIsInGameDotGo := rl.GetFontDefault()
	fontSize := float32(fontThatIsInGameDotGo.BaseSize) * 3.0

	rl.DrawFPS(10, 10)
	rl.DrawTextEx(fontThatIsInGameDotGo, fmt.Sprintf("%.6f", rl.GetFrameTime()),
		rl.NewVector2(10, 10+20*1), fontSize*2./3., 1, rl.Lime)
}

func Unload() {
	// TODO: Unload gameplay screen variables here!
	if rl.IsCursorHidden() {
		rl.EnableCursor() // without 3d ThirdPersonPerspective
	}
}

// Gameplay screen should finish?
func Finish() int {
	return finishScreen
}

func (p *Player) Update() {
	// Project the player as the camera target
	p.Position = camera.Target
	p.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(p.Position.X-p.Size.X/2, p.Position.Y-p.Size.Y/2, p.Position.Z-p.Size.Z/2),
		rl.NewVector3(p.Position.X+p.Size.X/2, p.Position.Y+p.Size.Y/2, p.Position.Z+p.Size.Z/2))

	// Wall collisions
	if p.BoundingBox.Min.X <= floor.BoundingBox.Min.X {
		isWallCollision = true
		p.Collisions.X = -1
	}
	if p.BoundingBox.Max.X >= floor.BoundingBox.Max.X {
		isWallCollision = true
		p.Collisions.X = 1
	}
	if p.BoundingBox.Min.Z <= floor.BoundingBox.Min.Z {
		isWallCollision = true
		p.Collisions.Z = -1
	}
	if p.BoundingBox.Max.Z >= floor.BoundingBox.Max.Z {
		isWallCollision = true
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
	rl.DrawCapsule(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/4, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
		p.Size.X/2, 16, 16, rl.Black)
	rl.DrawCapsuleWires(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/4, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
		p.Size.X/2, 16, 16, rl.Black)
	rl.DrawCylinderWiresEx(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/2, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/2, 0)),
		p.Size.X/2, p.Size.X/2, 16, rl.Black)
	if isWallCollision {
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
	drawXYZOrbitV(p.Position, 1.)
}

// drawXYZOrbitV draws perpendicular 3D circles to all 3 (x y z) axis.
func drawXYZOrbitV(pos rl.Vector3, radius float32) {
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, 1, 0), 90, xCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(1, 0, 0), 90, yCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, -1, 0), 0, zCol)
}

// drawWorldXYZAxis draws all 3 (x y z) axis intersecting at (0,0,0).
func drawWorldXYZAxis() {
	rl.DrawLine3D(rl.NewVector3(500, 0, 0), rl.NewVector3(-500, 0, 0), xCol)
	rl.DrawLine3D(rl.NewVector3(0, 500, 0), rl.NewVector3(0, -500, 0), yCol)
	rl.DrawLine3D(rl.NewVector3(0, 0, 500), rl.NewVector3(0, 0, -500), zCol)
}
