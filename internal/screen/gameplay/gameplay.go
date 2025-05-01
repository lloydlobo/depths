package gameplay

// See fog shader: https://github.com/mohsengreen1388/raylib-go-utility/blob/main/utility/fog.go

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/block"
	"example/depths/internal/common"
	"example/depths/internal/floor"
	"example/depths/internal/player"
	"example/depths/internal/projectile"
	"example/depths/internal/storage"
	"example/depths/internal/util/mathutil"
	"example/depths/internal/wall"
)

var (
	// Core data

	finishScreen  int
	framesCounter int32

	camera                 rl.Camera3D
	xFloor                 floor.Floor
	xPlayer                player.Player
	hasPlayerLeftDrillBase bool

	// Additional data

	xBlocks []block.Block
)

var (
	gPlayerRay             rl.Ray
	gPlayerRayCollision    rl.RayCollision
	playerForwardAimEndPos rl.Vector3 // Aim start is player position
	xProjectileSOA         projectile.ProjectileSOA
)

const projectileRadiusSphere = .05 // Duplicated.. but maybe wrong values
const projectileGuardDamage = (1.0 / 3.0) + .01

const (
	MaxGuards = int32(16)
)

// NPC
var (
	xGuardSOA GuardSOA
)

type GuardSOA struct { // size=1268 (0x4f4)
	Position    [MaxGuards]rl.Vector3
	Size        [MaxGuards]rl.Vector3
	BoundingBox [MaxGuards]rl.BoundingBox
	Collisions  [MaxGuards]rl.Quaternion

	Color           [MaxGuards]color.RGBA
	Rotation        [MaxGuards]float32 // (degrees) XZ plane
	Health          [MaxGuards]float32 // [0..1]
	IsActive        [MaxGuards]bool
	IsWallCollision [MaxGuards]bool
	IsWalk          [MaxGuards]bool

	CircularBufferIndex int32
}

func (gs *GuardSOA) Reset() {
	for i := range MaxGuards {
		gs.Position[i] = rl.Vector3{}
		gs.Size[i] = rl.Vector3{}
		gs.IsActive[i] = false
	}
	gs.CircularBufferIndex = 0
}

func (gs *GuardSOA) Emit(position, size rl.Vector3, rotationDegree float32) {
	gs.Position[gs.CircularBufferIndex] = position
	gs.Size[gs.CircularBufferIndex] = size

	gs.BoundingBox[gs.CircularBufferIndex] = common.GetBoundingBoxPositionSizeV(position, size)

	gs.Collisions[gs.CircularBufferIndex] = rl.Quaternion{}

	gs.Color[gs.CircularBufferIndex] = rl.White
	gs.Rotation[gs.CircularBufferIndex] = rotationDegree
	gs.Health[gs.CircularBufferIndex] = 1. // [0..1]
	gs.IsActive[gs.CircularBufferIndex] = true
	gs.IsWallCollision[gs.CircularBufferIndex] = false
	gs.IsWalk[gs.CircularBufferIndex] = true // IDK (should add timer)

	// Increment index: (ring like data structure / circular reusable buffer)
	gs.CircularBufferIndex = (gs.CircularBufferIndex + 1) % MaxGuards
}

var (
	// NOTE: AVOID using common.SavedgameSlotData.CurrentLevelID as reference
	// directly.. We must init levelID with it to maintain consistency for now
	levelID int32

	// Game stats

	hitCount int32
	hitScore int32

	money      int32
	experience int32
)

var (
	// FX control variables

	currentMusic  rl.Music
	previousMusic rl.Music
)

func Init() {
	framesCounter = 0
	finishScreen = 0

	xProjectileSOA.Reset()

	levelID = int32(common.SavedgameSlotData.CurrentLevelID)
	if levelID == 0 {
		panic("unexpected levelID")
	}

	// PERF: See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h
	// ViewCamera.target = position;
	// ViewCamera.position = Vector3Add(ViewCamera.target, Vector3{ 0, 0, CameraPullbackDistance });
	// ViewCamera.up = Vector3{ 0.0f, 1.0f, 0.0f };
	// ViewCamera.fovy = fovY;
	// ViewCamera.projection = CAMERA_PERSPECTIVE;
	cameraPullbackDistance := float32(cmp.Or(5, 3))

	camera = rl.Camera3D{
		Target: rl.NewVector3(0., .5, 0.),
		Position: cmp.Or(
			rl.Vector3Add(rl.NewVector3(0., .5, 0.), rl.NewVector3(0, 0, cameraPullbackDistance)),
			rl.NewVector3(0., 10., 10.),
		),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       15. * float32(cmp.Or(4., 3., 2.)),
		Projection: rl.CameraPerspective,
	}

	loadNewEntityData := func() {
		var mu sync.Mutex

		mu.Lock()
		defer mu.Unlock()

		finishScreen = 0
		framesCounter = 0

		// Order could be important
		player.InitPlayer(&xPlayer, camera)
		xFloor = floor.NewFloor(common.Vector3Zero, rl.NewVector3(16*2, 0.001*2, 9*2)) // 16:9 ratio // floor.InitFloor(&gameFloor)
		wall.InitWall()                                                                // NOTE: Empty func for convention

		hasPlayerLeftDrillBase = false
		xPlayer.IsPlayerWallCollision = false
	}
	loadNewAdditionalData := func() {
		var mu sync.Mutex

		mu.Lock()
		defer mu.Unlock()

		xBlocks = []block.Block{} // Clear
		block.InitBlocks(&xBlocks, block.GenerateRandomBlockPositions(xFloor))
	}
	loadNewLogicData := func() {
		var mu sync.Mutex

		mu.Lock()
		defer mu.Unlock()

		money = 1000
		experience = 0
		hitCount = 0
		hitScore = 0
	}

	const isNewGame = false

	/* xGuards = GuardSOA{
		Position:        [MaxGuards]rl.Vector3{},
		Size:            [MaxGuards]rl.Vector3{},
		BoundingBox:     [MaxGuards]rl.BoundingBox{},
		Collisions:      [MaxGuards]rl.Quaternion{},
		Rotation:        [MaxGuards]float32{},
		Health:          [MaxGuards]float32{},
		Color:           [MaxGuards]color.RGBA{},
		IsActive:        [MaxGuards]bool{},
		IsWallCollision: [MaxGuards]bool{},
		IsWalk:          [MaxGuards]bool{},
	} */
	xGuardSOA.Reset()

	// Core resources
	floor.SetupFloorModel()
	wall.SetupWallModel(common.OpenWorldRoom)
	player.SetupPlayerModel()                                                     // FIXME: in this func, use package common for models
	player.ToggleEquippedModels([player.MaxBoneSockets]bool{false, false, false}) // Unequip hat sword shield

	// Core data
	if !isNewGame {
		data, err := loadGameEntityData()
		if err == nil { // OK
			finishScreen = 0
			framesCounter = 0
			camera = data.Camera
			xFloor = data.XFloor
			xPlayer = data.XPlayer
			if true {
				hasPlayerLeftDrillBase = data.HasPlayerLeftDrillBase // If save game when far from drill and exit -> this will tell the reality
			} else {
				hasPlayerLeftDrillBase = false // How do we know?
			}
			xPlayer.IsPlayerWallCollision = false
			saveGameEntityData() // Save ASAP
		} else { // ERR
			slog.Warn(err.Error())
			loadNewEntityData()
		}
	} else {
		loadNewEntityData()
	}

	// Additional resources
	block.SetupBlockModels()

	// Additional data
	if !isNewGame {
		additionalGameData, err := loadAdditionalGameData()
		if err == nil { // OK
			xBlocks = make([]block.Block, len(additionalGameData.Blocks))
			copiedBlockCount := copy(xBlocks, additionalGameData.Blocks)
			if copiedBlockCount != 0 {
				log.Printf("blocks copied: %v", copiedBlockCount)
			} else {
				log.Panic("Incorrect saved file. Please delete it")
			}
			saveGameAdditionalData() // Save ASAP
		} else { // ERR
			slog.Warn(err.Error())
			loadNewAdditionalData()
		}
	} else {
		loadNewAdditionalData()
	}

	if !isNewGame {
		data, err := loadGameLogicData()
		if err == nil { // OK
			money = data.Money
			experience = data.Experience
			hitCount = data.HitCount
			hitScore = data.HitScore
			saveGameLogicData() // Save ASAP
		} else { // ERR
			slog.Warn(err.Error())
			loadNewLogicData()
		}
	} else {
		loadNewLogicData()
	}

	musicChoices := []rl.Music{common.Music.OpenWorld000, common.Music.OpenWorld001}
	tempMusic := musicChoices[rl.GetRandomValue(0, int32(len(musicChoices)-1))]
	if tempMusic != currentMusic {
		if rl.GetMusicTimePlayed(currentMusic) > 0 { // Already playing
			if !rl.IsMusicStreamPlaying(currentMusic) {
				rl.PlayMusicStream(currentMusic)
			}
		} else {
			if !rl.IsMusicStreamPlaying(tempMusic) {
				rl.PlayMusicStream(tempMusic)
				previousMusic = currentMusic
				currentMusic = tempMusic
			}
		}
	} else {
		counter, maxCounter := 0, 100
	GetRandomMusic:
		for {
			tempMusic = musicChoices[rl.GetRandomValue(0, int32(len(musicChoices)-1))]
			if tempMusic != currentMusic {
				break GetRandomMusic
			}
			if counter >= maxCounter {
				break GetRandomMusic
			}
			counter++
		}
		if rl.GetMusicTimePlayed(currentMusic) >= 0.5*rl.GetMusicTimeLength(currentMusic) { // Played 50% already
			rl.PlayMusicStream(tempMusic)
			previousMusic = currentMusic
			currentMusic = tempMusic
		} else {
			rl.PlayMusicStream(currentMusic) // Finally play the same music
		}
	}

	// TEMPORARY
	if false {
		slog.Warn("rl.PauseMusicStream(currentMusic)")
		rl.PauseMusicStream(currentMusic)
	}

	rl.DisableCursor() // for ThirdPersonPerspective
}

