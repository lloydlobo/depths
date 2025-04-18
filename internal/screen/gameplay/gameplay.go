package gameplay

import (
	"cmp"
	"fmt"
	"log"
	"log/slog"
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

	common.Shader.PBR = rl.LoadShader(
		filepath.Join("res", "shader", "glsl330_"+"pbr.vs"),
		filepath.Join("res", "shader", "glsl330_"+"pbr.fs"))

	// Load cubeTexture to be applied to the cubes sides (256x256 png)
	cubeTexture = rl.LoadTexture("res/cubicmap_atlas.png")

	InitFloor() // Init floor and friends

	player = NewPlayer()

	isPlayerWallCollision = false

	rl.SetMusicVolume(common.Music.Theme, float32(cmp.Or(0.125, 1.0)))

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

	dt := rl.GetFrameTime()
	_ = dt
	slog.Debug("Update", "dt", dt)

	rl.UpdateMusicStream(common.Music.Theme)

	// Save variables this frame
	oldCam := camera
	oldPlayer := player

	// Reset single frame flags/variables
	player.Collisions = rl.Quaternion{}
	isPlayerWallCollision = false

	rl.UpdateCamera(&camera, rl.CameraThirdPerson)

	player.Update()

	if isPlayerWallCollision {
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
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	// 3D World
	rl.BeginMode3D(camera)

	rl.ClearBackground(rl.Black)

	floor.Draw()
	player.Draw()

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
	// Draw cube with an applied texture
	vec := rl.Vector3{X: -2.0, Y: 2.0}
	DrawCubeTexture(cubeTexture, vec, 2.0, 4.0, 2.0, rl.White)

	// Draw cube with an applied texture, but only a defined rectangle piece of the texture
	rec := rl.Rectangle{
		Y:      float32(cubeTexture.Height) / 2.0,
		Width:  float32(cubeTexture.Width) / 2.0,
		Height: float32(cubeTexture.Height) / 2.0,
	}
	vec = rl.Vector3{X: 2.0, Y: 1.0}
	DrawCubeTextureRec(cubeTexture, rec, vec, 2.0, 2.0, 2.0, rl.White)
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
	rl.PushMatrix()
	// NOTE: Transformation is applied in inverse order (scale -> rotate -> translate)
	rl.Translatef(2.0, 0.0, 0.0)
	rl.Rotatef(45, 0, 1, 0)
	rl.Scalef(2.0, 2.0, 2.0)
	{
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
	}
	rl.PopMatrix()

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
