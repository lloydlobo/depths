package gameplay

import (
	"fmt"
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

func Init() {
	framesCounter = 0
	finishScreen = 0

	// See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h
	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 30., 30.),
		Target:     rl.NewVector3(0., 1+0.5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}

	player = Player{
		Position:   camera.Target,
		Size:       rl.NewVector3(1, 2, 1),
		Collisions: rl.NewQuaternion(0, 0, 0, 0),
	}
	player.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(camera.Target.X-player.Size.X/2, camera.Target.Y-player.Size.Y/2, camera.Target.Z-player.Size.Z/2),
		rl.NewVector3(camera.Target.X+player.Size.X/2, camera.Target.Y+player.Size.Y/2, camera.Target.Z+player.Size.Z/2))

	floor = Floor{
		Position: rl.Vector3Zero(),
		Size:     rl.NewVector3((20 * 2), 1, (20*2)*4/5),
	}
	floor.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(-floor.Size.X/2, -floor.Size.Y/2, -floor.Size.Z/2),
		rl.NewVector3(floor.Size.X/2, floor.Size.Y/2, floor.Size.Z/2))

	isWallCollision = false

	rl.SetMusicVolume(common.Music.Theme, 0.5)
	rl.PlayMusicStream(common.Music.Theme)

	rl.DisableCursor() // for ThirdPersonPerspective
}

func Update() {
	dt := rl.GetFrameTime()
	_ = dt

	rl.UpdateMusicStream(common.Music.Theme)

	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) /* || rl.IsGestureDetected(rl.GestureDrag) */ {
		finishScreen = 1
		rl.PlaySound(common.FX.Coin)
	}

	// Pick up item
	if rl.IsKeyDown(rl.KeyF) {
		log.Println("Picked up ...")
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
			rl.NewVector3(player.Position.X-player.Size.X/2,
				player.Position.Y-player.Size.Y/2, player.Position.Z-player.Size.Z/2),
			rl.NewVector3(player.Position.X+player.Size.X/2,
				player.Position.Y+player.Size.Y/2, player.Position.Z+player.Size.Z/2))

		camera.Target = oldCam.Target
		camera.Position = oldCam.Position
	}

	framesCounter++
}

func Draw() {
	rl.BeginMode3D(camera)
	rl.ClearBackground(rl.Black)

	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	player.Draw()

	rl.DrawCubeV(floor.Position, floor.Size, rl.DarkBrown)
	rl.DrawBoundingBox(floor.BoundingBox, rl.Brown)

	DrawXYZOrbitV(rl.Vector3Zero(), 2.)
	DrawWorldXYZAxis()

	rl.EndMode3D()

	// 2D HUD
	fontThatIsInGameDotGo := rl.GetFontDefault()
	fontSize := float32(fontThatIsInGameDotGo.BaseSize) * 3.0

	text := "[F] PICK UP"
	rl.DrawText(text, screenW/2-rl.MeasureText(text, 20)/2, screenH-20*2, 20, rl.White)

	rl.DrawFPS(10, 10)
	rl.DrawTextEx(fontThatIsInGameDotGo, fmt.Sprintf("%.6f", rl.GetFrameTime()),
		rl.NewVector2(10, 10+20*1), fontSize*2./3., 1, rl.Lime)
}

func Unload() {
	// TODO: Unload gameplay screen variables here!
	if rl.IsCursorHidden() {
		rl.EnableCursor() // without 3d ThirdPersonPerspective
	}
	// rl.UnloadMusicStream(music)
}

// Gameplay screen should finish?
func Finish() int {
	return finishScreen
}

type Floor struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
}

type Player struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
	Collisions  rl.Quaternion
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

	DrawXYZOrbitV(p.Position, 1.)
}

// DrawXYZOrbitV draws perpendicular 3D circles to all 3 (x y z) axis.
func DrawXYZOrbitV(pos rl.Vector3, radius float32) {
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, 1, 0), 90, xCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(1, 0, 0), 90, yCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, -1, 0), 0, zCol)
}

// DrawWorldXYZAxis draws all 3 (x y z) axis intersecting at (0,0,0).
func DrawWorldXYZAxis() {
	rl.DrawLine3D(rl.NewVector3(500, 0, 0), rl.NewVector3(-500, 0, 0), xCol)
	rl.DrawLine3D(rl.NewVector3(0, 500, 0), rl.NewVector3(0, -500, 0), yCol)
	rl.DrawLine3D(rl.NewVector3(0, 0, 500), rl.NewVector3(0, 0, -500), zCol)
}

const (
	screenTitleText    = "GAMEPLAY SCREEN" // This should be temporary during prototype
	screenSubtitleText = "continue"        // "press enter or tap to jump to ending screen"
)

var (
	framesCounter int32
	finishScreen  int

	camera rl.Camera3D

	player          Player
	floor           Floor
	isWallCollision bool
)

var (
	xCol = rl.Fade(rl.Red, .3)
	yCol = rl.Fade(rl.Green, .3)
	zCol = rl.Fade(rl.Green, .3)
)
