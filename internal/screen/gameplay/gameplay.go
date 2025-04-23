package gameplay

import (
	"cmp"
	"fmt"
	"log"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/block"
	"example/depths/internal/common"
	"example/depths/internal/floor"
	"example/depths/internal/player"
	"example/depths/internal/storage"
	"example/depths/internal/util/mathutil"
	"example/depths/internal/wall"
)

type Gameplay struct {
}

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

	blockArray []block.Block
	blockCount int32
)

var (
	checkedTexture rl.Texture2D
	checkedModel   rl.Model
)

func Init() {
	framesCounter = 0
	finishScreen = 0

	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 16., 16.),
		Target:     rl.NewVector3(0., .5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       15. * float32(cmp.Or(3., 4., 2.)),
		Projection: rl.CameraPerspective,
	} // See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h

	// These props/tiles/objects share the same dungeon texture
	// dungeonTexture = rl.LoadTexture(filepath.Join("res", "texture", "dungeon_texture.png"))
	// dungeonTexture = common.Model.OBJ.Colormap

	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	//			SCENES 0..3
	//						SCENES 0..3
	player.InitPlayer(&gamePlayer, camera)
	gamePlayer.IsPlayerWallCollision = false

	floor.InitFloor(&gameFloor)
	wall.InitWall()
	// - Avoid spawning where player is standing
	// - Randomly skip a position
	// - A noise map or simplex/perlin noise "can" serve better
	InitAllBlocks(block.GenerateRandomBlockPositions(gameFloor))
	//						SCENES 0..3
	//			SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3

	checkedImg := rl.GenImageChecked(100, 100, 1, 1, rl.ColorBrightness(rl.Black, .25), rl.ColorBrightness(rl.Black, .2))
	checkedTexture = rl.LoadTextureFromImage(checkedImg)
	rl.UnloadImage(checkedImg)
	checkedModel = rl.LoadModelFromMesh(rl.GenMeshPlane(100, 100, 10, 10))
	checkedModel.Materials.Maps.Texture = checkedTexture

	switch isPlay := true; isPlay {
	case true:
		rl.PlayMusicStream(common.Music.OpenWorld000)
	case false:
		rl.PauseMusicStream(common.Music.OpenWorld000)
	}

	rl.DisableCursor() // for ThirdPersonPerspective
}

func HandleUserInput() {
	if rl.IsKeyDown(rl.KeyF) {
		log.Println("[F] Picked up item")
	}

	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) { /* || rl.IsGestureDetected(rl.GestureDrag) */
		finishScreen = 1
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "rollover3.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "switch_33.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_interface-sounds", "Audio", "confirmation_001.ogg")))

	}
}

func Update() {
	rl.UpdateMusicStream(common.Music.OpenWorld000)

	HandleUserInput()

	// Save variables this frame
	oldCam := camera
	oldPlayer := gamePlayer

	// Reset flags/variables
	gamePlayer.Collisions = rl.Quaternion{}
	gamePlayer.IsPlayerWallCollision = false

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	gamePlayer.Update(camera, gameFloor)

	if gamePlayer.IsPlayerWallCollision {
		player.RevertPlayerAndCameraPositions(oldPlayer, &gamePlayer, oldCam, &camera)
	}

	for i := range blockCount {
		// Skip final broken/mined block residue
		if blockArray[i].State == block.MaxBlockState-1 {
			continue
		}
		if rl.CheckCollisionBoxes(
			common.GetBoundingBoxFromPositionSizeV(blockArray[i].Pos, blockArray[i].Size),
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
			player.RevertPlayerAndCameraPositions(oldPlayer, &gamePlayer, oldCam, &camera)

			// Trigger once while mining
			if (rl.IsKeyDown(rl.KeySpace) && framesCounter%16 == 0) ||
				(rl.IsMouseButtonDown(rl.MouseLeftButton) && framesCounter%16 == 0) {
				// Play mining sound with variations (s1:kick + s2:snare + s3:hollow-thock)
				state := blockArray[i].State
				s1 := common.FXS.ImpactsSoftMedium[rl.GetRandomValue(int32(state), int32(len(common.FXS.ImpactsSoftMedium)-1))]
				s2 := common.FXS.ImpactsGenericLight[rl.GetRandomValue(int32(state), int32(len(common.FXS.ImpactsGenericLight)-1))]
				s3 := common.FXS.ImpactsSoftHeavy[rl.GetRandomValue(int32(state), int32(len(common.FXS.ImpactsSoftHeavy)-1))]
				rl.SetSoundVolume(s1, float32(rl.GetRandomValue(7, 10))/10.)
				rl.SetSoundVolume(s2, float32(rl.GetRandomValue(4, 8))/10.)
				rl.SetSoundVolume(s3, float32(rl.GetRandomValue(1, 4))/10.)
				rl.PlaySound(s1)
				rl.PlaySound(s2)
				rl.PlaySound(s3)

				// Increment state
				blockArray[i].NextState()
			}
		}
	}

	// Move this in package player
	if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyD) {
		const fps = 60
		const framesInterval = fps / 3.0
		if framesCounter%int32(framesInterval) == 0 {
			if !rl.Vector3Equals(oldPlayer.Position, gamePlayer.Position) &&
				rl.Vector3Distance(oldCam.Position, gamePlayer.Position) > 1.0 &&
				(gamePlayer.Collisions.X == 0 && gamePlayer.Collisions.Z == 0) {
				rl.PlaySound(common.FXS.FootStepsConcrete[int(framesCounter)%len(common.FXS.FootStepsConcrete)])
			}
		}
	}

	framesCounter++
}

