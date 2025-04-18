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

	player          Player
	floor           Floor
	isWallCollision bool
)

var (
	xCol = rl.Fade(rl.Red, .3)
	yCol = rl.Fade(rl.Green, .3)
	zCol = rl.Fade(rl.Green, .3)
)

// TEMPORARY
//
//	TEMPORARY

var (
	// 							TEMPORARY
	// 							TEMPORARY
	cubeTexture rl.Texture2D
	floorModel  rl.Model
	//							TEMPORARY
	//							TEMPORARY
)

// 	TEMPORARY
//
// 	TEMPORARY

func Init() {
	framesCounter = 0
	finishScreen = 0

	// See also https://github.com/raylib-extras/extras-c/blob/main/cameras/rlTPCamera/rlTPCamera.h
	camera = rl.Camera3D{
		Position:   rl.NewVector3(0., 30., 30.),
		Target:     rl.NewVector3(0., 1+0.5, 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       45.0,
		Projection: rl.CameraPerspective,
	}

	common.Shader.PBR = rl.LoadShader(filepath.Join("res", "shader", "glsl330_"+"pbr.vs"), filepath.Join("res", "shader", "glsl330_"+"pbr.fs"))

	player = Player{
		Position:   camera.Target,
		Size:       rl.NewVector3(1, 2, 1),
		Collisions: rl.NewQuaternion(0, 0, 0, 0),
	}
	player.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(camera.Target.X-player.Size.X/2, camera.Target.Y-player.Size.Y/2, camera.Target.Z-player.Size.Z/2),
		rl.NewVector3(camera.Target.X+player.Size.X/2, camera.Target.Y+player.Size.Y/2, camera.Target.Z+player.Size.Z/2))

	floor = Floor{Position: rl.NewVector3(0, -0.5, 0), Size: rl.NewVector3(30, 1, 30)} // Load floor model mesh and assign material parameters
	floor.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(floor.Position.X-floor.Size.X/2, floor.Position.Y-floor.Size.Y/2, floor.Position.Z-floor.Size.Z/2),
		rl.NewVector3(floor.Position.X+floor.Size.X/2, floor.Position.Y+floor.Size.Y/2, floor.Position.Z+floor.Size.Z/2))
	floorMesh := rl.GenMeshPlane(floor.Size.X/5, floor.Size.Z/5, 10, 10) // [Scale up to 5x on draw] NOTE: A basic plane shape can be generated instead of being loaded from a model file
	floorModel = rl.LoadModelFromMesh(floorMesh)                         // rl.GenMeshTangents(&floorMesh) // TODO: Review tangents generation
	if isEnableLight := false; isEnableLight {
		floorModel.Materials.Shader = common.Shader.PBR
		floorModel.Materials.GetMap(rl.MapAlbedo).Color = rl.White
		floorModel.Materials.GetMap(rl.MapMetalness).Value = 0.0
		floorModel.Materials.GetMap(rl.MapRoughness).Value = 0.0
		floorModel.Materials.GetMap(rl.MapOcclusion).Value = 1.0
		floorModel.Materials.GetMap(rl.MapEmission).Color = rl.Black
	}
	floorModel.Materials.GetMap(rl.MapAlbedo).Texture = rl.LoadTexture(filepath.Join("res", "texture", "road_a.png"))
	floorModel.Materials.GetMap(rl.MapMetalness).Texture = rl.LoadTexture(filepath.Join("res", "texture", "road_mra.png"))
	floorModel.Materials.GetMap(rl.MapNormal).Texture = rl.LoadTexture(filepath.Join("res", "texture", "road_n.png"))

	isWallCollision = false

	rl.SetMusicVolume(common.Music.Theme, float32(cmp.Or(0.125, 1.0)))
	rl.PlayMusicStream(common.Music.Theme)
	cubeTexture = rl.LoadTexture("res/cubicmap_atlas.png") // 256x256 Load cubeTexture to be applied to the cubes sides

	rl.DisableCursor() // for ThirdPersonPerspective
}

