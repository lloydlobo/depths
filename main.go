package main

import (
	"fmt"
	"log"
	"math"
	"os"

	_ "github.com/gen2brain/raylib-go/easings"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	run()
}

const MaxPostproShaders = 12

const (
	FxGrayscale = iota
	FxPosterization
	FxDreamVision
	FxPixelizer
	FxCrossHatching
	FxCrossStitching
	FxPredatorView
	FxScanlines
	FxFisheye
	FxSobel
	FxBloom
	FxBlur
)

var postproShaderText = []string{
	"GRAYSCALE",
	"POSTERIZATION",
	"DREAM_VISION",
	"PIXELIZER",
	"CROSS_HATCHING",
	"CROSS_STITCHING",
	"PREDATOR_VIEW",
	"SCANLINES",
	"FISHEYE",
	"SOBEL",
	"BLOOM",
	"BLUR",
}

func run() {
	fps := int32(60)
	screenWidth := int32(800)
	screenHeight := int32(450)
	arenaWorldWidth := float32(20)  // X
	arenaWorldLength := float32(20) // Z

	rl.SetConfigFlags(rl.FlagMsaa4xHint) // Enable Multi Sampling Anti Aliasing 4x (if available)

	rl.InitWindow(screenWidth, screenHeight, "raylib [models] example - box collisions")

	var glslVersion uint32
	if os.Getenv("PLATFORM_DESKTOP") != "" {
		glslVersion = 330
	} else { // PLATFORM_ANDROID, PLATFORM_WEB
		glslVersion = 100
	}
	log.Println("GLSL_VERSION", glslVersion)

	// Load all postpro shaders
	// NOTE 1: All postpro shader use the base vertex shader (DEFAULT_VERTEX_SHADER)
	shaders := make([]rl.Shader, MaxPostproShaders)
	shaders[FxGrayscale] = rl.LoadShader("", "res/glsl330_grayscale.fs")
	shaders[FxPosterization] = rl.LoadShader("", "res/glsl330_posterization.fs")
	shaders[FxDreamVision] = rl.LoadShader("", "res/glsl330_dream_vision.fs")
	shaders[FxPixelizer] = rl.LoadShader("", "res/glsl330_pixelizer.fs")
	shaders[FxCrossHatching] = rl.LoadShader("", "res/glsl330_cross_hatching.fs")
	shaders[FxCrossStitching] = rl.LoadShader("", "res/glsl330_cross_stitching.fs")
	shaders[FxPredatorView] = rl.LoadShader("", "res/glsl330_predator.fs")
	shaders[FxScanlines] = rl.LoadShader("", "res/glsl330_scanlines.fs")
	shaders[FxFisheye] = rl.LoadShader("", "res/glsl330_fisheye.fs")
	shaders[FxSobel] = rl.LoadShader("", "res/glsl330_sobel.fs")
	shaders[FxBlur] = rl.LoadShader("", "res/glsl330_blur.fs")
	shaders[FxBloom] = rl.LoadShader("", "res/glsl330_bloom.fs")

	currentShader := FxSobel

	// Create a RenderTexture2D to be used for render to texture
	targetRenderTexture := rl.LoadRenderTexture(screenWidth, screenHeight)

	const defaultCameraFovy = 45.0
	camera := rl.Camera{}
	camera.Position = rl.NewVector3(0.0, arenaWorldWidth, arenaWorldLength)
	camera.Target = rl.NewVector3(0.0, -1.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = defaultCameraFovy
	camera.Projection = rl.CameraPerspective

	churchObj := rl.LoadModel("res/church.obj")                                  // Load OBJ model
	churchTexture := rl.LoadTexture("res/church_diffuse.png")                    // Load model texture
	rl.SetMaterialTexture(churchObj.Materials, rl.MapDiffuse, churchTexture)     // Set obj model diffuse texture
	churchPosition := rl.NewVector3(arenaWorldWidth/4, 0.0, -arenaWorldLength/4) // Set model position

	playerPosition := rl.NewVector3(0.0, 1.0, 2.0)
	playerSize := rl.NewVector3(1.0, 2.0, 1.0)
	playerColor := rl.RayWhite

	// See also https://github.com/Pakz001/Raylib-Examples/blob/master/ai/Example_-_Pattern_Movement.c
	enemyBoxPos := rl.NewVector3(-4.0, 1.0, 0.0)
	enemyBoxSize := rl.NewVector3(2.0, 2.0, 2.0)
	if true {
		enemyBoxPos = rl.NewVector3(-4.0, 1.0, 4.0)
		enemyBoxSize = rl.NewVector3(5, 2.0, 5)
	}

	enemySpherePos := rl.NewVector3(4.0, 0.0, 0.0)
	enemySphereSize := float32(1.5)
	if true {
		enemySpherePos = rl.NewVector3(-4.0, -0.4, -4.0)
		enemySphereSize = float32(2.5)
	}

	isCollision := false
	isOOBCollision := false

	isMartianManhunter := false
	martianManhunterFramesCounter := int32(0)
	martianManhunterMaxFrames := 4 * fps

	framesCounter := 0

	rl.DisableCursor()
	rl.SetTargetFPS(fps)

	for !rl.WindowShouldClose() {
		// Handle user input

		if rl.IsKeyPressed(rl.KeyRightBracket) {
			currentShader++
		} else if rl.IsKeyPressed(rl.KeyLeftBracket) {
			currentShader--
		}

		if currentShader >= MaxPostproShaders {
			currentShader = 0
		} else if currentShader < 0 {
			currentShader = MaxPostproShaders - 1
		}

		// Update

		// Store previous position to reuse as next postion on collision
		oldPlayerPos := playerPosition
		oldCamPos := camera.Position
		_ = oldCamPos

		// Move player
		const magnitude = float32(0.2)
		currMagnitude := magnitude
		isBoost := false
		isStrafe := false
		movement := rl.Vector3Zero()
		if rl.IsKeyDown(rl.KeyRight) {
			movement.X += 1 // Right
		}
		if rl.IsKeyDown(rl.KeyLeft) {
			movement.X -= 1 // Left
		}
		if rl.IsKeyDown(rl.KeyDown) {
			movement.Z += 1 // Backward
		}
		if rl.IsKeyDown(rl.KeyUp) {
			movement.Z -= 1 // Forward
		}
		if isMoveYPlane := true; isMoveYPlane {
			if rl.IsKeyDown(rl.KeySpace) {
				movement.Y += 1 // Up
				currMagnitude *= math.Phi * math.Phi
			}
			if rl.IsKeyDown(rl.KeyLeftControl) {
				movement.Y -= 1 // Down
			}
		}
		if rl.IsKeyDown(rl.KeyLeftShift) {
			isBoost = true
		}
		if rl.IsKeyDown(rl.KeyLeftAlt) {
			isStrafe = true
		}
		if !rl.Vector3Equals(movement, rl.Vector3Zero()) { // Vector3Length (XZ): 1.414 --diagonal-> 0.99999994
			movement = rl.Vector3Normalize(movement) // See also https://community.monogame.net/t/how-can-i-normalize-my-diagonal-movement/15276
		}
		if isBoost {
			currMagnitude *= math.Phi
		}
		if isStrafe {
			currMagnitude /= math.Phi
		}

		// Move player this frame
		playerPosition.X += movement.X * currMagnitude
		playerPosition.Y += movement.Y * currMagnitude
		playerPosition.Z += movement.Z * currMagnitude

		// Apply Gravity
		playerPosition.Y -= magnitude * (math.Phi / 2)

		// HACK: Check if player is touching an infinite floor
		if playerPosition.Y+playerSize.Y/2 < 2 {
			playerPosition.Y = playerSize.Y / 2
		}

		// Reset collision flags
		isCollision = false
		isOOBCollision = false

		// Check collisions player vs enemy-box
		if rl.CheckCollisionBoxes(
			rl.NewBoundingBox(
				rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
				rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2)),
			rl.NewBoundingBox(
				rl.NewVector3(enemyBoxPos.X-enemyBoxSize.X/2, enemyBoxPos.Y-enemyBoxSize.Y/2, enemyBoxPos.Z-enemyBoxSize.Z/2),
				rl.NewVector3(enemyBoxPos.X+enemyBoxSize.X/2, enemyBoxPos.Y+enemyBoxSize.Y/2, enemyBoxPos.Z+enemyBoxSize.Z/2)),
		) {
			isCollision = true
		}

		// Check collisions player vs enemy-sphere
		if rl.CheckCollisionBoxSphere(
			rl.NewBoundingBox(
				rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
				rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2)),
			enemySpherePos,
			enemySphereSize,
		) {
			isCollision = true
		}

		// Check collisions player vs arena bounds
		if playerPosition.X-playerSize.X/2 <= -arenaWorldWidth/2 || playerPosition.X+playerSize.X/2 >= arenaWorldWidth/2 {
			isOOBCollision = true
		}
		if playerPosition.Z-playerSize.Z/2 <= -arenaWorldLength/2 || playerPosition.Z+playerSize.Z/2 >= arenaWorldLength/2 {
			isOOBCollision = true
		}
		if isCollision || isOOBCollision {
			playerColor = rl.DarkGray
			camera.Fovy += (float32(rl.GetRandomValue(-10, 10)) / 20) / (2 * math.Pi) // Screenshake
		} else {
			playerColor = rl.Black
			camera.Fovy = defaultCameraFovy
		}
		if isCollision {
			deltaFovy := defaultCameraFovy - camera.Fovy
			deltaFovy = AbsF(deltaFovy)
			alpha := deltaFovy * deltaFovy
			if deltaFovy != 0 && alpha < 0.001 {
				isMartianManhunter = true
			}
			if isMartianManhunter {
				playerPosition = rl.Vector3Lerp(playerPosition, oldPlayerPos, 0.8)
			} else {
				if isStuck := !isMartianManhunter && camera.Fovy != defaultCameraFovy; isStuck {
					playerPosition = rl.Vector3Lerp(playerPosition, oldPlayerPos, 1-alpha)
				} else {
					playerPosition = oldPlayerPos
				}
			}
		}
		if isOOBCollision {
			playerPosition = oldPlayerPos
		}
		if martianManhunterFramesCounter > martianManhunterMaxFrames {
			isMartianManhunter = false
		}
		if isMartianManhunter {
			martianManhunterFramesCounter++
		} else if martianManhunterFramesCounter > 0 {
			martianManhunterFramesCounter--
			if martianManhunterFramesCounter <= 0 {
				martianManhunterFramesCounter = 0
			}
		}

		/* // Input vertex attributes (from vertex shader)
		   in vec2 fragTexCoord;
		   in vec4 fragColor;

		   // Input uniform values
		   uniform sampler2D texture0;
		   uniform vec4 colDiffuse;

		   // Output fragment color
		   out vec4 finalColor;

		   // NOTE: Add here your custom variables
		   uniform vec2 resolution = vec2(800, 450);
		*/
		{
			const uniformName = "resolution" // locIndex == 2
			locIndex := rl.GetShaderLocation(shaders[currentShader], uniformName)
			if camera.Fovy != defaultCameraFovy {
				value := []float32{float32(screenWidth), float32(screenHeight)}
				value[0] += (playerPosition.X - camera.Position.X) / arenaWorldWidth
				value[1] += (playerPosition.Z - camera.Position.Z) / arenaWorldLength
				alpha1 := ((defaultCameraFovy - camera.Fovy) / defaultCameraFovy)
				value[0] *= arenaWorldWidth * alpha1
				value[1] *= arenaWorldLength * alpha1
				value[0] += (defaultCameraFovy - camera.Fovy)
				value[1] += (defaultCameraFovy - camera.Fovy)
				value0Sign := AbsF(value[0] / value[0])
				value1Sign := AbsF(value[1] / value[1])
				// Note: Use rl.Lerp to tween value to avoid sudden animation jumpcut. Use cubic alpha2 for chaotic distortion
				if alpha2 := 1 - float32(martianManhunterFramesCounter)/float32(martianManhunterMaxFrames); alpha2 != 1 {
					value[0] = rl.Lerp(value[0], value0Sign*float32(screenWidth), SqrtF(alpha2))
					value[1] = rl.Lerp(value[1], value1Sign*float32(screenWidth), SqrtF(alpha2))
				} else {
					if value[0] > -64 && value[0] < 64 && value[0] != 0 { // value[0] = float32(screenWidth) / 2 * value0Sign
						value[0] = rl.Lerp(value[0], value0Sign*float32(screenWidth), 1/math.Phi)
					}
					if value[1] > -64 && value[1] < 64 && value[1] != 0 { // value[1] = float32(screenHeight) / 2 * value1Sign
						value[1] = rl.Lerp(value[1], value1Sign*float32(screenWidth), 1/math.Phi)
					}
				}
				rl.SetShaderValue(shaders[currentShader], locIndex, value, rl.ShaderUniformVec2)
			} else {
				value := []float32{float32(screenWidth), float32(screenHeight)}
				rl.SetShaderValue(shaders[currentShader], locIndex, value, rl.ShaderUniformVec2)
			}
		}

		framesCounter++

		// Draw

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		// Enable drawing to texture
		rl.BeginTextureMode(targetRenderTexture)

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)
		// Draw world
		{
			rl.DrawModel(churchObj, churchPosition, 0.25, rl.White) // Draw 3d model with texture

			// Draw floor
			rl.DrawCubeV(rl.NewVector3(0, -1, 0), rl.NewVector3(arenaWorldWidth, 2.0, arenaWorldLength), rl.White)
			rl.DrawCubeWiresV(rl.NewVector3(0, -1, 0), rl.NewVector3(arenaWorldWidth, 2.0, arenaWorldLength), rl.RayWhite)

			// Draw enemy-box
			rl.DrawCube(enemyBoxPos, enemyBoxSize.X, enemyBoxSize.Y, enemyBoxSize.Z, rl.Black)
			rl.DrawCubeWires(enemyBoxPos, enemyBoxSize.X, enemyBoxSize.Y, enemyBoxSize.Z, rl.DarkGray)

			// Draw enemy-sphere
			rl.DrawSphere(enemySpherePos, enemySphereSize, rl.Black)
			rl.DrawSphereWires(enemySpherePos, enemySphereSize, 16/4, 16/2, rl.DarkGray)

			// Draw player
			if martianManhunterFramesCounter > 0 {
				alpha := 1 - float32(martianManhunterFramesCounter/martianManhunterMaxFrames)
				rl.DrawCubeV(playerPosition, playerSize, rl.Fade(playerColor, alpha))
				rl.DrawCubeWiresV(playerPosition, playerSize, rl.Fade(playerColor, 1-alpha))
			} else {
				rl.DrawCubeV(playerPosition, playerSize, playerColor)
				rl.DrawCubeWiresV(playerPosition, playerSize, rl.Fade(playerColor, 0.382))
			}

			if false {
				rl.BeginBlendMode(rl.BlendCustom)
				rl.DrawGrid(int32(arenaWorldLength), 1.0)
				rl.EndBlendMode()
			}
		}
		rl.EndMode3D()

		rl.EndTextureMode()

		// Render previously generated texture using selected postpro shader
		rl.BeginShaderMode(shaders[currentShader])

		// NOTE: Render texture must be y-flipped due to default OpenGL coordinates (left-bottom)
		rl.DrawTextureRec(
			targetRenderTexture.Texture,
			rl.NewRectangle(0, 0, float32(targetRenderTexture.Texture.Width), float32(-targetRenderTexture.Texture.Height)),
			rl.NewVector2(0, 0),
			rl.White,
		)

		rl.EndShaderMode()

		if false {
			text := postproShaderText[currentShader]
			textW := rl.MeasureText(text, 10)
			rl.DrawRectangle(screenWidth/2-textW/2-10, 5, textW+10*2, 30, rl.Fade(rl.RayWhite, 0.7))
			rl.DrawText(text, screenWidth/2-textW/2, 15, 10, rl.Black)
		}

		rl.DrawText(fmt.Sprintln(martianManhunterFramesCounter), 10, 10, 10, rl.Gray)

		rl.DrawFPS(10, int32(rl.GetScreenHeight())-25)

		rl.EndDrawing()
	}

	// Unload all postpro shaders
	for i := 0; i < MaxPostproShaders; i++ {
		rl.UnloadShader(shaders[i])
	}

	rl.UnloadTexture(churchTexture)             // Unload texture
	rl.UnloadModel(churchObj)                   // Unload model
	rl.UnloadRenderTexture(targetRenderTexture) // Unload render texture

	rl.CloseWindow()
}

type ComaparableNumber interface {
	float32 | float64 | int | int32
}

type NumberType ComaparableNumber

func SqrtF[T NumberType](value T) float32 { return float32(math.Abs(float64(value))) }
func AbsF[T NumberType](value T) float32  { return float32(math.Abs(float64(value))) }
