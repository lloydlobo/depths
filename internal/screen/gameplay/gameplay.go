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

var (
	// dungeonTexture rl.Texture2D

	barrelModel         rl.Model
	barrelPositions     []rl.Vector3
	barrelBoundingBoxes []rl.BoundingBox
	barrelSize          rl.Vector3
	barrelCount         int32
)

const (
	DirtDSR DirtStoneRockState = iota
	RockDSR
	StoneDSR

	maxDirtStoneRockStates
)

var (
	dirtStoneRockArray  []DirtStoneRockObj
	dirtStoneRockCount  int32
	dirtStoneRockModels [maxDirtStoneRockStates]rl.Model
)

type DirtStoneRockState uint8
type DirtStoneRockObj struct {
	Pos      rl.Vector3
	Size     rl.Vector3
	Health   float32 // [0..1]
	State    DirtStoneRockState
	IsActive bool
}

func NewDirtStoneRockObj(pos, size rl.Vector3) DirtStoneRockObj {
	return DirtStoneRockObj{
		Pos:      pos,
		Size:     size,
		State:    DirtDSR,
		IsActive: true,
	}
}

func (o *DirtStoneRockObj) NextState() {
	o.State++
	if o.State >= maxDirtStoneRockStates {
		o.State = maxDirtStoneRockStates - 1
		o.IsActive = false
	}
}

func InitDirtStoneRockObjects(positions []rl.Vector3) {
	size := rl.NewVector3(1, 1, 1)
	for i := range positions {
		dirtStoneRockArray = append(dirtStoneRockArray, NewDirtStoneRockObj(positions[i], size))
		dirtStoneRockCount++
	}
	for i := range maxDirtStoneRockStates {
		switch i {
		case DirtDSR:
			dirtStoneRockModels[i] = common.Model.OBJ.Dirt
			rl.SetMaterialTexture(dirtStoneRockModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
		case RockDSR:
			dirtStoneRockModels[i] = common.Model.OBJ.Rocks
			rl.SetMaterialTexture(dirtStoneRockModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
		case StoneDSR:
			dirtStoneRockModels[i] = common.Model.OBJ.Stones
			rl.SetMaterialTexture(dirtStoneRockModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
		default:
			panic(fmt.Sprintf("unexpected gameplay.DirtStoneRockState: %#v", i))
		}
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
	InitDirtStoneRockObjects([]rl.Vector3{
		rl.NewVector3(2, 0, -8),
		rl.NewVector3(-3, 0, -6),
		rl.NewVector3(-8, 0, 5),
		rl.NewVector3(-5, 0, -4),
	})
	InitBarrels := func(positions []rl.Vector3) {
		barrelSize = rl.NewVector3(0.5, 0.5, 0.5)
		// barrelSize = rl.NewVector3(1.0, 1.0, 1.0)
		for _, pos := range positions {
			barrelPositions = append(barrelPositions, pos)
			barrelBoundingBoxes = append(barrelBoundingBoxes, common.GetBoundingBoxFromPositionSizeV(pos, barrelSize))
			barrelCount++
		}
		barrelModel = common.Model.OBJ.Barrel
		// barrelModel = common.Model.OBJ.Dirt
		rl.SetMaterialTexture(barrelModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
	}
	InitBarrels([]rl.Vector3{
		rl.NewVector3(-5, 0, -8),
		rl.NewVector3(-3, 0, -7),
		rl.NewVector3(4, 0, 7),
		rl.NewVector3(5, 0, -4),
	})

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
	for i := range barrelCount {
		if rl.CheckCollisionBoxes(barrelBoundingBoxes[i], player.BoundingBox) {
			RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)
		}
	}
	for i := range dirtStoneRockCount {
		if rl.CheckCollisionBoxes(
			common.GetBoundingBoxFromPositionSizeV(dirtStoneRockArray[i].Pos, dirtStoneRockArray[i].Size),
			player.BoundingBox,
		) {
			RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)
			// Trigger once
			if rl.IsKeyPressed(rl.KeyF) {
				dirtStoneRockArray[i].NextState()
			}
		}
	}

	framesCounter++
}

func Draw() {
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)
	rl.ClearBackground(rl.Black)
	{
		player.Draw()
		floor.Draw()
		if false {
			DrawWalls()
		}
		for _, pos := range barrelPositions { // Draw offgrid tiles
			rl.DrawModelEx(barrelModel, pos, rl.NewVector3(0, 1, 0), 0., rl.NewVector3(1, 1, 1), rl.White)
		}
		for i := range dirtStoneRockCount {
			chest := dirtStoneRockArray[i]
			rl.DrawModelEx(dirtStoneRockModels[chest.State], chest.Pos, rl.NewVector3(0, 1, 0), 0., rl.NewVector3(1, 1, 1), rl.White)
		}

		rl.DrawModel(checkedModel, rl.NewVector3(0., -.05, 0.), 1., rl.RayWhite)

		// Draw banners at floor corners
		floorBBMin := floor.BoundingBox.Min
		floorBBMax := floor.BoundingBox.Max
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X+1, 0, floorBBMin.Z+1), common.YAxis, 45, common.Vector3One, rl.White)  // leftback
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X-1, 0, floorBBMin.Z+1), common.YAxis, -45, common.Vector3One, rl.White) // rightback
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMax.X, 0, floorBBMax.Z), common.YAxis, 45, common.Vector3One, rl.White)      // rightfront
		rl.DrawModelEx(common.Model.OBJ.Banner, rl.NewVector3(floorBBMin.X, 0, floorBBMax.Z), common.YAxis, -45, common.Vector3One, rl.White)     // leftfront

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
