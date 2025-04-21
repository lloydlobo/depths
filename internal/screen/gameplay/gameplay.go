package gameplay

import (
	"cmp"
	"fmt"
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

var (
	framesCounter int32
	finishScreen  int

	camera rl.Camera3D

	player                Player
	floor                 Floor
	isPlayerWallCollision bool

	checkedTexture rl.Texture2D
	checkedModel   rl.Model

	// fxImpactsSoftHeavy    []rl.Sound
	// fxImpactsSoftMedium   []rl.Sound
	// fxImpactsGenericLight []rl.Sound
	// fxConcreteFootsteps []rl.Sound
)

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

const (
	DirtMineObjState MineObjState = iota
	RockMineObjState
	StoneMineObjState
	FloorDetailMineObjState // decorated floor tile

	maxMineObjState
)

var (
	mineObjArray  []MineObj
	mineObjCount  int32
	mineObjModels [maxMineObjState]rl.Model
)

type MineObjState uint8
type MineObj struct {
	Pos      rl.Vector3
	Size     rl.Vector3
	Rotn     float32
	Health   float32 // [0..1]
	State    MineObjState
	IsActive bool
}

func NewMineObj(pos, size rl.Vector3) MineObj {
	return MineObj{
		Pos:      pos,
		Size:     size,
		Rotn:     0.0,
		State:    DirtMineObjState,
		IsActive: true,
	}
}

func (o *MineObj) NextState() {
	o.State++
	if o.State >= maxMineObjState {
		o.State = maxMineObjState - 1
		o.IsActive = false
	}
}

func InitMineObjPositions() []rl.Vector3 {
	var positions []rl.Vector3 // 61% of maxPositions

	var (
		y    = (floor.BoundingBox.Min.Y + floor.BoundingBox.Max.Y) / 2.0
		bb   = floor.BoundingBox
		offX = float32(3)
		offZ = float32(3)
	)

	var (
		maxGridCells            = floor.Size.X * floor.Size.Z // just-in-case
		maxSkipLoopPositionOdds = int32(2)                    // if 2 -> 0,1,2 -> 1/3 odds
	)

NextCol:
	for x := bb.Min.X + 1; x < bb.Max.X; x++ {
	NextRow:
		for z := bb.Min.Z + 1; z < bb.Max.Z; z++ {
			if len(positions) >= int(maxGridCells) {
				break NextCol
			}
			// Reserve space for area in offset from origin
			for i := -offX; i <= offX; i++ {
				for k := -offZ; k <= offZ; k++ {
					if i == x && k == z {
						continue NextRow
					}
					if rl.Vector3Distance(rl.NewVector3(i, y, k), rl.NewVector3(x, y, z)) < (offX+offZ)/2 {
						continue NextRow

					}
				}
			}
			if rl.GetRandomValue(0, maxSkipLoopPositionOdds) == 0 {
				continue NextRow
			}
			positions = append(positions, rl.NewVector3(x, y, z))
		}
	}
	return positions
}

func InitAllMineObj(positions []rl.Vector3) {
	for i := range positions {
		size := rl.Vector3Multiply(
			rl.NewVector3(1, 1, 1),
			rl.NewVector3(
				float32(rl.GetRandomValue(88, 101))/100.,
				float32(rl.GetRandomValue(161, 2*161))/100.,
				float32(rl.GetRandomValue(88, 101))/100.))

		obj := NewMineObj(positions[i], size)
		obj.Rotn = cmp.Or(float32(rl.GetRandomValue(-50, 50)/10.), 0.)

		mineObjArray = append(mineObjArray, obj)
		mineObjCount++
	}
	for i := range maxMineObjState {
		switch i {
		case DirtMineObjState:
			mineObjModels[i] = common.Model.OBJ.Dirt
		case RockMineObjState:
			mineObjModels[i] = common.Model.OBJ.Rocks
		case StoneMineObjState:
			mineObjModels[i] = common.Model.OBJ.Stones
		case FloorDetailMineObjState:
			mineObjModels[i] = common.Model.OBJ.FloorDetail
		default:
			panic(fmt.Sprintf("unexpected gameplay.MineObjState: %#v", i))
		}
		rl.SetMaterialTexture(mineObjModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
	}
}

// TODO: Implement
//
//	scene_1.go
//	scene_2.go
//	scene_3.go
//	scene_4.go
//	scene_5.go
//	scene_6.go
type SceneManager struct {
	CurrentID  int32
	PreviousID int32
}

func (sm *SceneManager) SwitchTo(id int32) {
	sm.PreviousID = sm.CurrentID
	sm.CurrentID = id
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

	isPlayerWallCollision = false

	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	// SCENES 0..3
	//			SCENES 0..3
	//						SCENES 0..3
	InitPlayer()
	InitFloor()
	InitWall()
	// - Avoid spawning where player is standing
	// - Randomly skip a position
	// - A noise map or simplex/perlin noise "can" serve better
	InitAllMineObj(InitMineObjPositions())
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

	rl.SetMusicVolume(common.Music.Theme, float32(cmp.Or(1.0, 0.125)))

	rl.PlayMusicStream(common.Music.Theme)

	rl.DisableCursor() // for ThirdPersonPerspective
}

func HandleUserInput() {
	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) { /* || rl.IsGestureDetected(rl.GestureDrag) */
		finishScreen = 1
		rl.PlaySound(common.FX.Coin)
	}
	if rl.IsKeyDown(rl.KeyF) {
		log.Println("[F] Picked up item")
	}
}

func Update() {
	HandleUserInput()

	// Save variables this frame
	oldCam := camera
	oldPlayer := player

	// Reset flags/variables
	player.Collisions = rl.Quaternion{}
	isPlayerWallCollision = false

	rl.UpdateMusicStream(common.Music.Theme)
	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	player.Update()

	if isPlayerWallCollision {
		RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)
	}
	for i := range mineObjCount {
		// Skip final mined object residue
		if mineObjArray[i].State == maxMineObjState-1 {
			continue
		}
		if rl.CheckCollisionBoxes(
			common.GetBoundingBoxFromPositionSizeV(mineObjArray[i].Pos, mineObjArray[i].Size),
			player.BoundingBox,
		) {
			// FIND OUT WHERE PLAYER TOUCHED THE BOX
			// HACK
			//		HACK
			//			HACK
			player.Collisions.Z = 1
			dx := oldPlayer.Position.X - player.Position.X
			if dx < 0.0 { // new <- old
				player.Collisions.X = 1
			} else if dx > 0.0 { // new -> old
				player.Collisions.X = -1
			} else {
				if player.Collisions.X != 0 { // Placeholder (do not overwrite previous)
					player.Collisions.X = 0
				}
			}
			dz := oldPlayer.Position.Z - player.Position.Z
			if dz < 0.0 { // new <- old
				player.Collisions.Z = 1
			} else if dz > 0.0 { // new -> old
				player.Collisions.Z = -1
			} else {
				if player.Collisions.Z != 0 { // Placeholder (do not overwrite previous)
					player.Collisions.Z = 0
				}
			}
			//			HACK
			//		HACK
			// HACK
			RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)

			// Trigger once while mining
			if (rl.IsKeyDown(rl.KeySpace) && framesCounter%16 == 0) ||
				(rl.IsMouseButtonDown(rl.MouseLeftButton) && framesCounter%16 == 0) {
				// Play mining sound with variations (s1:kick + s2:snare + s3:hollow-thock)
				state := mineObjArray[i].State
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
				mineObjArray[i].NextState()
			}
		}
	}

	if rl.IsKeyDown(rl.KeyW) ||
		rl.IsKeyDown(rl.KeyA) ||
		rl.IsKeyDown(rl.KeyS) ||
		rl.IsKeyDown(rl.KeyD) {
		const fps = 60
		const framesInterval = fps / 3.0
		if framesCounter%int32(framesInterval) == 0 {
			if !rl.Vector3Equals(oldPlayer.Position, player.Position) &&
				rl.Vector3Distance(oldCam.Position, player.Position) > 1.0 &&
				(player.Collisions.X == 0 && player.Collisions.Z == 0) {
				rl.PlaySound(common.FXS.FootStepsConcrete[int(framesCounter)%len(common.FXS.FootStepsConcrete)])
			}
		}
	}

	framesCounter++
}

