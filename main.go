package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"image/color"
	"log/slog"
	"math"
	"os"

	"github.com/gen2brain/raylib-go/easings"
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
		arenaWidth       = float32(10) * math.Phi // X
		arenaLength      = float32(10) * 1        // Z
		arenaWidthRatio  = (arenaWidth / (arenaWidth + arenaLength))
		arenaLengthRatio = (arenaLength / (arenaWidth + arenaLength))
		camPosW          = (arenaWidth * (math.Phi + arenaLengthRatio)) * (1 - OneMinusInvMathPhi)
		camPosL          = (arenaLength * (math.Phi + arenaWidthRatio)) * (1 - OneMinusInvMathPhi)
	)

	camScrollEase := float32((float32(1.0) / float32(fps)) * 2.0) // 0.033

	camera := rl.Camera{
		Position:   rl.NewVector3(0.0, RoundEvenF(camPosW), RoundEvenF(camPosL)),
		Target:     rl.NewVector3(0.0, -1.0, 0.0),
		Up:         rl.NewVector3(0.0, 1.0, 0.0),
		Fovy:       float32(cmp.Or(45.0, 60.0, 30.0)), // Use higher Fovy to zoom out if following a target
		Projection: rl.CameraPerspective,
	}

	// Save initial settings for stabilizing custom naïve camera movement
	defaultCameraPosition := camera.Position
	defaultCameraTarget := camera.Target
	defaultCameraPositionTargetVector := rl.Vector3Subtract(defaultCameraPosition, defaultCameraTarget)
	defaultCameraPositionTargetDistance := rl.Vector3Distance(defaultCameraPosition, defaultCameraTarget)

	const sphereModelRadius = arenaWidth / math.Phi
	sphereMesh := rl.GenMeshSphere(sphereModelRadius, 16, 16)
	sphereModel := rl.LoadModelFromMesh(sphereMesh)

	fuelProgress := float32(1.0)
	shieldProgress := float32(1.0)

	isPlayerBoost := false
	isPlayerStrafe := false
	playerColor := rl.RayWhite
	playerJumpsLeft := 1
	playerPosition := rl.NewVector3(0.0, 1.0, 2.0)
	playerSize := rl.NewVector3(1.0, 2.0, 1.0)
	playerVelocity := rl.Vector3{}
	playerAirTimer := float32(0)

	_ = isPlayerBoost
	_ = isPlayerStrafe

	maxPlayerAirTime := float32(fps) / 2.0
	maxPlayerFreefallAirTime := maxPlayerAirTime * 3
	const movementMagnitude = float32(0.2)
	const playerJumpVelocity = 4 // 3..5
	const terminalVelocityLimiterAirFriction = movementMagnitude / math.Phi
	const terminalVelocityLimiterAirFrictionY = movementMagnitude / 2
	const terminalVelocityY = 5

	// FEAT: See also https://github.com/Pakz001/Raylib-Examples/blob/master/ai/Example_-_Pattern_Movement.c
	// Like Arrow shooter crazyggame,,, fruit dispenser

	// MaxResourceSOACapacity is the hardcoded capacity limit of each batch
	// items for ease of development and to avoid runtime heap allocation.
	const MaxResourceSOACapacity = 16

	// ResourceSOA is a struct of arrays that holds game components.
	// TODO: use omit empty json struct tag
	type ResourceSOA struct { // size=6824 (0x1aa8)
		PlatformBoundingBoxes    [MaxResourceSOACapacity]rl.BoundingBox
		PlatformDefaultPositions [MaxResourceSOACapacity]rl.Vector3
		PlatformModels           [MaxResourceSOACapacity]rl.Model
		PlatformPositions        [MaxResourceSOACapacity]rl.Vector3
		PlatformSizes            [MaxResourceSOACapacity]rl.Vector3
		PlatformMovementNormals  [MaxResourceSOACapacity]rl.Vector3
		PlatformCount            int

		FloorBoundingBoxes [MaxResourceSOACapacity]rl.BoundingBox
		FloorPositions     [MaxResourceSOACapacity]rl.Vector3
		FloorModels        [MaxResourceSOACapacity]rl.Model
		FloorSizes         [MaxResourceSOACapacity]rl.Vector3
		FloorCount         int

		HealBoxPositions [MaxResourceSOACapacity]rl.Vector3
		HealBoxSizes     [MaxResourceSOACapacity]rl.Vector3
		HealBoxCount     int

		DamageSpherePositions [MaxResourceSOACapacity]rl.Vector3
		DamageSphereSizes     [MaxResourceSOACapacity]float32
		DamageSphereCount     int

		TrampolineBoxPositions [MaxResourceSOACapacity]rl.Vector3
		TrampolineBoxSizes     [MaxResourceSOACapacity]rl.Vector3
		TrampolineBoxCount     int
	}

	type Entity struct {
		Pos   rl.Vector3 `json:"pos"`
		Size  rl.Vector3 `json:"size"`
		Color color.RGBA `json:"color"`
	}

	var resource ResourceSOA

	for _, data := range []Entity{
		{
			Pos:  rl.NewVector3(-4, 1, 0),
			Size: rl.NewVector3(2, 2, 2),
		},
		{
			Pos:  rl.NewVector3(0, 1, -4),
			Size: rl.NewVector3(2, 2, 2),
		},
	} {
		resource.HealBoxPositions[resource.HealBoxCount] = data.Pos
		resource.HealBoxSizes[resource.HealBoxCount] = data.Size
		resource.HealBoxCount++
	}

	{
		resource.DamageSpherePositions[resource.DamageSphereCount] = rl.NewVector3(4, 0, 0)
		resource.DamageSphereSizes[resource.DamageSphereCount] = 1.5
		resource.DamageSphereCount++

		resource.DamageSpherePositions[resource.DamageSphereCount] = rl.NewVector3(0, 0, 6)
		resource.DamageSphereSizes[resource.DamageSphereCount] = 1.5
		resource.DamageSphereCount++
	}

	{
		resource.TrampolineBoxPositions[resource.TrampolineBoxCount] = rl.NewVector3(0, 3, 6)
		resource.TrampolineBoxSizes[resource.TrampolineBoxCount] = rl.NewVector3(2, 0.25, 2)
		resource.TrampolineBoxCount++

		resource.TrampolineBoxPositions[resource.TrampolineBoxCount] = rl.NewVector3(0, 1, -9)
		resource.TrampolineBoxSizes[resource.TrampolineBoxCount] = rl.NewVector3(2, 0.25, 2)
		resource.TrampolineBoxCount++
	}

	// Setup floors
	setupFloorResource := func(pos, size rl.Vector3) {
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := GetBoundingBoxFromPositionSizeV(pos, size)
		resource.FloorBoundingBoxes[resource.FloorCount] = box
		resource.FloorModels[resource.FloorCount] = model
		resource.FloorPositions[resource.FloorCount] = pos
		resource.FloorSizes[resource.FloorCount] = size
		resource.FloorCount++
	}
	const floorThick = 1.0
	for _, data := range []Entity{
		{
			Pos:  rl.NewVector3(0, (playerPosition.Y-playerSize.Y/2)-(floorThick/2), 0),
			Size: rl.NewVector3(arenaWidth, floorThick, arenaLength),
		},
		{
			Pos:  rl.NewVector3(arenaWidth/math.Phi, -arenaWidth*1, arenaLength/8),
			Size: rl.NewVector3(arenaWidth/2, floorThick, arenaLength/2),
		},
		{
			Pos:  rl.NewVector3(-3*arenaWidth/4, -(playerSize.Y * 1), (arenaLength/1)+(playerSize.Z*2)),
			Size: rl.NewVector3(arenaWidth/2, floorThick, arenaLength/2),
		},
		{
			Pos:  rl.NewVector3(3*arenaWidth/4, -arenaWidth/2, -4*arenaLength/3.5),
			Size: rl.NewVector3(arenaWidth/2, floorThick, arenaLength/2),
		},
		{
			Pos:  rl.NewVector3(-2*arenaWidth/3, -((arenaWidth / 2) + (playerSize.Y * 4)), (-arenaLength/2)+(playerSize.Z*2)),
			Size: rl.NewVector3(arenaWidth/2, floorThick, arenaLength/2),
		},
	} {
		setupFloorResource(data.Pos, data.Size)
	}

	// Setup moving platforms
	setupPlatformResource := func(pos, size, movementNormal rl.Vector3) {
		if !IsUnitVec3(movementNormal) {
			panic(fmt.Sprintf("Invalid unit vector: movementNormal: %+v", movementNormal))
		}
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := GetBoundingBoxFromPositionSizeV(pos, size)
		resource.PlatformBoundingBoxes[resource.PlatformCount] = box
		resource.PlatformDefaultPositions[resource.PlatformCount] = pos
		resource.PlatformModels[resource.PlatformCount] = model
		resource.PlatformPositions[resource.PlatformCount] = pos
		resource.PlatformSizes[resource.PlatformCount] = size
		resource.PlatformMovementNormals[resource.PlatformCount] = movementNormal // Up/Down
		resource.PlatformCount++
	}
	const maxPlatformMoveAmplitude = float32(arenaWidth / 2) // Distance traveled
	const platformThick = 1.0
	for _, data := range []struct {
		Entity         Entity
		MovementNormal rl.Vector3
	}{
		{
			Entity: Entity{
				Pos:  rl.NewVector3(0, -4, -20),
				Size: rl.NewVector3(4, platformThick, 4),
			},
			MovementNormal: rl.NewVector3(0, 0, 0), /* Static */
		},
		{
			Entity: Entity{
				Pos:  rl.NewVector3(0, 2, 0),
				Size: rl.NewVector3(4, platformThick, 4),
			},
			MovementNormal: rl.NewVector3(1, 0, 0),
		},
		{
			Entity: Entity{
				Pos:  rl.NewVector3(-8, 4, -8),
				Size: rl.NewVector3(4, platformThick, 4),
			},
			MovementNormal: rl.NewVector3(0, 1, 0),
		},
		{
			Entity: Entity{
				Pos:  rl.NewVector3(4, -8, -12),
				Size: rl.NewVector3(4, platformThick, 4),
			},
			MovementNormal: rl.NewVector3(0, 0, 1),
		},
	} {
		setupPlatformResource(data.Entity.Pos, data.Entity.Size, data.MovementNormal)
	}

	var (
		isFloorCollision      bool
		isTrampolineCollision bool // Boosts upwards if jumped on
		isOOBCollision        bool // Out of worl bounds (unimplemented)
		isPlatformCollision   bool // Movable platforms
		isSafeSpotCollision   bool // Add health
		isUnsafeCollision     bool // Reduce health
		isWallCollision       bool // TODO: Slide/Y-axis-barrier
	)

	framesCounter := 0

	handlePlayerJump := func(velocityY float32) {
		if velocityY <= 0 {
			panic(fmt.Sprintf("Jumps must have a positive upwards Y velocity. Got: %f", velocityY))
		}

		playerVelocity.Y = velocityY
		playerAirTimer = maxPlayerAirTime

		playerJumpsLeft--
		if playerJumpVelocity < 0 {
			playerJumpsLeft = 0
		}
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
				handlePlayerJump(playerJumpVelocity)
			}
		}
		if rl.IsWindowResized() {
			screenWidth = int32(rl.GetScreenWidth())
			screenHeight = int32(rl.GetScreenHeight())
		}

		// Update

		dt := rl.GetFrameTime()                // Same as 1/float32(fps) if fps was consistent
		const progressRateDecay = 0.08         // Slows down change in frame time
		progressRate := dt * progressRateDecay // Rate of change in this world for aethetic taste

		// HACK: Store previous position to reuse as next postion on collision (quick position resets)
		oldPlayerPos := playerPosition
		oldCamPos := camera.Position
		oldCamTarget := camera.Target

		// Follow player center
		const smooth = 0.034
		camScrollEase = dt * 2.0
		camScrollEase = MinF(camScrollEase, smooth)
		if isQuickMovementYAxis := false; isQuickMovementYAxis {
			camera.Target.X = rl.Lerp(oldCamTarget.X, playerPosition.X, camScrollEase)
			camera.Target.Y = rl.Lerp(oldCamTarget.Y, playerPosition.Y, smooth*2)
			camera.Target.Z = rl.Lerp(oldCamTarget.Z, playerPosition.Z, camScrollEase)
			camera.Position.X = rl.Lerp(camera.Position.X, camera.Target.X+defaultCameraPositionTargetVector.X, 0.5+camScrollEase)
			camera.Position.Y = rl.Lerp(camera.Position.Y, camera.Target.Y+defaultCameraPositionTargetVector.Y, 0.8+camScrollEase/2)
			camera.Position.Z = rl.Lerp(camera.Position.Z, camera.Target.Z+defaultCameraPositionTargetVector.Z, 0.5+camScrollEase)
		} else {
			camera.Position = rl.Vector3Add(camera.Target, defaultCameraPositionTargetVector)
			camera.Target = rl.Vector3Lerp(oldCamTarget, playerPosition, camScrollEase)
		}

		_ = oldPlayerPos
		_ = oldCamPos

		// Reset collision flags
		isFloorCollision = false
		isTrampolineCollision = false
		isOOBCollision = false
		isPlatformCollision = false
		isSafeSpotCollision = false
		isUnsafeCollision = false
		isWallCollision = false

		// Check gameover events and conditions
		if fuelProgress <= 0 {
			playerPosition = rl.Vector3Zero()
		}
		if shieldProgress <= 0 {
			playerPosition = rl.Vector3Zero()
			playerVelocity = rl.Vector3Zero()
			playerAirTimer = 0
		}

		// Update player movement
		playerAirTimer += 1.0

		// Normalize input vector to avoid speeding up diagonally
		if !rl.Vector3Equals(playerMovementThisFrame, rl.Vector3Zero()) {
			playerMovementThisFrame = rl.Vector3Normalize(playerMovementThisFrame) // Vector3Length (XZ): 1.414 --diagonal-> 0.99999994
			fuelProgress -= 0.05 / float32(fps)                                    // See also https://community.monogame.net/t/how-can-i-normalize-my-diagonal-movement/15276
		}

		frameMovement := rl.Vector3Add(playerMovementThisFrame, playerVelocity)

		playerPosition.X += frameMovement.X * movementMagnitude
		playerPosition.Y += frameMovement.Y * movementMagnitude
		playerPosition.Z += frameMovement.Z * movementMagnitude

		// Check if player is safely standing on the floor
		// NOTE: SHOULD REARRANGE ORDER OF COLLISION CHECKS FOR FLOOR AND PLATFORM AND PLAYER.COLLISIONSTHISFRAME.W/Y
		for i := range resource.FloorCount {
			if rl.CheckCollisionBoxes(resource.FloorBoundingBoxes[i], GetBoundingBoxFromPositionSizeV(playerPosition, playerSize)) {
				isFloorCollision = true
				if canFloorSink := false; canFloorSink { // Only push floor down if player just jumped and landed on the floor
					isJumpLandingOnFloor := playerAirTimer >= maxPlayerAirTime
					isFallingWithFloor := playerAirTimer > 0
					if isJumpLandingOnFloor || isFallingWithFloor {
						resource.FloorPositions[i].Y += playerVelocity.Y * terminalVelocityLimiterAirFrictionY
						resource.FloorBoundingBoxes[i].Min.Y += playerVelocity.Y * terminalVelocityLimiterAirFrictionY
						resource.FloorBoundingBoxes[i].Max.Y += playerVelocity.Y * terminalVelocityLimiterAirFrictionY
					}
				}
				playerPosition.Y = playerSize.Y/2 + resource.FloorBoundingBoxes[i].Max.Y // HACK: Allow player to stand on the floor
				playerCollisionsThisFrame.W = 1
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

		for i := range resource.PlatformCount {
			oldPos := resource.PlatformPositions[i]
			// Update platform across movement normal type
			var f float32
			switch resource.PlatformMovementNormals[i] { // Only positive values accepted to keep it simple
			case rl.Vector3{X: 1, Y: 0, Z: 0}:
				t := float32(framesCounter)                                                         // Current Time
				b := float32(resource.PlatformDefaultPositions[i].X + maxPlatformMoveAmplitude/2.0) // Top(Beginning)
				c := float32(-maxPlatformMoveAmplitude)                                             // Bottom(Change)
				d := float32(fps) * 4                                                               // Duration
				f = easings.SineInOut(t, b, c, d)
				resource.PlatformPositions[i].X = f
				resource.PlatformBoundingBoxes[i].Min.X = resource.PlatformPositions[i].X - resource.PlatformSizes[i].X/2
				resource.PlatformBoundingBoxes[i].Max.X = resource.PlatformPositions[i].X + resource.PlatformSizes[i].X/2
			case rl.Vector3{X: 0, Y: 1, Z: 0}:
				t := float32(framesCounter)                                                         // Current Time
				b := float32(resource.PlatformDefaultPositions[i].Y + maxPlatformMoveAmplitude/2.0) // Top(Beginning)
				c := float32(-maxPlatformMoveAmplitude)                                             // Bottom(Change)
				d := float32(fps) * 4                                                               // Duration
				f = easings.SineInOut(t, b, c, d)
				resource.PlatformPositions[i].Y = f
				resource.PlatformBoundingBoxes[i].Min.Y = resource.PlatformPositions[i].Y - resource.PlatformSizes[i].Y/2 // platformThick/2
				resource.PlatformBoundingBoxes[i].Max.Y = resource.PlatformPositions[i].Y + resource.PlatformSizes[i].Y/2 // platformThick/2
			case rl.Vector3{X: 0, Y: 0, Z: 1}:
				t := float32(framesCounter)                                                         // Current Time
				b := float32(resource.PlatformDefaultPositions[i].Z + maxPlatformMoveAmplitude/2.0) // Top(Beginning)
				c := float32(-maxPlatformMoveAmplitude)                                             // Bottom(Change)
				d := float32(fps) * 4                                                               // Duration
				f = easings.SineInOut(t, b, c, d)
				resource.PlatformPositions[i].Z = f
				resource.PlatformBoundingBoxes[i].Min.Z = resource.PlatformPositions[i].Z - resource.PlatformSizes[i].Z/2
				resource.PlatformBoundingBoxes[i].Max.Z = resource.PlatformPositions[i].Z + resource.PlatformSizes[i].Z/2
			}
			// Check collisions between platform and player
			box := resource.PlatformBoundingBoxes
			isInsideXRange := playerPosition.X+playerSize.X/2 < box[i].Max.X && playerPosition.X-playerSize.X/2 > box[i].Min.X
			isInsideZRange := playerPosition.Z+playerSize.Z/2 < box[i].Max.Z && playerPosition.Z-playerSize.Z/2 > box[i].Min.Z
			isAboveYRange := playerPosition.Y+playerSize.Y/2 >= box[i].Max.Y && playerPosition.Y-playerSize.Y/2 >= box[i].Min.Y
			didPlatformMovePlayer := false
			const tolerance = platformThick                        // Avoid spamming isPlatformCollision as pure player size
			if isInsideXRange && isInsideZRange && isAboveYRange { // ... calculation does not handle changing bound tolerance in the same loop
				isPlatformCollision = true
				if rl.CheckCollisionBoxes(
					GetBoundingBoxFromPositionSizeV(playerPosition, rl.Vector3AddValue(playerSize, tolerance)),
					box[i],
				) {
					// Handled standing on playtform for Y-axis
					didPlatformMovePlayer = true
					playerPosition.Y = playerSize.Y/2 + box[i].Max.Y
					playerCollisionsThisFrame.W = 1
				}
			} else {
				if isPassFromUnderOrTouchEdges := rl.CheckCollisionBoxes(
					GetBoundingBoxFromPositionSizeV(playerPosition, playerSize),
					box[i],
				); isPassFromUnderOrTouchEdges {
					isPlatformCollision = true
					didPlatformMovePlayer = true
					playerCollisionsThisFrame.Y = 1
					playerPosition.Y = playerSize.Y/2 + box[i].Max.Y
					playerCollisionsThisFrame.Y = 0
					playerCollisionsThisFrame.W = 1
				}
			}
			// Update player lateral position while standing on moving platform
			if didPlatformMovePlayer {
				switch resource.PlatformMovementNormals[i] {
				case rl.Vector3{X: 1, Y: 0, Z: 0}:
					delta := (resource.PlatformPositions[i].X - oldPos.X)
					playerPosition.X += delta
				case rl.Vector3{X: 0, Y: 0, Z: 1}:
					delta := (resource.PlatformPositions[i].Z - oldPos.Z)
					playerPosition.Z += delta
				}
			}
		}
		for i := range resource.DamageSphereCount {
			if rl.CheckCollisionBoxSphere(
				GetBoundingBoxFromPositionSizeV(playerPosition, playerSize),
				resource.DamageSpherePositions[i],
				resource.DamageSphereSizes[i],
			) {
				isUnsafeCollision = true
				// Find perpendicular curve to XZ plane, i.e slope of circumference
				// WARN: Expect wonky animation, as bottom of player box when on a slope of sphere, may not collide with top tangent to sphere surface.
				height := resource.DamageSphereSizes[i]/2 + playerSize.Y
				dx := CosF(AbsF(playerPosition.X - resource.DamageSpherePositions[i].X))
				dz := CosF(AbsF(playerPosition.Z - resource.DamageSpherePositions[i].Z))
				dy := (dx*dx + dz*dz) * height
				dy = SqrtF(dy)
				dy = rl.Clamp(dy, 0, height)
				playerPosition.Y = resource.DamageSpherePositions[i].Y + dy
				playerCollisionsThisFrame.W = 1
			}
		}
		for i := range resource.HealBoxCount {
			box := GetBoundingBoxFromPositionSizeV(resource.HealBoxPositions[i], resource.HealBoxSizes[i])
			if rl.CheckCollisionBoxes(box, GetBoundingBoxFromPositionSizeV(playerPosition, playerSize)) {
				playerCollisionsThisFrame.W = 1
				isSafeSpotCollision = true
				playerPosition.Y = playerSize.Y/2 + box.Max.Y // HACK: Allow player to stand on the floor
			}
		}
		for i := range resource.TrampolineBoxCount {
			box := GetBoundingBoxFromPositionSizeV(resource.TrampolineBoxPositions[i], resource.TrampolineBoxSizes[i])
			if rl.CheckCollisionBoxes(box, GetBoundingBoxFromPositionSizeV(playerPosition, playerSize)) {
				isTrampolineCollision = true
				playerPosition.Y = playerSize.Y/2 + box.Max.Y // HACK: Allow player to stand on the floor
				if playerAirTimer <= maxPlayerAirTime {       // Do not activate when stepped on
					playerCollisionsThisFrame.W = 1
				} else {
					handlePlayerJump(playerJumpVelocity * 8)
					playerJumpsLeft++
				}
			}
		}

		// Highlight player color on interactions with different world objects
		switch {
		case isFloorCollision:
			playerColor = rl.Black
		case isOOBCollision:
			playerColor = rl.DarkGray
		case isPlatformCollision:
			playerColor = rl.Blue
		case isSafeSpotCollision:
			playerColor = rl.Lime
		case isTrampolineCollision:
			playerColor = rl.Maroon
		case isUnsafeCollision:
			playerColor = rl.Orange
		case isWallCollision:
			playerColor = rl.Fade(rl.Brown, 0.9)
		default: // Air-Time
			playerColor = rl.Fade(rl.Black, 0.8)
		}

		// Update player entity collisions
		// Entity: Update velocity
		playerVelocity.Y = MinF(terminalVelocityY, playerVelocity.Y-terminalVelocityLimiterAirFrictionY) // Decelerate if in air
		if playerCollisionsThisFrame.Y == 1 || playerCollisionsThisFrame.W == 1 {
			playerVelocity.Y = 0
		}
		if playerCollisionsThisFrame.W == 1 {
			playerAirTimer = 0
			playerJumpsLeft = 1
		}
		if playerAirTimer > maxPlayerAirTime*math.Phi && playerAirTimer < maxPlayerAirTime*math.Phi*math.Phi {
			playerVelocity.Y -= terminalVelocityLimiterAirFrictionY // Entity: Snappy bouncy jumps (Once)
		}

		// Update progress on collision, air-time, ...
		if isUnsafeCollision {
			shieldProgress -= progressRate
		}
		if isSafeSpotCollision {
			shieldProgress += progressRate * 2
			if shieldProgress >= 1.0 {
				shieldProgress = 1.0
			}
			fuelProgress += progressRate * 2
			if fuelProgress >= 1.0 {
				fuelProgress = 1.0
			}
		}
		if playerAirTimer > maxPlayerFreefallAirTime {
			shieldProgress -= PowF(progressRate*shieldProgress, maxPlayerFreefallAirTime/playerAirTimer)
		}
		if isDebugAllCollisions := false; isDebugAllCollisions {
			if (playerCollisionsThisFrame.X == 1 || playerCollisionsThisFrame.Z == 1) &&
				(!isOOBCollision && !isSafeSpotCollision && !isFloorCollision && !isPlatformCollision && !isWallCollision) {
				shieldProgress -= progressRate
			}
		}

		// Increment global frames counter tracker
		framesCounter++

		// Draw

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)

		// Draw interactive game resource objects
		for i := range resource.FloorCount {
			col := rl.ColorLerp(rl.Fade(rl.RayWhite, PowF(shieldProgress, 0.33)), rl.White, SqrtF(shieldProgress))
			rl.DrawModel(resource.FloorModels[i], resource.FloorPositions[i], 1.0, col)
			rl.DrawBoundingBox(resource.FloorBoundingBoxes[i], rl.Fade(rl.Black, 0.3))
		}
		for i := range resource.PlatformCount {
			rl.DrawModel(resource.PlatformModels[i], resource.PlatformPositions[i], 1.0, rl.SkyBlue)                                // Platform
			rl.DrawBoundingBox(resource.PlatformBoundingBoxes[i], rl.DarkBlue)                                                      // Platform outline
			rl.DrawCubeV(resource.PlatformDefaultPositions[i], rl.NewVector3(0.125, maxPlatformMoveAmplitude, 0.125), rl.LightGray) // Reference (y axis)
			rl.DrawPlane(resource.PlatformDefaultPositions[i], rl.NewVector2(0.5, 0.5), rl.Gray)                                    // Reference (midpoint plane)
		}
		for i := range resource.DamageSphereCount {
			rl.DrawSphere(resource.DamageSpherePositions[i], resource.DamageSphereSizes[i], rl.Gold)
			rl.DrawSphereWires(resource.DamageSpherePositions[i], resource.DamageSphereSizes[i], 8, 8, rl.Orange)
		}
		for i := range resource.HealBoxCount {
			rl.DrawCubeV(resource.HealBoxPositions[i], resource.HealBoxSizes[i], rl.Fade(rl.Green, 1.0))
			rl.DrawCubeWiresV(resource.HealBoxPositions[i], resource.HealBoxSizes[i], rl.Fade(rl.DarkGreen, 1.0))
		}
		for i := range resource.TrampolineBoxCount {
			rl.DrawCubeV(resource.TrampolineBoxPositions[i], resource.TrampolineBoxSizes[i], rl.Fade(rl.Red, 1.0))
			rl.DrawCubeWiresV(resource.TrampolineBoxPositions[i], resource.TrampolineBoxSizes[i], rl.Fade(rl.Maroon, 1.0))
		}

		// Draw player
		playerRadius := playerSize.X / 2
		playerStartPos := rl.NewVector3(playerPosition.X, playerPosition.Y-playerSize.Y/2+playerRadius, playerPosition.Z)
		playerEndPos := rl.NewVector3(playerPosition.X, playerPosition.Y+playerSize.Y/2-playerRadius, playerPosition.Z)
		rl.DrawCapsule(playerStartPos, playerEndPos, playerRadius, 16, 16, playerColor)
		rl.DrawCapsuleWires(playerStartPos, playerEndPos, playerRadius, 4, 6, rl.ColorLerp(playerColor, rl.Fade(rl.DarkGray, 0.8), 0.5))
		if isDebug := false; isDebug {
			rl.DrawCubeV(playerPosition, playerSize, playerColor)
			rl.DrawCubeWiresV(playerPosition, playerSize, playerColor)
		}

		// Draw destination prop sphere
		if true {
			centerPos := rl.NewVector3(0, -sphereModelRadius-arenaWidth*(8/4), -sphereModelRadius*2-arenaLength*3)
			rl.DrawSphere(centerPos, sphereModelRadius-0.02, rl.Fade(rl.LightGray, 0.5))
			rl.DrawModelEx(sphereModel, centerPos, rl.NewVector3(0, -1, 0), float32(framesCounter), rl.NewVector3(1, 1, 1), rl.White)
		}

		if false {
			rl.DrawGrid(int32(MinF(arenaWidth, arenaLength)), 1)
		}

		// Draw XYZ origin
		if true {
			rl.DrawCubeV(rl.Vector3Zero(), rl.NewVector3(8.0, 0.1, 0.1), rl.Red)
			rl.DrawCubeV(rl.Vector3Zero(), rl.NewVector3(0.1, 8.0, 0.1), rl.Green)
			rl.DrawCubeV(rl.Vector3Zero(), rl.NewVector3(0.1, 0.1, 8.0), rl.Blue)
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

		// Quick debug zone
		{
			text := fmt.Sprintf("playerAirTimer: %.2f\nplayerJumpsLeft: %d\n"+
				"playerPosition: %.2f\n"+
				"camera.Position: %.2f\ncamera.Target: %.2f\ncameraPositionTargetDistance: %.2f\n"+
				"cameraScrollEase: %.4f\n"+
				"defaultCameraPosition: %.2f\ndefaultCameraTarget: %.2f\ndefaultCameraPositionTargetDistance: %.2f\n",
				playerAirTimer, playerJumpsLeft,
				playerPosition,
				camera.Position, camera.Target, rl.Vector3Distance(camera.Position, camera.Target),
				camScrollEase,
				defaultCameraPosition, defaultCameraTarget, defaultCameraPositionTargetDistance,
			)
			rl.DrawText(text, int32(rl.GetScreenWidth())-10-rl.MeasureText(text, 10), 10, 10, rl.DarkGray)
		}

		rl.EndDrawing()
	}

	if false {
		jsonData, err := json.Marshal(resource) // jsonData, err := json.MarshalIndent(resource, "", " ") // Debug
		if err != nil {
			slog.Error(err.Error())
		}
		if true {
			fmt.Printf("jsonData: %+s\n", jsonData)
		}
		if true {
			if err := os.WriteFile("resource_assets.json", jsonData, 0644); err != nil {
				slog.Error(err.Error())
			}
		}
	}

	rl.CloseWindow()
}