func Update() {
	dt := rl.GetFrameTime()
	_ = dt

	rl.UpdateMusicStream(common.Music.Theme)

	// Press enter or tap to change to ending game screen
	if rl.IsKeyDown(rl.KeyF10) { /* || rl.IsGestureDetected(rl.GestureDrag) */
		finishScreen = 1
		rl.PlaySound(common.FX.Coin)
	}

	// Pick up item
	if rl.IsKeyDown(rl.KeyF) {
		log.Println("Picked up ...")
	}

	// Save variables this frame
	oldCam := camera
	oldPlayer := player

	// Reset single frame flags/variables
	player.Collisions = rl.Quaternion{}
	isWallCollision = false

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	player.Update()

	if isWallCollision {
		player.Position = oldPlayer.Position
		player.BoundingBox = rl.NewBoundingBox(
			rl.NewVector3(player.Position.X-player.Size.X/2,
				player.Position.Y-player.Size.Y/2, player.Position.Z-player.Size.Z/2),
			rl.NewVector3(player.Position.X+player.Size.X/2,
				player.Position.Y+player.Size.Y/2, player.Position.Z+player.Size.Z/2))
		camera.Target = oldCam.Target
		camera.Position = oldCam.Position
	}

	framesCounter++
}

func Draw() {
	rl.BeginMode3D(camera)
	rl.ClearBackground(rl.Black)

	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	player.Draw()

	// Set floor model texture tiling and emissive color parameters on shader
	if isEnableLight := false; isEnableLight {
		textureTilingLoc := rl.GetShaderLocation(floorModel.Materials.Shader, "tiling")
		emissiveColorLoc := rl.GetShaderLocation(floorModel.Materials.Shader, "emissiveColor")
		floorTextureTiling := []float32{.5, .5}
		rl.SetShaderValue(floorModel.Materials.Shader, textureTilingLoc, floorTextureTiling, rl.ShaderUniformVec2)

		floorEmissiveColorVector4 := rl.ColorNormalize(floorModel.Materials.GetMap(rl.MapEmission).Color)
		floorEmissiveColor := []float32{floorEmissiveColorVector4.X, floorEmissiveColorVector4.Y, floorEmissiveColorVector4.Z, floorEmissiveColorVector4.W}
		rl.SetShaderValue(floorModel.Materials.Shader, emissiveColorLoc, floorEmissiveColor, rl.ShaderUniformVec2)
	}
	rl.DrawModel(floorModel, rl.Vector3Zero(), 5.0, rl.Beige) // Draw floor model

	// rl.DrawCubeV(floor.Position, floor.Size, rl.DarkBrown)
	rl.DrawBoundingBox(floor.BoundingBox, rl.Brown)
	DrawXYZOrbitV(rl.Vector3Zero(), 2.)
	DrawWorldXYZAxis()

	rl.EndMode3D()

	// 2D HUD
	fontThatIsInGameDotGo := rl.GetFontDefault()
	fontSize := float32(fontThatIsInGameDotGo.BaseSize) * 3.0

	text := "[F] PICK UP"
	rl.DrawText(text, screenW/2-rl.MeasureText(text, 20)/2, screenH-20*2, 20, rl.White)

	rl.DrawFPS(10, 10)
	rl.DrawTextEx(fontThatIsInGameDotGo, fmt.Sprintf("%.6f", rl.GetFrameTime()),
		rl.NewVector2(10, 10+20*1), fontSize*2./3., 1, rl.Lime)
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

type Floor struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
}

type Player struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
	Collisions  rl.Quaternion
}

func (p *Player) Update() {
	// Project the player as the camera target
	p.Position = camera.Target

	p.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(p.Position.X-p.Size.X/2, p.Position.Y-p.Size.Y/2, p.Position.Z-p.Size.Z/2),
		rl.NewVector3(p.Position.X+p.Size.X/2, p.Position.Y+p.Size.Y/2, p.Position.Z+p.Size.Z/2))

	// Wall collisions
	if p.BoundingBox.Min.X <= floor.BoundingBox.Min.X {
		isWallCollision = true
		p.Collisions.X = -1
	}
	if p.BoundingBox.Max.X >= floor.BoundingBox.Max.X {
		isWallCollision = true
		p.Collisions.X = 1
	}
	if p.BoundingBox.Min.Z <= floor.BoundingBox.Min.Z {
		isWallCollision = true
		p.Collisions.Z = -1
	}
	if p.BoundingBox.Max.Z >= floor.BoundingBox.Max.Z {
		isWallCollision = true
		p.Collisions.Z = 1
	}

	// Floor collisions
	if p.BoundingBox.Min.Y <= floor.BoundingBox.Min.Y {
		p.Collisions.Y = 1 // Player head below floor
	}
	if p.BoundingBox.Max.Y >= floor.BoundingBox.Min.Y { // On floor
		p.Collisions.W = -1 // Allow walking freely
	}
}

