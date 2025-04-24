package drillroom

import (
	"cmp"
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	"example/depths/internal/floor"
	"example/depths/internal/player"
	"example/depths/internal/wall"
)

const (
	screenTitleText    = "DRILL ROOM"                                                                        // This should be temporary during prototype
	screenSubtitleText = "leave room: backspace swipe-left\nquit:          F10 pinch-out right-mouse-button" // "press enter or tap to jump to title screen"
)

var (
	// Core data

	finishScreen  int
	framesCounter int32

	levelID                int32
	camera                 rl.Camera3D
	gameFloor              floor.Floor
	gamePlayer             player.Player
	hasPlayerLeftDrillBase bool
)

var (
	// TODO: SEPARATE THIS FROM CORE DATA
	hitCount int32
	hitScore int32
)

func Init() {
	framesCounter = 0
	finishScreen = 0
	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 10., 10.),
		Target:     rl.NewVector3(0., .5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       15. * float32(cmp.Or(4., 3., 2.)),
		Projection: rl.CameraPerspective,
	} // See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h
	hasPlayerLeftDrillBase = false

	if !rl.IsMusicStreamPlaying(common.Music.DrillRoom000) {
		rl.PlayMusicStream(common.Music.DrillRoom000)
	}

	// Core resources
	player.SetupPlayerModel()
	floor.SetupFloorModel()
	wall.SetupWallModel(common.DrillRoom)

	// Core data
	player.InitPlayer(&gamePlayer, camera)
	gameFloor = floor.NewFloor(common.Vector3Zero, rl.NewVector3(10, 0.001*2, 10)) // 1:1 ratio

	// Unequip hat sword shield
	player.ToggleEquippedModels([player.MaxBoneSockets]bool{false, false, false})

	rl.DisableCursor() // For camera thirdperson view
}

func Update() {
	rl.UpdateMusicStream(common.Music.DrillRoom000)

	// Change to ENDING/GAMEPLAY screen
	if rl.IsKeyDown(rl.KeyF10) || rl.IsGestureDetected(rl.GesturePinchOut) ||
		rl.IsMouseButtonDown(rl.MouseButtonRight) {
		finishScreen = 1 // 1=>ending

		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "rollover3.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "switch33.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_interface-sounds", "Audio", "confirmation_001.ogg")))
	}
	if rl.IsKeyDown(rl.KeyBackspace) || rl.IsGestureDetected(rl.GestureSwipeLeft) {
		finishScreen = 2 // 2=>gameplay

		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("footstep0%d.ogg", rl.GetRandomValue(0, 9)))))  // 05
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "metalClick.ogg")))                                         // metalClick
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("creak%d.ogg", rl.GetRandomValue(1, 3)))))      // 3
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("doorClose_%d.ogg", rl.GetRandomValue(1, 4))))) // 4
	}

	// Update playerl leaving common.DrillRoom => common.Opcommon.OpenWorldRoom
	// (copied from gameplay.go => Update player─block collision+breaking/mining )

	bb1, bb2, bb3 := getGatesIntersectionBoundingBoxes() // bot barrier		(raywhite)
	updateIntersectGates(bb1, bb2, bb3)

	// Save variables this frame
	oldCam := camera
	oldPlayer := gamePlayer
	_ = oldCam
	_ = oldPlayer

	// Reset flags/variables
	gamePlayer.Collisions = rl.Quaternion{}
	gamePlayer.IsPlayerWallCollision = false

	// Update the game camera for this screen
	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	// Reset camera yaw(y-axis)/roll(z-axis) (on key [W] or [E])
	if got, want := camera.Up, (rl.Vector3{X: 0., Y: 1., Z: 0.}); !rl.Vector3Equals(got, want) {
		camera.Up = want
	}

	gamePlayer.Update(camera, gameFloor)
	if gamePlayer.IsPlayerWallCollision {
		player.RevertPlayerAndCameraPositions(&gamePlayer, oldPlayer, &camera, oldCam)
	}
}