func Update() {
	rl.UpdateMusicStream(currentMusic)

	// See https://github.com/lloydlobo/tinycreatures/blob/210c4a44ed62fbb08b5f003872e046c99e288bb9/src/main.lua#L624
	for i := range projectile.MaxProjectiles {
		if !xProjectileSOA.IsActive[i] {
			continue
		}

		isKillAnim := xProjectileSOA.TimeLeft[i] <= 0

		if isKillAnim {
			xProjectileSOA.IsActive[i] = false
		} else {
			playerProjectileSpeed := 10 * rl.GetFrameTime()

			angleRad := xProjectileSOA.Rotation[i] * rl.Deg2rad

			displacement := rl.NewVector3(mathutil.CosF(angleRad)*playerProjectileSpeed, 0, mathutil.SinF(angleRad)*playerProjectileSpeed)

			xProjectileSOA.Position[i] = rl.Vector3Add(xProjectileSOA.Position[i], displacement)

			pos := xProjectileSOA.Position[i]

			isOOBX := pos.X < xFloor.BoundingBox.Min.X || pos.X > xFloor.BoundingBox.Max.X
			isOOBZ := pos.Z < xFloor.BoundingBox.Min.Z || pos.Z > xFloor.BoundingBox.Max.Z

			if isOOBX || isOOBZ {
				xProjectileSOA.IsActive[i] = false
			}
		}
	}
	xProjectileSOA.FireRateTimer -= rl.GetFrameTime()

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

	UpdatePlayerRay()

	if xPlayer.IsPlayerWallCollision {
		player.RevertPlayerAndCameraPositions(&xPlayer, oldPlayer, &camera, oldCam)
	}

	// Play player weapon sounds
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) || rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		gPlayerRay = rl.NewRay(
			rl.Vector3{
				X: xPlayer.Position.X,
				Y: xPlayer.Position.Y + xPlayer.Size.Y/4,
				Z: xPlayer.Position.Z,
			},
			rl.Vector3Multiply(
				rl.GetCameraForward(&camera),
				rl.Vector3{X: 5., Y: .125 / 2., Z: 5.}, /* playerForwardEstimateMagnitude */
			),
		)

		if projectile.FireEntityProjectile(&xProjectileSOA, xPlayer.Position, xPlayer.Size, float32(xPlayer.Rotation+90)) {
			if n := int32(len(common.FXS.SciFiLaserSmall)); n > 0 {
				rl.PlaySound(common.FXS.SciFiLaserSmall[rl.GetRandomValue(0, n-1)])
			}
		} else {
			if false {
				if n := int32(len(common.FXS.InterfaceClick)); n > 0 {
					rl.PlaySound(common.FXS.InterfaceClick[rl.GetRandomValue(0, n-1)])
				}
			}
		}
	}

	// Update block and player interaction/mining
	// TODO: Find out where player touched the box
	// WARN: Should we clear out player collision
	for i := range xBlocks {
		if xBlocks[i].IsActive &&
			xBlocks[i].State < block.MaxBlockState-1 &&
			rl.CheckCollisionBoxes(xBlocks[i].GetBlockBoundingBox(), xPlayer.BoundingBox) {
			xPlayer.Collisions.X = -mathutil.AbsF(oldPlayer.Position.X - xPlayer.Position.X)
			xPlayer.Collisions.Z = -mathutil.SignF(oldPlayer.Position.Z - xPlayer.Position.Z)
			// NOTE: It is important that player touches the block first before mining
			player.RevertPlayerAndCameraPositions(&xPlayer, oldPlayer, &camera, oldCam)

			if rl.IsKeyDown(rl.KeySpace) {
				mineFasterIndex := 3 // Higher index ~= Faster mining
				mineFasterFrames := []int32{60, 52, 48, 40, 32, 24, 20, 16, 8}
				debounceRate := mineFasterFrames[mineFasterIndex]
				isDebounce := framesCounter%debounceRate != 0
				if !isDebounce {
					handleBlockOnMining(&xBlocks[i])
				}
			}
		}
	}

	for i := range projectile.MaxProjectiles {
		if xProjectileSOA.IsActive[i] {
			for j := range xBlocks {
				if xBlocks[j].IsActive &&
					xBlocks[j].State < block.MaxBlockState-1 &&
					common.CheckCollisionPointBox(xProjectileSOA.Position[i], xBlocks[j].GetBlockBoundingBox()) {
					xProjectileSOA.IsActive[i] = false

					handleBlockOnMining(&xBlocks[j])

					// Spawn a guard: 1 out of 4 chances => 1/4 or 25% to
					if rl.GetRandomValue(1, 4) == 1 {
						rotn := float32(xPlayer.Rotation)
						size := xBlocks[j].Size
						size = rl.Vector3Scale(size, .95)
						// Since position is on the floor. and model grows
						// upwards.. this is to keep bounding box logic consistent
						var pos rl.Vector3
						if isPatchedXBlocksOriginAndBounds := false; isPatchedXBlocksOriginAndBounds {
							pos = xBlocks[j].Position
						} else {
							pos = xBlocks[j].Position
							pos.Y += size.Y / 2
						}
						xGuardSOA.Emit(pos, size, rotn)
					}
					break
				}
			}
		}
	}

	if false { // PLACEHOLDER PROTOTYPE IN Draw() for now
		for i := range MaxGuards {
			if framesCounter%4 == 0 {
				xGuardSOA.Position[i].X += rl.GetFrameTime() * float32(rl.GetRandomValue(-1, 1))
				xGuardSOA.Position[i].Z += rl.GetFrameTime() * float32(rl.GetRandomValue(-1, 1))
			}
			xGuardSOA.BoundingBox[i] = common.GetBoundingBoxPositionSizeV(xGuardSOA.Position[i], xGuardSOA.Size[i])
		}
	}

	for i := range projectile.MaxProjectiles {
		if xProjectileSOA.IsActive[i] {
			for j := range MaxGuards {
				if xGuardSOA.IsActive[j] {
					if rl.CheckCollisionBoxSphere(
						xGuardSOA.BoundingBox[j],
						xProjectileSOA.Position[i],
						projectileRadiusSphere,
					) {
						xProjectileSOA.IsActive[i] = false
						xGuardSOA.Health[j] -= projectileGuardDamage
						if xGuardSOA.Health[j] <= 0. {
							xGuardSOA.Health[j] = 0.
							xGuardSOA.IsActive[j] = false
						}
					}
				}
			}
		}
	}

	{
		origin := xFloor.Position

		bb1 := common.GetBoundingBoxPositionSizeV(origin, rl.NewVector3(3, 2, 3)) // player is inside
		bb2 := common.GetBoundingBoxPositionSizeV(origin, rl.NewVector3(5, 2, 5)) // player is entering
		bb3 := common.GetBoundingBoxPositionSizeV(origin, rl.NewVector3(7, 2, 7)) // bot barrier

		isPlayerInsideBase := rl.CheckCollisionBoxes(xPlayer.BoundingBox, bb1)
		isPlayerEnteringBase := rl.CheckCollisionBoxes(xPlayer.BoundingBox, bb2)
		isPlayerInsideBotBarrier := rl.CheckCollisionBoxes(xPlayer.BoundingBox, bb3)

		canSwitchToDrillRoom := false

		if isPlayerInsideBotBarrier && !isPlayerEnteringBase && !isPlayerInsideBase {
			player.SetColor(rl.Blue)
		} else if isPlayerEnteringBase && !isPlayerInsideBase {
			player.SetColor(rl.Green)

			if hasPlayerLeftDrillBase { // STEP [2] ─ Wait a frame before switching // Avoid glitches (also quick dodge to not-exit)
				hasPlayerLeftDrillBase = false
				canSwitchToDrillRoom = true // Actual work done here
			}
		} else if isPlayerInsideBase {
			player.SetColor(rl.Red)
		} else { // => is outside bounds check
			player.SetColor(rl.RayWhite) // How to check non-binary logic.. more options.. unlike drill room

			// - RESET FLAG as soon as player leaves bounds check
			// - This is useful when player spawns near the drill room.
			// - This avoids re-entering drill base Immediately.
			if !hasPlayerLeftDrillBase {
				hasPlayerLeftDrillBase = true // STEP [1]
			}
		}

		// Do not allow entry till cargo capacity is full.. This is temporary.. to quickly develop a simple gameloop.
		// Later we add transactions and resource conversions
		if __IS_TEMPORARY__ := false; __IS_TEMPORARY__ {
			if canSwitchToDrillRoom {
				slog.Warn("OVERIDING ENTRY TO DRILL ROOM. (TEMPORARY)")
				canSwitchToDrillRoom = hitScore >= xPlayer.MaxCargoCapacity
			}
		}

		// NOTES:
		//	 - (gameplay ) saveScore?
		//	 - (common   )   how much resource is required to drill to next level
		//	 - (drillroom) how will you handle modifying currentLevelID in gamesave/slot/1.json?
		//	 - (drillroom) what decides
		//	 - Are we drilling asteroids in space?
		//		- Draw a protection barrier over the scene (like a firmament)
		if canSwitchToDrillRoom { // Play entry sounds
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("footstep0%d.ogg", rl.GetRandomValue(0, 9))))) // 05
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "metalClick.ogg")))                                        // metalClick
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("creak%d.ogg", rl.GetRandomValue(1, 3)))))     // 3
			rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("doorOpen_%d.ogg", rl.GetRandomValue(1, 2))))) // 2

			// Save screen state
			finishScreen = 2                      // 1=>ending 2=>drillroom
			camera.Up = rl.NewVector3(0., 1., 0.) // Reset yaw/pitch/roll
			saveGameEntityData()                  // (player,camera,...) 705 bytes
			saveGameAdditionalData()              // (blocks,...)        82871 bytes
			saveGameLogicData()
		}
	}

	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) /* || rl.IsGestureDetected(rl.GesturePinchOut) */ {
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "rollover3.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "switch33.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_interface-sounds", "Audio", "confirmation_001.ogg")))

		// Save screen state
		finishScreen = 1                      // 1=>ending 2=>drillroom
		camera.Up = rl.NewVector3(0., 1., 0.) // Reset yaw/pitch/roll
		saveGameEntityData()                  // (player,camera,...) 705 bytes
		saveGameAdditionalData()              // (blocks,...)        82871 bytes
		saveGameLogicData()
	}

	// TODO: Move this in package player (if possible)
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyD) {
		const fps = 60.0
		const framesInterval = fps / 2.
		if framesCounter%int32(framesInterval) == 0 {
			if !rl.Vector3Equals(oldPlayer.Position, xPlayer.Position) &&
				rl.Vector3Distance(oldCam.Position, xPlayer.Position) > 1.0 &&
				(xPlayer.Collisions.X == 0 && xPlayer.Collisions.Z == 0) {
				rl.PlaySound(common.FXS.ImpactFootStepsConcrete[int(framesCounter)%len(common.FXS.ImpactFootStepsConcrete)])
			}
		}
	}

	// Increment gameplay frames counter
	framesCounter++
}