var cachedCameraForward rl.Vector3

func Draw() {
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	rl.ClearBackground(rl.Black)

	{
		{
			var cameraViewMatrix rl.Matrix = rl.GetCameraMatrix(camera)
			var quat rl.Quaternion = rl.QuaternionFromMatrix(cameraViewMatrix)
			quatEulerPos := rl.QuaternionToEuler(quat)
			quatEulerLen := rl.Vector3Length(quatEulerPos)
			originOffset := rl.NewVector3(0., 5., 0.)
			position := rl.Vector3Add(quatEulerPos, originOffset)
			rl.DrawCubeWiresV(position, rl.NewVector3(.125, quatEulerLen, .125), rl.Violet)
			rl.DrawCubeWiresV(position, quatEulerPos, rl.Purple)

		}

		gamePlayer.Draw()

		// Draw player to camera direction
		{
			rl.DrawLine3D(gamePlayer.Position, common.Vector3Zero, rl.Red)
			rl.DrawLine3D(gamePlayer.Position, common.Vector3One, rl.Green)

			cameraForward := rl.GetCameraForward(&camera)
			cameraForwardProjectionVector3 := rl.Vector3Multiply(cameraForward, rl.NewVector3(9., .125/2., 9.))

			startPos := gamePlayer.Position
			endPos := rl.Vector3Add(gamePlayer.Position, cameraForwardProjectionVector3)
			if false {
				endPos.Y = startPos.Y // Maintain consistent y level as we cast parallel to XZ plane and perpendicular to Y axis
			}

			// Draw Rays
			rayCol := rl.Fade(rl.LightGray, .3)
			rl.DrawLine3D(startPos, endPos, rayCol)

			rayCol = rl.Fade(rayCol, .1)
			const maxRays = float32(8.)
			const rayGapFactor = 16 * maxRays
			for i := -maxRays; i < maxRays; i++ {
				rl.DrawLine3D(startPos, rl.Vector3Add(endPos, rl.NewVector3(i/rayGapFactor, .0, .0)), rayCol)
				rl.DrawLine3D(startPos, rl.Vector3Add(endPos, rl.NewVector3(.0, .0, i/rayGapFactor)), rayCol)
			}

			// Draw forward movement lookahead area
			rl.DrawCapsule(startPos, endPos, 2, 7, 7, rl.Fade(rl.Gray, .125/2))
			cachedCameraForward = cameraForward // Cache in update method

			{
				// Thanks to [hippocoder](https://discussions.unity.com/t/angle-between-camera-and-object/450430/9)
				// Best understand the tried and tested old school methods with
				// euler angles. This is the old way everyone has done since
				// time began. Behold the code in 2D (which you probably need
				// if you’re not bothered about all the angles):
				//
				//	function Angle2D(x1:float, y1:float, x2:float, y2:float) {
				// 		return Mathf.Atan2(y2-y1, x2-x1)*Mathf.Rad2Deg;
				// 	}
				//
				// What we do is use Atan2 and plug in two positions, a source
				// and a destination, which is converted to degrees. If you
				// want to use a different angle, just plug in x and z for
				// example or y and z. Have fun.
				Angle2D := func(x1, y1, x2, y2 float32) float32 {
					return mathutil.Atan2F(y2-y1, x2-x1) * rl.Rad2deg
				}
				degree := Angle2D(startPos.X, startPos.Z, endPos.X, endPos.Z)
				gamePlayer.Rotation = -90 + int32(degree)
			}
		}

		gameFloor.Draw()

		// Use walls to avoid infinite-map generation
		wall.DrawWalls(
			gameFloor.Position,
			gameFloor.Size,
			rl.NewVector3(1., cmp.Or(common.Phi, common.OneMinusInvPhi, float32(1.)), 1.),
		)

		// Draw drill
		const maxIndex = 2
		{ // Draw door gate entry logic before changing scene to drill base
			origin := common.Vector3Zero
			bb1 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(3, 2, 3)) // player is inside
			rl.DrawBoundingBox(bb1, rl.Red)
			bb2 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(5, 2, 5)) // player is entering
			rl.DrawBoundingBox(bb2, rl.Green)
			bb3 := common.GetBoundingBoxFromPositionSizeV(origin, rl.NewVector3(7, 2, 7)) // bot barrier
			rl.DrawBoundingBox(bb3, rl.Blue)
			{
				isPlayerInsideBase := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb1)
				isPlayerEnteringBase := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb2)
				isPlayerInsideBotBarrier := rl.CheckCollisionBoxes(gamePlayer.BoundingBox, bb3)

				if isPlayerInsideBotBarrier && !isPlayerEnteringBase && !isPlayerInsideBase {
					player.PlayerCol = rl.Blue
				} else if isPlayerEnteringBase && !isPlayerInsideBase {
					if hasPlayerLeftDrillBase { // HACK: Placeholder change scene check logic
						hasPlayerLeftDrillBase = false
						finishScreen = 2 // HACK: Placeholder to shift scene
						rl.PlaySound(common.FX.Coin)
					}
					player.PlayerCol = rl.Green
				} else if isPlayerInsideBase {
					player.PlayerCol = rl.Red
				} else {
					if !hasPlayerLeftDrillBase {
						hasPlayerLeftDrillBase = true
					}
					player.PlayerCol = rl.RayWhite
				}
			}
		}

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

		for i := range blockCount {
			obj := blockArray[i]
			rl.DrawModelEx(block.BlockModels[obj.State], obj.Pos,
				rl.NewVector3(0, 1, 0), obj.Rotn, obj.Size, rl.White)
		}

		if false {
			rl.DrawModel(checkedModel, rl.NewVector3(0., -.05, 0.), 1., rl.RayWhite)
		}
		if false {
			// Draw banners at floor corners
			floorBBMin := gameFloor.BoundingBox.Min
			floorBBMax := gameFloor.BoundingBox.Max
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X+1, 0, floorBBMin.Z+1), common.YAxis, 45, common.Vector3One, rl.White)  // leftback
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X-1, 0, floorBBMin.Z+1), common.YAxis, -45, common.Vector3One, rl.White) // rightback
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X, 0, floorBBMax.Z), common.YAxis, 45, common.Vector3One, rl.White)      // rightfront
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X, 0, floorBBMax.Z), common.YAxis, -45, common.Vector3One, rl.White)     // leftfront
		}
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
	// rl.UnloadMusicStream(music)
}