var hasLeftDrillBase bool

func Draw() {
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	rl.ClearBackground(rl.Black)

	{
		player.Draw()

		floor.Draw()

		// Use walls to avoid infinite-map generation
		DrawWalls(
			floor.Position,
			floor.Size,
			rl.NewVector3(
				1.,
				cmp.Or(common.Phi, common.OneMinusInvPhi, float32(1.)),
				1.,
			),
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
				isPlayerInsideBase := rl.CheckCollisionBoxes(player.BoundingBox, bb1)
				isPlayerEnteringBase := rl.CheckCollisionBoxes(player.BoundingBox, bb2)
				isPlayerInsideBotBarrier := rl.CheckCollisionBoxes(player.BoundingBox, bb3)
				if isPlayerInsideBotBarrier && !isPlayerEnteringBase && !isPlayerInsideBase {
					playerCol = rl.Blue
				} else if isPlayerEnteringBase && !isPlayerInsideBase {
					// HACK: Placeholder change scene check logic
					if hasLeftDrillBase {
						hasLeftDrillBase = false
						if false {
							Init()
						}
						finishScreen = 1 // HACK: Placeholder to shift scene
						rl.PlaySound(common.FX.Coin)
					}
					playerCol = rl.Green
				} else if isPlayerInsideBase {
					playerCol = rl.Red
				} else {
					if !hasLeftDrillBase {
						hasLeftDrillBase = true
					}
					playerCol = rl.White
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

		for i := range mineObjCount {
			obj := mineObjArray[i]
			rl.DrawModelEx(mineObjModels[obj.State], obj.Pos,
				rl.NewVector3(0, 1, 0), obj.Rotn, obj.Size, rl.White)
		}

		if false {
			rl.DrawModel(checkedModel, rl.NewVector3(0., -.05, 0.), 1., rl.RayWhite)
		}

		if false {
			// Draw banners at floor corners
			floorBBMin := floor.BoundingBox.Min
			floorBBMax := floor.BoundingBox.Max
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X+1, 0, floorBBMin.Z+1), common.YAxis, 45, common.Vector3One, rl.White)  // leftback
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X-1, 0, floorBBMin.Z+1), common.YAxis, -45, common.Vector3One, rl.White) // rightback
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X, 0, floorBBMax.Z), common.YAxis, 45, common.Vector3One, rl.White)      // rightfront
			rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X, 0, floorBBMax.Z), common.YAxis, -45, common.Vector3One, rl.White)     // leftfront
		}
		if false {
			// rl.DrawModel(common.Model.OBJ.Dirt, rl.NewVector3(0, common.Phi, 0), common.Phi, rl.White)
			rl.DrawModel(common.Model.OBJ.WoodStructure, rl.NewVector3(0, 0, 0), 1., rl.White)

			rl.DrawModel(common.Model.OBJ.CharacterHuman, common.Vector3One, 1., rl.Red)

			rl.DrawModel(common.Model.OBJ.Dirt, rl.NewVector3(4, 0, 4), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.Rocks, rl.NewVector3(5, 0, 4), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.Dirt, rl.NewVector3(6, 0, 4), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.Stones, rl.NewVector3(7, 0, 4), 1., rl.White)

			rl.DrawModel(common.Model.OBJ.Coin, rl.NewVector3(1, 0, 1), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.WoodSupport, rl.NewVector3(2, 0, 2), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.WoodStructure, rl.NewVector3(3, 0, 3), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.Dirt, rl.NewVector3(4, 0, 4), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.Rocks, rl.NewVector3(5, 0, 5), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.Stones, rl.NewVector3(6, 0, 6), 1., rl.White)
			rl.DrawModel(common.Model.OBJ.Trap, rl.NewVector3(7, 0, 7), 1., rl.White)
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
	return finishScreen
}