func Draw() {
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	BabyBlue := color.RGBA{R: 137, G: 207, B: 240, A: 255}
	rl.ClearBackground(rl.ColorBrightness(BabyBlue, -.85))

	xFloor.Draw()

	wall.DrawBatch(common.OpenWorldRoom, xFloor.Position, xFloor.Size, common.Vector3One)

	drawOuterDrillroom()

	for i := range xBlocks {
		xBlocks[i].Draw()

		if false { // DEBUG
			rl.DrawBoundingBox(xBlocks[i].GetBlockBoundingBox(), rl.Fade(rl.Gold, .3))
		}
	}
	xPlayer.Draw()
	DrawProjectiles()

	// ‥ Draw player to camera forward projected direction ray & area blob/blurb
	// TEMPORARY EXAMPLE TO SHOW RAY COLLISIONS
	rayTargetBoundingBox := common.GetBoundingBoxPositionSizeV(rl.NewVector3(0, 0, 0), rl.NewVector3(5, 5, 5)) // TEMPORARY
	gPlayerRayCollision = rl.GetRayCollisionBox(gPlayerRay, rayTargetBoundingBox)                              // Update
	if gPlayerRayCollision.Hit {
		startPos := rl.Vector3{X: gPlayerRay.Position.X, Y: gPlayerRay.Position.Y + xPlayer.Size.Y/4, Z: gPlayerRay.Position.Z}
		endPos := gPlayerRayCollision.Point
		rl.DrawLine3D(startPos, endPos, rl.SkyBlue)
	}
	if false {
		rl.DrawBoundingBox(rayTargetBoundingBox, rl.Blue)
	}

	for i := range MaxGuards {
		if !xGuardSOA.IsActive[i] {
			continue
		}

		startPos := rl.Vector3Add(xGuardSOA.Position[i], rl.NewVector3(0., -xGuardSOA.Size[i].Y/2, 0.)) // bottom
		endPos := rl.Vector3Add(xGuardSOA.Position[i], rl.NewVector3(0., xGuardSOA.Size[i].Y/2, 0.))    // top

		common.DrawXYZOrbitV(startPos, .1)              // bottom
		common.DrawXYZOrbitV(xGuardSOA.Position[i], .2) // center
		common.DrawXYZOrbitV(endPos, .1)                // top

		const radius = .25
		startPos.Y += radius
		endPos.Y -= radius

		model := common.ModelDungeonKit.OBJ.Barrel

		relativeModelPosition := xGuardSOA.Position[i]
		relativeModelPosition.Y -= xGuardSOA.Size[i].Y / 2

		const modelFloatInAirOffsetY = .0625
		relativeModelPosition.Y += modelFloatInAirOffsetY

		rl.DrawModelEx(model, relativeModelPosition, common.YAxis, 0, common.Vector3One, rl.Green)

		if false {
			rl.DrawBoundingBox(xGuardSOA.BoundingBox[i], rl.Fade(xGuardSOA.Color[i], .3))
		}
		if false {
			rings := int32(4)
			slices := int32(4)
			if framesCounter%4 == 0 {
				rings = int32(rl.Lerp(float32(rings), float32(rl.GetRandomValue(rings+1, 24)), .1))
				slices = int32(rl.Lerp(float32(slices), float32(rl.GetRandomValue(slices+1, 24)), .1))
			}
			rl.DrawSphereWires(xGuardSOA.Position[i], radius, rings, slices, rl.Red)
		}
	}

	for i := range MaxGuards {
		if !xGuardSOA.IsActive[i] {
			continue
		}
		for j := range MaxGuards {
			if !xGuardSOA.IsActive[j] {
				continue
			}
			if rl.CheckCollisionBoxes(xGuardSOA.BoundingBox[i], xGuardSOA.BoundingBox[j]) {
				// In a while loop
				counter := 0
				maxCounter := 100
				for counter < maxCounter && rl.CheckCollisionBoxes(
					common.GetBoundingBoxPositionSizeV(xGuardSOA.Position[i], xGuardSOA.Size[i]),
					common.GetBoundingBoxPositionSizeV(xGuardSOA.Position[j], xGuardSOA.Size[j]),
				) {
					if rl.GetRandomValue(1, 2) == 1 {
						xGuardSOA.Position[i].X = rl.Lerp(xGuardSOA.Position[i].X, xGuardSOA.Position[j].X, -0.05)
						xGuardSOA.Position[i].Z = rl.Lerp(xGuardSOA.Position[i].Z, xGuardSOA.Position[j].Z, -0.05)
					} else {
						xGuardSOA.Position[j].X = rl.Lerp(xGuardSOA.Position[j].X, xGuardSOA.Position[i].X, -0.05)
						xGuardSOA.Position[j].Z = rl.Lerp(xGuardSOA.Position[j].Z, xGuardSOA.Position[i].Z, -0.05)
					}

					counter++
				}
			}
		}
	}

	for i := range MaxGuards {
		if !xGuardSOA.IsActive[i] {
			continue
		}

		// Meander around
		if framesCounter%8 == 0 {
			xGuardSOA.Position[i].X = xGuardSOA.Position[i].X * rl.Lerp(1., 1.+xGuardSOA.Size[i].X*rl.GetFrameTime()*float32(rl.GetRandomValue(-1, 1)), .35)
			xGuardSOA.Position[i].Z = xGuardSOA.Position[i].Z * rl.Lerp(1., 1.+xGuardSOA.Size[i].Z*rl.GetFrameTime()*float32(rl.GetRandomValue(-1, 1)), .35)
		}

		xGuardSOA.BoundingBox[i] = common.GetBoundingBoxPositionSizeV(xGuardSOA.Position[i], xGuardSOA.Size[i])

		const lookahead = float32(12.)
		lookaheadSize := rl.Vector3Multiply(xGuardSOA.Size[i], rl.NewVector3(lookahead, 1, lookahead)) // Maintain y position
		lookaheadBounds := common.GetBoundingBoxPositionSizeV(xGuardSOA.Position[i], lookaheadSize)

		// Guard must dart towards player
		if rl.CheckCollisionBoxes(xPlayer.BoundingBox, lookaheadBounds) {
			alpha := rl.GetFrameTime() // Approach rate

			// Rush player once this is crossed
			distThreshold := float32(cmp.Or(
				mathutil.SqrtF(lookahead*common.InvPhi),
				lookahead/6.,
			))

			dist := rl.Vector3Distance(xGuardSOA.Position[i], xPlayer.Position)

			distCol := rl.Fade(rl.White, .1) // DEBUG - [1]
			distThresholdCol := rl.Fade(rl.Beige, .1)

			if dist <= distThreshold {
				f := dist / (lookahead / 2.)
				alpha += mathutil.SqrtF(f) / 8. // Jump scare

				distCol = rl.Fade(rl.Red, .1)
				distThresholdCol = rl.Fade(rl.Gold, .1)
			}

			if isDebugTweening := true; isDebugTweening {
				if false { // Above guard
					pos := xGuardSOA.Position[i]
					pos.Y += 1. + 2.*4.*alpha // 3D info
					rl.DrawSphereWires(pos, .5+2*alpha, 16, 16, rl.Red)
				}

				rl.DrawSphereWires(xGuardSOA.Position[i], dist, 8, 8, distCol) // DEBUG - [1]
				rl.DrawSphereWires(xGuardSOA.Position[i], distThreshold, 8, 8, distThresholdCol)
			}

			// Approach player
			// DONE: Avoid other guards from colliding with each other
			// TODO: Avoid guards from colliding with blocks/drillroom/etc..
			xGuardSOA.Position[i].X = rl.Lerp(xGuardSOA.Position[i].X, xPlayer.Position.X, alpha)
			xGuardSOA.Position[i].Z = rl.Lerp(xGuardSOA.Position[i].Z, xPlayer.Position.Z, alpha)
		}
		if rl.CheckCollisionBoxes(xPlayer.BoundingBox, xGuardSOA.BoundingBox[i]) {
			if true {
				guardOnPlayerDamage := rl.GetFrameTime() * 0.25
				xPlayer.Health = max(0.0, xPlayer.Health-guardOnPlayerDamage)
			} else {
				// See: DrawHeart references
				//   - 1.0 == 5 hearts
				//   - 0.0 == 0 hearts
				framesBeforeTakeDamage := int32(common.FPS * 2)
				if framesCounter%framesBeforeTakeDamage == 0 {
					guardOnPlayerDamage := float32(1.0 / 5.0)
					xPlayer.Health -= guardOnPlayerDamage
				}
			}
		}

		if false {
			rl.DrawBoundingBox(lookaheadBounds, rl.SkyBlue)
			rl.DrawBoundingBox(lookaheadBounds, rl.Blue)
		}
	}

	rl.EndMode3D()

	// =======================================================================
	// 2D World

	// Draw ray reticle on any 2D open space
	if gPlayerRayCollision.Hit {
		rl.DrawCircleV(rl.GetWorldToScreen(gPlayerRayCollision.Point, camera), 4, rl.Fade(rl.Gold, .3))
	} else {
		pos := rl.GetWorldToScreen(playerForwardAimEndPos, camera) // Draw a diamond
		rl.DrawRectanglePro(rl.NewRectangle(pos.X, pos.Y, 8, 8), rl.NewVector2(0, 0), 45, rl.Fade(rl.White, .1))
	}

	// Draw shooting/aiming ray reticle on block
	if closestBlockIndex := GetClosestMiningBlockIndexOnRayCollision(); closestBlockIndex > -1 && closestBlockIndex < len(xBlocks) {
		collision := rl.GetRayCollisionBox(gPlayerRay, xBlocks[closestBlockIndex].GetBlockBoundingBox())
		pos := rl.GetWorldToScreen(collision.Point, camera) // Draw a diamond
		rl.DrawRectanglePro(rl.NewRectangle(pos.X, pos.Y, 3, 3), rl.NewVector2(0, 0), 45, rl.Fade(rl.Green, .3))
	}

	// Draw depth meter
	{
		const gapX = 10

		var (
			totalLevels = len(common.SavedgameSlotData.AllLevelIDS)
			isShowText  bool
		)

		if rl.IsKeyDown(rl.KeyApostrophe) {
			isShowText = true
		}

		gapY := int32(mathutil.CeilF(float32(screenH) / float32(totalLevels))) // parts
		rl.DrawLine(screenW-gapX, gapY/2, screenW-gapX, screenH-gapY/2, rl.Gray)

		for i := range int32(totalLevels) {
			x := screenW - gapX
			y := gapY/2 + i*gapY
			rl.DrawLine(x, y, x-gapX/2, y, rl.Gray)

			radius := float32(6)

			if (i + 1) == levelID {
				col := rl.Orange

				if isShowText {
					col = rl.Red
				}
				rl.DrawCircle(x-int32(radius*2.5), y, radius, col)
			}
			if isShowText {
				rl.DrawText(fmt.Sprintf("%.2d", i+1), x-gapX*2-int32(radius*2), y-5, 10, rl.LightGray)
			}
		}
	}

	DrawHUD()

	if true { // Perf
		fontSize := float32(common.Font.Primary.BaseSize) * 3.0
		rl.DrawFPS(10, screenH-35)
		rl.DrawTextEx(common.Font.Primary, fmt.Sprintf("%.6f", rl.GetFrameTime()), rl.NewVector2(10, float32(screenH)-35-20*1), fontSize*2./3., 1, rl.Lime)
		rl.DrawTextEx(common.Font.Primary, fmt.Sprintf("%.3d", framesCounter), rl.NewVector2(10, float32(screenH)-35-20*2), fontSize*2./3., 1, rl.Lime)
	}
	if true { // Debug logic stats
		text := fmt.Sprintf("money: %.3d\nexperience: %.3d\n", money, experience)
		rl.DrawText(text, (screenW-10)-rl.MeasureText(text, 10), screenH-40, 10, rl.Green)
	}

}

