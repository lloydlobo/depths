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
	"example/depths/internal/currency"
	"example/depths/internal/floor"
	"example/depths/internal/hud"
	"example/depths/internal/player"
	"example/depths/internal/util/mathutil"
	"example/depths/internal/wall"
)

const (
	screenTitleText    = "DRILL"
	screenSubtitleText = "leave room: backspace swipe-left\nquit:          F10 pinch-out right-mouse-button"
)

var (
	// Core data

	finishScreen  int
	framesCounter int32

	camera                 rl.Camera3D
	xFloor                 floor.Floor
	xPlayer                player.Player // Player Entity or xPlayer
	hasPlayerLeftDrillBase bool
)

var (
	// NOTE: AVOID using common.SavedgameSlotData.CurrentLevelID as reference
	// directly.. We must init levelID with it to maintain consistency for now
	// FIXME: UNUSED <<<---------------------------------------
	levelID int32

	// WARN: DONT NEED IT HERE
	//       JUST READ DATA FROM FILE
	hitCount int32
	hitScore int32

	currencyItems [currency.MaxCurrencyTypes]currency.CurrencyItem
)

var (
	drillroomExitBoundingBox rl.BoundingBox
)

var (
	triggerChangeResourceCurrencyTypeState = currency.Copper + currency.CurrencyType(1) // 0:Copper + 1:Pearl
)

type TriggerType uint8

const (
	TriggerDigFaster TriggerType = iota
	TriggerDigHarder
	TriggerDigBigger
	TriggerDigMoveFaster
	TriggerGetTougher
	TriggerMakeResource
	TriggerChangeResource
	TriggerCarryMore
	TriggerStartDrill
	TriggerRefuelDrill

	MaxTriggerCount
)

