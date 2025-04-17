package gameplay

import (
	"cmp"
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
)

type Player struct {
	Position rl.Vector3
	Size     rl.Vector3
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
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}

	player = Player{
		Position: camera.Target,
		Size:     rl.NewVector3(1, 2, 1),
	}

	floor.Position = rl.Vector3Zero()
	floor.Size = rl.NewVector3(20, 1, 20)
	floor.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(-floor.Size.X/2, -floor.Size.Y/2, -floor.Size.Z/2),
		rl.NewVector3(floor.Size.X/2, floor.Size.Y/2, floor.Size.Z/2))

	rl.DisableCursor() // for ThirdPersonPerspective
}

func Update() {
	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) || rl.IsGestureDetected(rl.GestureDoubletap) {
		finishScreen = 1
		// rl.PlaySound(fxCoin)
	}

	// TODO: Update gameplay screen variables here!
	oldCam := camera
	dt := rl.GetFrameTime()

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	// Project player at camera target
	const tolerance = 1 // 1 cube unit apart
	dist := rl.Vector3Distance(player.Position, camera.Target)
	smooth := 1 / (dist + tolerance) // [0.44..0.98]
	player.Position = cmp.Or(
		// Either simulate easing (push/pull at start/stop movement)
		rl.Vector3Lerp(player.Position, rl.Vector3Lerp(oldCam.Target, camera.Target, smooth*(10.*dt)), smooth*(10.*dt)),
		// Or simulate snapping to the target
		camera.Target,
	)

	framesCounter++
}

func Draw() {
	// 3D World
	rl.BeginMode3D(camera)
	rl.ClearBackground(rl.RayWhite)

	// Floor
	rl.DrawCubeV(floor.Position, floor.Size, rl.White)
	rl.DrawBoundingBox(floor.BoundingBox, rl.LightGray)

	// Player
	rl.DrawCapsule(
		rl.Vector3Add(player.Position, rl.NewVector3(0, player.Size.Y/4, 0)),
		rl.Vector3Add(player.Position, rl.NewVector3(0, -player.Size.Y/4, 0)),
		player.Size.X/2, 16, 16, rl.Black)
	rl.DrawCapsuleWires(
		rl.Vector3Add(player.Position, rl.NewVector3(0, player.Size.Y/4, 0)),
		rl.Vector3Add(player.Position, rl.NewVector3(0, -player.Size.Y/4, 0)),
		player.Size.X/2, 16, 16, rl.Black)
	rl.DrawCylinderWiresEx(
		rl.Vector3Add(player.Position, rl.NewVector3(0, player.Size.Y/2, 0)),
		rl.Vector3Add(player.Position, rl.NewVector3(0, -player.Size.Y/2, 0)),
		player.Size.X/2, player.Size.X/2, 16, rl.Black)

	// Center orbit
	drawWorldXYZAxis()
	drawXYZOrbitV(rl.Vector3Zero(), 3.)

	drawXYZOrbitV(player.Position, 1.)

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

//
// var movement rl.Vector3
// if rl.IsKeyDown(rl.KeyW) {
// 	movement.Z += 1
// }
// if rl.IsKeyDown(rl.KeyS) {
// 	movement.Z -= 1
// }
// if rl.IsKeyDown(rl.KeyA) {
// 	movement.X -= 1
// }
// if rl.IsKeyDown(rl.KeyD) {
// 	movement.X += 1
// }
// // Simulate simple gravity
// if rl.IsKeyDown(rl.KeySpace) {
// 	movement.Y += 1
// }
// if rl.IsKeyDown(rl.KeyLeftControl) {
// 	movement.Y -= 1
// }
// if !rl.Vector3Equals(movement, rl.Vector3Zero()) {
// 	movement = rl.Vector3Normalize(movement)
// }
//
// // Limit max jump
// if camera.Target.Y > 1.5*3 {
// 	movement.Y = 0
// }
//
// camera.Target.Y += 0.2 * movement.Y
// if camera.Target.Y > 1.5 {
// 	camera.Target.Y -= 0.2 / 4
// 	if camera.Target.Y < 1.5 {
// 		camera.Target.Y = 1.5
// 	}
// }

// drawXYZOrbitV draws perpendicular 3D circles to all 3 (x y z) axis.
func drawXYZOrbitV(pos rl.Vector3, radius float32) {
	xCol, yCol, zCol := rl.Red, rl.Green, rl.Blue
	xCol, yCol, zCol = rl.Fade(xCol, .3), rl.Fade(yCol, .3), rl.Fade(zCol, .3)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, 1, 0), 90, xCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(1, 0, 0), 90, yCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, -1, 0), 0, zCol)
}

// drawWorldXYZAxis draws all 3 (x y z) axis intersecting at (0,0,0).
func drawWorldXYZAxis() {
	xCol, yCol, zCol := rl.Red, rl.Green, rl.Blue
	xCol, yCol, zCol = rl.Fade(xCol, .3), rl.Fade(yCol, .3), rl.Fade(zCol, .3)
	rl.DrawLine3D(rl.NewVector3(500, 0, 0), rl.NewVector3(-500, 0, 0), xCol)
	rl.DrawLine3D(rl.NewVector3(0, 500, 0), rl.NewVector3(0, -500, 0), yCol)
	rl.DrawLine3D(rl.NewVector3(0, 0, 500), rl.NewVector3(0, 0, -500), zCol)
}