type CurrencyType int32

const (
	Copper CurrencyType = iota
	Pearl               // or Iron
	Bronze
	Silver
	Ruby
	Gold
	Diamond
	Sapphire // Yellow/Blue
	// Platinum

	MaxCurrencyTypes
)

var currencyColorMap = map[CurrencyType]color.RGBA{
	Copper:   rl.Beige,
	Pearl:    rl.Gray,
	Bronze:   rl.Orange,
	Silver:   rl.Gray,
	Ruby:     rl.Maroon,
	Gold:     rl.Gold,
	Diamond:  rl.SkyBlue,
	Sapphire: rl.Yellow,
}

// currencyConversionMap maps Currency in Copper units.
var currencyConversionMap = map[CurrencyType]int32{
	Copper:   1,
	Pearl:    25,
	Bronze:   25,
	Silver:   30,
	Ruby:     35,
	Gold:     40,
	Diamond:  80,
	Sapphire: 80,
}

// DrawHUD draws the Heads-Up-Display on 2D screen.
func DrawHUD() {
	fontSize := float32(common.Font.Primary.BaseSize) * 3.0
	// Player stats: health / money / experience
	const (
		marginX    = 20
		marginY    = 20
		radius     = float32(20)
		marginLeft = float32(marginX * 2. / 3.)
	)
	var (
		cargoRatio = float32(xPlayer.CargoCapacity) / float32(xPlayer.MaxCargoCapacity)
	)
	// Set the transform matrix to where the HUD Stats are

	// Draw Health
	//   - 1.0 == 5 hearts
	//   - 0.0 == 0 hearts
	rl.PushMatrix()
	rl.Translatef(marginLeft, marginY, 0)
	radius0 := float32(15.)
	circlePos := rl.NewVector2(radius, 20*1)
	healthPartsCount := int32(rl.Clamp((xPlayer.Health*10.)/2., 0, 5))
	if healthPartsCount <= 1 { // FIXME: Use a better transition effect.. circle zoom out
		radius1 := radius0 * common.InvPhi
		f := rl.Clamp(xPlayer.Health, 0.00025, 1.0)
		healthCirclePos := rl.Vector2Subtract(circlePos, rl.NewVector2(0, radius0/4))
		outerCol := rl.Fade(rl.Maroon, 3*xPlayer.Health)
		innerCol := rl.Fade(rl.Red, 2*xPlayer.Health)
		if f <= 0.00025 {
			radius1 /= f
			f = max(0.1, mathutil.SqrtF(f)) // Black+Red splattered screen
			outerCol = rl.Fade(rl.ColorLerp(rl.Black, outerCol, f), max(0.1, 1000*f))
			innerCol = rl.Fade(rl.ColorLerp(rl.Black, innerCol, f), max(0.1, 1000*f))
			healthCirclePos = rl.Vector2Lerp(rl.NewVector2(float32(rl.GetScreenWidth())/2, float32(rl.GetScreenHeight())/2), healthCirclePos, f*f)
		}
		rl.DrawCircleV(healthCirclePos, radius1, outerCol)
		rl.DrawCircleSector(healthCirclePos, radius1, -90, -90+xPlayer.Health*360, cmp.Or(16, (healthPartsCount-1)*2), innerCol)
	}
	for i := range healthPartsCount {
		DrawHeart(rl.Vector2Add(circlePos, rl.NewVector2(2*radius0*float32(i), 0)), radius0)
	}
	// text := fmt.Sprintln(healthPartsCount)
	// textStrSize := rl.MeasureTextEx(common.Font.Primary, text, 10, 1.0)
	// playerScreenPosition := rl.GetWorldToScreen(rl.NewVector3(xPlayer.Position.X, xPlayer.Position.Y+0.5, xPlayer.Position.Z), camera)
	// if false {
	// 	rl.DrawCircleV(playerScreenPosition, radius0/2.0, rl.Fade(rl.Gray, 1.0+xPlayer.Health))
	// 	rl.DrawCircleSector(playerScreenPosition, radius0/2.0, -90, -90+xPlayer.Health*360, cmp.Or(16, (healthPartsCount-1)*2), rl.Fade(rl.LightGray, xPlayer.Health))
	// 	rl.DrawTextEx(common.Font.Primary, text, rl.NewVector2(playerScreenPosition.X-textStrSize.X/2.0, playerScreenPosition.Y-textStrSize.Y/3.0), 10, 1.0, rl.White)
	// }
	// if isEnableText := false; isEnableText {
	// 	text := fmt.Sprintf("%.0f", 100*xPlayer.Health)
	// 	stringSize := rl.MeasureTextEx(common.Font.Primary, text, fontSize*2./3., 1.)
	// 	textPos := rl.NewVector2(circlePos.X+radius0*2-stringSize.X/2, 20*1+radius0/2.-stringSize.Y)
	// 	rl.DrawTextEx(common.Font.Primary, text, textPos, fontSize*2./3., 1, rl.Red)
	// }
	rl.PopMatrix()

	rl.PushMatrix()
	rl.Translatef(marginLeft, marginY+20*3-radius/2, 0)
	// Draw Cargo Capacity - [1] circle sector meter
	circlePos = rl.NewVector2(radius, radius)
	if cargoRatio >= 1. {
		rl.DrawCircleGradient(int32(circlePos.X), int32(circlePos.Y), radius+3, rl.White, rl.Fade(rl.White, .1))
	}
	circleCutoutRec := rl.NewRectangle(radius/2., radius/2., radius, radius)
	rl.DrawRectangleRoundedLinesEx(circleCutoutRec, 1., 16, 0.5+radius/2., rl.DarkGray)
	rl.DrawCircleSector(circlePos, radius, -90, -90+360*cargoRatio, 16, cmp.Or(rl.White, rl.Gold))
	rl.DrawCircleV(circlePos, radius/2, rl.Fade(rl.Gold, cargoRatio))
	// Glass Half-Empty
	rl.DrawCircleV(circlePos, radius*max(.75, (1-cargoRatio)), rl.Fade(rl.Gold, 1.-cargoRatio))
	rl.DrawCircleV(circlePos, radius*max(.75, (1-cargoRatio)), rl.DarkGray)
	rl.DrawCircleV(circlePos, 8+8, rl.Gold)
	rl.DrawCircleV(circlePos, 8+4, rl.DarkGray)
	rl.DrawCircleV(circlePos, 8, rl.Gold)
	if false && cargoRatio >= 0.5 { // Glass Half-Full
		rl.DrawCircleV(circlePos, radius*cargoRatio, rl.Fade(rl.Gold, 1.0))
	}
	rl.PopMatrix()

	rl.PushMatrix()
	rl.Translatef(marginLeft+radius*2.25, marginY+20*3+radius, 0)
	// Draw Cargo Capacity - [2] meter text
	font := common.Font.Primary
	capText := fmt.Sprintf("%d", xPlayer.CargoCapacity)
	capStrLenX := rl.MeasureText(capText, int32(fontSize*2./3.))
	divideText := fmt.Sprintf("%s", strings.Repeat("-", len(capText)))
	divideStrLenX := rl.MeasureText(divideText, int32(fontSize)/2.)
	maxCapText := fmt.Sprintf(" %d", xPlayer.MaxCargoCapacity)
	rl.DrawTextEx(font, capText,
		rl.NewVector2(float32(capStrLenX)/2, -20-10/2), fontSize*2./3., 1, rl.White)
	rl.DrawTextEx(font, divideText,
		rl.NewVector2(float32(capStrLenX)/2+float32(divideStrLenX)/2, -(2*10)/1.5), fontSize/2., 0.0625, rl.Gray)
	rl.DrawTextEx(font, maxCapText,
		rl.NewVector2(0, -10/2), fontSize/2., 1, rl.Gray)
	rl.PopMatrix()

	rl.PushMatrix()
	rl.Translatef(marginLeft*.5, marginY+20*4+radius+20*.25, 0)
	type CurrencyItem struct {
		CurrencyType   CurrencyType
		OnHandCount    int
		CollectedCount int
	}

	// Hard-coded slice
	// FEATURES:
	//	- If player enters drillroom:
	//		- increment CollectedCount with OnHandCount
	//		- reset OnHandCount to 0
	currencyInventories := [MaxCurrencyTypes]CurrencyItem{
		Copper:   {CurrencyType: Copper, OnHandCount: 0, CollectedCount: 0},
		Pearl:    {CurrencyType: Pearl, OnHandCount: 0, CollectedCount: 0},
		Bronze:   {CurrencyType: Bronze, OnHandCount: 0, CollectedCount: 0},
		Silver:   {CurrencyType: Silver, OnHandCount: 0, CollectedCount: 0},
		Ruby:     {CurrencyType: Ruby, OnHandCount: 0, CollectedCount: 0},
		Gold:     {CurrencyType: Gold, OnHandCount: 0, CollectedCount: 0},
		Diamond:  {CurrencyType: Diamond, OnHandCount: 0, CollectedCount: 0},
		Sapphire: {CurrencyType: Sapphire, OnHandCount: 0, CollectedCount: 0},
	}

	currencyInventories[Copper].OnHandCount = int(hitScore)

	for i := range MaxCurrencyTypes {
		item := currencyInventories[i]

		const offset = (radius * 3)
		gapY := offset * float32(i)
		fontSize := (fontSize * 2. / 3.) - 2

		position := rl.NewVector2(circlePos.X, circlePos.Y+gapY)
		rl.DrawCircleV(position, min(fontSize/2, (radius*common.OneMinusInvPhi)), currencyColorMap[item.CurrencyType])

		{
			text := fmt.Sprintf("%d", item.CollectedCount)
			textStringSize := rl.MeasureTextEx(font, text, fontSize, 1)
			pos := rl.Vector2Add(position, rl.NewVector2(-textStringSize.X/2, textStringSize.Y*.8))
			rl.DrawTextEx(font, text, pos, fontSize, 1., rl.White)
		}

		if item.OnHandCount > 0 {
			text := fmt.Sprintf("+%d", item.OnHandCount)
			textStringSize := rl.MeasureTextEx(font, text, fontSize, 1)
			fontSize := fontSize - 2
			pos := rl.Vector2Add(position, rl.NewVector2(-textStringSize.X/2, textStringSize.Y*.8))
			pos = rl.Vector2Add(pos, rl.NewVector2(fontSize*1.5, -fontSize/1.5))
			rl.DrawTextEx(font, text, pos, fontSize, 1., rl.LightGray)
		}
	}
	rl.PopMatrix()
}