func (p Player) Draw() {
	col := rl.Beige
	rl.DrawCapsule(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/4, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
		p.Size.X/2, 16, 16, col)
	rl.DrawCapsuleWires(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/4, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
		p.Size.X/2, 16, 16, col)
	rl.DrawCylinderWiresEx(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/2, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/2, 0)),
		p.Size.X/2, p.Size.X/2, 16, col)
	if isWallCollision {
		rl.DrawBoundingBox(p.BoundingBox, rl.Red)
	} else {
		rl.DrawBoundingBox(p.BoundingBox, rl.LightGray)
	}

	if p.Collisions.X != 0 {
		pos := p.Position
		pos.X += p.Collisions.X * p.Size.X / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), xCol)
	}
	if p.Collisions.Y != 0 {
		pos := p.Position
		pos.Y += p.Collisions.Y * p.Size.Y / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), yCol)
	}
	if p.Collisions.Z != 0 {
		pos := p.Position
		pos.Z += p.Collisions.Z * p.Size.Z / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), zCol)
	}
	if p.Collisions.W != 0 { // Floor
		pos := p.Position
		pos.Y += p.Collisions.W * p.Size.Y / 2
		rl.DrawCubeV(pos, rl.Vector3Scale(p.Size, .5), yCol)
	}

	DrawXYZOrbitV(p.Position, 1.)
	draw_cube_texture_main()
}

// DrawXYZOrbitV draws perpendicular 3D circles to all 3 (x y z) axis.
func DrawXYZOrbitV(pos rl.Vector3, radius float32) {
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, 1, 0), 90, xCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(1, 0, 0), 90, yCol)
	rl.DrawCircle3D(pos, radius, rl.NewVector3(0, -1, 0), 0, zCol)
}

// DrawWorldXYZAxis draws all 3 (x y z) axis intersecting at (0,0,0).
func DrawWorldXYZAxis() {
	rl.DrawLine3D(rl.NewVector3(500, 0, 0), rl.NewVector3(-500, 0, 0), xCol)
	rl.DrawLine3D(rl.NewVector3(0, 500, 0), rl.NewVector3(0, -500, 0), yCol)
	rl.DrawLine3D(rl.NewVector3(0, 0, 500), rl.NewVector3(0, 0, -500), zCol)
}

func draw_cube_texture_main() {
	// screenWidth := int32(rl.GetScreenWidth())
	// screenHeight := int32(rl.GetScreenHeight())

	// rl.InitWindow(screenWidth, screenHeight, "raylib [models] example - draw cube texture")
	//
	// // Define the camera to look into our 3d world
	// camera := rl.Camera{
	// 	Position: rl.Vector3{
	// 		Y: 10.0,
	// 		Z: 10.0,
	// 	},
	// 	Target:     rl.Vector3{},
	// 	Up:         rl.Vector3{Y: 1.0},
	// 	Fovy:       45.0,
	// 	Projection: rl.CameraPerspective,
	// }

	// rl.SetTargetFPS(60) // Set our game to run at 60 frames-per-second

	// Main game loop
	// for !rl.WindowShouldClose() { // Detect window close button or ESC key
	// Update
	// TODO: Update your variables here

	// Draw
	// rl.BeginDrawing()
	// rl.ClearBackground(rl.RayWhite)
	// rl.BeginMode3D(camera)

	// Draw cube with an applied texture
	vec := rl.Vector3{
		X: -2.0,
		Y: 2.0,
	}
	DrawCubeTexture(cubeTexture, vec, 2.0, 4.0, 2.0, rl.White)

	// Draw cube with an applied texture, but only a defined rectangle piece of the texture
	rec := rl.Rectangle{
		Y:      float32(cubeTexture.Height) / 2.0,
		Width:  float32(cubeTexture.Width) / 2.0,
		Height: float32(cubeTexture.Height) / 2.0,
	}
	vec = rl.Vector3{
		X: 2.0,
		Y: 1.0,
	}
	DrawCubeTextureRec(cubeTexture, rec, vec, 2.0, 2.0, 2.0, rl.White)

	// rl.DrawGrid(10, 1.0) // Draw a grid
	// rl.EndMode3D()
	// rl.DrawFPS(10, 10)
	// rl.EndDrawing()
	// }

	// De-Initialization
	// rl.UnloadTexture(texture) // Unload texture
	// rl.CloseWindow()          // Close window and OpenGL context
}