func getGatesIntersectionBoundingBoxes() (bb1 rl.BoundingBox, bb2 rl.BoundingBox, bb3 rl.BoundingBox) {
	origin := gameFloor.Position
	size := gameFloor.Size
	size.Y = max(2, gamePlayer.Size.Y)
	wallThick := float32(2)
	bb1 = common.GetBoundingBoxFromPositionSizeV(origin, rl.Vector3Subtract(size, rl.NewVector3(3.5, 0, 3.5)))        // player is inside	 (red->green)
	bb2 = common.GetBoundingBoxFromPositionSizeV(origin, rl.Vector3Subtract(size, rl.NewVector3(1.5, 0, 1.5)))        // player is entering (green->blue)
	bb3 = common.GetBoundingBoxFromPositionSizeV(origin, rl.Vector3Add(size, rl.NewVector3(wallThick, 0, wallThick))) // outer barrier      (blue->white)
	return bb1, bb2, bb3
}
func updateIntersectGates(bb1 rl.BoundingBox, bb2 rl.BoundingBox, bb3 rl.BoundingBox) {
	isPlayerInsideBase := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb1)
	isPlayerEnteringBase := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb2)
	isPlayerInsideBotBarrier := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb3)

	// If blue exit
	if isPlayerInsideBotBarrier && !isPlayerEnteringBase && !isPlayerInsideBase {
		player.SetColor(rl.Blue)
		if !hasPlayerLeftDrillBase { // HACK: Placeholder change scene check logic
			hasPlayerLeftDrillBase = true
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("footstep0%d.ogg", rl.GetRandomValue(0, 9))))) // 05
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "metalClick.ogg")))                                        // metalClick
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("creak%d.ogg", rl.GetRandomValue(1, 3)))))     // 3
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("doorOpen_%d.ogg", rl.GetRandomValue(1, 2))))) // 2
			{                                                                                                                                            // Save screen state
				finishScreen = 2                      // 1=>ending 2=>drillroom
				camera.Up = rl.NewVector3(0., 1., 0.) // Reset yaw/pitch/roll
				// TODO: implement drillroom save/load functions (data and filenames)
				// saveCoreLevelState()                  // (player,camera,...) 705 bytes
				// saveAdditionalLevelState()            // (blocks,...)        82871 bytes
			}
		}
	} else if isPlayerEnteringBase && !isPlayerInsideBase {
		player.SetColor(rl.Green)
	} else if isPlayerInsideBase {
		player.SetColor(rl.Red)
	} else { // RESET FLAG just in case
		if hasPlayerLeftDrillBase {
			hasPlayerLeftDrillBase = false
		}
		player.SetColor(rl.RayWhite)
	}
}
func drawIntersectionGates(bb1 rl.BoundingBox, bb2 rl.BoundingBox, bb3 rl.BoundingBox) { // ‥ DEBUG: Draw drill door gate entry logic before changing scene to drill base
	rl.DrawBoundingBox(bb1, rl.Red)
	rl.DrawBoundingBox(bb2, rl.Green)
	rl.DrawBoundingBox(bb3, rl.Blue)
}

func Draw() {
	// TODO: Draw ending screen here!
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	rl.ClearBackground(rl.RayWhite)

	gamePlayer.Draw()
	gameFloor.Draw()
	wall.DrawBatch(common.DrillRoom, gameFloor.Position, gameFloor.Size, cmp.Or(rl.NewVector3(5, 2, 5), common.Vector3One))

	{ // TEMPORARY
		rl.DrawModel(
			common.Model.OBJ.Gate,
			rl.NewVector3(gameFloor.BoundingBox.Min.X+1, gameFloor.Position.Y, gameFloor.BoundingBox.Min.Z+1),
			1., rl.White)
		drawIntersectionGates(getGatesIntersectionBoundingBoxes())
	}

	rl.EndMode3D()

	// 2D World
	if false {
		rl.DrawRectangle(0, 0, screenW, screenH, rl.Fade(rl.Black, .8))
	}

	fontSize := float32(common.Font.Primary.BaseSize) * 3.0
	pos := rl.NewVector2(
		float32(screenW)/2.-float32(rl.MeasureText(screenTitleText, int32(fontSize)))/2.,
		float32(screenH)/12.)
	rl.DrawTextEx(common.Font.Primary, screenTitleText, pos, fontSize, 4, rl.Fade(rl.Black, .5))

	subtextSize := rl.MeasureTextEx(common.Font.Primary, screenSubtitleText, fontSize/2, 1)
	posX := int32(screenW)/2 - int32(subtextSize.X/2)
	posY := int32(screenH) - int32(subtextSize.Y*common.Phi)
	rl.DrawText(screenSubtitleText, posX, posY, int32(fontSize/2), rl.Gray)
}

func Unload() {
	// TODO: Unload gameplay screen variables here!
	if isTransToGameScreen := finishScreen == 2; !isTransToGameScreen && rl.IsCursorHidden() {
		rl.EnableCursor() // without 3d ThirdPersonPerspective
	}
	// Commented out as it hinders switching to drill room or
	// menu/ending (on pause/restart)
	//
	// rl.UnloadMusicStream(music)
}

// Drillroom screen should finish?
// NOTE: This is called each frame in main game loop
func Finish() int {
	return finishScreen
}
