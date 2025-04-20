package gameplay

import (
	"cmp"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"

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
//
//	TEMPORARY

var (
	floorRoadPBRModel   rl.Model
	floorTileLargeModel rl.Model

	wallModel       rl.Model
	wallCornerModel rl.Model

	boxLargeModel         rl.Model
	boxLargePositions     []rl.Vector3
	boxLargeBoundingBoxes []rl.BoundingBox
	boxLargeSize          = rl.NewVector3(1.5, 1.5, 1.5)
)

// 	TEMPORARY
//
// 	TEMPORARY

func Init() {
	framesCounter = 0
	finishScreen = 0

	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 30., 30.),
		Target:     rl.NewVector3(0., (1+0.5)-0.5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	} // See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h

	InitFloor() // Init floor and friends

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
		dungeonTexture := rl.LoadTexture(filepath.Join("res", "texture", "dungeon_texture.png"))
		floorTileLargeModel = rl.LoadModel(filepath.Join("res", "model", "obj", "floor_tile_large.obj"))
		rl.SetMaterialTexture(floorTileLargeModel.Materials, rl.MapDiffuse, dungeonTexture)

		wallModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall.obj"))
		rl.SetMaterialTexture(wallModel.Materials, rl.MapDiffuse, dungeonTexture)

		wallCornerModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall_corner.obj"))
		rl.SetMaterialTexture(wallCornerModel.Materials, rl.MapDiffuse, dungeonTexture)

		boxLargeModel = rl.LoadModel(filepath.Join("res", "model", "obj", "box_large.obj"))
		rl.SetMaterialTexture(boxLargeModel.Materials, rl.MapDiffuse, dungeonTexture)
	}

	for _, pos := range []rl.Vector3{
		rl.NewVector3(-5, 0, -8),
		rl.NewVector3(-3, 0, -7),
		rl.NewVector3(4, 0, 7),
		rl.NewVector3(5, 0, -4),
	} {
		boxLargePositions = append(boxLargePositions, pos)
		boxLargeBoundingBoxes = append(boxLargeBoundingBoxes, common.GetBoundingBoxFromPositionSizeV(pos, boxLargeSize))
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

	// Handle collisions

	// Update player to wall objects
	if isPlayerWallCollision {
		RevertPlayerAndCameraPositions(oldPlayer, &player, oldCam, &camera)
	}

	// Update player to props objects
	for _, bb := range boxLargeBoundingBoxes {
		if rl.CheckCollisionBoxes(player.BoundingBox, bb) {
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

	bgCol := cmp.Or(
		rl.Black,
		rl.ColorBrightness(rl.DarkPurple, -.9),
		rl.Gray,
		rl.RayWhite,
	)
	rl.ClearBackground(bgCol)

	player.Draw()

	// Draw floor
	if false {
		floor.Draw()
	} else {
		const floorModelScale = 4
		for x := float32(floor.BoundingBox.Min.X); x < float32(floor.BoundingBox.Max.X); x += floorModelScale {
			for z := float32(floor.BoundingBox.Min.Z); z < float32(floor.BoundingBox.Max.Z); z += floorModelScale {
				centerX, centerY, centerZ := x+(floorModelScale/4.)*2., float32(0.), z+(floorModelScale/4.)*2.
				rl.DrawModel(floorTileLargeModel, rl.NewVector3(centerX, centerY, centerZ), floorModelScale/4., rl.White)
			}
		}
	}
	if false {
		rl.DrawBoundingBox(floor.BoundingBox, rl.Purple)
	}
	DrawXYZOrbitV(rl.Vector3Zero(), 2.)
	DrawWorldXYZAxis()

	// Draw walls
	const wallLen = 4.
	const wallThick = 1. / 2.
	const wallBotY = 1. / 4.
	wallTint := rl.White
	wallPos := floor.Position
	floorSize := floor.Size
	for i := -float32(floorSize.Z/2) + wallLen/2; i < float32(floorSize.Z/2); i += wallLen { // Along Z axis
		pos1 := rl.NewVector3(
			wallPos.X-floorSize.X/2-wallThick,
			wallPos.Y+wallBotY,
			wallPos.Z+i)
		pos2 := rl.NewVector3(
			wallPos.X+floorSize.X/2+wallThick,
			wallPos.Y+wallBotY,
			wallPos.Z+i)
		rl.DrawModelEx(wallModel, pos1, rl.NewVector3(0, 1, 0), 90, common.Vector3One, wallTint) // -X +-Z
		rl.DrawModelEx(wallModel, pos2, rl.NewVector3(0, 1, 0), 90, common.Vector3One, wallTint) // +X +-Z
	}
	for i := -float32(floorSize.X/2) + wallLen/2; i < float32(floorSize.X/2); i += wallLen { // Along X axis
		pos1 := rl.NewVector3(
			wallPos.X-i,
			wallPos.Y+wallBotY,
			wallPos.Z-floorSize.Z/2-wallThick)
		pos2 := rl.NewVector3(
			wallPos.X+i,
			wallPos.Y+wallBotY,
			wallPos.Z+floorSize.Z/2+wallThick)
		rl.DrawModelEx(wallModel, pos1, rl.NewVector3(0, 1, 0), 180, common.Vector3One, wallTint) // +-X -Z
		rl.DrawModelEx(wallModel, pos2, rl.NewVector3(0, 1, 0), 180, common.Vector3One, wallTint) // +-X +Z
	}
	bottomLeft := rl.NewVector3(wallPos.X-floorSize.X/2-wallThick, wallPos.Y+wallBotY, wallPos.Z+floorSize.Z/2+wallThick)
	bottomRight := rl.NewVector3(wallPos.X+floorSize.X/2+wallThick, wallPos.Y+wallBotY, wallPos.Z+floorSize.Z/2+wallThick)
	topRight := rl.NewVector3(wallPos.X+floorSize.X/2+wallThick, wallPos.Y+wallBotY, wallPos.Z-floorSize.Z/2-wallThick)
	topLeft := rl.NewVector3(wallPos.X-floorSize.X/2-wallThick, wallPos.Y+wallBotY, wallPos.Z-floorSize.Z/2-wallThick)
	rl.DrawModelEx(wallCornerModel, topRight, rl.NewVector3(0, 1, 0), 0, common.Vector3One, wallTint)
	rl.DrawModelEx(wallCornerModel, topLeft, rl.NewVector3(0, 1, 0), 90, common.Vector3One, wallTint)
	rl.DrawModelEx(wallCornerModel, bottomLeft, rl.NewVector3(0, 1, 0), 180, common.Vector3One, wallTint)
	rl.DrawModelEx(wallCornerModel, bottomRight, rl.NewVector3(0, 1, 0), 270, common.Vector3One, wallTint)

	// Draw offgrid tiles
	for _, pos := range boxLargePositions {
		rl.DrawModel(boxLargeModel, pos, 1., rl.White)
	}

	if false {
		for i := range light.MaxLights {
			lightColor := rl.NewColor(
				uint8(light.Lights[i].Color[0]*255),
				uint8(light.Lights[i].Color[1]*255),
				uint8(light.Lights[i].Color[2]*255),
				uint8(light.Lights[i].Color[3]*255))
			if light.Lights[i].EnabledBinary == 1 {
				rl.DrawSphereEx(light.Lights[i].Position, .2, 8, 8, lightColor)
			} else {
				rl.DrawSphereWires(light.Lights[i].Position, .2, 8, 8, rl.Fade(lightColor, .3))
			}
		}
	}
	if false {
		DrawCubicmaps()
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