func Unload() {
	// TODO: Unload gameplay screen variables here!
	if rl.IsCursorHidden() {
		rl.EnableCursor() // without 3d ThirdPersonPerspective
	}
}

// Gameplay screen should finish?
// NOTE: This is called each frame in main game loop
func Finish() int {
	return finishScreen
}

// Set next block state
// Update score
// Play mining impacts with variations (s1:kick + s2:snare + s3:hollow-thock)
func handleBlockOnMining(b *block.Block) {
	if b.State == block.DirtBlockState { // First state
		soundName := "handleSmallLeather"
		if rl.GetRandomValue(0, 1) == 0 {
			soundName += "2"
		}
		v := rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", soundName+".ogg"))
		rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10))/40.0)
		rl.SetSoundVolume(v, 0.5)
		rl.PlaySound(v)
	}
	if b.State > block.DirtBlockState {
		v := rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("cloth%d.ogg", min(block.MaxBlockState-1, max(1, b.State+1)))))
		rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10)/(2*10)))
		rl.SetSoundVolume(v, 0.0625)
		rl.PlaySound(v)
	}
	if rl.GetRandomValue(0, 1) == 0 && b.State > block.DirtBlockState {
		v := rl.LoadSound(filepath.Join("res", "fx", "kenney_impact-sounds", "Audio", fmt.Sprintf("impactMining_00%d.ogg", min(block.MaxBlockState-1, b.State))))
		rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10)/(2*10)))
		rl.SetSoundVolume(v, 2.00)
		rl.PlaySound(v)
	}
	if b.State < block.MaxBlockState-1 /* framesCounter%int32(state+1) == 0 */ { // Higher states are small items.. So no need for bass
		s1 := common.FXS.ImpactsSoftMedium[rl.GetRandomValue(int32(b.State), int32(len(common.FXS.ImpactsSoftMedium)-1))]
		s2 := common.FXS.ImpactsGenericLight[rl.GetRandomValue(int32(b.State), int32(len(common.FXS.ImpactsGenericLight)-1))]
		s3 := common.FXS.ImpactsSoftHeavy[rl.GetRandomValue(int32(b.State), int32(len(common.FXS.ImpactsSoftHeavy)-1))]
		rl.SetSoundVolume(s1, float32(rl.GetRandomValue(7, 10))/10.)
		rl.SetSoundVolume(s2, float32(rl.GetRandomValue(4, 8))/10.)
		rl.SetSoundVolume(s3, float32(rl.GetRandomValue(1, 4))/10.)
		rl.PlaySound(s1)
		rl.PlaySound(s2)
		rl.PlaySound(s3)
	}

	// Update stats
	hitCount++

	{
		const finalState = (block.MaxBlockState - 1)
		const cargoCapacityUnitPerIncrement = 2

		canIncrementScore := b.State == finalState-1

		if canIncrementScore {
			hitScore += cargoCapacityUnitPerIncrement
			xPlayer.CargoCapacity = min(xPlayer.MaxCargoCapacity, xPlayer.CargoCapacity+cargoCapacityUnitPerIncrement)
		}
		if canIncrementScore { // FIXME: Record.. hitCount and hitScore to save game.. and load and update directly
			if hitCount/hitScore != int32(finalState) {
				msg := fmt.Sprintf("expect for %d hits, score to incrementby 1. (except if counter started from an already semi-mined block)", finalState)
				if isEnablePerfectionist := false; isEnablePerfectionist {
					panic(msg)
				}
				slog.Warn(msg)
			}
		}
	}

	// Increment state on successful mining action
	b.NextState()
}

