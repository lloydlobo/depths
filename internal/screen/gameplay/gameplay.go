package gameplay

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
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/block"
	"example/depths/internal/common"
	"example/depths/internal/floor"
	"example/depths/internal/player"
	"example/depths/internal/storage"
	"example/depths/internal/util/mathutil"
	"example/depths/internal/wall"
)

var (
	// Core data

	finishScreen  int
	framesCounter int32

	levelID                int32
	camera                 rl.Camera3D
	xFloor                 floor.Floor
	xPlayer                player.Player
	hasPlayerLeftDrillBase bool

	// Additional data

	blocks []block.Block
)

var (
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

	// Other

	checkedTexture rl.Texture2D
	checkedModel   rl.Model
)

func Init() {
	framesCounter = 0
	finishScreen = 0

	levelID = int32(common.SavedgameSlotData.CurrentLevelID)
	if levelID == 0 {
		panic("unexpected levelID")
	}

	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 10., 10.),
		Target:     rl.NewVector3(0., .5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       15. * float32(cmp.Or(4., 3., 2.)),
		Projection: rl.CameraPerspective,
	} // See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h

	// INIT WOULD REQUIRE A LEVEL ID? OR CREATE MULTIPLE gameplay000.go files
	//
	// InitWorld loads resources each time it is called.
	//
	// Prefers loading data from saved game files if any, else generates new data.
	//
	// If new game flag is turned on, generates new game.
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

		blocks = []block.Block{} // Clear
		block.InitBlocks(&blocks, block.GenerateRandomBlockPositions(xFloor))
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

	// Core resources
	floor.SetupFloorModel()
	wall.SetupWallModel(common.OpenWorldRoom)
	player.SetupPlayerModel() // FIXME: in this func, use package common for models
	player.ToggleEquippedModels([player.MaxBoneSockets]bool{false, true, true})

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
			blocks = make([]block.Block, len(additionalGameData.Blocks))
			copiedBlockCount := copy(blocks, additionalGameData.Blocks)
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
			fmt.Printf("data: %v\n", data)
			saveGameLogicData() // Save ASAP
		} else { // ERR
			slog.Warn(err.Error())
			loadNewLogicData()
		}
	} else {
		loadNewLogicData()
	}

	{
		checkedImg := rl.GenImageChecked(100, 100, 1, 1, rl.ColorBrightness(rl.Black, .24), rl.ColorBrightness(rl.Black, .20))
		checkedTexture = rl.LoadTextureFromImage(checkedImg)
		rl.UnloadImage(checkedImg)
		checkedModel = rl.LoadModelFromMesh(rl.GenMeshPlane(100, 100, 10, 10))
		checkedModel.Materials.Maps.Texture = checkedTexture
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
		rl.PauseMusicStream(currentMusic)
	}

	rl.DisableCursor() // for ThirdPersonPerspective
}

