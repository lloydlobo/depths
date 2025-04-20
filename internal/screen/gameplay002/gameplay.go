package gameplay

import (
	"cmp"
	"fmt"
	"log"
	"log/slog"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	"example/depths/internal/light"
)

var (
	framesCounter int32
	finishScreen  int

	camera rl.Camera3D

	player                Player
	floor                 Floor
	isPlayerWallCollision bool

	textureTilingLoc     int32
	emissiveColorLoc     int32
	emissiveIntensityLoc int32
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

	boxLargeModel         rl.Model
	boxLargePositions     []rl.Vector3
	boxLargeBoundingBoxes []rl.BoundingBox
	boxLargeSize          rl.Vector3
	boxLargeCount         int32
)

const (
	GoldChestEmpty GoldChestState = iota
	GoldChestFull
	// GoldChestFullClosed
	// GoldChestEmptyClosed

	maxGoldChestStates
)

var (
	goldChests      []GoldChest
	goldChestCount  int32
	goldChestModels [maxGoldChestStates]rl.Model
)

type GoldChestState uint8
type GoldChest struct {
	Pos      rl.Vector3
	Size     rl.Vector3
	Health   float32 // [0..1]
	State    GoldChestState
	IsActive bool
}

func NewGoldChest(pos, size rl.Vector3) GoldChest {
	return GoldChest{
		Pos:      pos,
		Size:     size,
		State:    GoldChestFull,
		IsActive: true,
	}
}

func InitGoldChests() {
	positions := []rl.Vector3{
		rl.NewVector3(2, 0, -12),
		rl.NewVector3(-6, 0, -6),
		rl.NewVector3(-8, 0, 12),
		rl.NewVector3(-9, 0, -4),
	}
	size := rl.NewVector3(1, 1, 1)
	for i := range positions {
		goldChests = append(goldChests, NewGoldChest(positions[i], size))
		goldChestCount++
	}
	for i := range maxGoldChestStates {
		// dir := filepath.Join("res", "model", "obj")
		switch i {
		case GoldChestEmpty:
			// goldChestModels[i] = rl.LoadModel(filepath.Join(dir, "chest.obj"))

			goldChestModels[i] = common.Model.OBJ.Stones
			rl.SetMaterialTexture(goldChestModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
		case GoldChestFull:
			// goldChestModels[i] = rl.LoadModel(filepath.Join(dir, "chest_gold.obj"))
			goldChestModels[i] = common.Model.OBJ.Chest
			rl.SetMaterialTexture(goldChestModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
		default:
			panic(fmt.Sprintf("unexpected gameplay.GoldChestState: %#v", i))
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
		Target:     rl.NewVector3(0., (1+.5)-.5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       15. * float32(cmp.Or(4., 3., 2.)),
		Projection: rl.CameraPerspective,
	} // See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h

	// These props/tiles/objects share the same dungeon texture
	// dungeonTexture = rl.LoadTexture(filepath.Join("res", "texture", "dungeon_texture.png"))
	{
		InitFloor()
		InitWall()
		InitGoldChests()
	}

	player = NewPlayer()

	isPlayerWallCollision = false

	rl.SetMusicVolume(common.Music.Theme, float32(cmp.Or(1.0, 0.125)))

	rl.PlayMusicStream(common.Music.Theme)

	rl.DisableCursor() // for ThirdPersonPerspective

	// Get location for shader parameters that can be modified in real time
	emissiveIntensityLoc = rl.GetShaderLocation(common.Shader.PBR, "emissivePower")
	emissiveColorLoc = rl.GetShaderLocation(common.Shader.PBR, "emissiveColor")
	textureTilingLoc = rl.GetShaderLocation(common.Shader.PBR, "tiling")

	{ // KayKit_DungeonRemastered_1.1_FREE/Assets/textures/dungeon_texture.png
		// boxLargeModel = rl.LoadModel(filepath.Join("res", "model", "obj", "box_large.obj"))
		boxLargeModel = common.Model.OBJ.Barrel
		rl.SetMaterialTexture(boxLargeModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
	}

	boxLargeSize = rl.NewVector3(1.5, 1.5, 1.5)
	for _, pos := range []rl.Vector3{
		rl.NewVector3(-5, 0, -8),
		rl.NewVector3(-3, 0, -7),
		rl.NewVector3(4, 0, 7),
		rl.NewVector3(5, 0, -4),
	} {
		boxLargePositions = append(boxLargePositions, pos)
		boxLargeBoundingBoxes = append(boxLargeBoundingBoxes, common.GetBoundingBoxFromPositionSizeV(pos, boxLargeSize))
		boxLargeCount++
	}
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

	if false {
		// Check key inputs to enable/disable lights
		if rl.IsKeyPressed(rl.KeyOne) {
			light.Lights[2].EnabledBinary = light.GetToggledEnabledBinary(2)
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			light.Lights[1].EnabledBinary = light.GetToggledEnabledBinary(1)
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			light.Lights[3].EnabledBinary = light.GetToggledEnabledBinary(3)
		}
		if rl.IsKeyPressed(rl.KeyFour) {
			light.Lights[0].EnabledBinary = light.GetToggledEnabledBinary(0)
		}
	}
}

func Update() {
	rl.UpdateMusicStream(common.Music.Theme)

	HandleUserInput()

	// Save variables this frame
	oldCam := camera
	oldPlayer := player

	dt := rl.GetFrameTime()
	slog.Debug("Update", "dt", dt)

	// Reset single frame flags/variables
	player.Collisions = rl.Quaternion{}
	isPlayerWallCollision = false

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	// Update the shader with the camera view vector (points towards { 0.0f, 0.0f, 0.0f })
	camPos := [3]float32{oldCam.Position.X, oldCam.Position.Y, oldCam.Position.Z}
	rl.SetShaderValue(common.Shader.PBR, common.Shader.PBR.GetLocation(rl.ShaderLocVectorView), camPos[:], rl.ShaderUniformVec3)

	if false {
		// Update light values on shader (actually, only enable/disable them)
		for i := range light.MaxLights {
			light.UpdateLight(common.Shader.PBR, light.Lights[i])
		}
	}

	player.Update()

	// Wall to Player collisions
	if isPlayerWallCollision {
		RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)
	}

	// LargeBox to Player collisions
	for i := range boxLargeCount {
		if rl.CheckCollisionBoxes(boxLargeBoundingBoxes[i], player.BoundingBox) {
			RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)
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

	player.Draw()
	floor.Draw()
	DrawWalls()
	for _, pos := range boxLargePositions { // Draw offgrid tiles
		rl.DrawModel(boxLargeModel, pos, 1., rl.White)
	}
	for i := range goldChestCount {
		chest := goldChests[i]
		rl.DrawModel(goldChestModels[chest.State], chest.Pos, 1.0, rl.White)
	}

	cubeSize := rl.NewVector3(.9, .7, .9)
	rl.DrawCubeV(rl.NewVector3(4, 1, 4), cubeSize, rl.Brown)
	rl.DrawCubeV(rl.NewVector3(5, 1, 4), cubeSize, rl.Brown)
	rl.DrawCubeV(rl.NewVector3(5, 1, 5), cubeSize, rl.DarkBrown)

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
