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
	"example/depths/internal/util/mathutil"
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
	// WARN: DONT NEED IT HERE
	//       JUST READ DATA FROM FILE
	hitCount int32
	hitScore int32
)

var (
	drillroomExitBoundingBox rl.BoundingBox
)

var (
	triggerBoundingBoxes       [MaxTriggerCount]rl.BoundingBox
	triggerSensorBoundingBoxes [MaxTriggerCount]rl.BoundingBox
	triggerPositions           [MaxTriggerCount]rl.Vector3
	triggerLabels              [MaxTriggerCount]string
	triggerScreenPositions     [MaxTriggerCount]rl.Vector2
	isPlayerNearTriggerSensors [MaxTriggerCount]bool

	triggerModels [MaxTriggerCount]rl.Model
)

const (
	MaxTriggerCount int32 = 10
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
	// triggerPositions = []rl.Vector3{
	// 	// Corners
	// 	rl.NewVector3(-2, triggerPosY, -2),
	// 	rl.NewVector3(+2, triggerPosY, -2),
	// 	rl.NewVector3(-2, triggerPosY, +2),
	// 	rl.NewVector3(+2, triggerPosY, +2),
	//
	// 	// Sides
	// 	rl.NewVector3(-2, triggerPosY, +0),
	// 	rl.NewVector3(+2, triggerPosY, +0),
	// 	rl.NewVector3(+0, triggerPosY, -2),
	// 	rl.NewVector3(+0, triggerPosY, +2),
	// }
	kx := (floorEntity.Size.X / (1. * math.Pi)) - 1.
	kz := (floorEntity.Size.Z / (1. * math.Pi)) - 2.

	// 45 degree tangent lines (use cos/sin??)
	dx := float32(0.15 + triggerSize.X)
	dz := float32(0.40 + triggerSize.Z)

	for i, v := range [MaxTriggerCount]struct {
		Position rl.Vector3
		Label    string
	}{
		0: {Position: rl.NewVector3(-kx, triggerPosY, -kz), Label: "DIG FASTER"},              // NW				|-- Upper corners --|
		1: {Position: rl.NewVector3(+kx, triggerPosY, -kz), Label: "CARRY MORE"},              // NE
		2: {Position: rl.NewVector3(-kx+dx, triggerPosY, -kz-dz), Label: "DIG HARDER"},        // NW -> NE		|--Upper arcs--|
		3: {Position: rl.NewVector3(-kx+dx+dx, triggerPosY, -kz-dz-dz), Label: "DIG BIGGER"},  // NW -> NE -> NE
		4: {Position: rl.NewVector3(+kx-dx, triggerPosY, -kz-dz), Label: "GET TOUGHER"},       // NE -> NW
		5: {Position: rl.NewVector3(+kx-dx-dx, triggerPosY, -kz-dz-dz), Label: "MOVE FASTER"}, // NE -> NW -> NW
		6: {Position: rl.NewVector3(-kx, triggerPosY, +kz), Label: "REFUEL DRILL"},            // SW				|-- Lower corners --|
		7: {Position: rl.NewVector3(+kx, triggerPosY, +kz), Label: "MAKE RESOURCE"},           // SE
		8: {Position: rl.NewVector3(-kx+dx, triggerPosY, +kz+dz), Label: "START DRILL"},       // SW -> SE		|-- Lower arcs --|
		9: {Position: rl.NewVector3(+kx-dx, triggerPosY, +kz+dz), Label: "CHANGE RESOURCE"},   // SE -> SW
	} {
		triggerPositions[i] = v.Position

		triggerLabels[i] = v.Label

		const text3DOffsetY = .5
		triggerScreenPositions[i] =
			rl.GetWorldToScreen(rl.NewVector3(v.Position.X,
				v.Position.Y+text3DOffsetY, v.Position.Z), camera)

		triggerBoundingBoxes[i] =
			common.GetBoundingBoxFromPositionSizeV(v.Position, triggerSize)

		triggerSensorBoundingBoxes[i] =
			common.GetBoundingBoxFromPositionSizeV(v.Position,
				rl.Vector3Scale(triggerSize, 2))

		isPlayerNearTriggerSensors[i] = false

		dir := filepath.Join("res", "kenney_prototype-kit", "Models", "OBJ format")
		model := rl.LoadModel(filepath.Join(dir, "weapon-shield.obj"))
		texture := rl.LoadTexture(filepath.Join(dir, "Textures", "colormap.png"))
		rl.SetMaterialTexture(model.Materials, rl.MapDiffuse, texture)
		triggerModels[i] = model
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
	for i := range MaxTriggerCount {
		isPlayerNearTriggerSensors[i] = rl.CheckCollisionBoxes(
			playerEntity.BoundingBox,
			triggerSensorBoundingBoxes[i],
		)

		if rl.CheckCollisionBoxes(
			playerEntity.BoundingBox,
			triggerBoundingBoxes[i],
		) {
			playerEntity.Collisions.X = 1
			playerEntity.Collisions.Z = 1
			player.RevertPlayerAndCameraPositions(&playerEntity, oldPlayer, &camera, oldCam)
		}
	}
	const TriggerBeginDrillingID = 8
	if isPlayerNearTriggerSensors[TriggerBeginDrillingID] {
		if rl.IsKeyPressed(rl.KeyF) {
			panic("[F] works")
		}
	}

	// Update screen position after accumulating all player entity collisions with trigger
	for i := range MaxTriggerCount {
		pos := triggerPositions
		cam := camera
		if playerEntity.Collisions.X != 0 || playerEntity.Collisions.Z != 0 {
			cam = oldCam // Avoid glitching text position on player's X/Z movement
		}
		triggerScreenPositions[i] = rl.GetWorldToScreen(rl.NewVector3(pos[i].X, pos[i].Y+.5, pos[i].Z), cam)
	}

	// Change to ENDING screen
	if rl.IsKeyDown(rl.KeyF10) || rl.IsGestureDetected(rl.GesturePinchOut) ||
		rl.IsMouseButtonDown(rl.MouseButtonRight) {

		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "rollover3.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "switch33.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_interface-sounds", "Audio", "confirmation_001.ogg")))

		finishScreen = 1                      // 1=>ending
		camera.Up = rl.NewVector3(0., 1., 0.) // Reset yaw/pitch/roll
		// TODO: implement drillroom save/load functions (data and filenames)
		// saveCoreLevelState()                  // (player,camera,...) 705 bytes
		// saveAdditionalLevelState()            // (blocks,...)        82871 bytes
	}
	// Change to GAMEPLAY screen
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

	// Increment drillroom frames counter
	framesCounter++
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
	if false {
		rl.DrawBoundingBox(drillroomExitBoundingBox, rl.Green)
	}
	wall.DrawBatch(common.DrillRoom, floorEntity.Position, floorEntity.Size, cmp.Or(rl.NewVector3(5, 2, 5), common.Vector3One))

	for i := range MaxTriggerCount {
		// Circular model shape --expand-> to 1x1x1 bounding box
		const k = 1. + common.OneMinusInvPhi
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
		if false {
			rl.DrawBoundingBox(triggerBoundingBoxes[i], rl.Fade(rl.SkyBlue, 0.1))
			rl.DrawBoundingBox(triggerSensorBoundingBoxes[i], rl.Fade(col, 0.1))
		}
	}

	rl.EndMode3D()

	// 2D World

	if false { // TEMPORARY
		rl.DrawRectangle(0, 0, screenW, screenH, rl.Fade(rl.Black, .2))
	}

	// One second fade in duration when fps==60
	alpha := min(1., float32(framesCounter)/60.)
	col := rl.Fade(rl.White, alpha)

	// See https://www.raylib.com/examples/core/loader.html?name=core_world_screen
	for i := range MaxTriggerCount {
		text := fmt.Sprintf("%d", i)
		const fontSize = 16
		stringSize := rl.MeasureTextEx(common.Font.Primary, text, fontSize, 2)
		x := triggerScreenPositions[i].X - stringSize.X
		y := triggerScreenPositions[i].Y - stringSize.Y
		// Gradually fade in text wait for a second to reset World to Screen Coordinates
		rl.DrawTextEx(common.Font.Primary, text, rl.NewVector2(x, y), fontSize, 2, col)
	}
	instructionPosY := float32(screenH) - 40
	for i := range MaxTriggerCount {
		textCol := rl.Fade(rl.Black, .6)
		bgCol := rl.RayWhite
		if isPlayerNearTriggerSensors[i] {
			fontSize := float32(common.Font.Primary.BaseSize) * 2
			const maxLabelLenForFontSizeX2 = 148
			text := triggerLabels[i]
			pos := rl.NewVector2(float32(screenW)/2-maxLabelLenForFontSizeX2*2./3., instructionPosY)
			rl.DrawRectangleRounded(rl.NewRectangle(pos.X-2, pos.Y-2, fontSize+4, fontSize+4), .3, 16, textCol)
			rl.DrawText("F", int32(pos.X)+2+2+1, int32(pos.Y)+2, int32(fontSize)-2, bgCol)
			rl.DrawTextEx(common.Font.Primary, text, rl.Vector2{X: pos.X + fontSize + fontSize/2 + 1, Y: pos.Y}, fontSize, 2, textCol)
			break // Avoid overlapping text
		}
	}

	fontSize := float32(common.Font.Primary.BaseSize) * 3.
	if f := float32(framesCounter) / 60.; (alpha >= 1.) && (f > 2. && f < 1000.) {
		delta := mathutil.PowF(float32(rl.GetTime()), 1.5-(2.0/f))
		delta *= rl.GetFrameTime()
		alpha = max(0., alpha-delta)
	} else if f >= 1000. {
		alpha = 0.
	} else { // Initial delay on screen start
		alpha *= .5 * f
	}
	pos := rl.NewVector2(float32(screenW)/2.-float32(rl.MeasureText(screenTitleText, int32(fontSize*common.Phi)))/2., float32(screenH)/16.)
	rl.DrawTextEx(common.Font.Primary, screenTitleText, pos, (fontSize * common.Phi), 4, rl.Fade(rl.Black, .5*(alpha)))
	pos = rl.NewVector2(float32(screenW)/2.-float32(rl.MeasureText("ROOM", int32(fontSize)))/2., float32(screenH)/16.+(fontSize*common.Phi))
	rl.DrawTextEx(common.Font.Primary, "ROOM", pos, fontSize, common.Phi, rl.Fade(rl.Gray, .7*(alpha)))

	{
		fontSize := float32(20. - 9.)
		subtextSize := rl.MeasureTextEx(common.Font.Primary, screenSubtitleText, fontSize, 1)
		rl.DrawTextEx(common.Font.Primary, screenSubtitleText,
			rl.NewVector2(float32(screenW)/2-subtextSize.X/2, min(instructionPosY-40-fontSize, float32(screenH)-subtextSize.Y*3)),
			fontSize, 1, rl.Fade(rl.Gray, 1.0*alpha))
	}

	rl.DrawText(fmt.Sprint(rl.GetFrameTime()), 10, 30, 20, rl.Green)
	rl.DrawText(fmt.Sprint(framesCounter), 10, 50, 20, rl.Green)
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