func GetBoundingBoxFromPositionSizeV(pos, size rl.Vector3) rl.BoundingBox {
	return rl.NewBoundingBox(
		rl.NewVector3(pos.X-size.X/2, pos.Y-size.Y/2, pos.Z-size.Z/2),
		rl.NewVector3(pos.X+size.X/2, pos.Y+size.Y/2, pos.Z+size.Z/2),
	)
}

// Copied from Go's cmp.Ordered
// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// Note that floating-point types may contain NaN ("not-a-number") values.
// An operator such as == or < will always report false when
// comparing a NaN value with any other value, NaN or not.
// See the [Compare] function for a consistent way to compare NaN values.
type OrderedNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// NumberType typecast to avoid casting `OrderedNumber` interface when used.
type NumberType OrderedNumber

func AbsF[T NumberType](x T) float32       { return float32(math.Abs(float64(x))) }
func SqrtF[T NumberType](x T) float32      { return float32(math.Sqrt(float64(x))) }
func CosF[T NumberType](x T) float32       { return float32(math.Cos(float64(x))) }
func SinF[T NumberType](x T) float32       { return float32(math.Sin(float64(x))) }
func SignF[T NumberType](x T) float32      { return cmp.Or(float32(math.Abs(float64(x))/float64(x)), 0) }
func FloorF[T NumberType](x T) float32     { return float32(math.Floor(float64(x))) }
func CeilF[T NumberType](x T) float32      { return float32(math.Ceil(float64(x))) }
func RoundI[T NumberType](x T) int32       { return int32(math.Round(float64(x))) }
func RoundF[T NumberType](x T) float32     { return float32(math.Round(float64(x))) }
func RoundEvenF[T NumberType](x T) float32 { return float32(math.RoundToEven(float64(x))) }