// Gameplay screen should finish?
func Finish() int {
	// PERF: Find way to reduce size.
	//       Size of "additional level state" is 117x times size of "core level state"
	saveCoreLevelState()       // 705 bytes   (player,camera,...)
	saveAdditionalLevelState() // 82871 bytes (blocks,...)

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
			"blockCount": int32(0),
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

func saveAdditionalLevelState() {
	data := storage.GameStorageLevelJSON{
		Version: "0.0.0-blocks",
		LevelID: levelID,
		Data:    map[string]any{"blockArray": blockArray, "blockCount": blockCount},
	}
	storage.SaveStorageLevelEx(data, "blocks")
}

func saveCoreLevelState() {
	data := storage.GameStorageLevelJSON{
		Version: "0.0.0",
		LevelID: levelID,
		Data: map[string]any{
			"camera":                 camera,
			"finishScreen":           finishScreen,
			"framesCounter":          framesCounter,
			"gameFloor":              gameFloor,
			"gamePlayer":             gamePlayer,
			"hasPlayerLeftDrillBase": hasPlayerLeftDrillBase,
		},
	}
	storage.SaveStorageLevel(data)
}

func InitAllBlocks(positions []rl.Vector3) {
	for i := range positions {
		size := rl.Vector3Multiply(
			rl.NewVector3(1, 1, 1),
			rl.NewVector3(
				float32(rl.GetRandomValue(88, 101))/100.,
				float32(rl.GetRandomValue(100, 300))/100.,
				float32(rl.GetRandomValue(88, 101))/100.))

		obj := block.NewBlock(positions[i], size)
		obj.Rotn = cmp.Or(float32(rl.GetRandomValue(-50, 50)/10.), 0.)

		blockArray = append(blockArray, obj)
		blockCount++
	}
	for i := range block.MaxBlockState {
		switch i {
		case block.DirtBlockState:
			block.BlockModels[i] = common.Model.OBJ.Dirt
		case block.RockBlockState:
			block.BlockModels[i] = common.Model.OBJ.Rocks
		case block.StoneBlockState:
			block.BlockModels[i] = common.Model.OBJ.Stones
		case block.FloorDetailBlockState:
			block.BlockModels[i] = common.Model.OBJ.FloorDetail
		default:
			panic(fmt.Sprintf("unexpected gameplay.BlockState: %#v", i))
		}
		rl.SetMaterialTexture(block.BlockModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
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
