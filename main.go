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

	const arenaWidth = float32(10) * 3  // X
	const arenaLength = float32(10) * 3 // Z

	camera := rl.Camera{}
	camera.Position = rl.NewVector3(0.0, arenaWidth, arenaLength)
	camera.Target = rl.NewVector3(0.0, -1.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	const sphereModelRadius = arenaWidth / math.Phi
	sphereMesh := rl.GenMeshSphere(sphereModelRadius, 16, 16)
	sphereModel := rl.LoadModelFromMesh(sphereMesh)

	isPlayerBoost := false
	isPlayerStrafe := false
	_ = isPlayerBoost
	_ = isPlayerStrafe
	playerColor := rl.RayWhite
	playerJumpsLeft := 1
	playerPosition := rl.NewVector3(0.0, 1.0, 2.0)
	playerSize := rl.NewVector3(1.0, 2.0, 1.0)
	playerVelocity := rl.Vector3{}
	playerAirTimer := float32(0)

	maxPlayerAirTime := 5 * float32(fps)
	const movementMagnitude = float32(0.2)
	const playerJumpVelocity = 3
	const terminalVelocityY = 5
	// # if max: 0.1333333333.. (makes jumping possible to 3x player height) # else use min for easy floaty feel
	// self._terminal_limiter_air_friction: Final = max(0.1, ((pre.TILE_SIZE * 0.5) / (pre.FPS_CAP)))
	const terminalVelocityLimiterAirFriction = movementMagnitude / math.Phi
	const terminalVelocityLimiterAirFrictionY = movementMagnitude / 2

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

	handlePlayerJump := func() {
		playerJumpsLeft--
		playerVelocity.Y = playerJumpVelocity
		playerAirTimer = maxPlayerAirTime
	}

	rl.SetTargetFPS(fps)

	for !rl.WindowShouldClose() {
		// Handle user input events

		playerMovementThisFrame := rl.Vector3{}
		playerCollisionsThisFrame := rl.Vector4{}

		if rl.IsKeyDown(rl.KeyRight) {
			playerMovementThisFrame.X += 1 // Right
		}
		if rl.IsKeyDown(rl.KeyLeft) {
			playerMovementThisFrame.X -= 1 // Left
		}
		if rl.IsKeyDown(rl.KeyDown) {
			playerMovementThisFrame.Z += 1 // Backward
		}
		if rl.IsKeyDown(rl.KeyUp) {
			playerMovementThisFrame.Z -= 1 // Forward
		}
		if rl.IsKeyDown(rl.KeyLeftShift) {
			isPlayerBoost = true
		}
		if rl.IsKeyDown(rl.KeyLeftAlt) {
			isPlayerStrafe = true
		}
		if rl.IsKeyDown(rl.KeySpace) { // Jump
			if playerJumpsLeft > 0 {
				handlePlayerJump()
			}
		}

		// Update

		// Store previous position to reuse as next postion on collision
		oldPlayerPos := playerPosition
		oldCamPos := camera.Position
		_ = oldCamPos

		// Update player
		{
			playerAirTimer++
			// currMagnitude := movementMagnitude

			// Normalize input vector to avoid speeding up diagonally
			if !rl.Vector3Equals(playerMovementThisFrame, rl.Vector3Zero()) { // Vector3Length (XZ): 1.414 --diagonal-> 0.99999994
				playerMovementThisFrame = rl.Vector3Normalize(playerMovementThisFrame) // See also https://community.monogame.net/t/how-can-i-normalize-my-diagonal-movement/15276
			}

			// Copied from https://github.com/lloydlobo/tiptoe/blob/master/src/internal/entities.py
			// """
			// Update the entity's position based on physics and collisions.
			// Physics: movement via collision detection 2 part axis method handle one
			// axis at a time for predictable resolution also pygame-ce allows
			// calculating floats with Rects
			// Note: For each X and Y axis movement, we update x and y position as int
			// as pygame rect don't handle it as of now.
			// """
			// # Compute players input based movement with entity velocity
			// frame_movement: pg.Vector2 = movement + self.velocity
			frameMovement := rl.Vector3Add(playerMovementThisFrame, playerVelocity)
			{
				playerPosition.X += frameMovement.X * movementMagnitude
				if isTouchXPlaneEdges := playerPosition.X-playerSize.X/2 < -arenaWidth/2 ||
					playerPosition.X+playerSize.X/2 > arenaWidth/2; isTouchXPlaneEdges {
					playerCollisionsThisFrame.X = 1
				}
				playerPosition.Y += frameMovement.Y * movementMagnitude
				if false {
					playerCollisionsThisFrame.Y = 1
				}
				playerPosition.Z += frameMovement.Z * movementMagnitude
				if isTouchZPlaneEdges := playerPosition.Z-playerSize.Z/2 < -arenaLength/2 ||
					playerPosition.Z+playerSize.Z/2 > arenaLength/2; isTouchZPlaneEdges {
					playerCollisionsThisFrame.Z = 1
				}

				// HACK: Gravity: Check if player is touching an infinite floor
				isTouchFloor := playerPosition.Y+playerSize.Y/2 < 2
				if isTouchFloor {
					playerCollisionsThisFrame.W = 1
				}

				// # Entity: Update velocity
				playerVelocity.Y = MinF(terminalVelocityY, playerVelocity.Y-terminalVelocityLimiterAirFrictionY) // Decelerate if in air

				// # Entity: Handle velocity based on collisions up or down
				if playerCollisionsThisFrame.Y == 1 || playerCollisionsThisFrame.W == 1 {
					playerVelocity.Y = 0 // self.Velocity = 0
				}

				// # Entity:Player: Handle velocity based on collisions
				if playerCollisionsThisFrame.Y == 1 || playerCollisionsThisFrame.W == 1 {
					// ... or rl.QuaternionLength(playerCollisionsThisFrame) > 0
					playerAirTimer = 0
					playerPosition.Y = playerSize.Y / 2 // Fix to ground
					playerJumpsLeft = 1
				}

				if playerAirTimer > maxPlayerAirTime {
					// fmt.Printf("playerAirTimer: %v\n", playerAirTimer)
					// panic(playerAirTimer)
					// playerVelocity.Y -= terminalVelocityLimiterAirFrictionY
				}

				if maxAirTimeToGameOver := maxPlayerAirTime * float32(fps); playerAirTimer > maxAirTimeToGameOver {
					panic("Unimplemented: playerAirTimer > maxAirTimeToGameOver")
				}
			}

			// Apply Gravity
			playerVelocity.Y -= terminalVelocityLimiterAirFrictionY

			// Normalize velocity along XZ plane (width and length)
			// Only for player (remove for other entities)!!!!!
			if playerVelocity.X > 0 {
				playerVelocity.X = MaxF(0, playerVelocity.X-terminalVelocityLimiterAirFriction)
			} else {
				playerVelocity.X = MinF(0, playerVelocity.X+terminalVelocityLimiterAirFriction)
			}
			if playerVelocity.Z > 0 {
				playerVelocity.Z = MaxF(0, playerVelocity.Z-terminalVelocityLimiterAirFriction)
			} else {
				playerVelocity.Z = MinF(0, playerVelocity.Z+terminalVelocityLimiterAirFriction)
			}
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

// From cmp.Ordered
type OrderedNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}
type NumberType OrderedNumber

func MaxF[T NumberType](x T, y T) float32 { return float32(max(float64(x), float64(y))) }
func MinF[T NumberType](x T, y T) float32 { return float32(min(float64(x), float64(y))) }

