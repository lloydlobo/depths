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
	// TODO: Update gameplay screen variables here!
	oldCam := camera
	dt := rl.GetFrameTime()

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	// Press enter or tap to change to ENDING screen
	if rl.IsKeyDown(rl.KeyF10) || rl.IsGestureDetected(rl.GestureDoubletap) {
		finishScreen = 1
		// rl.PlaySound(fxCoin)
	}

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

	rl.EndMode3D()

	// 2D HUD
	fontThatIsInGameDotGo := rl.GetFontDefault()

	fontSize := float32(fontThatIsInGameDotGo.BaseSize) * 3.0
	if false {
		pos := rl.NewVector2(
			float32(rl.GetScreenWidth())/2-float32(rl.MeasureText(screenTitleText, int32(fontSize)))/2,
			float32(rl.GetScreenHeight())/2.25,
		)
		rl.DrawTextEx(fontThatIsInGameDotGo, screenTitleText, pos, fontSize, 4, rl.Orange)

		posX := int32(rl.GetScreenWidth())/2 - rl.MeasureText(screenSubtitleText, 20)/2
		posY := int32(rl.GetScreenHeight()) / 2
		rl.DrawText(screenSubtitleText, posX, posY, 20, rl.Orange)
	}
	rl.DrawTextEx(fontThatIsInGameDotGo, fmt.Sprintln(rl.GetFrameTime()), rl.NewVector2(10, 10), fontSize, 4, rl.Orange)

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