// File Struct
var (
	triggerBoundingBoxes       [MaxTriggerCount]rl.BoundingBox
	triggerSensorBoundingBoxes [MaxTriggerCount]rl.BoundingBox
	triggerPositions           [MaxTriggerCount]rl.Vector3
	triggerLabels              [MaxTriggerCount]string
	triggerScreenPositions     [MaxTriggerCount]rl.Vector2
	isPlayerNearTriggerSensors [MaxTriggerCount]bool
	isTriggerActive            [MaxTriggerCount]bool

	triggerModels [MaxTriggerCount]rl.Model
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

	levelID = int32(common.SavedgameSlotData.CurrentLevelID)

	if !rl.IsMusicStreamPlaying(common.Music.DrillRoom000) {
		rl.PlayMusicStream(common.Music.DrillRoom000)
	}

	// Core resources
	player.SetupPlayerModel()
	player.ToggleEquippedModels([player.MaxBoneSockets]bool{false, false, false}) // Unequip hat sword shield
	floor.SetupFloorModel()
	wall.SetupWallModel(common.DrillRoom)

	// Core data
	player.InitPlayer(&xPlayer, camera)
	xFloor = floor.NewFloor(common.Vector3Zero, rl.NewVector3(10, 0.001*2, 10)) // 1:1 ratio

	// Layout copied from https://annekatran.itch.io/dig-and-delve
	triggerSize := rl.NewVector3(.5, .5, .5)
	triggerPosY := triggerSize.Y / 2.

	kx := (xFloor.Size.X / (1. * math.Pi)) - 1.
	kz := (xFloor.Size.Z / (1. * math.Pi)) - 2.

	// 45 degree tangent lines (use cos/sin??)
	dx := float32(0.15 + triggerSize.X)
	dz := float32(0.40 + triggerSize.Z)

	for i, v := range []struct {
		Position rl.Vector3
		Label    string
	}{
		// Clockwise  starting from 9 o'clock
		TriggerDigFaster:      {Position: rl.NewVector3(-kx, triggerPosY, -kz), Label: "DIG FASTER"},              // NW
		TriggerDigHarder:      {Position: rl.NewVector3(-kx+dx, triggerPosY, -kz-dz), Label: "DIG HARDER"},        // NW -> NE
		TriggerDigBigger:      {Position: rl.NewVector3(-kx+dx+dx, triggerPosY, -kz-dz-dz), Label: "DIG BIGGER"},  // NW -> NE -> NE
		TriggerDigMoveFaster:  {Position: rl.NewVector3(+kx-dx-dx, triggerPosY, -kz-dz-dz), Label: "MOVE FASTER"}, // NE -> NW -> NW
		TriggerGetTougher:     {Position: rl.NewVector3(+kx-dx, triggerPosY, -kz-dz), Label: "GET TOUGHER"},       // NE -> NW
		TriggerMakeResource:   {Position: rl.NewVector3(+kx, triggerPosY, +kz), Label: "MAKE RESOURCE"},           // SE
		TriggerChangeResource: {Position: rl.NewVector3(+kx-dx, triggerPosY, +kz+dz), Label: "CHANGE RESOURCE"},   // SE -> SW
		TriggerCarryMore:      {Position: rl.NewVector3(+kx, triggerPosY, -kz), Label: "CARRY MORE"},              // NE
		TriggerStartDrill:     {Position: rl.NewVector3(-kx+dx, triggerPosY, +kz+dz), Label: "START DRILL"},       // SW -> SE
		TriggerRefuelDrill:    {Position: rl.NewVector3(-kx, triggerPosY, +kz), Label: "REFUEL DRILL"},            // SW
	} {
		triggerPositions[i] = v.Position

		triggerLabels[i] = v.Label

		const text3DOffsetY = .5
		triggerScreenPositions[i] =
			rl.GetWorldToScreen(rl.NewVector3(v.Position.X,
				v.Position.Y+text3DOffsetY, v.Position.Z), camera)

		triggerBoundingBoxes[i] =
			common.GetBoundingBoxPositionSizeV(v.Position, triggerSize)

		triggerSensorBoundingBoxes[i] =
			common.GetBoundingBoxPositionSizeV(v.Position,
				rl.Vector3Scale(triggerSize, 2))

		isPlayerNearTriggerSensors[i] = false
		isTriggerActive[i] = true

		dir := filepath.Join("res", "kenney_prototype-kit", "Models", "OBJ format")
		texture := rl.LoadTexture(filepath.Join(dir, "Textures", "colormap.png"))
		var model rl.Model
		switch TriggerType(i) {
		case TriggerStartDrill:
			model = rl.LoadModel(filepath.Join(dir, "button-floor-round.obj"))
		case TriggerChangeResource:
			model = rl.LoadModel(filepath.Join(dir, "lever-double.obj"))
		default:
			// model = rl.LoadModel(filepath.Join(dir, "column-rounded-low.obj"))
			// model = rl.LoadModel(filepath.Join(dir, "column-triangle-low.obj"))
			// model = rl.LoadModel(filepath.Join(dir, "column-low-low.obj"))
			model = rl.LoadModel(filepath.Join(dir, "weapon-shield.obj"))
			triggerModels[i] = model
		}
		rl.SetMaterialTexture(model.Materials, rl.MapDiffuse, texture)
		triggerModels[i] = model
	}

	// TEMPORARY
	if __IS_TEMPORARY__ := false; __IS_TEMPORARY__ {
		for i := range MaxTriggerCount {
			isTriggerActive[i] = false
		}
		isTriggerActive[TriggerStartDrill] = true // Only enable drill
	}

	// Compute once
	drillroomExitBoundingBox = common.GetBoundingBoxPositionSizeV(
		xFloor.Position,
		rl.Vector3Subtract(xFloor.Size, rl.NewVector3(1+xPlayer.Size.X/2, -xPlayer.Size.Y*2, 1+xPlayer.Size.Z/2)),
	)

	currency.LoadCurrencyItems(&currencyItems)
	fmt.Printf("currencyItems: %v\n", currencyItems)

	// For camera thirdperson view
	rl.DisableCursor()
}

