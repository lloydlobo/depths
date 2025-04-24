package gameplay

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
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

	levelID int32

	finishScreen  int
	framesCounter int32

	camera                 rl.Camera3D
	gameFloor              floor.Floor
	gamePlayer             player.Player
	hasPlayerLeftDrillBase bool

	// Additional data

	blocks []block.Block
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

	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 8., 8.),
		Target:     rl.NewVector3(0., .5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       15. * float32(cmp.Or(3., 4., 2.)),
		Projection: rl.CameraPerspective,
	} // See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h

	// INIT WOULD REQUIRE A LEVEL ID? OR CREATE MULTIPLE gameplay000.go files
	loadNewCoreData := func() {
		var mu sync.Mutex

		mu.Lock()
		defer mu.Unlock()

		finishScreen = 0
		framesCounter = 0

		// Order is maybe important
		player.InitPlayer(&gamePlayer, camera)
		floor.InitFloor(&gameFloor)
		wall.InitWall() // NOTE: Empty func for convention

		hasPlayerLeftDrillBase = false
		gamePlayer.IsPlayerWallCollision = false
	}

	loadNewAdditionalData := func() {
		var mu sync.Mutex

		mu.Lock()
		defer mu.Unlock()

		blocks = []block.Block{} // Clear
		block.InitBlocks(&blocks, block.GenerateRandomBlockPositions(gameFloor))
	}

	const isNewGame = false

	// Core resources
	floor.SetupFloorModel()
	wall.SetupWallModel()
	player.SetupPlayerModel() // FIXME: in this func, use package common for models

	// Core data
	if !isNewGame {
		data, err := loadCoreGameData()
		if err == nil { // OK
			finishScreen = 0
			framesCounter = 0
			camera = data.Camera
			gameFloor = data.GameFloor
			gamePlayer = data.GamePlayer
			if false {
				hasPlayerLeftDrillBase = data.HasPlayerLeftDrillBase
			} else {
				hasPlayerLeftDrillBase = false
			}
			gamePlayer.IsPlayerWallCollision = false
			saveCoreLevelState() // Save ASAP
		} else { // ERR
			slog.Warn(err.Error())
			loadNewCoreData()
		}
	} else {
		loadNewCoreData()
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
			saveAdditionalLevelState() // Save ASAP
		} else { // ERR
			slog.Warn(err.Error())
			loadNewAdditionalData()
		}
	} else {
		loadNewAdditionalData()
	}

	{
		checkedImg := rl.GenImageChecked(100, 100, 1, 1, rl.ColorBrightness(rl.Black, .24), rl.ColorBrightness(rl.Black, .20))
		checkedTexture = rl.LoadTextureFromImage(checkedImg)
		rl.UnloadImage(checkedImg)
		checkedModel = rl.LoadModelFromMesh(rl.GenMeshPlane(100, 100, 10, 10))
		checkedModel.Materials.Maps.Texture = checkedTexture
	}

	musicChoices := []rl.Music{
		common.Music.OpenWorld000,
		common.Music.OpenWorld001,
	}
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
		isLoop := true
		for isLoop {
			tempMusic = musicChoices[rl.GetRandomValue(0, int32(len(musicChoices)-1))]
			if tempMusic != currentMusic {
				isLoop = false
				break // Just in case
			}
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
	if true {
		rl.PauseMusicStream(currentMusic)
	}

	rl.DisableCursor() // for ThirdPersonPerspective
}

