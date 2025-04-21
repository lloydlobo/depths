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

	fxSoftHeavyImpacts    []rl.Sound
	fxSoftMediumImpacts   []rl.Sound
	fxGenericLightImpacts []rl.Sound

	fxFootsteps []rl.Sound
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
	Rotn     float32
	Health   float32 // [0..1]
	State    DirtStoneRockState
	IsActive bool
}

func NewDirtStoneRockObj(pos, size rl.Vector3) DirtStoneRockObj {
	return DirtStoneRockObj{
		Pos:      pos,
		Size:     size,
		Rotn:     0.0,
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
	for i := range positions {
		size := rl.Vector3Multiply(
			rl.NewVector3(1, 1, 1),
			rl.NewVector3(
				float32(rl.GetRandomValue(94, 98))/100.,
				float32(rl.GetRandomValue(60, 62))/100.,
				float32(rl.GetRandomValue(94, 98))/100.))
		obj := NewDirtStoneRockObj(positions[i], size)
		obj.Rotn = cmp.Or(float32(rl.GetRandomValue(-30, 30)/10.), 0.)
		dirtStoneRockArray = append(dirtStoneRockArray, obj)
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

	var (
		fxAudioDir = filepath.Join("res", "fx", "kenney_impact-sounds", "Audio")
	)
	fxFootsteps = []rl.Sound{
		rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_000.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_001.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_002.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_003.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_004.ogg")),
	}
	fxSoftHeavyImpacts = []rl.Sound{
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_000.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_001.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_002.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_003.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_004.ogg")),
	}
	fxSoftMediumImpacts = []rl.Sound{
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_000.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_001.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_002.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_003.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_004.ogg")),
	}
	fxGenericLightImpacts = []rl.Sound{
		rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_000.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_001.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_002.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_003.ogg")),
		rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_004.ogg")),
	}

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
		offsetFromPlayerPos := float32(4.) // FIXME: This won't work if player is not at (0,0,0)
	NextCol:
		for x := floor.BoundingBox.Min.X + 1; x < floor.BoundingBox.Max.X; x++ {
		NextRow:
			for z := floor.BoundingBox.Min.Z + 1; z < floor.BoundingBox.Max.Z; z++ {
				for i := -offsetFromPlayerPos; i <= offsetFromPlayerPos; i++ {
					for k := -offsetFromPlayerPos; k <= offsetFromPlayerPos; k++ {
						if i == x && k == z {
							continue NextRow
						}
					}
				}
				if rl.GetRandomValue(0, maxSkipPosOdds) == 0 {
					continue NextRow
				}
				if len(a) >= int(maxPositions) {
					break NextCol
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
				state := dirtStoneRockArray[i].State
				s1 := fxSoftMediumImpacts[rl.GetRandomValue(int32(state), int32(len(fxSoftMediumImpacts)-1))]
				s2 := fxGenericLightImpacts[rl.GetRandomValue(int32(state), int32(len(fxGenericLightImpacts)-1))]
				s3 := fxSoftHeavyImpacts[rl.GetRandomValue(int32(state), int32(len(fxSoftHeavyImpacts)-1))]
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
				rl.PlaySound(fxFootsteps[int(framesCounter)%len(fxFootsteps)])
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
		DrawWalls(
			floor.Position,
			floor.Size,
			rl.NewVector3(1., cmp.Or(float32(1.), common.OneMinusInvPhi), 1.),
		) // Use walls to avoid infinite-map generation

		// Draw drill
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

		for i := range dirtStoneRockCount {
			obj := dirtStoneRockArray[i]
			rl.DrawModelEx(dirtStoneRockModels[obj.State], obj.Pos,
				rl.NewVector3(0, 1, 0), obj.Rotn, obj.Size, rl.White)
		}

		rl.DrawModel(checkedModel, rl.NewVector3(0., -.05, 0.), 1., rl.RayWhite)

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
