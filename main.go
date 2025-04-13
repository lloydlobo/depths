package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"

	"github.com/gen2brain/raylib-go/easings"
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
		arenaWidth       = float32(15) * math.Phi // X
		arenaLength      = float32(15) * 1        // Z
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
	type ResourceSOA struct { // size=6824 (0x1aa8)
		PlatformBoundingBoxes    [MaxResourceSOACapacity]rl.BoundingBox
		PlatformDefaultPositions [MaxResourceSOACapacity]rl.Vector3
		PlatformModels           [MaxResourceSOACapacity]rl.Model
		PlatformPositions        [MaxResourceSOACapacity]rl.Vector3
		PlatformSizes            [MaxResourceSOACapacity]rl.Vector3
		PlatformMoveNormals      [MaxResourceSOACapacity]rl.Vector3
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

	var resource ResourceSOA

	{
		resource.HealBoxPositions[resource.HealBoxCount] = rl.NewVector3(-4, 1, 0)
		resource.HealBoxSizes[resource.HealBoxCount] = rl.NewVector3(2, 2, 2)
		resource.HealBoxCount++
	}
	{
		resource.HealBoxPositions[resource.HealBoxCount] = rl.NewVector3(0, 1, -4)
		resource.HealBoxSizes[resource.HealBoxCount] = rl.NewVector3(2, 2, 2)
		resource.HealBoxCount++
	}

	{
		resource.DamageSpherePositions[resource.DamageSphereCount] = rl.NewVector3(4, 0, 0)
		resource.DamageSphereSizes[resource.DamageSphereCount] = 1.5
		resource.DamageSphereCount++
	}
	{
		resource.DamageSpherePositions[resource.DamageSphereCount] = rl.NewVector3(0, 0, 6)
		resource.DamageSphereSizes[resource.DamageSphereCount] = 1.5
		resource.DamageSphereCount++
	}

	{
		resource.TrampolineBoxPositions[resource.TrampolineBoxCount] = rl.NewVector3(0, 3, 6)
		resource.TrampolineBoxSizes[resource.TrampolineBoxCount] = rl.NewVector3(2, 0.25, 2)
		resource.TrampolineBoxCount++
	}
	{
		resource.TrampolineBoxPositions[resource.TrampolineBoxCount] = rl.NewVector3(0, 1, -9)
		resource.TrampolineBoxSizes[resource.TrampolineBoxCount] = rl.NewVector3(2, 0.25, 2)
		resource.TrampolineBoxCount++
	}

	// anticlockwise: 0 -> 1 -> 2 -> 3 TLBR
	const wallThick = 1 / 2.0
	walls := [4]rl.BoundingBox{}
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

	// Setup moving platforms
	const maxPlatformTravelAmplitude = float32(arenaWidth / 2) // Distance traveled
	const platformThick = 0.25 * 4
	{
		origin := rl.NewVector3(2*arenaWidth/3, 0, -arenaLength/2)
		size := rl.NewVector3(CeilF(playerSize.X*PowF(math.Phi, 4)), platformThick, CeilF(playerSize.Z*PowF(math.Phi, 4)))
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-size.X/2, origin.Y-size.Y/2, origin.Z-size.Z/2),
			rl.NewVector3(origin.X+size.X/2, origin.Y+size.Y/2, origin.Z+size.Z/2))
		resource.PlatformBoundingBoxes[resource.PlatformCount] = box
		resource.PlatformDefaultPositions[resource.PlatformCount] = origin
		resource.PlatformModels[resource.PlatformCount] = model
		resource.PlatformPositions[resource.PlatformCount] = origin
		resource.PlatformSizes[resource.PlatformCount] = size
		resource.PlatformMoveNormals[resource.PlatformCount] = rl.NewVector3(0, 1, 0) // Up/Down
		resource.PlatformCount++
	}
	{
		origin := rl.NewVector3(2*arenaWidth/3, 0, -arenaLength/2*2)
		size := rl.NewVector3(CeilF(playerSize.X*PowF(math.Phi, 4)), platformThick, CeilF(playerSize.Z*PowF(math.Phi, 4)))
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-size.X/2, origin.Y-size.Y/2, origin.Z-size.Z/2),
			rl.NewVector3(origin.X+size.X/2, origin.Y+size.Y/2, origin.Z+size.Z/2))
		resource.PlatformBoundingBoxes[resource.PlatformCount] = box
		resource.PlatformDefaultPositions[resource.PlatformCount] = origin
		resource.PlatformModels[resource.PlatformCount] = model
		resource.PlatformPositions[resource.PlatformCount] = origin
		resource.PlatformSizes[resource.PlatformCount] = size
		resource.PlatformMoveNormals[resource.PlatformCount] = rl.NewVector3(0, 0, 1) // Front/Back
		resource.PlatformCount++
	}

	// Setup floors
	const floorThick = 0.5 * 2
	{
		origin := rl.NewVector3(0, (playerPosition.Y-playerSize.Y/2)-(floorThick/2), 0)
		size := rl.NewVector3(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-size.X/2, origin.Y-size.Y/2, origin.Z-size.Z/2),
			rl.NewVector3(origin.X+size.X/2, origin.Y+size.Y/2, origin.Z+size.Z/2),
		)
		resource.FloorBoundingBoxes[resource.FloorCount] = box
		resource.FloorModels[resource.FloorCount] = model
		resource.FloorPositions[resource.FloorCount] = origin
		resource.FloorSizes[resource.FloorCount] = size
		resource.FloorCount++
	}
	{
		origin := rl.NewVector3(arenaWidth/math.Phi, (playerPosition.Y-playerSize.Y/2)-(floorThick/2)-arenaWidth*1, arenaLength/8)
		size := rl.NewVector3(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)
		resource.FloorBoundingBoxes[resource.FloorCount] = box
		resource.FloorModels[resource.FloorCount] = model
		resource.FloorPositions[resource.FloorCount] = origin
		resource.FloorSizes[resource.FloorCount] = size
		resource.FloorCount++
	}
	{
		origin := rl.NewVector3(-3*arenaWidth/4, (playerPosition.Y-playerSize.Y/2)-(floorThick/2)-(playerSize.Y*1), (arenaLength/1)+(playerSize.Z*2))
		size := rl.NewVector3(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)
		resource.FloorBoundingBoxes[resource.FloorCount] = box
		resource.FloorModels[resource.FloorCount] = model
		resource.FloorPositions[resource.FloorCount] = origin
		resource.FloorSizes[resource.FloorCount] = size
		resource.FloorCount++
	}
	{
		origin := rl.NewVector3(3*arenaWidth/4, (playerPosition.Y-playerSize.Y/2-floorThick/2)-arenaWidth/2, -4*arenaLength/3.5)
		size := rl.NewVector3(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)
		resource.FloorBoundingBoxes[resource.FloorCount] = box
		resource.FloorModels[resource.FloorCount] = model
		resource.FloorPositions[resource.FloorCount] = origin
		resource.FloorSizes[resource.FloorCount] = size
		resource.FloorCount++
	}
	{
		origin := rl.NewVector3(-2*arenaWidth/3, (playerPosition.Y-playerSize.Y/2)-(floorThick/2)-((arenaWidth/2)+(playerSize.Y*4)), (-arenaLength/2)+(playerSize.Z*2))
		size := rl.NewVector3(arenaWidth, floorThick, arenaLength)
		model := rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		box := rl.NewBoundingBox(
			rl.NewVector3(origin.X-arenaWidth/2, origin.Y-floorThick/2, origin.Z-arenaLength/2),
			rl.NewVector3(origin.X+arenaWidth/2, origin.Y+floorThick/2, origin.Z+arenaLength/2),
		)
		resource.FloorBoundingBoxes[resource.FloorCount] = box
		resource.FloorModels[resource.FloorCount] = model
		resource.FloorPositions[resource.FloorCount] = origin
		resource.FloorSizes[resource.FloorCount] = size
		resource.FloorCount++
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

				// Only push floor down if player just jumped and landed on the floor
				if isFloorSinking := false; isFloorSinking {
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
			if isMovePlatformVerticaly := true; isMovePlatformVerticaly {
				t := float32(framesCounter)                                                           // Current Time
				b := float32(resource.PlatformDefaultPositions[i].Y + maxPlatformTravelAmplitude/2.0) // Top(Beginning)
				c := float32(-maxPlatformTravelAmplitude)                                             // Bottom(Change)
				d := float32(fps) * 4                                                                 // Duration
				dy := easings.SineInOut(t, b, c, d)
				resource.PlatformPositions[i].Y = dy
				resource.PlatformBoundingBoxes[i].Min.Y = resource.PlatformPositions[i].Y - platformThick/2
				resource.PlatformBoundingBoxes[i].Max.Y = resource.PlatformPositions[i].Y + platformThick/2
			}
			box := resource.PlatformBoundingBoxes
			isInsideXRange := playerPosition.X+playerSize.X/2 < box[i].Max.X && playerPosition.X-playerSize.X/2 > box[i].Min.X
			isInsideZRange := playerPosition.Z+playerSize.Z/2 < box[i].Max.Z && playerPosition.Z-playerSize.Z/2 > box[i].Min.Z
			isAboveYRange := playerPosition.Y+playerSize.Y/2 >= box[i].Max.Y && playerPosition.Y-playerSize.Y/2 >= box[i].Min.Y
			const tolerance = platformThick // Avoid spamming isPlatformCollision as pure player size calculation does not handle changing bound tolerance in the same loop
			if isInsideXRange && isInsideZRange && isAboveYRange {
				isPlatformCollision = true
				if rl.CheckCollisionBoxes(
					GetBoundingBoxFromPositionSizeV(playerPosition, rl.Vector3AddValue(playerSize, tolerance)),
					box[i],
				) {
					playerPosition.Y = playerSize.Y/2 + box[i].Max.Y
					playerCollisionsThisFrame.W = 1
				}
			} else if isPassFromUnderOrTouchEdges := rl.CheckCollisionBoxes(
				GetBoundingBoxFromPositionSizeV(playerPosition, playerSize),
				box[i],
			); isPassFromUnderOrTouchEdges {
				isPlatformCollision = true
				playerCollisionsThisFrame.Y = 1
				playerPosition.Y = playerSize.Y/2 + box[i].Max.Y
				playerCollisionsThisFrame.Y = 0
				playerCollisionsThisFrame.W = 1
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
		if false {
			for i := range walls {
				if !isWallCollision && rl.CheckCollisionBoxes(rl.NewBoundingBox(
					rl.NewVector3(playerPosition.X-playerSize.X/2, playerPosition.Y-playerSize.Y/2, playerPosition.Z-playerSize.Z/2),
					rl.NewVector3(playerPosition.X+playerSize.X/2, playerPosition.Y+playerSize.Y/2, playerPosition.Z+playerSize.Z/2)), walls[i]) {
					isWallCollision = true
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
			rl.DrawModel(resource.PlatformModels[i], resource.PlatformPositions[i], 1.0, rl.SkyBlue)                                  // Platform
			rl.DrawBoundingBox(resource.PlatformBoundingBoxes[i], rl.DarkBlue)                                                        // Platform outline
			rl.DrawCubeV(resource.PlatformDefaultPositions[i], rl.NewVector3(0.125, maxPlatformTravelAmplitude, 0.125), rl.LightGray) // Reference (y axis)
			rl.DrawPlane(resource.PlatformDefaultPositions[i], rl.NewVector2(0.5, 0.5), rl.Gray)                                      // Reference (midpoint plane)
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
		if false {
			// Draw walls
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

	{
		jsonData, err := json.MarshalIndent(resource, "", "  ")
		if err != nil {
			slog.Error(err.Error())
		}
		if false {
			fmt.Printf("jsonData: %b\n", jsonData)
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

const (
	InvMathPhi         = 1 / math.Phi
	OneMinusInvMathPhi = 1 - InvMathPhi
)