func DrawProjectiles() {
	for i := range projectile.MaxProjectiles {
		if !xProjectileSOA.IsActive[i] {
			continue
		}

		col := rl.White
		col = rl.Fade(col, .2)

		const maxTrailLength = 3. // Projectile trail
		const maxTrailThick = .08 // Radius
		const radius0 = maxTrailThick * common.InvPhi

		// rl.DrawSphere(projectiles.Position[i], radius0, col) // Projectile Head
		rl.DrawSphereWires(xProjectileSOA.Position[i], radius0, 16, 16, rl.Fade(col, .1))

		timeFactor := (xProjectileSOA.TimeLeft[i] / projectile.MaxTimeLeft)

		angle := xProjectileSOA.Rotation[i] * rl.Deg2rad
		dist := rl.Vector3Distance(xPlayer.Position, xProjectileSOA.Position[i])
		radius1 := float32(maxTrailThick * timeFactor)
		trailLength := float32(maxTrailLength)

		// Avoid passing projectile trail through the player body
		// itself when animation just started
		if dist < maxTrailLength {
			trailLength = dist
		}

		currPos := xProjectileSOA.Position[i]
		prevPos := rl.Vector3{
			X: xProjectileSOA.Position[i].X - mathutil.CosF(angle)*trailLength,
			Y: 0,
			Z: xProjectileSOA.Position[i].Z - mathutil.SinF(angle)*trailLength}

		rl.DrawCylinderEx(prevPos, currPos, (radius1/4)/timeFactor, radius1, 16, col)
	}
}

