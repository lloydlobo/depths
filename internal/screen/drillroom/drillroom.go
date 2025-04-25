// TODO: Package drillroom only makes sense if player has a limited cargo capacity
package drillroom

import (
	"cmp"
	"fmt"
	"image/color"
	"math"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	"example/depths/internal/floor"
	"example/depths/internal/player"
	"example/depths/internal/wall"
)

const (
	screenTitleText    = "DRILL"                                                                             // This should be temporary during prototype
	screenSubtitleText = "leave room: backspace swipe-left\nquit:          F10 pinch-out right-mouse-button" // "press enter or tap to jump to title screen"
)

var (
	// Core data

	finishScreen  int
	framesCounter int32

	levelID                int32
	camera                 rl.Camera3D
	floorEntity            floor.Floor
	playerEntity           player.Player // Player Entity or xPlayer
	hasPlayerLeftDrillBase bool
)

var (
	// TODO: SEPARATE THIS FROM CORE DATA
	hitCount int32
	hitScore int32
)

var (
	drillroomExitBoundingBox rl.BoundingBox
)

var (
	triggerBoundingBoxes       []rl.BoundingBox
	triggerSensorBoundingBoxes []rl.BoundingBox
	triggerPositions           []rl.Vector3
	isPlayerNearTriggerSensors []bool
	triggerCount               int32

	triggerModels []rl.Model
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
	player.InitPlayer(&playerEntity, camera)
	floorEntity = floor.NewFloor(common.Vector3Zero, rl.NewVector3(10, 0.001*2, 10)) // 1:1 ratio

	// Layout copied from https://annekatran.itch.io/dig-and-delve
	triggerSize := rl.NewVector3(.5, .5, .5)
	triggerPosY := triggerSize.Y / 2.
	triggerPositions = []rl.Vector3{
		// Corners
		rl.NewVector3(-2, triggerPosY, -2),
		rl.NewVector3(+2, triggerPosY, -2),
		rl.NewVector3(-2, triggerPosY, +2),
		rl.NewVector3(+2, triggerPosY, +2),

		// Sides
		rl.NewVector3(-2, triggerPosY, +0),
		rl.NewVector3(+2, triggerPosY, +0),
		rl.NewVector3(+0, triggerPosY, -2),
		rl.NewVector3(+0, triggerPosY, +2),
	}
	kx := (floorEntity.Size.X / (1. * math.Pi)) - 1.
	kz := (floorEntity.Size.Z / (1. * math.Pi)) - 2.

	// 45 degree tangent lines (use cos/sin??)
	dx := float32(0.15 + triggerSize.X)
	dz := float32(0.40 + triggerSize.Z)

	triggerPositions = []rl.Vector3{
		// Upper corners
		rl.NewVector3(-kx, triggerPosY, -kz), // NW
		rl.NewVector3(+kx, triggerPosY, -kz), // NE

		// Upper arcs
		rl.NewVector3(-kx+dx, triggerPosY, -kz-dz),       // NW -> NE
		rl.NewVector3(-kx+dx+dx, triggerPosY, -kz-dz-dz), // NW -> NE -> NE
		rl.NewVector3(+kx-dx, triggerPosY, -kz-dz),       // NE -> NW
		rl.NewVector3(+kx-dx-dx, triggerPosY, -kz-dz-dz), // NE -> NW -> NW

		// Lower corners
		rl.NewVector3(-kx, triggerPosY, +kz), // SW
		rl.NewVector3(+kx, triggerPosY, +kz), // SE

		// Start drill trigger
		rl.NewVector3(-kx+dx, triggerPosY, +kz+dz), // SW -> SE
		// Change resource trigger
		rl.NewVector3(+kx-dx, triggerPosY, +kz+dz), // SE -> SW

	}

	triggerCount = int32(len(triggerPositions))
	for i := range triggerCount {
		triggerBoundingBoxes = append(triggerBoundingBoxes, common.GetBoundingBoxFromPositionSizeV(triggerPositions[i], triggerSize))
	}
	for i := range triggerCount {
		triggerSensorBoundingBoxes = append(triggerSensorBoundingBoxes, common.GetBoundingBoxFromPositionSizeV(triggerPositions[i], rl.Vector3Scale(triggerSize, 2)))
	}
	for range triggerCount {
		isPlayerNearTriggerSensors = append(isPlayerNearTriggerSensors, false)
	}
	for range triggerCount {
		dir := filepath.Join("res", "kenney_prototype-kit", "Models", "OBJ format")
		model := rl.LoadModel(filepath.Join(dir, "weapon-shield.obj"))
		texture := rl.LoadTexture(filepath.Join(dir, "Textures", "colormap.png"))
		rl.SetMaterialTexture(model.Materials, rl.MapDiffuse, texture)
		triggerModels = append(triggerModels, model)
	}

	// Unequip hat sword shield
	player.ToggleEquippedModels([player.MaxBoneSockets]bool{false, false, false})

	// Compute once
	drillroomExitBoundingBox = common.GetBoundingBoxFromPositionSizeV(
		floorEntity.Position,
		rl.Vector3Subtract(floorEntity.Size, rl.NewVector3(1+playerEntity.Size.X/2, -playerEntity.Size.Y*2, 1+playerEntity.Size.Z/2)),
	)

	// For camera thirdperson view
	rl.DisableCursor()
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
		// Play exit sounds
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("footstep0%d.ogg", rl.GetRandomValue(0, 9)))))  // 05
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "metalClick.ogg")))                                         // metalClick
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("creak%d.ogg", rl.GetRandomValue(1, 3)))))      // 3
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("doorClose_%d.ogg", rl.GetRandomValue(1, 4))))) // 4

		// Save screen state
		finishScreen = 2                      // 1=>ending 2=>gameplay(openworldroom)
		camera.Up = rl.NewVector3(0., 1., 0.) // Reset yaw/pitch/roll
		// TODO: implement drillroom save/load functions (data and filenames)
		// saveCoreLevelState()                  // (player,camera,...) 705 bytes
		// saveAdditionalLevelState()            // (blocks,...)        82871 bytes
	}

	// PERF: Just check if player is not colliding wit floor bounding box * scale of 0.9

	// Save variables this frame
	oldCam := camera
	oldPlayer := playerEntity

	// Reset flags/variables
	playerEntity.Collisions = rl.Quaternion{}
	playerEntity.IsPlayerWallCollision = false

	// Update the game camera for this screen
	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	// Reset camera yaw(y-axis)/roll(z-axis) (on key [W] or [E])
	if got, want := camera.Up, (rl.Vector3{X: 0., Y: 1., Z: 0.}); !rl.Vector3Equals(got, want) {
		camera.Up = want
	}

	playerEntity.Update(camera, floorEntity)
	if playerEntity.IsPlayerWallCollision {
		player.RevertPlayerAndCameraPositions(&playerEntity, oldPlayer, &camera, oldCam)
	}

	// Update playerl leaving common.DrillRoom => common.Opcommon.OpenWorldRoom
	if !rl.CheckCollisionBoxes(playerEntity.BoundingBox, drillroomExitBoundingBox) { // Is exiting
		if !hasPlayerLeftDrillBase { // STEP [2] // Avoid glitches (also quick dodge to not-exit)
			hasPlayerLeftDrillBase = true

			// Play exit sounds
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("footstep0%d.ogg", rl.GetRandomValue(0, 9)))))  // 05
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "metalClick.ogg")))                                         // metalClick
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("creak%d.ogg", rl.GetRandomValue(1, 3)))))      // 3
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("doorClose_%d.ogg", rl.GetRandomValue(1, 4))))) // 4

			// Save screen state
			finishScreen = 2                      // 1=>ending 2=>gameplay(openworldroom)
			camera.Up = rl.NewVector3(0., 1., 0.) // Reset yaw/pitch/roll
			// TODO: implement drillroom save/load functions (data and filenames)
			// saveCoreLevelState()                  // (player,camera,...) 705 bytes
			// saveAdditionalLevelState()            // (blocks,...)        82871 bytes
		}
	} else { // Is still inside
		if !hasPlayerLeftDrillBase { // RESET FLAG (just-in-case)
			hasPlayerLeftDrillBase = false // STEP [1] (maybe)
		}
	}

	// Recheck binary logic
	if hasPlayerLeftDrillBase {
		player.SetColor(rl.Blue)
	} else {
		player.SetColor(rl.Green)
	}

	// Check player collisions with instruments
	for i := range triggerCount {
		isPlayerNearTriggerSensors[i] = rl.CheckCollisionBoxes(playerEntity.BoundingBox, triggerSensorBoundingBoxes[i])
		if rl.CheckCollisionBoxes(playerEntity.BoundingBox, triggerBoundingBoxes[i]) {
			player.RevertPlayerAndCameraPositions(&playerEntity, oldPlayer, &camera, oldCam)
		}
	}
}