func MaxF[T NumberType](x T, y T) float32 { return float32(max(float64(x), float64(y))) }
func MinF[T NumberType](x T, y T) float32 { return float32(min(float64(x), float64(y))) }
func PowF[T NumberType](x T, y T) float32 { return float32(math.Pow(float64(x), float64(y))) }
func MaxI[T NumberType](x T, y T) int32   { return int32(max(float64(x), float64(y))) }
func MinI[T NumberType](x T, y T) int32   { return int32(min(float64(x), float64(y))) }

func manhattanV2(a, b rl.Vector2) float32 { return AbsF(b.X-a.X) + AbsF(b.Y-a.Y) }
func manhattanV3(a, b rl.Vector3) float32 { return AbsF(b.X-a.X) + AbsF(b.Y-a.Y) + AbsF(b.Z-a.Z) }

var (
	Vector3OneLength = rl.Vector3Length(rl.Vector3One())
	Vector2OneLength = rl.Vector2Length(rl.Vector2One())
)

func IsUnitVec3(v rl.Vector3) bool { return rl.Vector3Length(v) <= Vector3OneLength }
func IsUnitVec2(v rl.Vector2) bool { return rl.Vector2Length(v) <= Vector2OneLength }

const (
	InvMathPhi         = 1 / math.Phi
	OneMinusInvMathPhi = 1 - InvMathPhi
)