// InitWorld loads resources each time it is called.
// Prefers loading data from saved game files if any, else generates new data.
// If new game flag is turned on, generates new game.
func InitWorld() {
	const isNewGame = false

	var mu sync.Mutex
	var wg sync.WaitGroup

	{
		checkedImg := rl.GenImageChecked(100, 100, 1, 1, rl.ColorBrightness(rl.Black, .24), rl.ColorBrightness(rl.Black, .20))
		checkedTexture = rl.LoadTextureFromImage(checkedImg)
		rl.UnloadImage(checkedImg)

		checkedModel = rl.LoadModelFromMesh(rl.GenMeshPlane(100, 100, 10, 10))
		checkedModel.Materials.Maps.Texture = checkedTexture
	}

	// Core resources
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()

		floor.SetupFloorModel()
		wall.SetupWallModel()
	}()

	player.SetupPlayerModel() // FIXME: This has File i/o logic.. Should use resources loaded common to load models apriori

	// Core data
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()
		loadNewData := func() {
			var mu sync.Mutex

			mu.Lock()
			defer mu.Unlock()

			finishScreen = 0  //data.FinishScreen
			framesCounter = 0 // data.FramesCounter
			player.InitPlayer(&gamePlayer, camera)
			floor.InitFloor(&gameFloor)
			wall.InitWall()
			hasPlayerLeftDrillBase = false
			gamePlayer.IsPlayerWallCollision = false
		}
		if isNewGame {
			loadNewData()
		} else {
			data, err := loadCoreGameData()
			if err != nil {
				slog.Warn(err.Error()) // WARN: improve error handling
				loadNewData()
			} else { // Resume from saved state
				finishScreen = 0  // data.FinishScreen
				framesCounter = 0 // data.FramesCounter
				camera = data.Camera
				gameFloor = data.GameFloor
				gamePlayer = data.GamePlayer
				if false { // Should recalculate if player positiion is inside drill.. else default to false
					hasPlayerLeftDrillBase = data.HasPlayerLeftDrillBase
				} else {
					hasPlayerLeftDrillBase = false
				}
				gamePlayer.IsPlayerWallCollision = false
				saveCoreLevelState()
			}
		}
	}()

	// Additional resources
	if true { // FIXES: Unloaded huge blocks saved file
		block.SetupBlockModels()
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()

			block.SetupBlockModels()
		}()
	}

	loadNewAdditionalData := func() {
		var mu sync.Mutex
		mu.Lock()
		defer mu.Unlock()
		if len(blocks) >= 0 {
			blocks = []block.Block{} // Clear all blocks
		}
		block.InitBlocks(&blocks, block.GenerateRandomBlockPositions(gameFloor))
	}

	// Additional data
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()

		if isNewGame {
			loadNewAdditionalData()
		} else {
			additionalGameData, err := loadAdditionalGameData()
			fmt.Printf("additionalGameData: %v\n", additionalGameData)
			if err != nil {
				slog.Warn(err.Error()) // WARN: improve error handling
				loadNewAdditionalData()
			} else { // Resume from saved state
				blocks = make([]block.Block, len(additionalGameData.Blocks)) // PERF: Make this a static array and use a counter
				copiedBlockCount := copy(blocks, additionalGameData.Blocks)
				log.Println(fmt.Sprintf("blocks copied: %v", copiedBlockCount))
				saveAdditionalLevelState()
			}
		}
	}()

	// Wait blocks until the WaitGroup counter is zero
	wg.Wait()
}

func HandleUserInput() {
	if rl.IsKeyDown(rl.KeyF) {
		log.Println("[F] Picked up item")
	}

	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) { /* || rl.IsGestureDetected(rl.GestureDrag) */
		finishScreen = 1
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "rollover3.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "switch33.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_interface-sounds", "Audio", "confirmation_001.ogg")))
	}
}

