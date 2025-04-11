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

	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "raylib [models] example - box collisions")
	// rl.ToggleFullscreen()

	const arenaWidth = float32(10) * math.Phi             // X
	const arenaLength = float32(10) * math.Phi * math.Phi // Z

	const arenaWallHeight = 2

	camera := rl.Camera{}
	camera.Position = rl.NewVector3(0.0, arenaWidth, arenaLength)
	camera.Target = rl.NewVector3(0.0, -1.0, 0.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	const sphereModelRadius = arenaWidth / math.Phi
	sphereMesh := rl.GenMeshSphere(sphereModelRadius, 16, 16)
	sphereModel := rl.LoadModelFromMesh(sphereMesh)

	fuelProgress := float32(1.0)
	shieldProgress := float32(1.0)

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

	maxPlayerAirTime := float32(fps) / 2.0
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

	// anticlockwise: 0 -> 1 -> 2 -> 3 TLBR
	walls := [4]rl.BoundingBox{}
	const wallThick = 1 / 2.0
	{
		w := arenaWidth / 2
		l := arenaLength / 2
		walls = [4]rl.BoundingBox{
			rl.NewBoundingBox(rl.NewVector3(-w, 0, -l), rl.NewVector3(w, arenaWallHeight, -l+wallThick)),
			rl.NewBoundingBox(rl.NewVector3(-w, 0, -l), rl.NewVector3(-w+wallThick, arenaWallHeight, l)),
			rl.NewBoundingBox(rl.NewVector3(-w, 0, l-wallThick), rl.NewVector3(w, arenaWallHeight, l)),
			rl.NewBoundingBox(rl.NewVector3(w-wallThick, 0, -l), rl.NewVector3(w, arenaWallHeight, l)),
		}
	}

	floorOrigin := rl.NewVector3(0, -1, 0)
	const floorThick = 1
	floorMesh := rl.GenMeshPlane(arenaWidth, arenaLength, 3, 3)
	floorModel := rl.LoadModelFromMesh(floorMesh)
	floorBoundingBox := rl.NewBoundingBox(rl.NewVector3(-arenaWidth/2, floorOrigin.Y, -arenaLength/2), rl.NewVector3(arenaWidth/2, floorOrigin.Y+floorThick, arenaLength/2))
	_ = floorModel
	_ = floorBoundingBox

	isCollision := false
	isOOBCollision := false

	martianManhunterTriggerFactor := float32(0.0)
	const maxMartianManhunterTriggerFactor = 45.0

	isMartianManhunter := false
	martianManhunterFramesCounter := int32(0)
	maxMartianManhunterFramesCounter := int32(4 * fps)

	framesCounter := 0

	handlePlayerJump := func() {
		playerJumpsLeft--
		playerVelocity.Y = playerJumpVelocity
		playerAirTimer = maxPlayerAirTime
	}

	rl.DisableCursor()
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
			playerAirTimer += 1.0

			// Normalize input vector to avoid speeding up diagonally
			// See also https://community.monogame.net/t/how-can-i-normalize-my-diagonal-movement/15276
			// Vector3Length (XZ): 1.414 --diagonal-> 0.99999994
			if !rl.Vector3Equals(playerMovementThisFrame, rl.Vector3Zero()) {
				playerMovementThisFrame = rl.Vector3Normalize(playerMovementThisFrame)
				fuelProgress -= 0.05 / float32(fps)
			}

			if fuelProgress <= 0 {
				playerPosition = rl.Vector3Zero() // Gameover
			}
			if shieldProgress <= 0 {
				playerPosition = rl.Vector3Zero() // Gameover
			}

			frameMovement := rl.Vector3Add(playerMovementThisFrame, playerVelocity)
			{
				playerPosition.X += frameMovement.X * movementMagnitude
				if isTouchXPlaneEdges := playerPosition.X-playerSize.X/2 < -arenaWidth/2 || playerPosition.X+playerSize.X/2 > arenaWidth/2; isTouchXPlaneEdges {
					playerCollisionsThisFrame.X = 1
				}
				playerPosition.Y += frameMovement.Y * movementMagnitude
				if false {
					playerCollisionsThisFrame.Y = 1
				}
				playerPosition.Z += frameMovement.Z * movementMagnitude
				if isTouchZPlaneEdges := playerPosition.Z-playerSize.Z/2 < -arenaLength/2 || playerPosition.Z+playerSize.Z/2 > arenaLength/2; isTouchZPlaneEdges {
					playerCollisionsThisFrame.Z = 1
				}
				// HACK: Gravity: Check if player is touching the infinite floor
				if false {
					isTouchFloor := playerPosition.Y+playerSize.Y/2 < 2
					if isTouchFloor {
						playerCollisionsThisFrame.W = 1
					}
				} else {
					playerBox := rl.NewBoundingBox(
						rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
						rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2))

					if rl.CheckCollisionBoxes(playerBox, floorBoundingBox) {
						playerCollisionsThisFrame.W = 1
					}
				}

				// # Entity: Update velocity
				playerVelocity.Y = MinF(terminalVelocityY, playerVelocity.Y-terminalVelocityLimiterAirFrictionY) // Decelerate if in air
				// # Entity: Handle velocity based on collisions up or down
				if playerCollisionsThisFrame.Y == 1 || playerCollisionsThisFrame.W == 1 {
					playerVelocity.Y = 0 // self.Velocity = 0
				}
				// # Entity:Player: Handle velocity based on collisions
				if playerCollisionsThisFrame.Y == 1 || playerCollisionsThisFrame.W == 1 {
					if playerAirTimer > 0 && playerJumpsLeft == 0 {
					}
					playerAirTimer = 0
					playerPosition.Y = playerSize.Y / 2 // Fix to ground
					playerJumpsLeft = 1
				}
				// Snappy bouncy jumps
				if playerAirTimer > maxPlayerAirTime*math.Phi && playerAirTimer < maxPlayerAirTime*math.Phi*math.Phi { // Once
					playerVelocity.Y -= terminalVelocityLimiterAirFrictionY
				}
			}

			// Apply Gravity
			playerVelocity.Y -= terminalVelocityLimiterAirFrictionY

			// Normalize velocity along XZ plane (width and length) Only for player (remove for other entities)!!!!!
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

		// Reset collision flags
		isCollision = false
		isOOBCollision = false

		// Check collisions player vs enemy-box
		playerBox := rl.NewBoundingBox(
			rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
			rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2))
		enemyBoundingBox := rl.NewBoundingBox(
			rl.NewVector3(enemyBoxPos.X-enemyBoxSize.X/2, enemyBoxPos.Y-enemyBoxSize.Y/2, enemyBoxPos.Z-enemyBoxSize.Z/2),
			rl.NewVector3(enemyBoxPos.X+enemyBoxSize.X/2, enemyBoxPos.Y+enemyBoxSize.Y/2, enemyBoxPos.Z+enemyBoxSize.Z/2))

		if rl.CheckCollisionBoxes(playerBox, enemyBoundingBox) {
			isCollision = true
		}
		if rl.CheckCollisionBoxSphere(playerBox, enemySpherePos, enemySphereSize) {
			isCollision = true
		}
		for i := range walls {
			if rl.CheckCollisionBoxes(playerBox, walls[i]) {
				isCollision = true
			}
		}

		// Check collisions player vs arena outer bounds (security check)
		if playerPosition.X-playerSize.X/2 <= -arenaWidth/2 || playerPosition.X+playerSize.X/2 >= arenaWidth/2 {
			isOOBCollision = true
		}
		if playerPosition.Z-playerSize.Z/2 <= -arenaLength/2 || playerPosition.Z+playerSize.Z/2 >= arenaLength/2 {
			isOOBCollision = true
		}
		const offsetTrigger = 2.0
		if isCollision || isOOBCollision {
			playerColor = rl.DarkGray
			martianManhunterTriggerFactor += (float32(rl.GetRandomValue(-offsetTrigger, offsetTrigger)) / offsetTrigger * 2) / (2 * math.Pi) // Screenshake
		} else {
			playerColor = rl.Fade(rl.Black, 0.9)
			martianManhunterTriggerFactor = maxMartianManhunterTriggerFactor
		}
		if isCollision {
			deltaFovy := 45 - martianManhunterTriggerFactor
			deltaFovy = float32(math.Abs(float64(deltaFovy)))
			alpha := deltaFovy * deltaFovy
			if deltaFovy != 0 && alpha < 0.0001*offsetTrigger {
				isMartianManhunter = true
			}
			if isMartianManhunter {
				playerPosition = rl.Vector3Lerp(playerPosition, oldPlayerPos, 0.8)
			} else {
				if isStuck := !isMartianManhunter && martianManhunterTriggerFactor != maxMartianManhunterTriggerFactor; isStuck {
					playerPosition = rl.Vector3Lerp(playerPosition, oldPlayerPos, 1-alpha)
				} else {
					playerPosition = oldPlayerPos
				}
			}
		}
		if false {
			if isOOBCollision {
				playerPosition = oldPlayerPos
			}
		}
		if martianManhunterFramesCounter > maxMartianManhunterFramesCounter {
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

		if playerCollisionsThisFrame.X == 1 || playerCollisionsThisFrame.Z == 1 {
			shieldProgress -= 0.1 / float32(fps)
		}

		framesCounter++

		// Draw

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)

		for i := range walls {
			vecMin := walls[i].Min
			vecMax := walls[i].Max
			const amount = 0.5 // Lerp(min, max, 0.5)
			size := rl.Vector3{X: vecMax.X - vecMin.X, Y: vecMax.Y - vecMin.Y, Z: vecMax.Z - vecMin.Z}
			origin := rl.Vector3{
				X: vecMin.X + amount*(vecMax.X-vecMin.X),
				Y: vecMin.Y + amount*(vecMax.Y-vecMin.Y),
				Z: vecMin.Z + amount*(vecMax.Z-vecMin.Z),
			}
			_ = size
			_ = origin
			rl.DrawCubeV(origin, size, rl.Fade(rl.White, 0.125/2))
			rl.DrawBoundingBox(walls[i], rl.Fade(rl.LightGray, 0.4))
		}

		// Draw floor
		rl.DrawCubeV(rl.Vector3{X: floorOrigin.X, Y: floorOrigin.Y - 0.125, Z: floorOrigin.Z}, rl.NewVector3(arenaWidth, 2.0, arenaLength), rl.Fade(rl.White, 0.125/2))
		rl.DrawCubeWiresV(floorOrigin, rl.NewVector3(arenaWidth, 2.0, arenaLength), rl.Fade(rl.LightGray, 0.7))
		rl.DrawBoundingBox(floorBoundingBox, rl.Fade(rl.LightGray, 0.7))
		rl.DrawModel(floorModel, rl.NewVector3(floorOrigin.X, floorOrigin.Y+1, floorOrigin.Z), 1.0, rl.Fade(rl.White, 0.9))
		if false {
			rl.DrawPlane(rl.NewVector3(floorOrigin.X, floorOrigin.Y, floorOrigin.Z), rl.NewVector2(arenaWidth, arenaLength), rl.Fade(rl.White, 0.3))
		}

		// Draw enemy-box
		rl.DrawCube(enemyBoxPos, enemyBoxSize.X, enemyBoxSize.Y, enemyBoxSize.Z, rl.Black)
		rl.DrawCubeWires(enemyBoxPos, enemyBoxSize.X, enemyBoxSize.Y, enemyBoxSize.Z, rl.Black)

		// Draw enemy-sphere
		rl.DrawSphere(enemySpherePos, enemySphereSize, rl.Black)
		rl.DrawSphereWires(enemySpherePos, enemySphereSize, 16/4, 16/2, rl.Black)

		// Draw player
		{
			playerRadius := playerSize.X / 2
			startPos := rl.NewVector3(playerPosition.X, playerPosition.Y-playerSize.Y/2+playerRadius, playerPosition.Z)
			endPos := rl.NewVector3(playerPosition.X, playerPosition.Y+playerSize.Y/2-playerRadius, playerPosition.Z)
			if martianManhunterFramesCounter > 0 {
				alpha := float32(martianManhunterFramesCounter / maxMartianManhunterFramesCounter)
				rl.DrawCubeV(playerPosition, playerSize, rl.Fade(playerColor, 0.5+alpha/4))
				rl.DrawCapsule(startPos, endPos, playerRadius, 16, 16, rl.Fade(playerColor, 0.5+alpha/4))
			} else {
				rl.DrawCapsule(startPos, endPos, playerRadius, 16, 16, rl.Black)
				rl.DrawCapsuleWires(startPos, endPos, playerRadius, 4, 6, rl.Fade(rl.DarkGray, 0.8))
			}
		}

		// Draw prop sphere
		if false {
			rl.DrawSphere(rl.NewVector3(0, -sphereModelRadius, -sphereModelRadius*2), sphereModelRadius-0.02, rl.Fade(rl.LightGray, 0.5))
			rl.DrawModelEx(sphereModel, rl.NewVector3(0, -sphereModelRadius, -sphereModelRadius*2), rl.NewVector3(0, -1, 0), float32(framesCounter), rl.NewVector3(1, 1, 1), rl.White)
		}

		// rl.DrawGrid(4*int32(MaxF(arenaWidth, arenaLength)), 1/4.0)

		rl.EndMode3D()

		rl.DrawRectangle(10, 20, 100, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20}, rl.Vector2{X: fuelProgress * 100, Y: 20}, rl.DarkGray)
		rl.DrawText("Fuel", 10+5, 21, 20, rl.White)
		rl.DrawText(fmt.Sprintf("%.0f", 100*fuelProgress), 90+5, 20+5*2, 10, rl.White)

		rl.DrawRectangle(10, 20+20, 100, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20 + 20}, rl.Vector2{X: shieldProgress * 100, Y: 20}, rl.DarkGray)
		rl.DrawText("Shield", 10+5, 21+20, 20, rl.White)
		rl.DrawText(fmt.Sprintf("%.0f", 100*shieldProgress), 90+5, 20+20+5*2, 10, rl.White)

		rl.DrawRectangle(10, 20+20+20, 100, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20 + 20 + 20}, rl.Vector2{X: 100 * (1 - (float32(martianManhunterFramesCounter) / float32(maxMartianManhunterFramesCounter))), Y: 20}, rl.DarkGray)
		rl.DrawText("Depth", 10+5, 20+20+20, 20, rl.White)
		rl.DrawText(fmt.Sprintln(martianManhunterFramesCounter), 90+5, 20+20+20+5*2, 10, rl.White)

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

