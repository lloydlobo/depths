package main

import (
	"cmp"
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

	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagWindowResizable)                        // Config flags must be set before InitWindow
	rl.InitWindow(screenWidth, screenHeight, "raylib [models] example - box collisions") // Initialize Window and OpenGL Graphics
	rl.SetWindowState(rl.FlagVsyncHint | rl.FlagInterlacedHint | rl.FlagWindowHighdpi)   // Window state must be set after InitWindow

	// rl.ToggleFullscreen()
	rl.SetWindowMinSize(800, 450) // Prevents my window manager shrinking this to 2x1 units window size

	const (
		arenaWallHeight  = 1
		arenaWidth       = float32(20) * math.Phi // X
		arenaLength      = float32(20) * 1        // Z
		arenaWidthRatio  = (arenaWidth / (arenaWidth + arenaLength))
		arenaLengthRatio = (arenaLength / (arenaWidth + arenaLength))
		camPosW          = (arenaWidth * (math.Phi + arenaLengthRatio)) * (1 - OneMinusInvMathPhi)
		camPosL          = (arenaLength * (math.Phi + arenaWidthRatio)) * (1 - OneMinusInvMathPhi)
	)

	camera := rl.Camera{
		Position:   rl.NewVector3(0.0, camPosW, camPosL),
		Target:     rl.NewVector3(0.0, -1.0, 0.0),
		Up:         rl.NewVector3(0.0, 1.0, 0.0),
		Fovy:       float32(cmp.Or(45.0, 60.0, 30.0)),
		Projection: rl.CameraPerspective,
	}

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
	maxPlayerFreefallAirTime := maxPlayerAirTime * 3
	const movementMagnitude = float32(0.2)
	const playerJumpVelocity = 4 // 3
	const terminalVelocityY = 5
	// # if max: 0.1333333333.. (makes jumping possible to 3x player height) # else use min for easy floaty feel
	// self._terminal_limiter_air_friction: Final = max(0.1, ((pre.TILE_SIZE * 0.5) / (pre.FPS_CAP)))
	const terminalVelocityLimiterAirFriction = movementMagnitude / math.Phi
	const terminalVelocityLimiterAirFrictionY = movementMagnitude / 2

	// FEAT: See also https://github.com/Pakz001/Raylib-Examples/blob/master/ai/Example_-_Pattern_Movement.c
	// Like Arrow shooter crazyggame,,, fruit dispenser

	safeOrangeBoxCount := 0
	var safeOrangeBoxPositions []rl.Vector3
	var safeOrangeBoxSizes []rl.Vector3
	{
		safeOrangeBoxPositions = append(safeOrangeBoxPositions, rl.NewVector3(-4, 1, 0))
		safeOrangeBoxSizes = append(safeOrangeBoxSizes, rl.NewVector3(2, 2, 2))
		safeOrangeBoxCount++
	}
	{
		safeOrangeBoxPositions = append(safeOrangeBoxPositions, rl.NewVector3(0, 1, -4))
		safeOrangeBoxSizes = append(safeOrangeBoxSizes, rl.NewVector3(2, 2, 2))
		safeOrangeBoxCount++
	}

	unsafeRedSphereCount := 0
	var unsafeRedSpherePositions []rl.Vector3
	var unsafeRedSphereSizes []float32
	{
		unsafeRedSpherePositions = append(unsafeRedSpherePositions, rl.NewVector3(4, 0, 0))
		unsafeRedSphereSizes = append(unsafeRedSphereSizes, 1.5)
		unsafeRedSphereCount++
	}
	{
		unsafeRedSpherePositions = append(unsafeRedSpherePositions, rl.NewVector3(0, 0, 6))
		unsafeRedSphereSizes = append(unsafeRedSphereSizes, 1.5)
		unsafeRedSphereCount++
	}

	// safeOrangeBoxPos := rl.NewVector3(-4.0, 1.0, 0.0)
	// safeOrangeBoxSize := rl.NewVector3(2.0, 2.0, 2.0)
	// if true {
	// 	safeOrangeBoxPos = rl.NewVector3(-4.0, 1.0, 4.0)
	// 	safeOrangeBoxSize = rl.NewVector3(5, 2.0/2, 5)
	// }

	// unsafeRedSpherePos := rl.NewVector3(4.0, 0.0, 0.0)
	// unsafeRedSphereSize := float32(1.5)
	// if true {
	// 	unsafeRedSpherePos = rl.NewVector3(-4.0, -0.4, -4.0)
	// 	unsafeRedSphereSize = float32(2.5)
	// }

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

	// Setup floors
	floorCount := 0
	const floorThick = 0.5 * 2
	var floorBoundingBoxes []rl.BoundingBox
	var floorOrigins []rl.Vector3
	var floorModels []rl.Model
	var floorMeshes []rl.Mesh
	{
		origin := rl.NewVector3(0, (playerPosition.Y-playerSize.Y/2)-(floorThick/2), 0)

		mesh := rl.GenMeshCube(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(mesh)
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)

		floorOrigins = append(floorOrigins, origin)
		floorBoundingBoxes = append(floorBoundingBoxes, box)
		floorModels = append(floorModels, model)
		floorMeshes = append(floorMeshes, mesh)
		floorCount++
	}
	{
		origin := rl.NewVector3(arenaWidth/math.Phi, (playerPosition.Y-playerSize.Y/2)-(floorThick/2)-arenaWidth*1, arenaLength/8)

		mesh := rl.GenMeshCube(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(mesh)
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)

		floorOrigins = append(floorOrigins, origin)
		floorBoundingBoxes = append(floorBoundingBoxes, box)
		floorModels = append(floorModels, model)
		floorMeshes = append(floorMeshes, mesh)
		floorCount++
	}
	{
		origin := rl.NewVector3(-3*arenaWidth/4, (playerPosition.Y-playerSize.Y/2)-(floorThick/2)-(playerSize.Y*1), (arenaLength/1)+(playerSize.Z*2))

		mesh := rl.GenMeshCube(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(mesh)
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)

		floorOrigins = append(floorOrigins, origin)
		floorBoundingBoxes = append(floorBoundingBoxes, box)
		floorModels = append(floorModels, model)
		floorMeshes = append(floorMeshes, mesh)
		floorCount++
	}
	{
		origin := rl.NewVector3(3*arenaWidth/4, (playerPosition.Y-playerSize.Y/2-floorThick/2)-arenaWidth/2, -4*arenaLength/3.5)

		mesh := rl.GenMeshCube(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(mesh)
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)

		floorOrigins = append(floorOrigins, origin)
		floorBoundingBoxes = append(floorBoundingBoxes, box)
		floorModels = append(floorModels, model)
		floorMeshes = append(floorMeshes, mesh)
		floorCount++
	}
	{
		origin := rl.NewVector3(-2*arenaWidth/3, (playerPosition.Y-playerSize.Y/2)-(floorThick/2)-((arenaWidth/2)+(playerSize.Y*4)), (-arenaLength/2)+(playerSize.Z*2))

		mesh := rl.GenMeshCube(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(mesh)
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)

		floorOrigins = append(floorOrigins, origin)
		floorBoundingBoxes = append(floorBoundingBoxes, box)
		floorModels = append(floorModels, model)
		floorMeshes = append(floorMeshes, mesh)
		floorCount++
	}

	isUnsafeCollision := false
	isSafeSpotCollision := false
	isFloorCollision := false
	isOOBCollision := false
	isWallCollision := false

	framesCounter := 0

	handlePlayerJump := func() {
		playerJumpsLeft--
		playerVelocity.Y = playerJumpVelocity
		playerAirTimer = maxPlayerAirTime
	}

	rl.DisableCursor()
	rl.SetTargetFPS(fps)

	for !rl.WindowShouldClose() {
		// Reset collision flags
		isUnsafeCollision = false
		isOOBCollision = false
		isSafeSpotCollision = false
		isFloorCollision = false
		isWallCollision = false

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

		// Update player movement
		playerAirTimer += 1.0

		// Normalize input vector to avoid speeding up diagonally
		if !rl.Vector3Equals(playerMovementThisFrame, rl.Vector3Zero()) {
			playerMovementThisFrame = rl.Vector3Normalize(playerMovementThisFrame) // Vector3Length (XZ): 1.414 --diagonal-> 0.99999994
			fuelProgress -= 0.05 / float32(fps)                                    // See also https://community.monogame.net/t/how-can-i-normalize-my-diagonal-movement/15276
		}

		if fuelProgress <= 0 {
			playerPosition = rl.Vector3Zero() // Gameover
		}
		if shieldProgress <= 0 {
			playerPosition = rl.Vector3Zero() // Gameover
			playerAirTimer = 0
			playerVelocity = rl.Vector3Zero()
		}

		frameMovement := rl.Vector3Add(playerMovementThisFrame, playerVelocity)
		{
			playerPosition.X += frameMovement.X * movementMagnitude
			if false {
				if isTouchXPlaneEdges := playerPosition.X-playerSize.X/2 < -arenaWidth/2 || playerPosition.X+playerSize.X/2 > arenaWidth/2; isTouchXPlaneEdges {
					playerCollisionsThisFrame.X = 1
				}
			}
			playerPosition.Y += frameMovement.Y * movementMagnitude
			if false {
				playerCollisionsThisFrame.Y = 1
			}
			playerPosition.Z += frameMovement.Z * movementMagnitude
			if false {
				if isTouchZPlaneEdges := playerPosition.Z-playerSize.Z/2 < -arenaLength/2 || playerPosition.Z+playerSize.Z/2 > arenaLength/2; isTouchZPlaneEdges {
					playerCollisionsThisFrame.Z = 1
				}
			}

			// Check if player is safely standing on the floor
			playerBox := rl.NewBoundingBox(
				rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
				rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2))

			for i := range floorCount {
				if rl.CheckCollisionBoxes(playerBox, floorBoundingBoxes[i]) {
					isFloorCollision = true
					playerCollisionsThisFrame.W = 1
					if isFloorSinking := false; isFloorSinking { // Only push floor down if player just jumped and landed on the floor
						isJumpLandingOnFloor := playerAirTimer >= maxPlayerAirTime
						isFallingWithFloor := playerAirTimer > 0
						if isJumpLandingOnFloor || isFallingWithFloor {
							floorOrigins[i].Y += playerVelocity.Y * terminalVelocityLimiterAirFrictionY
							floorBoundingBoxes[i].Min.Y += playerVelocity.Y * terminalVelocityLimiterAirFrictionY
							floorBoundingBoxes[i].Max.Y += playerVelocity.Y * terminalVelocityLimiterAirFrictionY
						}
					}
					playerPosition.Y = playerSize.Y/2 + floorBoundingBoxes[i].Max.Y // HACK: Allow player to stand on the floor
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
				playerAirTimer = 0
				// HACK: This is handled directly by floorBoundingBoxes collision check interations
				if false {
					// Detect the floor the player is touching and get bounding box top offset from player bottom
					playerPosition.Y = playerSize.Y / 2 // Fix to ground
				}
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

		// Check collisions player vs enemy-box

		// safeOrangeBoundingBox := rl.NewBoundingBox(
		// 	rl.NewVector3(safeOrangeBoxPos.X-safeOrangeBoxSize.X/2, safeOrangeBoxPos.Y-safeOrangeBoxSize.Y/2, safeOrangeBoxPos.Z-safeOrangeBoxSize.Z/2),
		// 	rl.NewVector3(safeOrangeBoxPos.X+safeOrangeBoxSize.X/2, safeOrangeBoxPos.Y+safeOrangeBoxSize.Y/2, safeOrangeBoxPos.Z+safeOrangeBoxSize.Z/2))

		for i := range safeOrangeBoxCount {
			pos := safeOrangeBoxPositions[i]
			size := safeOrangeBoxSizes[i]
			box := rl.NewBoundingBox(
				rl.NewVector3(pos.X-size.X/2, pos.Y-size.Y/2, pos.Z-size.Z/2),
				rl.NewVector3(pos.X+size.X/2, pos.Y+size.Y/2, pos.Z+size.Z/2))
			if rl.CheckCollisionBoxes(box, rl.NewBoundingBox(
				rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
				rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2))) {
				isSafeSpotCollision = true
				playerPosition.Y = playerSize.Y/2 + box.Max.Y // HACK: Allow player to stand on the floor
			}
		}

		for i := range unsafeRedSphereCount {
			if rl.CheckCollisionBoxSphere(rl.NewBoundingBox(
				rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
				rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2)), unsafeRedSpherePositions[i], unsafeRedSphereSizes[i]) {
				isUnsafeCollision = true

				// Find perpendicular curve to XZ plane, i.e slope of circumference
				// WARN: Expect wonky animation, as bottom of player box when on a slope of sphere,
				// may not collide with top tangent to sphere surface.
				dx := CosF(AbsF(playerPosition.X - unsafeRedSpherePositions[i].X))
				dz := CosF(AbsF(playerPosition.Z - unsafeRedSpherePositions[i].Z))
				height := unsafeRedSphereSizes[i]/2 + playerSize.Y
				dy := (dx*dx + dz*dz) * height
				dy = SqrtF(dy)
				dy = rl.Clamp(dy, 0, height)

				playerPosition.Y = unsafeRedSpherePositions[i].Y + dy
			}
		}

		if false {
			for i := range walls {
				if !isWallCollision && rl.CheckCollisionBoxes(rl.NewBoundingBox(
					rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
					rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2)), walls[i]) {
					isWallCollision = true
				}
			}
		}

		const offsetTrigger = 2.0

		switch {
		case isUnsafeCollision:
			playerColor = rl.DarkBlue
		case isOOBCollision:
			playerColor = rl.DarkGray
		case isSafeSpotCollision:
			playerColor = rl.Fade(rl.Green, 0.9)
		case isWallCollision: // TODO: Figure out how to make player wall slide
			playerColor = rl.Fade(rl.Brown, 0.9)
			playerPosition = oldPlayerPos
			panic("unimplemented")
		default:
			playerColor = rl.Fade(rl.Black, 0.9)
		}

		// Update progress
		progressRate := 0.1 / float32(fps)
		if (playerCollisionsThisFrame.X == 1 || playerCollisionsThisFrame.Z == 1) &&
			(!isOOBCollision && !isSafeSpotCollision && !isFloorCollision) {
			shieldProgress -= progressRate
		}
		if playerAirTimer > maxPlayerFreefallAirTime {
			shieldProgress -= PowF(progressRate*shieldProgress, maxPlayerFreefallAirTime/playerAirTimer)
		}
		if isUnsafeCollision {
			shieldProgress -= progressRate
		}
		if isSafeSpotCollision {
			shieldProgress += progressRate / 2
			if shieldProgress >= 1.0 {
				shieldProgress = 1.0
			}
			fuelProgress += progressRate / 2
			if fuelProgress >= 1.0 {
				fuelProgress = 1.0
			}
		}

		framesCounter++

		// Draw

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)

		// Draw walls
		if false {
			for i := range walls {
				max := walls[i].Max
				min := walls[i].Min
				const t = 1 / 2 // Interpolate t==1/2
				size := rl.NewVector3(max.X-min.X, max.Y-min.Y, max.Z-min.Z)
				origin := rl.NewVector3(min.X+t*(max.X-min.X), min.Y+t*(max.Y-min.Y), min.Z+t*(max.Z-min.Z))
				rl.DrawCubeV(origin, size, rl.Fade(rl.White, 0.125/2))
				rl.DrawBoundingBox(walls[i], rl.Fade(rl.LightGray, 0.4))
			}
		}

		// Draw floors
		for i := range floorCount {
			col := rl.ColorLerp(rl.Fade(rl.RayWhite, PowF(shieldProgress, 0.33)), rl.White, SqrtF(shieldProgress))
			rl.DrawModel(floorModels[i], floorOrigins[i], 1.0, col)
			rl.DrawBoundingBox(floorBoundingBoxes[i], rl.Fade(rl.LightGray, 0.7))
		}

		// Draw enemy-box
		// rl.DrawCube(safeOrangeBoxPos, safeOrangeBoxSize.X, safeOrangeBoxSize.Y, safeOrangeBoxSize.Z, rl.Fade(rl.Orange, 1.0))
		// rl.DrawCubeWires(safeOrangeBoxPos, safeOrangeBoxSize.X, safeOrangeBoxSize.Y, safeOrangeBoxSize.Z, rl.Fade(rl.Orange, 1.0))
		for i := range safeOrangeBoxCount {
			pos := safeOrangeBoxPositions[i]
			size := safeOrangeBoxSizes[i]
			rl.DrawCubeV(pos, size, rl.Fade(rl.Orange, 1.0))
			rl.DrawCubeWiresV(pos, size, rl.Fade(rl.Gold, 1.0))
		}

		// Draw enemy-sphere
		// rl.DrawSphere(unsafeRedSpherePos, unsafeRedSphereSize, rl.Red)
		// rl.DrawSphereWires(unsafeRedSpherePos, unsafeRedSphereSize, 16/4, 16/2, rl.Red)

		for i := range unsafeRedSphereCount {
			rl.DrawSphere(unsafeRedSpherePositions[i], unsafeRedSphereSizes[i], rl.Red)
			rl.DrawSphereWires(unsafeRedSpherePositions[i], unsafeRedSphereSizes[i], 8, 8, rl.Maroon)
		}

		// Draw player
		playerRadius := playerSize.X / 2
		playerStartPos := rl.NewVector3(playerPosition.X, playerPosition.Y-playerSize.Y/2+playerRadius, playerPosition.Z)
		playerEndPos := rl.NewVector3(playerPosition.X, playerPosition.Y+playerSize.Y/2-playerRadius, playerPosition.Z)
		rl.DrawCubeV(playerPosition, playerSize, playerColor)
		rl.DrawCapsule(playerStartPos, playerEndPos, playerRadius, 16, 16, playerColor)
		rl.DrawCapsuleWires(playerStartPos, playerEndPos, playerRadius, 4, 6, rl.ColorLerp(playerColor, rl.Fade(rl.DarkGray, 0.8), 0.5))

		// Draw prop sphere
		if false {
			rl.DrawSphere(rl.NewVector3(0, -sphereModelRadius, -sphereModelRadius*2), sphereModelRadius-0.02, rl.Fade(rl.LightGray, 0.5))
			rl.DrawModelEx(sphereModel, rl.NewVector3(0, -sphereModelRadius, -sphereModelRadius*2), rl.NewVector3(0, -1, 0), float32(framesCounter), rl.NewVector3(1, 1, 1), rl.White)
		}

		if false {
			rl.DrawGrid(4*int32(MinF(arenaWidth, arenaLength)), 1/4.0)
		}

		rl.EndMode3D()

		// Draw HUD
		rl.DrawRectangle(10, 20, 100, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20}, rl.Vector2{X: fuelProgress * 100, Y: 20}, rl.DarkGray)
		rl.DrawText("Fuel", 10+5, 21, 20, rl.White)
		rl.DrawText(fmt.Sprintf("%.0f", fuelProgress*100), 90+5, 20+5*2, 10, rl.White)

		rl.DrawRectangle(10, 20+20, 100, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20 + 20}, rl.Vector2{X: shieldProgress * 100, Y: 20}, rl.DarkGray)
		rl.DrawText("Shield", 10+5, 21+20, 20, rl.White)
		rl.DrawText(fmt.Sprintf("%.0f", shieldProgress*100), 90+5, 20+20+5*2, 10, rl.White)

		rl.DrawFPS(10, int32(rl.GetScreenHeight())-25)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}

