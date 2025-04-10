package main

import (
	"fmt"
	"math"

	_ "github.com/gen2brain/raylib-go/easings"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Checklist
//   - Ensure on fullscreen toggle, the proportion stays same, and the world is
//     scaled by Raylib 3d camera mode
func main() {
	fps := int32(60)
	screenWidth := int32(800)
	screenHeight := int32(450)

	screenWidth = int32(rl.GetScreenWidth())
	screenHeight = int32(rl.GetScreenHeight())

	rl.InitWindow(screenWidth, screenHeight, "raylib [models] example - box collisions")
	rl.ToggleFullscreen()

	const arenaWidth = float32(20)  // X
	const arenaLength = float32(20) // Z

	camera := rl.Camera{}
	camera.Position = rl.NewVector3(0.0, arenaWidth, arenaLength)
	camera.Target = rl.NewVector3(0.0, -1.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	const sphereModelRadius = arenaWidth / math.Phi
	sphereMesh := rl.GenMeshSphere(sphereModelRadius, 16, 16)
	sphereModel := rl.LoadModelFromMesh(sphereMesh)

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
	martianManhunterMaxFrames := int32(4 * fps)

	framesCounter := 0

	rl.SetTargetFPS(fps)

	for !rl.WindowShouldClose() {
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
		playerPosition.X += movement.X * currMagnitude
		playerPosition.Y += movement.Y * currMagnitude
		playerPosition.Z += movement.Z * currMagnitude

		// Apply Gravity
		playerPosition.Y -= magnitude * (math.Phi / 2)
		// HACK: Check if player is touching an infinite floor
		if playerPosition.Y+playerSize.Y/2 < 2 {
			playerPosition.Y = playerSize.Y / 2
		}
		// camera.Position.Z = rl.Lerp(camera.Position.Z, arenaLength, math.Phi-1)
		// camera.Position.Z = rl.Lerp(camera.Position.Z, playerPosition.Z, math.Phi-1)

		// camera.Target.X = rl.Lerp(camera.Target.X, playerPosition.X, 0.1)
		// camera.Target.Y = rl.Lerp(camera.Target.Y, playerPosition.Y, 0.1)
		// camera.Target.Z = rl.Lerp(camera.Target.Z, playerPosition.Z, 0.1)

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
		if playerPosition.X-playerSize.X/2 <= -arenaWidth/2 || playerPosition.X+playerSize.X/2 >= arenaWidth/2 {
			isOOBCollision = true
		}
		if playerPosition.Z-playerSize.Z/2 <= -arenaLength/2 || playerPosition.Z+playerSize.Z/2 >= arenaLength/2 {
			isOOBCollision = true
		}
		if isCollision || isOOBCollision {
			playerColor = rl.DarkGray
			camera.Fovy += (float32(rl.GetRandomValue(-10, 10)) / 20) / (2 * math.Pi) // Screenshake
		} else {
			playerColor = rl.Black
			camera.Fovy = 45.0
		}
		if isCollision {
			deltaFovy := 45 - camera.Fovy
			deltaFovy = float32(math.Abs(float64(deltaFovy)))
			alpha := deltaFovy * deltaFovy
			if deltaFovy != 0 && alpha < 0.001 {
				isMartianManhunter = true
			}
			if isMartianManhunter {
				playerPosition = rl.Vector3Lerp(playerPosition, oldPlayerPos, 0.8)
			} else {
				if isStuck := !isMartianManhunter && camera.Fovy != 45.0; isStuck {
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

		framesCounter++

		// Draw

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)

		// Draw floor
		rl.DrawCubeV(rl.NewVector3(0, -1, 0), rl.NewVector3(arenaWidth, 2.0, arenaLength), rl.White)
		rl.DrawCubeWiresV(rl.NewVector3(0, -1, 0), rl.NewVector3(arenaWidth, 2.0, arenaLength), rl.RayWhite)

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
			rl.DrawCubeWiresV(playerPosition, playerSize, rl.ColorBrightness(playerColor, 1-alpha))
		} else {
			rl.DrawCubeV(playerPosition, playerSize, playerColor)
			rl.DrawCubeWiresV(playerPosition, playerSize, rl.ColorBrightness(playerColor, 0.382))
		}

		// Draw prop shpere
		if false {
			rl.DrawSphere(rl.NewVector3(0, -sphereModelRadius, -sphereModelRadius*2), sphereModelRadius-0.02, rl.Fade(rl.LightGray, 0.5))
			rl.DrawModelEx(sphereModel, rl.NewVector3(0, -sphereModelRadius, -sphereModelRadius*2), rl.NewVector3(0, -1, 0), float32(framesCounter), rl.NewVector3(1, 1, 1), rl.White)
		}

		// rl.DrawGrid(int32(arenaWidth), 1.0)

		rl.EndMode3D()

		rl.DrawText(fmt.Sprintln(martianManhunterFramesCounter), 10, 10, 10, rl.Gray)

		// rl.DrawText("Move player with cursors to collide", 220, 40, 20, rl.Gray)

		rl.DrawFPS(10, int32(rl.GetScreenHeight())-25)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}