func Update() {
	rl.UpdateMusicStream(common.Music.DrillRoom000)

	// PERF: Just check if player is not colliding wit floor bounding box * scale of 0.9

	// Save variables this frame
	oldCam := camera
	oldPlayer := xPlayer

	// Reset flags/variables
	xPlayer.Collisions = rl.Quaternion{}
	xPlayer.IsPlayerWallCollision = false

	// Update the game camera for this screen
	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	// Reset camera yaw(y-axis)/roll(z-axis) (on key [W] or [E])
	if got, want := camera.Up, (rl.Vector3{X: 0., Y: 1., Z: 0.}); !rl.Vector3Equals(got, want) {
		camera.Up = want
	}

	xPlayer.Update(camera, xFloor)
	if xPlayer.IsPlayerWallCollision {
		player.RevertPlayerAndCameraPositions(&xPlayer, oldPlayer, &camera, oldCam)
	}

	// Update playerl leaving common.DrillRoom => common.Opcommon.OpenWorldRoom
	if !rl.CheckCollisionBoxes(xPlayer.BoundingBox, drillroomExitBoundingBox) { // Is exiting
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

			currency.HandleWalletToBankTransaction(&currencyItems)
			currency.SaveCurrencyItems(currencyItems)

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
		if rl.CheckCollisionBoxes(
			xPlayer.BoundingBox,
			triggerBoundingBoxes[i],
		) {
			xPlayer.Collisions.X = 1
			xPlayer.Collisions.Z = 1
			player.RevertPlayerAndCameraPositions(&xPlayer, oldPlayer, &camera, oldCam)
		}

		// Disable everything apart from "Start drill" trigger for now
		if __IS_TEMPORARY__ := false; __IS_TEMPORARY__ {
			if !isTriggerActive[i] {

				continue
			}
		}

		isPlayerNearTriggerSensors[i] = rl.CheckCollisionBoxes(xPlayer.BoundingBox, triggerSensorBoundingBoxes[i])
	}

	for i := range MaxTriggerCount {
		if isPlayerNearTriggerSensors[i] && rl.IsKeyPressed(rl.KeyF) {
			HandleTriggerOnPlayerPressF(TriggerType(i))
		}
	}

	// Update screen position after accumulating all player entity collisions with trigger
	for i := range MaxTriggerCount {
		pos := triggerPositions
		cam := camera
		if xPlayer.Collisions.X != 0 || xPlayer.Collisions.Z != 0 {
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

	// TODO: Move this in package player (if possible)
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyD) {
		const fps = 60.0
		const framesInterval = fps / 2.5
		if framesCounter%int32(framesInterval) == 0 {
			if !rl.Vector3Equals(oldPlayer.Position, xPlayer.Position) &&
				rl.Vector3Distance(oldCam.Position, xPlayer.Position) > 1.0 &&
				(xPlayer.Collisions.X == 0 && xPlayer.Collisions.Z == 0) {
				rl.PlaySound(common.FXS.ImpactFootStepsConcrete[int(framesCounter)%len(common.FXS.ImpactFootStepsConcrete)])
			}
		}
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

	xPlayer.Draw()
	xFloor.Draw()
	{
		scale := cmp.Or(rl.NewVector3(5, 2, 5), common.Vector3One)
		wall.DrawBatch(common.DrillRoom, xFloor.Position, xFloor.Size, scale)
	}

	for i := range MaxTriggerCount {
		// Circular model shape --expand-> to 1x1x1 bounding box
		const k = 1. + common.OneMinusInvPhi

		var (
			scale rl.Vector3
			col   color.RGBA
		)

		if isPlayerNearTriggerSensors[i] {
			scale = rl.Vector3{X: k * 1.25, Y: k * 1.25, Z: k * 1.25}
			col = rl.Purple
		} else {
			scale = rl.Vector3{X: k, Y: k, Z: k}
			col = rl.Pink
		}

		if true {
			rl.DrawBoundingBox(triggerBoundingBoxes[i], rl.Fade(rl.SkyBlue, 0.1))
			rl.DrawBoundingBox(triggerSensorBoundingBoxes[i], rl.Fade(col, 0.1))
		}

		rl.PushMatrix()
		rl.Translatef(triggerPositions[i].X, triggerPositions[i].Y, triggerPositions[i].Z)
		switch TriggerType(i) {
		case TriggerCarryMore:

		case TriggerChangeResource:
			const resourceRadius = 0.22 / 2 // Hologram
			const padInlineStart = 0.32
			col := currency.ToColorMap[triggerChangeResourceCurrencyTypeState]
			if false {
				rl.DrawSphereEx(rl.NewVector3(padInlineStart, 1, 0), resourceRadius, 6, 6, col)
			}

		case TriggerDigBigger:
		case TriggerDigFaster:
		case TriggerDigHarder:
		case TriggerDigMoveFaster:
		case TriggerGetTougher:
		case TriggerMakeResource:

		case TriggerRefuelDrill:
			rl.Scalef(scale.X, scale.Y, scale.Z) // WARN: This works till PopMatrix()

			const y1 = (1.0 / (8.0 * 2))
			const h1 = (1.0 / math.Pi)
			const y2 = y1 + h1
			const h2 = 0.0625
			const radius1 = 0.20
			const radius2 = radius1 * 0.8

			// NOTE: Maintain the draw order to avoid top part of the cylinder adding with bottom part of frustum like cover
			rl.DrawCylinderEx(rl.NewVector3(0.0, y2, 0.0), rl.NewVector3(0.0, y2+h2, 0.0), radius1, radius2, 32, rl.Fade(rl.DarkGray, 0.2)) // Frustum Top-cover
			rl.DrawCylinderEx(rl.NewVector3(0.0, y1, 0.0), rl.NewVector3(0.0, y1+h1, 0.0), radius1, radius1, 32, rl.Fade(rl.DarkGray, 0.2)) // Cylindric Sides

		case TriggerStartDrill:

		default:
			panic("unexpected drillroom.TriggerType")
		}
		rl.PopMatrix()

		rl.DrawModelEx(triggerModels[i], triggerPositions[i], common.YAxis, 0., scale, rl.White)

	}

	rl.EndMode3D()

	// 2D World

	// Gradually fade in text wait for a second to reset World to Screen Coordinates
	// One second fade in duration when fps==60
	alpha := min(1., float32(framesCounter)/60.)
	col := rl.Fade(rl.White, alpha)

	// Draw trigger index
	// See https://www.raylib.com/examples/core/loader.html?name=core_world_screen
	if false {
		for i := range MaxTriggerCount {
			text := fmt.Sprintf("%d", i)
			stringSize := rl.MeasureTextEx(common.Font.SourGummy, text, float32(common.Font.SourGummy.BaseSize), 2)
			position := rl.NewVector2(triggerScreenPositions[i].X-stringSize.X, triggerScreenPositions[i].Y-stringSize.Y)
			rl.DrawTextEx(common.Font.SourGummy, text, position, float32(common.Font.SourGummy.BaseSize), 2, col)
		}
	}

	// Draw description over for each trigger
	for i := range MaxTriggerCount {
		srcPos := triggerPositions[i]                                                           // World 3d
		dstPos0 := rl.GetWorldToScreen(rl.NewVector3(srcPos.X, srcPos.Y+1.0, srcPos.Z), camera) // Screen 2d
		dstPos1 := rl.GetWorldToScreen(rl.NewVector3(srcPos.X, srcPos.Y+0.5, srcPos.Z), camera) // Screen 2d
		_ = dstPos1

		rl.PushMatrix()
		rl.Translatef(dstPos0.X, dstPos0.Y, 0)

		switch TriggerType(i) {
		case TriggerCarryMore:
		case TriggerChangeResource:
		case TriggerDigBigger:
		case TriggerDigFaster:
		case TriggerDigHarder:
		case TriggerDigMoveFaster:
		case TriggerGetTougher:
		case TriggerMakeResource:
			var availableCopperQuantity int32 // co base currency::Copper
			// TEMPORARY
			availableCopperQuantity = int32(30)

			if triggerChangeResourceCurrencyTypeState == currency.Copper { // Skip over base currency copper
				panic(fmt.Sprintf("expected drillroom.TriggerType %d to be skipped", triggerChangeResourceCurrencyTypeState))
			}

			var (
				pixelSize = float32(screenW) / float32(screenH)
				spacing   = float32(1.5)

				currencyID           = triggerChangeResourceCurrencyTypeState
				currencyCol          = currency.ToColorMap[currencyID]
				currencyString       = currency.ToStringMap[currencyID]
				currencyToCopperUnit = currency.ToCopperUnitsMap[currencyID]
				convertedAmount      = availableCopperQuantity / currencyToCopperUnit

				debugText     = fmt.Sprintf("%s::%d %dco=>%d", currencyString, currencyID, currencyToCopperUnit, convertedAmount)
				debugStrSize  = rl.MeasureTextEx(common.Font.SourGummy, debugText, float32(common.Font.SourGummy.BaseSize), spacing)
				debugPosition = rl.NewVector2(0-debugStrSize.X/2, 0-float32(common.Font.SourGummy.BaseSize*2)-debugStrSize.Y/2)
				_             = debugPosition

				actualText     = fmt.Sprintln(currencyToCopperUnit) // This much is required for 1 of currency to change into
				actualStrSize  = rl.MeasureTextEx(common.Font.SourGummy, actualText, float32(common.Font.SourGummy.BaseSize), spacing)
				actualPosition = rl.NewVector2(0+0*pixelSize*5-actualStrSize.X/2, 0-actualStrSize.Y/4)

				iconLargePosition  = rl.NewVector2(actualPosition.X+actualStrSize.X/2, actualPosition.Y-8*2)
				iconPosition       = rl.NewVector2(actualPosition.X+actualStrSize.X+8*common.Phi, 0*actualPosition.Y)
				iconLargeRadius    = float32(8 + 8/2)
				iconSmallRadius    = float32(8)
				segmentsRingBuffer = []int32{3, 4, 5, 6}
				segments           = segmentsRingBuffer[int(currencyID)%len(segmentsRingBuffer)]
				startAngle         = float32(currencyID)*15 + float32(segments)*15
			)

			// Create unique shapes for each currency
			rl.DrawRing(iconLargePosition, 0, iconLargeRadius, startAngle, 360+startAngle, segments, rl.Fade(currencyCol, 0.7)) // Other
			// Draw Copper Quantity required for switched resource type and draw copper icon next to it
			rl.DrawRing(iconPosition, 0, iconSmallRadius, 0, 360, 6, rl.Fade(currency.ToColorMap[currency.Copper], 0.7))
			var availableCol color.RGBA
			if convertedAmount < 1 {
				availableCol = rl.Purple
			} else {
				availableCol = rl.White
			}
			rl.DrawTextEx(common.Font.SourGummy, actualText, actualPosition, float32(common.Font.SourGummy.BaseSize), spacing, rl.Fade(availableCol, 0.8))

		case TriggerRefuelDrill:
			refuelGoalCurrencyTypes := []currency.CurrencyType{
				currency.Copper,
				currency.Pearl,
				currency.Bronze,
				currency.Silver,
				currency.Ruby,
				currency.Gold,
				currency.Diamond,
				currency.Sapphire,
			}

			type TriggerRefuelDrillDataSOA struct {
				Currency    [currency.MaxCurrencyTypes * 2]currency.CurrencyType
				CopperUnits [currency.MaxCurrencyTypes * 2]int32
			}

			triggerData := TriggerRefuelDrillDataSOA{
				Currency: [currency.MaxCurrencyTypes * 2]currency.CurrencyType{
					currency.Pearl, currency.Pearl, currency.Bronze,
					currency.Bronze, currency.Silver, currency.Silver,
					currency.Ruby, currency.Ruby, currency.Gold, currency.Gold,
					currency.Diamond, currency.Diamond, currency.Sapphire,
					currency.Sapphire, currency.Sapphire, currency.Sapphire,
				},
				CopperUnits: [currency.MaxCurrencyTypes * 2]int32{
					80, 90, 100, 110, 120, 130, 150, 175, 180, 190, 200, 210,
					220, 230, 240, 255,
				},
			}

			id := triggerData.Currency[levelID]
			if id > currency.CurrencyType(len(common.SavedgameSlotData.AllLevelIDS)) {
				panic(fmt.Sprintf("%s", "id > currency.CurrencyType(len(common.SavedgameSlotData.AllLevelIDS))"))
			}

			refuelGoalCurrencyType := refuelGoalCurrencyTypes[id]
			const multiplier = common.Phi
			var (
				pixelSize = float32(screenW) / float32(screenH)
				spacing   = float32(1.5)

				currencyID     = triggerChangeResourceCurrencyTypeState
				currencyCol    = currency.ToColorMap[currencyID]
				currencyString = currency.ToStringMap[currencyID]

				currencyToCopperUnit = currency.ToCopperUnitsMap[refuelGoalCurrencyType]

				// refuelGoal    = int32(mathutil.FloorF(float32(id)*multiplier*100*float32(currencyToCopperUnit))) / 100
				refuelGoal    = triggerData.CopperUnits[id]
				actualText    = fmt.Sprintln(refuelGoal) // This much is required for 1 of currency to change into
				actualStrSize = rl.MeasureTextEx(common.Font.SourGummy, actualText, float32(common.Font.SourGummy.BaseSize), spacing)

				actualPosition     = rl.NewVector2(0+0*pixelSize*5-actualStrSize.X/2, 0-actualStrSize.Y/4)
				iconSmallPosition  = rl.NewVector2(actualPosition.X+actualStrSize.X+8*common.Phi, 0*actualPosition.Y)
				iconLargePosition  = rl.NewVector2(actualPosition.X+actualStrSize.X/2, actualPosition.Y-8*2)
				iconPosition       = rl.NewVector2(actualPosition.X+actualStrSize.X+8*common.Phi, 0*actualPosition.Y)
				iconLargeRadius    = float32(8 + 8/2)
				iconSmallRadius    = float32(8)
				segmentsRingBuffer = []int32{3, 4, 5, 6}
				segments           = segmentsRingBuffer[int(refuelGoalCurrencyType)%len(segmentsRingBuffer)]
				startAngle         = float32(currencyID)*15 + float32(segments)*15

				_ = currencyCol
				_ = currencyString
				_ = iconSmallPosition
				_ = iconLargePosition
				_ = iconPosition
				_ = iconLargeRadius
				_ = iconSmallRadius
				_ = startAngle

				convertedAmount = currencyToCopperUnit
			)
			rl.DrawRing(iconLargePosition, 0, iconLargeRadius, startAngle, 360+startAngle, segments, rl.Fade(currency.ToColorMap[refuelGoalCurrencyType], 0.7)) // Other

			// rl.DrawRing(iconPosition, 0, iconSmallRadius, 0, 360, 6, rl.Fade(currency.CurrencyColorMap[refuelGoalCurrencyType], 0.7))
			var availableCol color.RGBA
			if convertedAmount < 1 {
				availableCol = rl.Purple
			} else {
				availableCol = rl.White
			}

			rl.DrawTextEx(common.Font.SourGummy, actualText, actualPosition, float32(common.Font.SourGummy.BaseSize), spacing, rl.Fade(availableCol, 0.8))

		case TriggerStartDrill:

		default:
			panic("unexpected drillroom.TriggerType")
		}
		rl.PopMatrix()
	}

	// Draw description on HUD for each trigger
	instructionPosY := float32(screenH) - 40
	for i := range MaxTriggerCount {
		textCol := rl.Fade(rl.Black, .6)
		bgCol := rl.RayWhite

		if isPlayerNearTriggerSensors[i] {
			const maxLabelLenForFontSizeX2 = 148

			fontSize := float32(common.Font.SourGummy.BaseSize) * 2
			text := triggerLabels[i]
			pos := rl.NewVector2(float32(screenW)/2-maxLabelLenForFontSizeX2*2./3., instructionPosY)

			rl.DrawRectangleRounded(rl.NewRectangle(pos.X-2, pos.Y-2, fontSize+4, fontSize+4), .3, 16, textCol)
			rl.DrawTextEx(common.Font.SourGummy, "F", rl.NewVector2(pos.X+2+2+1, pos.Y+2), fontSize-2, 1.0, bgCol)
			rl.DrawTextEx(common.Font.SourGummy, text, rl.Vector2{X: pos.X + fontSize + fontSize/2 + 1, Y: pos.Y}, fontSize, 2, textCol)

			break // Avoid overlapping text
		}
	}

	hud.DrawHUD(xPlayer, hitScore, currencyItems)

	if f := float32(framesCounter) / 60.; (alpha >= 1.) && (f > 2. && f < 1000.) {
		delta := mathutil.PowF(float32(rl.GetTime()), 1.5-(2.0/f))
		delta *= rl.GetFrameTime()
		alpha = max(0., alpha-delta)
	} else if f >= 1000. {
		alpha = 0.
	} else { // Initial delay on screen start
		alpha *= .5 * f
	}

	{
		font := common.Font.SourGummy
		strSize := rl.MeasureTextEx(font, screenTitleText, float32(font.BaseSize), 1.0)
		pos := rl.NewVector2(float32(screenW)/2-strSize.X/2, float32(screenH)/16-strSize.Y/2)
		rl.DrawTextEx(font, screenTitleText, pos, float32(font.BaseSize), 1.0, rl.Fade(rl.Black, 0.5+0.5*(alpha)))
	}

	{
		font := common.Font.SourGummy
		text := "ROOM"
		strSize := rl.MeasureTextEx(font, text, float32(font.BaseSize), 1.0)
		pos := rl.NewVector2(float32(screenW)/2-strSize.X/2, float32(screenH)/16-strSize.Y/2+float32(font.BaseSize))
		rl.DrawTextEx(font, text, pos, float32(font.BaseSize), 1.0, rl.Fade(rl.Gray, 0.5+0.7*(alpha)))
	}

	{
		fontSize := float32(20. - 9.)
		subtextSize := rl.MeasureTextEx(common.Font.SourGummy, screenSubtitleText, fontSize, 1)
		position := rl.NewVector2(float32(screenW)/2-subtextSize.X/2, min(instructionPosY-40-fontSize, float32(screenH)-subtextSize.Y*3))
		rl.DrawTextEx(common.Font.SourGummy, screenSubtitleText, position, fontSize, 1.0, rl.Fade(rl.Gray, 1.0*alpha))
	}

	if true {
		rl.DrawText(fmt.Sprint(rl.GetFrameTime()), 10, 30, 20, rl.Green)
		rl.DrawText(fmt.Sprint(framesCounter), 10, 50, 20, rl.Green)
	}
}

func Unload() {
	// TODO: Unload gameplay screen variables here!
	// 1 is ending screen
	if isTransToGameScreen := finishScreen == 1; !isTransToGameScreen && rl.IsCursorHidden() {
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

func HandleTriggerOnPlayerPressF(i TriggerType) {
	switch i {

	case TriggerDigFaster:
		rl.PlaySound(common.FX.InterfaceBong)

	case TriggerDigHarder:
		rl.PlaySound(common.FX.InterfaceBong)

	case TriggerDigBigger:
		rl.PlaySound(common.FX.InterfaceBong)

	case TriggerDigMoveFaster:
		rl.PlaySound(common.FX.InterfaceBong)

	case TriggerGetTougher:
		rl.PlaySound(common.FX.InterfaceBong)

	case TriggerMakeResource:
		rl.PlaySound(common.FX.InterfaceBong)

	case TriggerChangeResource:
		common.PlayRandomFromSounds(common.FXS.InterfaceClick)
		rl.PlaySound(common.FX.InterfaceScratch)
		triggerChangeResourceCurrencyTypeState = triggerChangeResourceCurrencyTypeState.Next()
		if triggerChangeResourceCurrencyTypeState == currency.Copper { // Skip over base currency copper
			triggerChangeResourceCurrencyTypeState++
		}

	case TriggerCarryMore:
		rl.PlaySound(common.FX.InterfaceBong)

	case TriggerStartDrill:
		var canDrill bool

		if __IS_TEMPORARY__ := true; __IS_TEMPORARY__ {
			if isSuccess := true; isSuccess { // Force success
				canDrill = hitCount == 0
			} else {
				canDrill = hitCount == xPlayer.MaxCargoCapacity
			}
		}

		if !canDrill {
			rl.PlaySound(common.FX.InterfaceErrorSemiDown)
			rl.PlaySound(common.FX.InterfaceBong)
		} else {
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_sci-fi-sounds", "Audio", "lowFrequency_explosion_000.ogg")))
			common.PlayRandomFromSounds(common.FXS.InterfaceConfirmation)

			// Transition to next level/screen
			// NOTE: Why does this feel so hacky? ^_^
			// NOTE: IDs are non-zero (unsigned) integers
			// currLevelID := common.SavedgameSlotData.CurrentLevelID
			finalLevelID := uint8(len(common.SavedgameSlotData.AllLevelIDS))
			common.SavedgameSlotData.CurrentLevelID = min(finalLevelID, uint8(levelID)+1)

			if uint8(levelID) >= finalLevelID {
				finishScreen = 1 // => ending (gameover)
			} else {
				finishScreen = 2 // => gameplay (next-level)
				common.SavedgameSlotData.UnlockedLevelIDS = append(common.SavedgameSlotData.UnlockedLevelIDS, uint8(levelID))
			}
		}

	case TriggerRefuelDrill:
		rl.PlaySound(common.FX.InterfaceBong)

	default:
		panic("unexpected drillroom.TriggerType")
	}
}
