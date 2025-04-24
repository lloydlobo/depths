package drillroom

import (
	"cmp"
	"fmt"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	"example/depths/internal/floor"
	"example/depths/internal/player"
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

	if !rl.IsMusicStreamPlaying(common.Music.DrillRoom000) {
		rl.PlayMusicStream(common.Music.DrillRoom000)
	}

	// Core resources
	player.SetupPlayerModel()
	floor.SetupFloorModel()

	// Core data
	player.InitPlayer(&gamePlayer, camera)
	gameFloor = floor.NewFloor(common.Vector3Zero, rl.NewVector3(4*3, 0.001*2, 3*3))

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

	// Save variables this frame
	oldCam := camera
	oldPlayer := gamePlayer
	_ = oldCam
	_ = oldPlayer

	gamePlayer.Update(camera, gameFloor)

	if true {
		rl.UpdateCamera(&camera, rl.CameraThirdPerson)
	}
}

func Draw() {
	// TODO: Draw ending screen here!
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	rl.ClearBackground(rl.RayWhite)

	rl.DrawModel(common.Model.OBJ.Gate, common.Vector3Zero, 1., rl.White)
	rl.DrawModel(common.Model.OBJ.Barrel, common.Vector3Zero, 1., rl.White)

	gamePlayer.Draw()
	gameFloor.Draw()

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
func Finish() int {
	// saveCoreLevelState()       // (player,camera,...) 705 bytes
	// saveAdditionalLevelState() // (blocks,...)        82871 bytes
	return finishScreen
}
