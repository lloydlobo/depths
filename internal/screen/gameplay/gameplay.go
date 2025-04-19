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

	wallModel rl.Model
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

	rl.SetMusicVolume(common.Music.Theme, float32(cmp.Or(0.125, 1.0)))

	rl.PlayMusicStream(common.Music.Theme)

	rl.DisableCursor() // for ThirdPersonPerspective

	// Get location for shader parameters that can be modified in real time
	emissiveIntensityLoc = rl.GetShaderLocation(common.Shader.PBR, "emissivePower")
	emissiveColorLoc = rl.GetShaderLocation(common.Shader.PBR, "emissiveColor")
	textureTilingLoc = rl.GetShaderLocation(common.Shader.PBR, "tiling")

	{
		// KayKit_DungeonRemastered_1.1_FREE/Assets/textures/dungeon_texture.png
		dungeonTexture := rl.LoadTexture(filepath.Join("res", "texture", "dungeon_texture.png"))
		floorTileLargeModel = rl.LoadModel(filepath.Join("res", "model", "obj", "floor_tile_large.obj"))
		rl.SetMaterialTexture(floorTileLargeModel.Materials, rl.MapDiffuse, dungeonTexture)

		// KayKit_DungeonRemastered_1.1_FREE/Assets/textures/dungeon_texture.png
		wallModel = rl.LoadModel(filepath.Join("res", "model", "obj", "wall.obj"))
		rl.SetMaterialTexture(wallModel.Materials, rl.MapDiffuse, dungeonTexture)
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
	HandleUserInput()
	rl.UpdateMusicStream(common.Music.Theme)

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

	if isPlayerWallCollision {
		player.Position = oldPlayer.Position
		player.BoundingBox = rl.NewBoundingBox(
			rl.NewVector3(player.Position.X-player.Size.X/2, player.Position.Y-player.Size.Y/2, player.Position.Z-player.Size.Z/2),
			rl.NewVector3(player.Position.X+player.Size.X/2, player.Position.Y+player.Size.Y/2, player.Position.Z+player.Size.Z/2))
		camera.Target = oldCam.Target
		camera.Position = oldCam.Position
	}

	framesCounter++
}

func Draw() {
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	bgCol := cmp.Or(rl.ColorBrightness(rl.DarkPurple, -.9), rl.Black, rl.RayWhite)
	rl.ClearBackground(bgCol)

	player.Draw()

	// Draw floor
	{
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
	}

	// Draw walls
	const wallLen = 4
	const wallThick = 1. / 4.
	wallTint := rl.White
	wallPos := floor.Position
	floorSize := floor.Size

	for i := -float32(floorSize.Z/2) + wallLen/2; i < float32(floorSize.Z/2); i += wallLen { // Along Z axis
		rl.DrawModelEx(wallModel, rl.NewVector3(wallPos.X-floorSize.X/2-wallThick, wallPos.Y, wallPos.Z+i),
			rl.NewVector3(0, 1, 0), 90, common.Vector3One, wallTint)
		rl.DrawModelEx(wallModel, rl.NewVector3(wallPos.X+floorSize.X/2+wallThick, wallPos.Y, wallPos.Z+i),
			rl.NewVector3(0, 1, 0), 90, common.Vector3One, wallTint)
	}
	for i := -float32(floorSize.X/2) + wallLen/2; i < float32(floorSize.X/2); i += wallLen { // Along X axis
		rl.DrawModelEx(wallModel, rl.NewVector3(wallPos.X-i, wallPos.Y, wallPos.Z-floorSize.X/2-wallThick),
			rl.NewVector3(0, 1, 0), 180, common.Vector3One, wallTint)
		rl.DrawModelEx(wallModel, rl.NewVector3(wallPos.X+i, wallPos.Y, wallPos.Z+floorSize.X/2+wallThick),
			rl.NewVector3(0, 1, 0), 180, common.Vector3One, wallTint)
	}

	// rl.DrawModelEx(wallModel, rl.NewVector3(wallPos.X, wallPos.Y, wallPos.Z+floorSize.Z/2), rl.NewVector3(0, 1, 0), 180, common.Vector3One, common.ZAxisColor)
	// rl.DrawModelEx(wallModel, rl.NewVector3(wallPos.X, wallPos.Y, wallPos.Z-floorSize.Z/2), rl.NewVector3(0, 1, 0), 180, common.Vector3One, common.ZAxisColor)

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

	if true {
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