func Update() {
	rl.UpdateMusicStream(currentMusic)

	if rl.IsKeyDown(rl.KeyF) {
		log.Println("[F] Picked up item")
	}

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

	for i := range blocks {
		// Skip final broken/mined block residue
		if blocks[i].State == block.MaxBlockState-1 {
			continue
		}
		if rl.CheckCollisionBoxes(
			common.GetBoundingBoxFromPositionSizeV(blocks[i].Pos, blocks[i].Size),
			xPlayer.BoundingBox,
		) {
			// FIND OUT WHERE PLAYER TOUCHED THE BOX
			// HACK
			//		HACK
			//			HACK
			xPlayer.Collisions.Z = 1
			dx := oldPlayer.Position.X - xPlayer.Position.X
			if dx < 0.0 { // new <- old
				xPlayer.Collisions.X = 1
			} else if dx > 0.0 { // new -> old
				xPlayer.Collisions.X = -1
			} else {
				if xPlayer.Collisions.X != 0 { // Placeholder (do not overwrite previous)
					xPlayer.Collisions.X = 0
				}
			}
			dz := oldPlayer.Position.Z - xPlayer.Position.Z
			if dz < 0.0 { // new <- old
				xPlayer.Collisions.Z = 1
			} else if dz > 0.0 { // new -> old
				xPlayer.Collisions.Z = -1
			} else {
				if xPlayer.Collisions.Z != 0 { // Placeholder (do not overwrite previous)
					xPlayer.Collisions.Z = 0
				}
			}
			//			HACK
			//		HACK
			// HACK

			player.RevertPlayerAndCameraPositions(&xPlayer, oldPlayer, &camera, oldCam)

			// Trigger once while mining
			if (rl.IsKeyDown(rl.KeySpace) && framesCounter%16 == 0) ||
				(rl.IsMouseButtonDown(rl.MouseLeftButton) && framesCounter%16 == 0) {
				// Play player weapon sounds
				if true {
					v := rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "drawKnife3.ogg"))
					rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10)/(2*10)))
					rl.SetSoundVolume(v, 0.1)
					rl.PlaySound(v)
				}
				// Play mining impacts with variations (s1:kick + s2:snare + s3:hollow-thock)
				state := blocks[i].State
				if state == block.DirtBlockState { // First state
					soundName := "handleSmallLeather"
					if rl.GetRandomValue(0, 1) == 0 {
						soundName += "2"
					}
					v := rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", soundName+".ogg"))
					rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10))/40.0)
					rl.SetSoundVolume(v, 0.5)
					rl.PlaySound(v)
				}
				if true {
					v := rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", fmt.Sprintf("cloth%d.ogg", min(block.MaxBlockState-1, max(1, state+1)))))
					rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10)/(2*10)))
					rl.SetSoundVolume(v, 0.0625)
					rl.PlaySound(v)
				}
				if framesCounter%int32(state+1) == 0 { // Higher states are small items.. So no need for bass
					s1 := common.FXS.ImpactsSoftMedium[rl.GetRandomValue(int32(state), int32(len(common.FXS.ImpactsSoftMedium)-1))]
					s2 := common.FXS.ImpactsGenericLight[rl.GetRandomValue(int32(state), int32(len(common.FXS.ImpactsGenericLight)-1))]
					s3 := common.FXS.ImpactsSoftHeavy[rl.GetRandomValue(int32(state), int32(len(common.FXS.ImpactsSoftHeavy)-1))]
					rl.SetSoundVolume(s1, float32(rl.GetRandomValue(7, 10))/10.)
					rl.SetSoundVolume(s2, float32(rl.GetRandomValue(4, 8))/10.)
					rl.SetSoundVolume(s3, float32(rl.GetRandomValue(1, 4))/10.)
					rl.PlaySound(s1)
					rl.PlaySound(s2)
					rl.PlaySound(s3)
				}
				if rl.GetRandomValue(0, 1) == 0 && state > block.DirtBlockState {
					v := rl.LoadSound(filepath.Join("res", "fx", "kenney_impact-sounds", "Audio", fmt.Sprintf("impactMining_00%d.ogg", min(block.MaxBlockState-1, state))))
					rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10)/(2*10)))
					rl.SetSoundVolume(v, 2.00)
					rl.PlaySound(v)
				}

				// Update stats
				hitCount++

				const finalState = (block.MaxBlockState - 1)
				canIncrementScore := state == finalState-1

				if canIncrementScore {
					hitScore++
					xPlayer.CargoCapacity = min(xPlayer.MaxCargoCapacity, xPlayer.CargoCapacity+1)

					// FIXME: Record.. hitCount and hitScore to save game.. and load and update directly
					if hitCount/hitScore != int32(finalState) {
						msg := fmt.Sprintf("expect for %d hits, score to incrementby 1. (except if counter started from an already semi-mined block)", finalState)
						if isEnablePerfectionist := false; isEnablePerfectionist {
							panic(msg)
						} else {
							slog.Warn(msg)
						}
					}
				}
				// Increment state on successful mining action
				blocks[i].NextState()
			}
		}
	}

	// Update player─block collision+breaking/mining
	{
		origin := xFloor.Position
		bb1 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(3, 2, 3)) // player is inside
		bb2 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(5, 2, 5)) // player is entering
		bb3 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(7, 2, 7)) // bot barrier
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
		// - (gameplay ) saveScore?
		// - (common   )   how much resource is required to drill to next level
		// - (drillroom) how will you handle modifying currentLevelID in gamesave/slot/1.json?
		// - (drillroom) what decides
		// - Are we drilling asteroids in space?
		//	- Draw a protection barrier over the scene (like a firmament)
		if canSwitchToDrillRoom {
			// Play entry sounds
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
	if rl.IsKeyDown(rl.KeyF10) || rl.IsGestureDetected(rl.GesturePinchOut) {
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
		const framesInterval = fps / 2.5
		if framesCounter%int32(framesInterval) == 0 {
			if !rl.Vector3Equals(oldPlayer.Position, xPlayer.Position) &&
				rl.Vector3Distance(oldCam.Position, xPlayer.Position) > 1.0 &&
				(xPlayer.Collisions.X == 0 && xPlayer.Collisions.Z == 0) {
				rl.PlaySound(common.FXS.FootStepsConcrete[int(framesCounter)%len(common.FXS.FootStepsConcrete)])
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

	rl.ClearBackground(rl.Black)

	if false { // ‥ Draw pseudo-infinite(ish) floor backdrop
		rl.DrawModel(checkedModel, rl.NewVector3(0., -.05, 0.), 1., rl.RayWhite)
	}

	xFloor.Draw()

	wall.DrawBatch(common.OpenWorldRoom, xFloor.Position, xFloor.Size, common.Vector3One)

	{
		type BlockResourceType uint8
		const (
			DefaultBlockResource BlockResourceType = iota
			CopperBlockResource
			SilverBlockResource
			GoldBlockResource
		)
		drawSpecialBlock := func(block block.Block, typ BlockResourceType) {
			var col color.RGBA
			switch typ {
			case CopperBlockResource:
				col = rl.DarkPurple
			case SilverBlockResource:
				col = rl.Maroon
			case GoldBlockResource:
				col = rl.Orange
			default:
				panic(fmt.Sprintf("unexpected gameplay.BlockResourceType: %#v", typ))
			}
			if true {
				rl.DrawSphereWires(block.Pos, -.0625+common.InvPhi*rl.Vector3Length(block.Size)/math.Pi, 8, 8, rl.Fade(col, .35))
			} else {
				rl.DrawCapsuleWires(
					rl.Vector3Add(block.Pos, rl.NewVector3(.0625, block.Size.Y/2, .0625)),
					rl.Vector3Subtract(block.Pos, rl.NewVector3(.0625, block.Size.Y/4, .0625)),
					-.0625+common.InvPhi*rl.Vector3Length(block.Size)/math.Pi,
					2, 2, rl.Fade(col, .1))
			}
		}

		rl.BeginBlendMode(rl.BlendAlphaPremultiply) // For special block

		for i := range blocks {
			blocks[i].Draw()

			if false {
				if i%11 == 0 && blocks[i].IsActive && blocks[i].State != block.MaxBlockState-1 {
					drawSpecialBlock(blocks[i], CopperBlockResource)
				} else if i%13 == 0 && blocks[i].IsActive && blocks[i].State != block.MaxBlockState-1 {
					drawSpecialBlock(blocks[i], SilverBlockResource)
				} else if i%23 == 0 && blocks[i].IsActive && blocks[i].State != block.MaxBlockState-1 {
					drawSpecialBlock(blocks[i], GoldBlockResource)
				}
			}
		}

		rl.EndBlendMode()
	}

	xPlayer.Draw()
	if false { // ‥ Draw player to camera forward projected direction ray & area blob/blurb
		const maxRays = float32(8. * 2)
		const rayGapFactor = 16 * maxRays
		rayCol := rl.Fade(rl.Yellow, .3)
		startPos := xPlayer.Position // NOTE: startPos.Y and endPos.Y may fluctuate
		endPos := rl.Vector3Add(
			xPlayer.Position,
			rl.Vector3Multiply(
				rl.GetCameraForward(&camera),
				rl.NewVector3(9., .125/2., 9.),
			),
		)
		rl.DrawLine3D(startPos, endPos, rayCol) // Draw middle ray
		rayCol = rl.Fade(rayCol, .1)
		for i := -maxRays; i < maxRays; i++ { // Draw spread-out rays
			rl.DrawLine3D(startPos, rl.Vector3Add(endPos, rl.NewVector3(i/rayGapFactor, .0, .0)), rayCol)
			rl.DrawLine3D(startPos, rl.Vector3Add(endPos, rl.NewVector3(.0, .0, i/rayGapFactor)), rayCol)
		}
		rl.DrawCapsule(startPos, endPos, 2, 7, 7, rl.Fade(rl.Gray, .125/2)) // Draw forward movement lookahead area
	}

	{ // ‥ Draw drillroom entry
		const maxIndex = 2
		wallScale := rl.NewVector3(1., 1., 1.)
		for i := float32(-maxIndex + 1); i < maxIndex; i++ {
			var model rl.Model
			var y float32
			model = common.ModelDungeonKit.OBJ.Column
			y = 0.
			rl.DrawModelEx(model, rl.NewVector3(i, y, maxIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
			rl.DrawModelEx(model, rl.NewVector3(i, y, -maxIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
			rl.DrawModelEx(model, rl.NewVector3(maxIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
			rl.DrawModelEx(model, rl.NewVector3(-maxIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
			model = common.ModelDungeonKit.OBJ.Wall
			y = 1. + .125*.5
			rl.DrawModelEx(model, rl.NewVector3(i, y, maxIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
			rl.DrawModelEx(model, rl.NewVector3(i, y, -maxIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
			rl.DrawModelEx(model, rl.NewVector3(maxIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
			rl.DrawModelEx(model, rl.NewVector3(-maxIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
			model = common.ModelDungeonKit.OBJ.Column
			y = 2. + .125*.5
			rl.DrawModelEx(model, rl.NewVector3(i, y, maxIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
			rl.DrawModelEx(model, rl.NewVector3(i, y, -maxIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
			rl.DrawModelEx(model, rl.NewVector3(maxIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
			rl.DrawModelEx(model, rl.NewVector3(-maxIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
		}

		if false { // ‥ DEBUG: Draw drill door gate entry logic before changing scene to drill base
			origin := common.Vector3Zero
			origin = xFloor.Position
			bb1 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(3, 2, 3)) // player is inside
			bb2 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(5, 2, 5)) // player is entering
			bb3 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(7, 2, 7)) // bot barrier
			rl.DrawBoundingBox(bb1, rl.Red)
			rl.DrawBoundingBox(bb2, rl.Green)
			rl.DrawBoundingBox(bb3, rl.Blue)
		}
	}

	if false { // ‥ Draw banners at floor corners
		floorBBMin := xFloor.BoundingBox.Min
		floorBBMax := xFloor.BoundingBox.Max
		rl.DrawModelEx(common.ModelDungeonKit.OBJ.Banner, rl.NewVector3(floorBBMin.X+1, 0, floorBBMin.Z+1), common.YAxis, 45, common.Vector3One, rl.White)  // leftback
		rl.DrawModelEx(common.ModelDungeonKit.OBJ.Banner, rl.NewVector3(floorBBMax.X-1, 0, floorBBMin.Z+1), common.YAxis, -45, common.Vector3One, rl.White) // rightback
		rl.DrawModelEx(common.ModelDungeonKit.OBJ.Banner, rl.NewVector3(floorBBMax.X, 0, floorBBMax.Z), common.YAxis, 45, common.Vector3One, rl.White)      // rightfront
		rl.DrawModelEx(common.ModelDungeonKit.OBJ.Banner, rl.NewVector3(floorBBMin.X, 0, floorBBMax.Z), common.YAxis, -45, common.Vector3One, rl.White)     // leftfront
	}

	rl.EndMode3D()

	// 2D World

	// Draw depth meter
	{
		const (
			totalLevels = 8
			gapX        = 10
		)
		var (
			isShowText bool
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
			radius := float32(4)
			if (i + 1) == levelID {
				col := rl.Gray
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

	fontSize := float32(common.Font.Primary.BaseSize) * 3.0
	text := "[F] PICK UP"
	rl.DrawText(text, screenW/2-rl.MeasureText(text, 20)/2, screenH-20*2, 20, rl.White)

	// Player
	{
		rl.DrawTextEx(common.Font.Primary, fmt.Sprintf("%.0f", 100*xPlayer.Health), rl.NewVector2(10, 10+20*1), fontSize*2./3., 1, rl.Red)
		const radius = 20
		const marginLeft = 8
		cargoRatio := (float32(xPlayer.CargoCapacity) / float32(xPlayer.MaxCargoCapacity))
		circlePos := rl.NewVector2(10+radius, 10+20*3+radius)
		if cargoRatio >= 1. {
			rl.DrawCircleGradient(int32(circlePos.X), int32(circlePos.Y), radius+3, rl.White, rl.Fade(rl.White, .1))
		}
		circleCutoutRec := rl.NewRectangle(10+radius/2., 10+20*3+radius/2., radius, radius)
		rl.DrawRectangleRoundedLinesEx(circleCutoutRec, 1., 16, 0.5+radius/2., rl.DarkGray)
		rl.DrawCircleSector(circlePos, radius, -90, -90+360*cargoRatio, 16, rl.Gold)
		rl.DrawCircleV(circlePos, radius/2, rl.Fade(rl.Gold, cargoRatio))
		// Glass Half-Empty
		rl.DrawCircleV(circlePos, radius*max(.5, (1-cargoRatio)), rl.Fade(rl.Gold, 1.-cargoRatio))
		rl.DrawCircleV(circlePos, radius*max(.5, (1-cargoRatio)), rl.DarkGray)
		// Glass Half-Full
		if cargoRatio >= 0.5 {
			rl.DrawCircleV(circlePos, radius*cargoRatio, rl.Fade(rl.Gold, 1.0))
		}
		rl.DrawTextEx(common.Font.Primary, fmt.Sprintf("%d", xPlayer.CargoCapacity), rl.NewVector2(10+radius*2+marginLeft, 10+radius/2+20*3-1), fontSize*2./3., 1, rl.Gold)
		rl.DrawTextEx(common.Font.Primary, fmt.Sprintf("/%d", xPlayer.MaxCargoCapacity), rl.NewVector2(10+radius*2+marginLeft, 10+radius/2+20*4-1), fontSize*2./4., 1, rl.Gray)
	}

	// Perf
	rl.DrawFPS(10, screenH-35)
	rl.DrawTextEx(common.Font.Primary, fmt.Sprintf("%.6f", rl.GetFrameTime()), rl.NewVector2(10, float32(screenH)-35-20*1), fontSize*2./3., 1, rl.Lime)
	rl.DrawTextEx(common.Font.Primary, fmt.Sprintf("%.3d", framesCounter), rl.NewVector2(10, float32(screenH)-35-20*2), fontSize*2./3., 1, rl.Lime)

	// Debug Score
	text = fmt.Sprintf("hitScore: %.3d\nhitCount: %.3d\n", hitScore, hitCount)
	rl.DrawText(text, (screenW-10)-rl.MeasureText(text, 10), screenH-40, 10, rl.Green)

}

func Unload() {
	// TODO: Unload gameplay screen variables here!
	if rl.IsCursorHidden() {
		rl.EnableCursor() // without 3d ThirdPersonPerspective
	}

	// Commented out as it hinders switching to drill room or menu/ending (on pause/restart)
	// rl.UnloadMusicStream(music)
}

// Gameplay screen should finish?
// NOTE: This is called each frame in main game loop
func Finish() int {
	return finishScreen
}

// TEMPORARY
//		TEMPORARY
//			TEMPORARY
//				TEMPORARY
//					TEMPORARY
//						TEMPORARY
//							TEMPORARY
//								TEMPORARY
//									TEMPORARY
//										TEMPORARY
//											TEMPORARY
//												TEMPORARY
//													TEMPORARY
//														TEMPORARY

// // Migration/Reference/Example
// // Refactors huge blocks data state from main game data state
// var GameAdditionalLevelDataVersions = map[string]map[string]any{
// 	"0.0.0-blocks": {
// 		"version": "0.0.0-blocks",
// 		"levelID": levelID,
// 		"data": map[string]any{
// 			"blockArray": []block.Block{},
// 		},
// 	},
// 	"0.0.1": {},
// 	"0.0.2": {},
// }
//
// // Migration/Reference/Example
// var GameCoreLevelDataVersions = map[string]map[string]any{
// 	"0.0.0": {
// 		"version": "0.0.0",
// 		"levelID": levelID,
// 		"data": map[string]any{
// 			"camera":                 rl.Camera3D{},
// 			"finishScreen":           int(0),
// 			"framesCounter":          int32(0),
// 			"gameFloor":              floor.Floor{},
// 			"gamePlayer":             player.Player{},
// 			"hasPlayerLeftDrillBase": false,
// 		},
// 	},
// 	"0.0.1": {},
// 	"0.0.2": {},
// }
//

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
		Blocks: blocks,
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

//														TEMPORARY
//													TEMPORARY
//												TEMPORARY
//											TEMPORARY
//										TEMPORARY
//									TEMPORARY
//								TEMPORARY
//							TEMPORARY
//						TEMPORARY
//					TEMPORARY
//				TEMPORARY
//			TEMPORARY
//		TEMPORARY
// TEMPORARY
