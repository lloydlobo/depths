package gameplay

import (
	"cmp"
	"fmt"
	"log"
	"path/filepath"

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

	fxImpactSoftHeavy    []rl.Sound
	fxImpactSoftMedium   []rl.Sound
	fxImpactGenericLight []rl.Sound
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
	DirtDSR DirtStoneRockState = iota
	RockDSR
	StoneDSR
	FloorDetailTileDSR

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
		case FloorDetailTileDSR:
			dirtStoneRockModels[i] = common.Model.OBJ.FloorDetail
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

	// - Avoid spawning where player is standing
	// - Randomly skip a position
	// - A noise map or simplex/perlin noise "can" serve better
	InitMiningObjectPositions := func() []rl.Vector3 {
		var a []rl.Vector3 // 61% of maxPositions
		var (
			maxPositions   = floor.Size.X * floor.Size.Z
			maxSkipPosOdds = int32(3)
			y              = (floor.BoundingBox.Min.Y + floor.BoundingBox.Max.Y) / 2.0
		)
		offsetFromPlayerPos := float32(5.) // FIXME: This won't work if player is not at (0,0,0)
	NextCol:
		for x := floor.BoundingBox.Min.X + 1; x < floor.BoundingBox.Max.X; x++ {
		NextRow:
			for z := floor.BoundingBox.Min.Z + 1; z < floor.BoundingBox.Max.Z; z++ {
				if rl.GetRandomValue(0, maxSkipPosOdds) == 0 {
					continue NextRow
				}
				if len(a) >= int(maxPositions) {
					break NextCol
				}
				for i := -offsetFromPlayerPos; i <= offsetFromPlayerPos; i++ {
					for k := -offsetFromPlayerPos; k <= offsetFromPlayerPos; k++ {
						if i == x && k == z {
							continue NextRow
						}
					}
				}
				a = append(a, rl.NewVector3(x, y, z))
			}
		}
		return a
	}
	InitDirtStoneRockObjects(InitMiningObjectPositions())

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

	sfxDir := filepath.Join("res", "fx", "kenney_impact-sounds", "Audio")
	fxImpactSoftHeavy = []rl.Sound{
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_heavy_000.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_heavy_001.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_heavy_002.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_heavy_003.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_heavy_004.ogg")),
	}
	fxImpactSoftMedium = []rl.Sound{
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_medium_000.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_medium_001.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_medium_002.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_medium_003.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactSoft_medium_004.ogg")),
	}
	fxImpactGenericLight = []rl.Sound{
		rl.LoadSound(filepath.Join(sfxDir, "impactGeneric_light_000.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactGeneric_light_001.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactGeneric_light_002.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactGeneric_light_003.ogg")),
		rl.LoadSound(filepath.Join(sfxDir, "impactGeneric_light_004.ogg")),
	}

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
	for i := range dirtStoneRockCount {
		// Skip final mined object residue
		if dirtStoneRockArray[i].State == maxDirtStoneRockStates-1 {
			continue
		}
		if rl.CheckCollisionBoxes(
			common.GetBoundingBoxFromPositionSizeV(dirtStoneRockArray[i].Pos, dirtStoneRockArray[i].Size),
			player.BoundingBox,
		) {
			RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)

			// Trigger once while mining
			if rl.IsKeyPressed(rl.KeySpace) {
				// Play mining sound with variations (s1:kick + s2:snare + s3:hollow-thock)
				state := dirtStoneRockArray[i].State
				s1 := fxImpactSoftMedium[rl.GetRandomValue(int32(state), int32(len(fxImpactSoftMedium)-1))]
				s2 := fxImpactGenericLight[rl.GetRandomValue(int32(state), int32(len(fxImpactGenericLight)-1))]
				s3 := fxImpactSoftHeavy[rl.GetRandomValue(int32(state), int32(len(fxImpactSoftHeavy)-1))]
				rl.SetSoundVolume(s1, float32(rl.GetRandomValue(7, 10))/10.)
				rl.SetSoundVolume(s2, float32(rl.GetRandomValue(4, 8))/10.)
				rl.SetSoundVolume(s3, float32(rl.GetRandomValue(1, 4))/10.)
				rl.PlaySound(s1)
				rl.PlaySound(s2)
				rl.PlaySound(s3)

				// Increment state
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
		if true {
			pos := floor.Position
			size := rl.Vector3Multiply(floor.Size, rl.NewVector3(.21, 1., .21))
			DrawWalls(pos, size)
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