// Update and Set ray each frame
//
//	intersection using the slab method
//	https://tavianator.com/2011/ray_box.html#:~:text=The%20fastest%20method%20for%20performing,remains%2C%20it%20intersected%20the%20box.
//
//	bool RayIntersectRect(Rectangle rect, Vector2 origin, Vector2 direction, Vector2* point) {}
//	bool CheckCollisionRay2dCircle(Ray2d ray, Vector2 center, float radius, Vector2* intersection) {}
//
// See https://github.com/raylib-extras/examples-c/blob/6ed2ac244d961239b1695d0b6a729f6fd7bc209b/ray2d_rect_intersection/ray2d_rect_intersection.c
func UpdatePlayerRay() {
	cameraForward := rl.GetCameraForward(&camera)
	playerForwardEstimateMagnitude := rl.Vector3{X: 5., Y: .125 / 2., Z: 5.} // HACK: Projection
	playerReticlePosition := rl.Vector3Multiply(cameraForward, playerForwardEstimateMagnitude)
	gPlayerRay = rl.NewRay(rl.Vector3{X: xPlayer.Position.X, Y: xPlayer.Position.Y /* + xPlayer.Size.Y/4 */, Z: xPlayer.Position.Z}, playerReticlePosition)
	playerForwardAimEndPos = rl.Vector3Add(xPlayer.Position, playerReticlePosition)
}

func GetClosestMiningBlockIndexOnRayCollision() int {
	var (
		index           = -1
		minimumDistance = float32(math.MaxFloat32)
	)
	for i := range xBlocks {
		if !xBlocks[i].IsActive || xBlocks[i].State >= (block.MaxBlockState-1) { // for max==4 -> where last is 3 , only allow 0,1,2
			continue
		}
		if rc := rl.GetRayCollisionBox(gPlayerRay, xBlocks[i].GetBlockBoundingBox()); rc.Hit {
			temp := minimumDistance
			minimumDistance = min(rc.Distance, minimumDistance)

			if minimumDistance < temp {
				index = i
			}
		}
	}
	return index
}