func Update() {
	rl.UpdateMusicStream(currentMusic)

	HandleUserInput()

	// Save variables this frame
	oldCam := camera
	oldPlayer := gamePlayer

	// Reset flags/variables
	gamePlayer.Collisions = rl.Quaternion{}
	gamePlayer.IsPlayerWallCollision = false

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	gamePlayer.Update(camera, gameFloor)
	// Update player rotation.. based on camera forward projection
	{
		startPos := gamePlayer.Position
		endPos := rl.Vector3Add(gamePlayer.Position, rl.GetCameraForward(&camera))
		degree := mathutil.Angle2D(startPos.X, startPos.Z, endPos.X, endPos.Z)
		gamePlayer.Rotation = -90 + int32(degree)
	}

	if gamePlayer.IsPlayerWallCollision {
		player.RevertPlayerAndCameraPositions(&gamePlayer, oldPlayer, &camera, oldCam)
	}

	for i := range blocks {
		// Skip final broken/mined block residue
		if blocks[i].State == block.MaxBlockState-1 {
			continue
		}
		if rl.CheckCollisionBoxes(
			common.GetBoundingBoxFromPositionSizeV(blocks[i].Pos, blocks[i].Size),
			gamePlayer.BoundingBox,
		) {
			// FIND OUT WHERE PLAYER TOUCHED THE BOX
			// HACK
			//		HACK
			//			HACK
			gamePlayer.Collisions.Z = 1
			dx := oldPlayer.Position.X - gamePlayer.Position.X
			if dx < 0.0 { // new <- old
				gamePlayer.Collisions.X = 1
			} else if dx > 0.0 { // new -> old
				gamePlayer.Collisions.X = -1
			} else {
				if gamePlayer.Collisions.X != 0 { // Placeholder (do not overwrite previous)
					gamePlayer.Collisions.X = 0
				}
			}
			dz := oldPlayer.Position.Z - gamePlayer.Position.Z
			if dz < 0.0 { // new <- old
				gamePlayer.Collisions.Z = 1
			} else if dz > 0.0 { // new -> old
				gamePlayer.Collisions.Z = -1
			} else {
				if gamePlayer.Collisions.Z != 0 { // Placeholder (do not overwrite previous)
					gamePlayer.Collisions.Z = 0
				}
			}
			//			HACK
			//		HACK
			// HACK
			player.RevertPlayerAndCameraPositions(&gamePlayer, oldPlayer, &camera, oldCam)

			// Trigger once while mining
			if (rl.IsKeyDown(rl.KeySpace) && framesCounter%16 == 0) ||
				(rl.IsMouseButtonDown(rl.MouseLeftButton) && framesCounter%16 == 0) {
				// Play player weapon sounds
				{
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
					alpha := 0.5 + float32(rl.GetRandomValue(-10, 10))/40.0
					rl.SetSoundPan(v, alpha)
					rl.SetSoundVolume(v, 0.5)
					rl.PlaySound(v)
				}
				if true {
					v := rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio",
						fmt.Sprintf("cloth%d.ogg", min(block.MaxBlockState-1, max(1, state+1)))))
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
					v := rl.LoadSound(filepath.Join("res", "fx", "kenney_impact-sounds", "Audio",
						fmt.Sprintf("impactMining_00%d.ogg", min(block.MaxBlockState-1, state))))
					rl.SetSoundPan(v, 0.5+float32(rl.GetRandomValue(-10, 10)/(2*10)))
					rl.SetSoundVolume(v, 2.00)
					rl.PlaySound(v)
				}

				// Increment state
				blocks[i].NextState()
			}
		}
	}

	// Update player─block collision+breaking/mining
	{
		origin := common.Vector3Zero
		bb1 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(3, 2, 3)) // player is inside
		bb2 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(5, 2, 5)) // player is entering
		bb3 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(7, 2, 7)) // bot barrier

		isPlayerInsideBase := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb1)
		isPlayerEnteringBase := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb2)
		isPlayerInsideBotBarrier := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb3)

		if isPlayerInsideBotBarrier && !isPlayerEnteringBase && !isPlayerInsideBase {
			player.SetColor(rl.Blue)
		} else if isPlayerEnteringBase && !isPlayerInsideBase {
			if hasPlayerLeftDrillBase { // HACK: Placeholder change scene check logic
				hasPlayerLeftDrillBase = false
				finishScreen = 2 // HACK: Placeholder to shift scene
				rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "footstep05.ogg")))
				rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "metalClick.ogg")))
				rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "creak3.ogg")))
				rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_rpg-audio", "Audio", "doorOpen_2.ogg")))
			}
			player.SetColor(rl.Green)
		} else if isPlayerInsideBase {
			player.SetColor(rl.Red)
		} else {
			if !hasPlayerLeftDrillBase {
				hasPlayerLeftDrillBase = true
			}
			player.SetColor(rl.RayWhite)
		}
	}

	// Move this in package player
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyD) {
		const fps = 60.0
		const framesInterval = fps / 2.5
		if framesCounter%int32(framesInterval) == 0 {
			if !rl.Vector3Equals(oldPlayer.Position, gamePlayer.Position) &&
				rl.Vector3Distance(oldCam.Position, gamePlayer.Position) > 1.0 &&
				(gamePlayer.Collisions.X == 0 && gamePlayer.Collisions.Z == 0) {
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

	gameFloor.Draw()

	if false { // ‥ Draw pseudo-infinite(ish) floor backdrop
		rl.DrawModel(checkedModel, rl.NewVector3(0., -.05, 0.), 1., rl.RayWhite)
	}

	wall.DrawBatch(gameFloor.Position, gameFloor.Size, common.Vector3One)

	for i := range blocks {
		blocks[i].Draw()
	}

	gamePlayer.Draw()
	{ // ‥ Draw player to camera forward projected direction
		const maxRays = float32(8.)
		const rayGapFactor = 16 * maxRays
		rayCol := rl.Fade(rl.LightGray, .3)
		startPos := gamePlayer.Position // NOTE: startPos.Y and endPos.Y may fluctuate
		endPos := rl.Vector3Add(
			gamePlayer.Position,
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

	{ // ‥ Draw drill
		const maxIndex = 2
		wallScale := rl.NewVector3(1., 1., 1.)
		for i := float32(-maxIndex + 1); i < maxIndex; i++ {
			var model rl.Model
			var y float32
			model = common.Model.OBJ.Column
			y = 0.
			rl.DrawModelEx(model, rl.NewVector3(i, y, maxIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
			rl.DrawModelEx(model, rl.NewVector3(i, y, -maxIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
			rl.DrawModelEx(model, rl.NewVector3(maxIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
			rl.DrawModelEx(model, rl.NewVector3(-maxIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
			model = common.Model.OBJ.Wall
			y = 1. + .125*.5
			rl.DrawModelEx(model, rl.NewVector3(i, y, maxIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
			rl.DrawModelEx(model, rl.NewVector3(i, y, -maxIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
			rl.DrawModelEx(model, rl.NewVector3(maxIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
			rl.DrawModelEx(model, rl.NewVector3(-maxIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
			model = common.Model.OBJ.Column
			y = 2. + .125*.5
			rl.DrawModelEx(model, rl.NewVector3(i, y, maxIndex), common.YAxis, 0., wallScale, rl.White)    // +-X +Z
			rl.DrawModelEx(model, rl.NewVector3(i, y, -maxIndex), common.YAxis, 180., wallScale, rl.White) // +-X -Z
			rl.DrawModelEx(model, rl.NewVector3(maxIndex, y, i), common.YAxis, 90., wallScale, rl.White)   // +X +-Z
			rl.DrawModelEx(model, rl.NewVector3(-maxIndex, y, i), common.YAxis, -90., wallScale, rl.White) // -X +-Z
		}

		if false { // ‥ DEBUG: Draw drill door gate entry logic before changing scene to drill base
			origin := common.Vector3Zero
			bb1 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(3, 2, 3)) // player is inside
			bb2 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(5, 2, 5)) // player is entering
			bb3 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(7, 2, 7)) // bot barrier

			rl.DrawBoundingBox(bb1, rl.Red)
			rl.DrawBoundingBox(bb2, rl.Green)
			rl.DrawBoundingBox(bb3, rl.Blue)
		}
	}

	if true { // ‥ Draw banners at floor corners
		floorBBMin := gameFloor.BoundingBox.Min
		floorBBMax := gameFloor.BoundingBox.Max
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X+1, 0, floorBBMin.Z+1), common.YAxis, 45, common.Vector3One, rl.White)  // leftback
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X-1, 0, floorBBMin.Z+1), common.YAxis, -45, common.Vector3One, rl.White) // rightback
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X, 0, floorBBMax.Z), common.YAxis, 45, common.Vector3One, rl.White)      // rightfront
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X, 0, floorBBMax.Z), common.YAxis, -45, common.Vector3One, rl.White)     // leftfront
	}

	if false { //  ‥ DEBUG: Draw camera movement gimble-like interpretation
		var cameraViewMatrix rl.Matrix = rl.GetCameraMatrix(camera)
		var quat rl.Quaternion = rl.QuaternionFromMatrix(cameraViewMatrix)
		quatEulerPos := rl.QuaternionToEuler(quat)
		quatEulerLen := rl.Vector3Length(quatEulerPos)
		originOffset := rl.NewVector3(0., 5., 0.)
		position := rl.Vector3Add(quatEulerPos, originOffset)
		rl.DrawCubeWiresV(position, rl.NewVector3(.125, quatEulerLen, .125), rl.Violet)
		rl.DrawCubeWiresV(position, quatEulerPos, rl.Purple)
	}

	rl.EndMode3D()

	// 2D World
	fontSize := float32(common.Font.Primary.BaseSize) * 3.0
	text := "[F] PICK UP"
	rl.DrawFPS(10, 10)
	rl.DrawText(text, screenW/2-rl.MeasureText(text, 20)/2, screenH-20*2, 20, rl.White)
	rl.DrawTextEx(common.Font.Primary,
		fmt.Sprintf("%.6f", rl.GetFrameTime()),
		rl.NewVector2(10, 10+20*1),
		fontSize*2./3., 1, rl.Lime)
	rl.DrawText(fmt.Sprintf("camera forward:%.2f\ncamera right:%.2f\n",
		rl.GetCameraForward(&camera), rl.GetCameraRight(&camera)), screenW-200,
		screenH-40, 10, rl.Green)
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
func Finish() int {
	//
	// PERF: Find way to reduce size. => Size of "additional level state" is
	//       117x times size of "core level state"
	//
	saveCoreLevelState()       // (player,camera,...) 705 bytes
	saveAdditionalLevelState() // (blocks,...)        82871 bytes

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

// Migration/Reference/Example
// Refactors huge blocks data state from main game data state
var GameAdditionalLevelDataVersions = map[string]map[string]any{
	"0.0.0-blocks": {
		"version": "0.0.0-blocks",
		"levelID": levelID,
		"data": map[string]any{
			"blockArray": []block.Block{},
		},
	},
	"0.0.1": {},
	"0.0.2": {},
}

// Migration/Reference/Example
var GameCoreLevelDataVersions = map[string]map[string]any{
	"0.0.0": {
		"version": "0.0.0",
		"levelID": levelID,
		"data": map[string]any{
			"camera":                 rl.Camera3D{},
			"finishScreen":           int(0),
			"framesCounter":          int32(0),
			"gameFloor":              floor.Floor{},
			"gamePlayer":             player.Player{},
			"hasPlayerLeftDrillBase": false,
		},
	},
	"0.0.1": {},
	"0.0.2": {},
}

type GameCoreData struct {
	Camera                 rl.Camera3D   `json:"camera"`
	FinishScreen           int           `json:"finishScreen"`
	FramesCounter          int32         `json:"framesCounter"`
	GameFloor              floor.Floor   `json:"gameFloor"`
	GamePlayer             player.Player `json:"gamePlayer"`
	HasPlayerLeftDrillBase bool          `json:"hasPlayerLeftDrillBase"`
}

type GameAdditionalData struct {
	Blocks []block.Block `json:"blocks"`
}

func saveCoreLevelState() {
	input := GameCoreData{
		Camera:                 camera,
		FinishScreen:           finishScreen,
		FramesCounter:          framesCounter,
		GameFloor:              gameFloor,
		GamePlayer:             gamePlayer,
		HasPlayerLeftDrillBase: hasPlayerLeftDrillBase,
	}
	var b []byte
	bb := bytes.NewBuffer(b)
	{
		enc := json.NewEncoder(bb)
		if err := enc.Encode(input); err != nil {
			panic(fmt.Errorf("encode level: %w", err))
		}
	}
	dataJSON := storage.GameStorageLevelJSON{
		Version: "0.0.0",
		LevelID: levelID,
		Data:    bb.Bytes(),
	}
	storage.SaveStorageLevel(dataJSON)

}

func saveAdditionalLevelState() {
	input := GameAdditionalData{
		Blocks: blocks,
	}
	var b []byte
	bb := bytes.NewBuffer(b)
	{
		enc := json.NewEncoder(bb)
		if err := enc.Encode(input); err != nil {
			panic(fmt.Errorf("encode level: %w", err))
		}
	}
	data := storage.GameStorageLevelJSON{
		Version: "0.0.0-blocks",
		LevelID: levelID,
		Data:    bb.Bytes(),
	}
	storage.SaveStorageLevelEx(data, "blocks")
}

func loadCoreGameData() (*GameCoreData, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	saveDir := filepath.Join(cwd, "storage")
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(levelID))+".json")

	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", name, err)
	}

	dest := &storage.GameStorageLevelJSON{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&dest); err != nil {
		return nil, fmt.Errorf("decode level: %w", err)
	}

	// return dest,nil
	// Upto here.. same as storage.LoadStorageLevel

	switch version := dest.Version; version {
	case "0.0.0":
		var v *GameCoreData
		if err := json.Unmarshal(dest.Data, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("invalid game core data version %q", version)
	}
}

func loadAdditionalGameData() (*GameAdditionalData, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	saveDir := filepath.Join(cwd, "storage")
	name := filepath.Join(saveDir, "level_"+strconv.Itoa(int(levelID))+"_blocks"+".json")

	f, err := os.OpenFile(name, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("create %q: %w", name, err)
	}

	dest := &storage.GameStorageLevelJSON{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&dest); err != nil {
		return nil, fmt.Errorf("decode level: %w", err)
	}

	// return dest,nil
	// Upto here.. same as storage.LoadStorageLevel

	switch version := dest.Version; version {
	case "0.0.0-blocks":
		var v *GameAdditionalData
		if err := json.Unmarshal(dest.Data, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("invalid game additional data version %q", version)
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