// From cmp.Ordered
// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
//
// Note that floating-point types may contain NaN ("not-a-number") values.
// An operator such as == or < will always report false when
// comparing a NaN value with any other value, NaN or not.
// See the [Compare] function for a consistent way to compare NaN values.
type OrderedNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// To avoid casting each time `OrderedNumber` interface it is used
type NumberType OrderedNumber

func MaxF[T NumberType](x T, y T) float32 { return float32(max(float64(x), float64(y))) }
func MinF[T NumberType](x T, y T) float32 { return float32(min(float64(x), float64(y))) }
func AbsF[T NumberType](x T) float32      { return float32(math.Abs(float64(x))) }
func SqrtF[T NumberType](x T) float32     { return float32(math.Sqrt(float64(x))) }
func PowF[T NumberType](x T, y T) float32 { return float32(math.Pow(float64(x), float64(y))) }
func SinF[T NumberType](x T) float32      { return float32(math.Sin(float64(x))) }
func CosF[T NumberType](x T) float32      { return float32(math.Cos(float64(x))) }

func manhattanVector2(a, b rl.Vector2) float32 { return AbsF(b.X-a.X) + AbsF(b.Y-a.Y) }
func manhattanVector3(a, b rl.Vector3) float32 { return AbsF(b.X-a.X) + AbsF(b.Y-a.Y) + AbsF(b.Z-a.Z) }

const (
	InvMathPhi         = 1 / math.Phi
	OneMinusInvMathPhi = 1 - InvMathPhi
)