func Draw() {
	// TODO: Draw ending screen here!
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	rl.ClearBackground(rl.RayWhite)

	playerEntity.Draw()
	floorEntity.Draw()
	rl.DrawBoundingBox(drillroomExitBoundingBox, rl.Green)
	wall.DrawBatch(common.DrillRoom, floorEntity.Position, floorEntity.Size, cmp.Or(rl.NewVector3(5, 2, 5), common.Vector3One))

	for i := range triggerCount {
		// Circular model shape --expand-> to 1x1x1 bounding box
		const k = 1. + common.OneMinusInvPhi // math.Pi / 2.
		var scale rl.Vector3
		var col color.RGBA

		if isPlayerNearTriggerSensors[i] {
			scale = rl.Vector3{X: k * 1.25, Y: k * 1.25, Z: k * 1.25}
			col = rl.Purple
		} else {
			scale = rl.Vector3{X: k, Y: k, Z: k}
			col = rl.Pink
		}
		rl.DrawModelEx(triggerModels[i], triggerPositions[i], common.YAxis, 0., scale, rl.White)

		if true {
			rl.DrawBoundingBox(triggerBoundingBoxes[i], rl.Fade(rl.SkyBlue, 0.1))
			rl.DrawBoundingBox(triggerSensorBoundingBoxes[i], rl.Fade(col, 0.1))
		}
	}

	rl.EndMode3D()

	// 2D World
	if false {
		rl.DrawRectangle(0, 0, screenW, screenH, rl.Fade(rl.Black, .8))
	}

	fontSize := float32(common.Font.Primary.BaseSize) * 3.0

	pos := rl.NewVector2(
		float32(screenW)/2.-float32(rl.MeasureText(screenTitleText, int32(fontSize*common.Phi)))/2.,
		float32(screenH)/16.,
	)
	rl.DrawTextEx(common.Font.Primary, screenTitleText, pos, (fontSize * common.Phi), 4, rl.Fade(rl.Black, .4))

	pos = rl.NewVector2(
		float32(screenW)/2.-float32(rl.MeasureText("ROOM", int32(fontSize)))/2.,
		float32(screenH)/16.+(fontSize*common.Phi),
	)
	rl.DrawTextEx(common.Font.Primary, "ROOM", pos, fontSize, common.Phi, rl.Fade(rl.LightGray, .7))

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