func drawOuterDrillroom() {
	const maxDrillWallIndex = 2
	wallScale := rl.NewVector3(1., 1., 1.)
	for i := float32(-maxDrillWallIndex + 1); i < maxDrillWallIndex; i++ {
		var model rl.Model
		var y float32
		model = common.ModelDungeonKit.OBJ.Column
		y = 0.
		rl.DrawModelEx(model, rl.NewVector3(i, y, maxDrillWallIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
		rl.DrawModelEx(model, rl.NewVector3(i, y, -maxDrillWallIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
		rl.DrawModelEx(model, rl.NewVector3(maxDrillWallIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
		rl.DrawModelEx(model, rl.NewVector3(-maxDrillWallIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
		model = common.ModelDungeonKit.OBJ.Wall
		y = 1. + .125*.5
		rl.DrawModelEx(model, rl.NewVector3(i, y, maxDrillWallIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
		rl.DrawModelEx(model, rl.NewVector3(i, y, -maxDrillWallIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
		rl.DrawModelEx(model, rl.NewVector3(maxDrillWallIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
		rl.DrawModelEx(model, rl.NewVector3(-maxDrillWallIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
		model = common.ModelDungeonKit.OBJ.Column
		y = 2. + .125*.5
		rl.DrawModelEx(model, rl.NewVector3(i, y, maxDrillWallIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
		rl.DrawModelEx(model, rl.NewVector3(i, y, -maxDrillWallIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
		rl.DrawModelEx(model, rl.NewVector3(maxDrillWallIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
		rl.DrawModelEx(model, rl.NewVector3(-maxDrillWallIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
	}
}

type GameEntityData struct {
	LevelID int32 `json:"levelID"`

	Camera                 rl.Camera3D   `json:"camera"`
	FinishScreen           int           `json:"finishScreen"`
	FramesCounter          int32         `json:"framesCounter"`
	XFloor                 floor.Floor   `json:"xFloor"`
	XPlayer                player.Player `json:"xPlayer"`
	HasPlayerLeftDrillBase bool          `json:"hasPlayerLeftDrillBase"`
}

type GameAdditionalData struct {
	LevelID int32

	Blocks []block.Block `json:"blocks"`
}

type GameLogicData struct {
	LevelID int32

	Money      int32 `json:"money"`
	Experience int32 `json:"experience"`
	HitScore   int32 `json:"hitScore"`
	HitCount   int32 `json:"hitCount"`
}

const (
	entityGameDataVersionSuffix     = "entity"
	additionalGameDataVersionSuffix = "additional"
	logicGameDataVersionSuffix      = "logic"
)

func saveGameLogicData() {
	const suffix = logicGameDataVersionSuffix
	input := GameLogicData{
		LevelID: levelID,

		Money:      1000,
		Experience: 0,
		HitScore:   hitScore,
		HitCount:   hitCount,
	}
	var b []byte
	bb := bytes.NewBuffer(b)
	{
		enc := json.NewEncoder(bb)
		if err := enc.Encode(input); err != nil {
			panic(fmt.Errorf("encode game %s level data: %w", suffix, err))
		}
	}
	dataJSON := storage.GameStorageLevelJSON{
		Version: "0.0.0" + "-" + suffix,
		LevelID: levelID,
		Data:    bb.Bytes(),
	}
	storage.SaveStorageLevelEx(dataJSON, suffix)
}
func saveGameEntityData() {
	const suffix = entityGameDataVersionSuffix
	input := GameEntityData{
		LevelID: levelID,

		Camera:                 camera,
		FinishScreen:           finishScreen,
		FramesCounter:          framesCounter,
		XFloor:                 xFloor,
		XPlayer:                xPlayer,
		HasPlayerLeftDrillBase: hasPlayerLeftDrillBase,
	}
	var b []byte
	bb := bytes.NewBuffer(b)
	{
		enc := json.NewEncoder(bb)
		if err := enc.Encode(input); err != nil {
			panic(fmt.Errorf("encode game %s level data: %w", suffix, err))
		}
	}
	dataJSON := storage.GameStorageLevelJSON{
		Version: "0.0.0" + "-" + suffix,
		LevelID: levelID,
		Data:    bb.Bytes(),
	}
	storage.SaveStorageLevelEx(dataJSON, suffix)
}
func saveGameAdditionalData() {
	const suffix = additionalGameDataVersionSuffix
	input := GameAdditionalData{
		Blocks: xBlocks,
	}
	var b []byte
	bb := bytes.NewBuffer(b)
	{
		enc := json.NewEncoder(bb)
		if err := enc.Encode(input); err != nil {
			panic(fmt.Errorf("encode game %s level data: %w", suffix, err))
		}
	}
	data := storage.GameStorageLevelJSON{
		Version: "0.0.0" + "-" + suffix,
		LevelID: levelID,
		Data:    bb.Bytes(),
	}
	storage.SaveStorageLevelEx(data, suffix)
}

func loadGameLogicData() (*GameLogicData, error) {
	const suffix = logicGameDataVersionSuffix

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	saveDir := filepath.Join(cwd, "storage")
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(levelID))+"_"+suffix+".json")

	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", name, err)
	}

	dest := &storage.GameStorageLevelJSON{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&dest); err != nil {
		return nil, fmt.Errorf("decode level: %w", err)
	}

	switch version := dest.Version; version {
	case "0.0.0" + "-" + suffix:
		var v *GameLogicData
		if err := json.Unmarshal(dest.Data, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("invalid game %s data version %q", suffix, version)
	}
}
func loadGameEntityData() (*GameEntityData, error) {
	const suffix = entityGameDataVersionSuffix

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	saveDir := filepath.Join(cwd, "storage")
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(levelID))+"_"+suffix+".json")

	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", name, err)
	}

	dest := &storage.GameStorageLevelJSON{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&dest); err != nil {
		return nil, fmt.Errorf("decode level: %w", err)
	}
	// return dest,nil // => Upto here.. same as storage.LoadStorageLevel

	switch version := dest.Version; version {
	case "0.0.0" + "-" + suffix:
		var v *GameEntityData
		err := json.Unmarshal(dest.Data, &v)
		return v, err
	default:
		return nil, fmt.Errorf("invalid game %s data version %q", suffix, version)
	}
}
func loadAdditionalGameData() (*GameAdditionalData, error) {
	const suffix = additionalGameDataVersionSuffix

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	saveDir := filepath.Join(cwd, "storage")
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(levelID))+"_"+suffix+".json")

	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", name, err)
	}

	dest := &storage.GameStorageLevelJSON{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&dest); err != nil {
		return nil, fmt.Errorf("decode level: %w", err)
	}

	switch version := dest.Version; version {
	case "0.0.0" + "-" + suffix:
		var v *GameAdditionalData
		if err := json.Unmarshal(dest.Data, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("invalid game %s data version %q", suffix, version)
	}
}

// LOGIC

// Conversion rate
func logicGameCurrencyConversionPrototype() {
	// Copied from block.go
	// ========================================================================
	type BlockResourceType uint8

	const (
		DefaultBlockResource BlockResourceType = iota
		CopperBlockResource
		SilverBlockResource
		GoldBlockResource
	)
	// ------------------------------------------------------------------------

	// - copper to iron => 25:1
	// - copper to bronze => 25:1
	// - copper to silver => 30:1
	// - copper to ruby => 35:1
	// - copper to gold => 40:1
	// - copper to diamond => 80:1
	// - copper to sapphire => 80:1

	// https://en.wikipedia.org/wiki/Hierarchy_of_precious_substances
	type Currency int32

	const (
		Copper   Currency = 1
		Pearl    Currency = 25 // or Iron
		Bronze   Currency = 25
		Silver   Currency = 30
		Ruby     Currency = 35
		Gold     Currency = 40
		Diamond  Currency = 80
		Sapphire Currency = 80 // Yellow/Blue
		// Platinum
	)

	// Traditional manifestations
	// Jubilees have a hierarchy of years:
	//
	// Years	Precious Material	Example
	// 25	Silver	Silver Jubilee
	// 40	Ruby	Ruby Jubilee
	// 50	Gold	Golden Jubilee
	// 60	Diamond	Diamond Jubilee
	// 65	Sapphire	Sapphire Jubilee
	// 70	Platinum	Platinum Jubilee

}

func DrawHeart(position rl.Vector2, radius float32) {
	if isDrawBackdropCircle := false; isDrawBackdropCircle {
		rl.DrawCircleV(position, radius, rl.Fade(rl.Red, .1))
	}
	offsetX := radius / math.Pi
	l := rl.NewVector2(position.X-offsetX, position.Y-radius/2.)
	r := rl.NewVector2(position.X+offsetX, position.Y-radius/2.)
	ll := rl.NewVector2(l.X-offsetX*math.Pi/2, l.Y+radius/(math.Pi*2))
	rr := rl.NewVector2(r.X+offsetX*math.Pi/2, r.Y+radius/(math.Pi*2))
	bot := rl.NewVector2(position.X, position.Y+radius/(math.Pi/2))
	if isShowLines := false; isShowLines {
		rl.DrawTriangleLines(bot, rr, ll, rl.Red)
		rl.DrawCircleLinesV(l, radius/2., rl.Red)
		rl.DrawCircleLinesV(r, radius/2., rl.Red)
	} else {
		rl.DrawTriangle(bot, rr, ll, rl.Red)
		rl.DrawCircleV(l, radius/2., rl.Red)
		rl.DrawCircleV(r, radius/2., rl.Red)
	}
}