// DrawCubeTexture draws a textured cube
// NOTE: Cube position is the center position
func DrawCubeTexture(texture rl.Texture2D, position rl.Vector3, width, height, length float32, color rl.Color) {
	x := position.X
	y := position.Y
	z := position.Z

	// Set desired texture to be enabled while drawing following vertex data
	rl.SetTexture(texture.ID)

	// Vertex data transformation can be defined with the commented lines,
	// but in this example we calculate the transformed vertex data directly when calling rlVertex3f()
	// rl.PushMatrix()
	// NOTE: Transformation is applied in inverse order (scale -> rotate -> translate)
	//rl.Translatef(2.0, 0.0, 0.0)
	//rl.Rotatef(45, 0, 1, )
	//rl.Scalef(2.0, 2.0, 2.0)

	rl.Begin(rl.Quads)
	rl.Color4ub(color.R, color.G, color.B, color.A)
	// Front Face
	rl.Normal3f(0.0, 0.0, 1.0) // Normal Pointing Towards Viewer
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2) // Top Left Of The Texture and Quad
	// Back Face
	rl.Normal3f(0.0, 0.0, -1.0) // Normal Pointing Away From Viewer
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2) // Bottom Left Of The Texture and Quad
	// Top Face
	rl.Normal3f(0.0, 1.0, 0.0) // Normal Pointing Up
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad.
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad Bottom Face
	rl.Normal3f(0.0, -1.0, 0.0)                    // Normal Pointing Down
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2) // Top Left Of The Texture and Quad
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	// Right face
	rl.Normal3f(1.0, 0.0, 0.0) // Normal Pointing Right
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2) // Top Left Of The Texture and Quad
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	// Left Face
	rl.Normal3f(-1.0, 0.0, 0.0) // Normal Pointing Left
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad

	rl.End()
	//rl.PopMatrix()

	rl.SetTexture(0)
}

// DrawCubeTextureRec draws a cube with texture piece applied to all faces
func DrawCubeTextureRec(texture rl.Texture2D, source rl.Rectangle, position rl.Vector3, width, height,
	length float32, color rl.Color) {

	x := position.X
	y := position.Y
	z := position.Z

	texWidth := float32(texture.Width)
	texHeight := float32(texture.Height)

	// Set desired texture to be enabled while drawing following vertex data
	rl.SetTexture(texture.ID)

	// We calculate the normalized texture coordinates for the desired texture-source-rectangle
	// It means converting from (tex.width, tex.height) coordinates to [0.0f, 1.0f] equivalent
	rl.Begin(rl.Quads)
	rl.Color4ub(color.R, color.G, color.B, color.A)

	// Front face
	rl.Normal3f(0.0, 0.0, 1.0)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)

	// Back face
	rl.Normal3f(0.0, 0.0, -1.0)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)

	// Top face
	rl.Normal3f(0.0, 1.0, 0.0)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)

	// Bottom face
	rl.Normal3f(0.0, -1.0, 0.0)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)

	// Right face
	rl.Normal3f(1.0, 0.0, 0.0)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)

	// Left face
	rl.Normal3f(-1.0, 0.0, 0.0)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)

	rl.End()

	rl.SetTexture(0)
}
