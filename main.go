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

	rl.SetConfigFlags(rl.FlagMsaa4xHint | rl.FlagWindowResizable) // Config flags must be set before InitWindow

	rl.InitWindow(screenWidth, screenHeight, "raylib [models] example - box collisions") // Initialize Window and OpenGL Graphics

	rl.SetWindowState(rl.FlagVsyncHint | rl.FlagInterlacedHint | rl.FlagWindowHighdpi) // Window state must be set after InitWindow
	rl.SetWindowMinSize(800, 450)                                                      // Prevents my window manager shrinking this to 2x1 units window size

	const (
		playerSizeY      = 2.0
		arenaW           = float32(playerSizeY * 3)  // X
		arenaL           = float32(playerSizeY * 3)  // Z
		arenaH           = float32(playerSizeY * 12) // Y (For reference of screen)
		floorThick       = float32(playerSizeY)      // NOTE: Easier to move vertically between platforms if thicker
		arenaWidthRatio  = float32(arenaW / (arenaW + arenaL))
		arenaLengthRatio = float32(arenaL / (arenaW + arenaL))
		arenaWallHeight  = float32(1)
		camPosW          = float32(arenaW*(Phi+arenaLengthRatio)) * (1 - OneMinusInvMathPhi)
		camPosL          = float32(arenaL*(Phi+arenaWidthRatio)) * (1 - OneMinusInvMathPhi)
	)

	var (
		mouseRay          rl.Ray
		rayMouseCollision rl.RayCollision

		playerCameraRay          rl.Ray
		playerCameraRayCollision rl.RayCollision
	)

	isPlayerBoost := false
	isPlayerStrafe := false
	playerColor := rl.RayWhite
	playerJumpsLeft := 1
	playerPosition := rl.NewVector3(0.0, 1.0, 2.0)
	playerSize := rl.NewVector3(1.0, playerSizeY, 1.0)
	playerVelocity := rl.Vector3{}
	playerAirTimer := float32(0)
	playerRotationNormal := rl.NewVector3(0, -1, 0)
	playerRotation := rl.NewVector4(0, 0, 0, 0)
	playerModel := rl.LoadModelFromMesh(rl.GenMeshCube(playerSize.X/2, playerSize.Y/2, playerSize.Z/2))

	camScrollEase := float32((float32(1.0) / float32(fps)) * 2.0) // 0.033

	camera := rl.Camera{
		Position: cmp.Or(
			rl.NewVector3(0., 20., 20.),
			rl.NewVector3(0., camPosW/2, camPosW*Phi),
			rl.NewVector3(0., 10., 10.),
		),
		Target:     rl.NewVector3(0., -1., 0.),
		Up:         rl.NewVector3(0., 1., 0.),
		Fovy:       float32(cmp.Or(60., 30., 45.)),
		Projection: rl.CameraPerspective,
	}

	// Save initial settings for stabilizing custom naïve camera movement
	defaultCameraPosition := camera.Position
	defaultCameraTarget := camera.Target
	defaultCameraPositionTargetVector := rl.Vector3Subtract(defaultCameraPosition, defaultCameraTarget)
	_ = defaultCameraPositionTargetVector
	defaultCameraPositionTargetDistance := rl.Vector3Distance(defaultCameraPosition, defaultCameraTarget)

	// Copied from https://github.com/lloydlobo/ChunkMinerGame/blob/main/Src/Drill.odin
	type UpgradeType uint8
	const (
		UpgradeNone UpgradeType = iota
		UpgradeFuel
		UpgradeHull
		UpgradeSpeed
		UpgradePrimary
		UpgradeCargo
	)
	type UpgradeProp struct {
		Price int32
		Value int32
	}
	const upgradePropCap = 5
	var UpgradeProps = make(map[UpgradeType][upgradePropCap]UpgradeProp)
	UpgradeProps[UpgradeNone] = [upgradePropCap]UpgradeProp{}
	UpgradeProps[UpgradeFuel] = [upgradePropCap]UpgradeProp{
		{Price: 0, Value: 100},
		{Price: 200, Value: 200},
		{Price: 4000, Value: 400},
		{Price: 25000, Value: 8000},
		{Price: 50000, Value: 1600},
	}
	UpgradeProps[UpgradeHull] = [upgradePropCap]UpgradeProp{
		{Price: 0, Value: 100},
		{Price: 400, Value: 200},
		{Price: 8000, Value: 400},
		{Price: 25000, Value: 800},
		{Price: 50000, Value: 1600},
	}
	UpgradeProps[UpgradeSpeed] = [upgradePropCap]UpgradeProp{
		// Value is move cooldown in milliseconds
		{Price: 0, Value: 220},
		{Price: 250, Value: 200},
		{Price: 8000, Value: 180},
		{Price: 25000, Value: 150},
		{Price: 50000, Value: 120},
	}
	UpgradeProps[UpgradePrimary] = [upgradePropCap]UpgradeProp{
		// Value is drill time in milliseconds to remove one voxel health
		// Additionally at least Upgrade 1 is required to drill ROCK and Upgrda2 for ROCK2
		{Price: 0, Value: 400},
		{Price: 250, Value: 250},
		{Price: 8000, Value: 100},
		{Price: 20000, Value: 50},
		{Price: 80000, Value: 25},
	}
	UpgradeProps[UpgradeCargo] = [upgradePropCap]UpgradeProp{
		{Price: 0, Value: 10},
		{Price: 200, Value: 20},
		{Price: 400, Value: 60},
		{Price: 800, Value: 80},
		{Price: 1000, Value: 100},
	}
	// For balancing purposes we want more than one fuel per dollar.
	// That would require a PRICE_PER_UNIT of <1 which we can't do with integers.
	// Instead we just invert the meaning of the unit. UNIT_PER_DOLLAR sounds a bit weird but works.
	type PropPrice int32
	const (
		fuelUnitPerDollar    = PropPrice(3)
		repairUnitPerDollar  = PropPrice(2)
		horizontalClimbPrice = PropPrice(20000)
	)

	// See https://ldjam.com/events/ludum-dare/57/depthshift
	// Controls (Keyboard & Mouse / Xbox Controller):
	//
	// - Move => WASD / Left Stick
	// - Look => Mouse move / Right Stick
	// - Jump => Space / A
	// - Pickup / Drop / Interact => Left Mouse / X
	// - Shoot => Right Mouse / B
	// - Change Focus depth => Scroll Wheel / LB & RB
	// - Pause Menu => Escape / Start

	fuelProgress := float32(1.0)
	shieldProgress := float32(1.0)

	_ = isPlayerBoost
	_ = isPlayerStrafe

	maxPlayerAirTime := float32(fps) / 2.0
	maxPlayerFreefallAirTime := maxPlayerAirTime * 3
	const movementMagnitude = float32(0.2)
	const playerJumpVelocity = 4 // 3..5
	const terminalVelocityLimiterAirFriction = movementMagnitude / Phi
	const terminalVelocityLimiterAirFrictionY = movementMagnitude / 2
	const terminalVelocityY = 5

	// FEAT: See also https://github.com/Pakz001/Raylib-Examples/blob/master/ai/Example_-_Pattern_Movement.c
	// Like Arrow shooter crazyggame,,, fruit dispenser

	// MaxResourceSOACapacity is the hardcoded capacity limit of each batch
	// items for ease of development and to avoid runtime heap allocation.
	const MaxResourceSOACapacity = 32

	// GameResourceSOA is a struct of arrays that holds game components.
	// TODO: use omit empty json struct tag
	type GameResourceSOA struct { // size=6824 (0x1aa8)
		PlatformBoundingBoxes      [MaxResourceSOACapacity]rl.BoundingBox
		PlatformDefaultPositions   [MaxResourceSOACapacity]rl.Vector3
		PlatformModels             [MaxResourceSOACapacity]rl.Model
		PlatformPositions          [MaxResourceSOACapacity]rl.Vector3
		PlatformSizes              [MaxResourceSOACapacity]rl.Vector3
		PlatformMovementNormals    [MaxResourceSOACapacity]rl.Vector3
		PlatformMovementAmplitudes [MaxResourceSOACapacity]float32 // World units maxPlatformMoveAmplitude
		PlatformAtIsActive         [MaxResourceSOACapacity]bool
		PlatformCount              int

		FloorBoundingBoxes [MaxResourceSOACapacity]rl.BoundingBox
		FloorPositions     [MaxResourceSOACapacity]rl.Vector3
		FloorModels        [MaxResourceSOACapacity]rl.Model
		FloorSizes         [MaxResourceSOACapacity]rl.Vector3
		FloorAtIsActive    [MaxResourceSOACapacity]bool
		FloorCount         int

		HealBoxPositions  [MaxResourceSOACapacity]rl.Vector3
		HealBoxSizes      [MaxResourceSOACapacity]rl.Vector3
		HealBoxAtIsActive [MaxResourceSOACapacity]bool
		HealBoxCount      int

		DamageSpherePositions  [MaxResourceSOACapacity]rl.Vector3
		DamageSphereSizes      [MaxResourceSOACapacity]float32
		DamageSphereAtIsActive [MaxResourceSOACapacity]bool
		DamageSphereCount      int

		TrampolineBoxPositions  [MaxResourceSOACapacity]rl.Vector3
		TrampolineBoxSizes      [MaxResourceSOACapacity]rl.Vector3
		TrampolineBoxAtIsActive [MaxResourceSOACapacity]bool
		TrampolineBoxCount      int
	}

	type Entity struct {
		Pos  rl.Vector3 `json:"pos"`
		Size rl.Vector3 `json:"size"`
	}

	var resource GameResourceSOA

	// Vertical Slice (Level prototyping)

	// Setup floors
	setupFloorResource := func(pos, size rl.Vector3) {
		resource.FloorBoundingBoxes[resource.FloorCount] = GetBoundingBoxFromPositionSizeV(pos, size)
		resource.FloorModels[resource.FloorCount] = rl.LoadModelFromMesh(rl.GenMeshCube(size.X, size.Y, size.Z))
		resource.FloorPositions[resource.FloorCount] = pos
		resource.FloorSizes[resource.FloorCount] = size
		resource.FloorAtIsActive[resource.FloorCount] = true
		resource.FloorCount++
	}
	playerStartPosY := (playerPosition.Y - playerSize.Y/2)
	padH := playerStartPosY - (floorThick / 2)
	_ = padH

	{
		offset := float32(InvMathPhi - OneMinusInvMathPhi)
		// Layout inspired by 2D game Trench https://ldjam.com/events/ludum-dare/57/trench
		// floor color changes on contact // like piano tiles
		for _, data := range []Entity{
			/* L0: Initial floor */
			// {Pos: rl.NewVector3(0, padH, 0), Size: rl.NewVector3(W/PowF(Phi, 1), floorThick, L/PowF(Phi, 1))},

			/* L1: Next depth level floor splits */
			{
				Pos:  rl.NewVector3(-offset-arenaW/PowF(Phi, 1)/4, -arenaH/2, -offset-arenaW/PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW/PowF(Phi, 1)/2, floorThick, arenaL/PowF(Phi, 1)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW/PowF(Phi, 1)/4, -arenaH/2, -offset-arenaW/PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW/PowF(Phi, 1)/2, floorThick, arenaL/PowF(Phi, 1)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW/PowF(Phi, 1)/4, -arenaH/2, offset+arenaW/PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW/PowF(Phi, 1)/2, floorThick, arenaL/PowF(Phi, 1)/2),
			},
			{
				Pos:  rl.NewVector3(-offset-arenaW/PowF(Phi, 1)/4, -arenaH/2, offset+arenaW/PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW/PowF(Phi, 1)/2, floorThick, arenaL/PowF(Phi, 1)/2),
			},

			/* L2: Next depth level floor splits */
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 0)/4, -arenaH*PowF(Phi, 0), -offset-arenaW*PowF(Phi, 0)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 0)/2, floorThick, arenaL*PowF(Phi, 0)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 0)/4, -arenaH*PowF(Phi, 0), -offset-arenaW*PowF(Phi, 0)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 0)/2, floorThick, arenaL*PowF(Phi, 0)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 0)/4, -arenaH*PowF(Phi, 0), offset+arenaW*PowF(Phi, 0)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 0)/2, floorThick, arenaL*PowF(Phi, 0)/2),
			},
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 0)/4, -arenaH*PowF(Phi, 0), offset+arenaW*PowF(Phi, 0)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 0)/2, floorThick, arenaL*PowF(Phi, 0)/2),
			},

			/* L3: ... */
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 1)/4, -arenaH*PowF(Phi, 1), -offset-arenaW*PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 1)/2, floorThick, arenaL*PowF(Phi, 1)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 1)/4, -arenaH*PowF(Phi, 1), -offset-arenaW*PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 1)/2, floorThick, arenaL*PowF(Phi, 1)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 1)/4, -arenaH*PowF(Phi, 1), offset+arenaW*PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 1)/2, floorThick, arenaL*PowF(Phi, 1)/2),
			},
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 1)/4, -arenaH*PowF(Phi, 1), offset+arenaW*PowF(Phi, 1)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 1)/2, floorThick, arenaL*PowF(Phi, 1)/2),
			},

			/* L4: .... */
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 2)/4, -arenaH*PowF(Phi, 2), -offset-arenaW*PowF(Phi, 2)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 2)/2, floorThick, arenaL*PowF(Phi, 2)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 2)/4, -arenaH*PowF(Phi, 2), -offset-arenaW*PowF(Phi, 2)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 2)/2, floorThick, arenaL*PowF(Phi, 2)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 2)/4, -arenaH*PowF(Phi, 2), offset+arenaW*PowF(Phi, 2)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 2)/2, floorThick, arenaL*PowF(Phi, 2)/2),
			},
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 2)/4, -arenaH*PowF(Phi, 2), offset+arenaW*PowF(Phi, 2)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 2)/2, floorThick, arenaL*PowF(Phi, 2)/2),
			},

			/* L5: ..... {Pos: rl.NewVector3(0, -H*PowF(Phi, 3), 0), Size: rl.NewVector3(W*PowF(Phi, 3), floorThick, L*PowF(Phi, 3))} */
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 3)/4, -arenaH*PowF(Phi, 3), -offset-arenaW*PowF(Phi, 3)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 3)/2, floorThick, arenaL*PowF(Phi, 3)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 3)/4, -arenaH*PowF(Phi, 3), -offset-arenaW*PowF(Phi, 3)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 3)/2, floorThick, arenaL*PowF(Phi, 3)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 3)/4, -arenaH*PowF(Phi, 3), offset+arenaW*PowF(Phi, 3)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 3)/2, floorThick, arenaL*PowF(Phi, 3)/2),
			},
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 3)/4, -arenaH*PowF(Phi, 3), offset+arenaW*PowF(Phi, 3)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 3)/2, floorThick, arenaL*PowF(Phi, 3)/2),
			},

			/* L6: ...... */
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 4)/4, -arenaH*PowF(Phi, 4), -offset-arenaW*PowF(Phi, 4)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 4)/2, floorThick, arenaL*PowF(Phi, 4)/2),
			},
			{
				Pos:  rl.NewVector3(arenaW*PowF(Phi, 4)/4, -arenaH*PowF(Phi, 4), -offset-arenaW*PowF(Phi, 4)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 4)/2, floorThick, arenaL*PowF(Phi, 4)/2),
			},
			{
				Pos:  rl.NewVector3(offset+arenaW*PowF(Phi, 4)/4, -arenaH*PowF(Phi, 4), offset+arenaW*PowF(Phi, 4)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 4)/2, floorThick, arenaL*PowF(Phi, 4)/2),
			},
			{
				Pos:  rl.NewVector3(-offset-arenaW*PowF(Phi, 4)/4, -arenaH*PowF(Phi, 4), offset+arenaW*PowF(Phi, 4)/4),
				Size: rl.NewVector3(arenaW*PowF(Phi, 4)/2, floorThick, arenaL*PowF(Phi, 4)/2),
			},
		} {
			setupFloorResource(data.Pos, data.Size)
		}
	}
	// Setup moving platforms
	// NOTE: Easier to move if they are parallel and in similar position
	// directions... laid together.. just varied y axis
	const platformThick = 1.0
	const _maxPlatformMoveAmplitude = float32(arenaW / 2) // Distance traveled
	setupPlatformResource := func(pos, size rl.Vector3, movementNormal rl.Vector3, movementAmplitude float32) {
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
		resource.PlatformMovementNormals[resource.PlatformCount] = movementNormal       // Up/Down
		resource.PlatformMovementAmplitudes[resource.PlatformCount] = movementAmplitude // World unit
		resource.PlatformAtIsActive[resource.PlatformCount] = true
		resource.PlatformCount++
	}
	for _, data := range []struct {
		Entity            Entity
		MovementNormal    rl.Vector3
		MovementAmplitude float32
	}{
		{
			Entity: Entity{
				Pos:  rl.NewVector3(-arenaW*PowF(Phi, 0), -arenaH*PowF(Phi, 0), -arenaL*PowF(Phi, 0)),
				Size: rl.NewVector3(arenaW*PowF(Phi, 0), floorThick, arenaL*PowF(Phi, 0)),
			},
			MovementNormal:    rl.NewVector3(0, 1, 0),
			MovementAmplitude: -arenaH*PowF(Phi, 1) - floorThick,
		},
		{
			Entity: Entity{
				Pos:  rl.NewVector3(-arenaW*PowF(Phi, 1), -arenaH*PowF(Phi, 1), -arenaL*PowF(Phi, 1)),
				Size: rl.NewVector3(arenaW*PowF(Phi, 1), floorThick, arenaL*PowF(Phi, 1)),
			},
			MovementNormal:    rl.NewVector3(0, 1, 0),
			MovementAmplitude: -arenaH*PowF(Phi, 2) - floorThick,
		},
		{
			Entity: Entity{
				Pos:  rl.NewVector3(-arenaW*PowF(Phi, 2), -arenaH*PowF(Phi, 2), -arenaL*PowF(Phi, 2)),
				Size: rl.NewVector3(arenaW*PowF(Phi, 2), floorThick, arenaL*PowF(Phi, 2)),
			},
			MovementNormal:    rl.NewVector3(0, 1, 0),
			MovementAmplitude: -arenaH*PowF(Phi, 3) - floorThick,
		},
		{
			Entity: Entity{
				Pos:  rl.NewVector3(-arenaW*PowF(Phi, 3), -arenaH*PowF(Phi, 3), -arenaL*PowF(Phi, 3)),
				Size: rl.NewVector3(arenaW*PowF(Phi, 3), floorThick, arenaL*PowF(Phi, 3)),
			},
			MovementNormal:    rl.NewVector3(0, 1, 0),
			MovementAmplitude: -arenaH*PowF(Phi, 4) - floorThick,
		},
	} {
		setupPlatformResource(data.Entity.Pos, data.Entity.Size, data.MovementNormal, data.MovementAmplitude)
	}
	for _, data := range []Entity{
		{rl.NewVector3(-4, -arenaH*2, 0), rl.NewVector3(2, 2, 2)},
		{rl.NewVector3(arenaW/4, -arenaH*8, -arenaL*2), rl.NewVector3(2, 2, 2)},
	} {
		resource.HealBoxPositions[resource.HealBoxCount] = data.Pos
		resource.HealBoxSizes[resource.HealBoxCount] = data.Size
		resource.HealBoxAtIsActive[resource.HealBoxCount] = true
		resource.HealBoxCount++
	}
	for _, data := range []Entity{
		{rl.NewVector3(4.0, -arenaH*2, 0.0), rl.NewVector3(1.5, 1.5, 1.5)},
		{rl.NewVector3(0.0, -arenaH*5, -arenaL), rl.NewVector3(1.0, 1.0, 1.0)},
	} {
		resource.DamageSpherePositions[resource.DamageSphereCount] = data.Pos
		resource.DamageSphereSizes[resource.DamageSphereCount] = data.Size.Y // Radius
		resource.DamageSphereAtIsActive[resource.DamageSphereCount] = true
		resource.DamageSphereCount++
	}
	for _, data := range []Entity{
		{rl.NewVector3(0.0, playerStartPosY+(floorThick/2), 5.0), rl.NewVector3(2.0, 0.25, 2.0)},
		{rl.NewVector3(0.0, 1.0, -9.0), rl.NewVector3(2.0, 0.25, 2.0)},
	} {
		resource.TrampolineBoxPositions[resource.TrampolineBoxCount] = data.Pos
		resource.TrampolineBoxSizes[resource.TrampolineBoxCount] = data.Size
		resource.TrampolineBoxAtIsActive[resource.TrampolineBoxCount] = true
		resource.TrampolineBoxCount++
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

		if rl.IsKeyDown(rl.KeyD) || /* rl.IsKeyDown(rl.KeyRight) || */ rl.IsKeyDown(rl.KeyL) {
			playerMovementThisFrame.X += 1 // Right
		}
		if rl.IsKeyDown(rl.KeyA) || /* rl.IsKeyDown(rl.KeyLeft) ||  */ rl.IsKeyDown(rl.KeyJ) {
			playerMovementThisFrame.X -= 1 // Left
		}
		if rl.IsKeyDown(rl.KeyS) || /* rl.IsKeyDown(rl.KeyDown) || */ rl.IsKeyDown(rl.KeyK) {
			playerMovementThisFrame.Z += 1 // Backward
		}
		if rl.IsKeyDown(rl.KeyW) || /* rl.IsKeyDown(rl.KeyUp) ||  */ rl.IsKeyDown(rl.KeyI) {
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
		mousePos := rl.GetMousePosition()
		if rl.IsMouseButtonPressed(rl.MouseRightButton) {
			if rl.IsCursorHidden() {
				rl.EnableCursor()
			} else {
				rl.DisableCursor()
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

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			if !rayMouseCollision.Hit {
				mouseRay = rl.GetScreenToWorldRay(mousePos, camera)
				// Check collision between ray and origin box
				rayMouseCollision = rl.GetRayCollisionBox(mouseRay, GetBoundingBoxFromPositionSizeV(rl.Vector3Zero(), rl.Vector3One()))
			} else {
				rayMouseCollision.Hit = false
			}
		}

		// Follow player center
		const smooth = 0.034
		camScrollEase = dt * 2.0
		// camScrollEase *= 2.8 // Smooth (trying this out)
		camScrollEase = MinF(camScrollEase, smooth)
		camScrollEase *= 2
		camera.Target = rl.Vector3Lerp(oldCamTarget, playerPosition, camScrollEase)
		rl.UpdateCamera(&camera, rl.CameraThirdPerson)

		// rl.UpdateCamera(&camera, rl.CameraFirstPerson)
		// playerPosition = camera.Target

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

			if isEnable := true; isEnable {
				fuelProgress -= 0.05 / float32(fps) // See also https://community.monogame.net/t/how-can-i-normalize-my-diagonal-movement/15276
			}
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
				t := float32(framesCounter)                                                                       // Current Time
				b := float32(resource.PlatformDefaultPositions[i].X + resource.PlatformMovementAmplitudes[i]/2.0) // Top(Beginning)
				c := float32(-resource.PlatformMovementAmplitudes[i])                                             // Bottom(Change)
				d := float32(fps) * 4                                                                             // Duration
				f = easings.SineInOut(t, b, c, d)
				resource.PlatformPositions[i].X = f
				resource.PlatformBoundingBoxes[i].Min.X = resource.PlatformPositions[i].X - resource.PlatformSizes[i].X/2
				resource.PlatformBoundingBoxes[i].Max.X = resource.PlatformPositions[i].X + resource.PlatformSizes[i].X/2
			case rl.Vector3{X: 0, Y: 1, Z: 0}:
				t := float32(framesCounter)                                                                       // Current Time
				b := float32(resource.PlatformDefaultPositions[i].Y + resource.PlatformMovementAmplitudes[i]/2.0) // Top(Beginning)
				c := float32(-resource.PlatformMovementAmplitudes[i])                                             // Bottom(Change)
				d := float32(fps) * 4                                                                             // Duration
				f = easings.SineInOut(t, b, c, d)
				resource.PlatformPositions[i].Y = f
				resource.PlatformBoundingBoxes[i].Min.Y = resource.PlatformPositions[i].Y - resource.PlatformSizes[i].Y/2 // platformThick/2
				resource.PlatformBoundingBoxes[i].Max.Y = resource.PlatformPositions[i].Y + resource.PlatformSizes[i].Y/2 // platformThick/2
			case rl.Vector3{X: 0, Y: 0, Z: 1}:
				t := float32(framesCounter)                                                                       // Current Time
				b := float32(resource.PlatformDefaultPositions[i].Z + resource.PlatformMovementAmplitudes[i]/2.0) // Top(Beginning)
				c := float32(-resource.PlatformMovementAmplitudes[i])                                             // Bottom(Change)
				d := float32(fps) * 4                                                                             // Duration
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
			playerColor = rl.White
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
		if playerAirTimer > maxPlayerAirTime*Phi && playerAirTimer < maxPlayerAirTime*Phi*Phi {
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
		if isEnable := true; isEnable {
			if playerAirTimer > maxPlayerFreefallAirTime {
				shieldProgress -= PowF(progressRate*shieldProgress, maxPlayerFreefallAirTime/playerAirTimer)
			}
		}
		if isDebugAllCollisions := false; isDebugAllCollisions {
			if (playerCollisionsThisFrame.X == 1 || playerCollisionsThisFrame.Z == 1) &&
				(!isOOBCollision && !isSafeSpotCollision && !isFloorCollision && !isPlatformCollision && !isWallCollision) {
				shieldProgress -= progressRate
			}
		}

		cameraPlayerCrossProduct := rl.Vector3CrossProduct(rl.Vector3Normalize(camera.Position), rl.Vector3Normalize(playerPosition))
		cameraPlayerDotProduct := rl.Vector3DotProduct(rl.Vector3Normalize(camera.Position), rl.Vector3Normalize(playerPosition))
		cameraPlayerDotProductRadian := math.Acos(float64(cameraPlayerDotProduct))
		cameraPlayerCrossProduct = rl.Vector3CrossProduct(rl.Vector3Normalize(playerPosition), rl.Vector3Normalize(camera.Position))
		cameraPlayerDotProductRadianDegree := cameraPlayerDotProductRadian * rl.Rad2deg
		_ = cameraPlayerDotProductRadianDegree

		playerCameraRay = rl.NewRay(playerPosition, cameraPlayerCrossProduct)
		playerCameraRay = rl.GetScreenToWorldRay(rl.GetWorldToScreen(playerPosition, camera), camera)

		if false {
			// Chaos rotate on y axis... spiral down
			// playerPosition = rl.Vector3RotateByAxisAngle(playerPosition, rl.NewVector3(playerPosition.X, 0, playerPosition.Z), dt*movementMagnitude)

			playerPosition = rl.Vector3Lerp(
				playerPosition,
				rl.Vector3RotateByAxisAngle(
					playerPosition,
					rl.NewVector3(oldPlayerPos.X, 0, oldPlayerPos.Z),
					dt*movementMagnitude,
				),
				1.0,
			)
			playerPosition = rl.Vector3Lerp(
				playerPosition,
				rl.Vector3RotateByAxisAngle(
					playerPosition,
					rl.NewVector3(oldPlayerPos.X+playerSize.X/2, oldPlayerPos.Y+playerSize.Y/2, oldPlayerPos.Z+playerSize.Z/2),
					dt*movementMagnitude,
				),
				1.0,
			)
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
			rl.DrawBoundingBox(resource.FloorBoundingBoxes[i], rl.Fade(rl.LightGray, 0.3))
		}
		for i := range resource.PlatformCount {
			rl.DrawModel(resource.PlatformModels[i], resource.PlatformPositions[i], 1.0, rl.Black) // Platform
			rl.DrawBoundingBox(resource.PlatformBoundingBoxes[i], rl.DarkGray)                     // Platform outline
			if false {
				magnitude := resource.PlatformMovementAmplitudes[i]
				amplitude := rl.NewVector3(magnitude, magnitude, magnitude)
				normalAxis := rl.Vector3Multiply(resource.PlatformMovementNormals[i], amplitude)
				normalAxisSize := rl.Vector3AddValue(normalAxis, 0.125*0.5)
				rl.DrawCubeV(resource.PlatformDefaultPositions[i], normalAxisSize, rl.Fade(rl.White, 0.8)) // Reference (y axis)
				size := rl.Vector3Invert(rl.Vector3Add(resource.PlatformMovementNormals[i], Vector3One))   // [0 1 0] => [1 1 .5]
				rl.DrawCubeV(resource.PlatformDefaultPositions[i], size, rl.Fade(rl.White, 0.8))           // Reference (midpoint plane trick)
			}
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
		{
			playerRadius := playerSize.X / 2
			playerStartPos := rl.NewVector3(playerPosition.X, playerPosition.Y-playerSize.Y/2+playerRadius, playerPosition.Z)
			playerEndPos := rl.NewVector3(playerPosition.X, playerPosition.Y+playerSize.Y/2-playerRadius, playerPosition.Z)
			rl.DrawCapsule(playerStartPos, playerEndPos, playerRadius, 4, 4, playerColor)
			rl.DrawCapsuleWires(playerStartPos, playerEndPos, playerRadius, 4*2, 6*2, rl.ColorLerp(playerColor, rl.Fade(rl.DarkGray, 0.8), 0.5))
			if isDebug := false; isDebug {
				rl.DrawCubeV(playerPosition, playerSize, playerColor)
				rl.DrawCubeWiresV(playerPosition, playerSize, playerColor)
			}
			if isDebug := false; isDebug {
				oldPlayerStartPos := rl.NewVector3(oldPlayerPos.X, oldPlayerPos.Y-playerSize.Y/2+playerRadius, oldPlayerPos.Z)
				oldPlayerEndPos := rl.NewVector3(oldPlayerPos.X, oldPlayerPos.Y+playerSize.Y/2-playerRadius, oldPlayerPos.Z)
				rl.DrawCapsule(oldPlayerStartPos, oldPlayerEndPos, playerRadius, 16, 16, rl.DarkGray)
				rl.DrawCapsuleWires(oldPlayerStartPos, oldPlayerEndPos, playerRadius, 4, 6, rl.ColorLerp(playerColor, rl.Fade(rl.DarkGray, 0.8), 0.5))
			}
			if false {
				if false {
					rl.DrawModelWiresEx(playerModel, playerPosition, playerRotationNormal, rl.QuaternionLength(playerRotation)*0+dt+80*rl.Deg2rad, rl.NewVector3(2, 2, 2), rl.Black)
				}
				rl.DrawModelEx(playerModel,
					rl.Vector3RotateByAxisAngle(playerPosition, playerRotationNormal, dt*100),
					playerRotationNormal,
					rl.QuaternionLength(playerRotation)*0+dt+80*rl.Deg2rad,
					rl.NewVector3(2, 2, 2), rl.Black)
			}
			if false {
				if !rl.IsCursorHidden() && rl.IsCursorOnScreen() {
					pos := rl.Vector3{X: rl.GetMouseDelta().X, Y: 0., Z: rl.GetMouseDelta().Y}
					pos = mouseRay.Position
					rayDir := mouseRay.Direction // Ray direction
					pos = rl.Vector3CrossProduct(pos, rayDir)
					rl.DrawRay(mouseRay, rl.White)
					rl.DrawModel(resource.FloorModels[0], pos, 1.0, rl.White)
				}
			}
			if false { // TODO: HOW TO FIND ANGLE?
				dirUnitVector := rl.Vector3Normalize(rl.Vector3CrossProduct(camera.Position, camera.Target))
				rl.DrawLine3D(camera.Target, dirUnitVector, rl.Gold)
			}
		}

		if false {
			rl.DrawGrid(int32(MinF(arenaW, arenaL)*InvMathPhi), 1)
		}

		// Draw orbital XYZ origins
		DrawXYZOrbitAxisV(Vector3Zero, 12.0, 0.05, 0.3)                // Level Center
		DrawXYZOrbitAxisV(playerPosition, playerSize.Y*Phi, 0.05, 0.3) // Player Center
		for i := range MaxResourceSOACapacity {
			if false {
				if resource.PlatformAtIsActive[i] {
					DrawXYZOrbitAxisV(resource.PlatformPositions[i], rl.Vector3Length(resource.PlatformSizes[i]), 0.05, 0.5)
					DrawXYZOrbitAxisV(resource.PlatformDefaultPositions[i], rl.Vector3Length(resource.PlatformSizes[i]), 0.05, 0.5/2)
				}
				if resource.FloorAtIsActive[i] {
					DrawXYZOrbitAxisV(resource.FloorPositions[i], rl.Vector3Length(resource.FloorSizes[i]), 0.05, 0.5)
				}
				if resource.HealBoxAtIsActive[i] {
					DrawXYZOrbitAxisV(resource.HealBoxPositions[i], rl.Vector3Length(resource.HealBoxSizes[i]), 0.05, 0.5)
				}
				if resource.DamageSphereAtIsActive[i] {
					DrawXYZOrbitAxisV(resource.DamageSpherePositions[i], resource.DamageSphereSizes[i]*2.0, 0.05, 0.5)
				}
				if resource.TrampolineBoxAtIsActive[i] {
					DrawXYZOrbitAxisV(resource.TrampolineBoxPositions[i], rl.Vector3Length(resource.TrampolineBoxSizes[i]), 0.05, 0.5)
				}
			}
		}

		// DEBUG
		if false {
			rl.DrawRay(mouseRay, rl.Gold)
			rl.DrawCubeV(mouseRay.Position, rl.Vector3One(), rl.Gold)
			rl.DrawCapsuleWires(playerPosition, rl.Vector3Lerp(camera.Position, playerPosition, 0.5), 0.125, 4, 4, rl.Fade(rl.SkyBlue, .3))
			if playerCameraRayCollision.Hit {
				rl.DrawCubeV(playerCameraRay.Position, rl.Vector3One(), rl.SkyBlue)
			}
		}

		rl.EndMode3D()

		// Draw HUD
		rl.DrawRectangle(10, 20, 200, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20}, rl.Vector2{X: fuelProgress * 200, Y: 20}, rl.DarkGray)
		rl.DrawText("Fuel", 10+5, 21, 20, rl.White)
		text := fmt.Sprintf("%.0f", fuelProgress*100)
		rl.DrawText(text, 200-rl.MeasureText(text, 10)/2, 20+5*2, 10, rl.White)

		rl.DrawRectangle(10, 20+20, 200, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20 + 20}, rl.Vector2{X: shieldProgress * 200, Y: 20}, rl.DarkGray)
		rl.DrawText("Hull", 10+5, 21+20, 20, rl.White)
		text = fmt.Sprintf("%.0f", shieldProgress*100)
		rl.DrawText(text, 200-rl.MeasureText(text, 10)/2, 20+20+5*2, 10, rl.White)

		rl.DrawRectangle(10, 20+20+20, 200, 20, rl.Fade(rl.Black, 0.9))
		rl.DrawRectangleV(rl.Vector2{X: 10, Y: 20 + 20 + 20}, rl.Vector2{X: shieldProgress * 200, Y: 20}, rl.DarkGray)
		rl.DrawText("Depth", 10+5, 21+20+20, 20, rl.White)
		text = fmt.Sprintf("%.0f", playerPosition.Y)
		rl.DrawText(text, 200-rl.MeasureText(text, 10)*2/3, 20+20+20+5*2, 10, rl.White)

		rl.DrawFPS(10, int32(rl.GetScreenHeight())-25)

		// Quick debug zone
		text = fmt.Sprintf(
			"playerAirTimer: %.2f\nplayerJumpsLeft: %d\nplayerPosition: %.2f\n"+
				"camera.Position: %.2f\ncamera.Target: %.2f\ncameraPositionTargetDistance: %.2f\n"+
				"cameraScrollEase: %.4f\n"+
				"defaultCameraPosition: %.2f\ndefaultCameraTarget: %.2f\ndefaultCameraPositionTargetDistance: %.2f\n"+
				"mousePos: %.4f\n",
			playerAirTimer, playerJumpsLeft, playerPosition,
			camera.Position, camera.Target, rl.Vector3Distance(camera.Position, camera.Target),
			camScrollEase,
			defaultCameraPosition, defaultCameraTarget, defaultCameraPositionTargetDistance,
			mousePos,
		)

		debugWidth := rl.MeasureText(text, 10)
		debugXPos := int32(rl.GetScreenWidth()) - 10 - debugWidth
		rl.DrawRectangle(debugXPos-5, 10-5, debugWidth+5*2, debugWidth*2/3, rl.Fade(rl.Blue, 0.3))
		rl.DrawText(text, debugXPos, 10, 10, rl.White)

		rl.EndDrawing()
	}

	if false {
		jsonData, err := json.Marshal(resource)
		if err != nil {
			slog.Error(err.Error())
		}
		err = os.WriteFile("resource_assets.json", jsonData, 0644)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	rl.CloseWindow()
}

// alpha goes from 0.0f to 1.0f
func DrawXYZOrbitAxisV(pos rl.Vector3, maxSize, thick, alpha float32) {
	var (
		Red   = color.RGBA{R: 230, G: 41, B: 55, A: uint8(alpha * 255)} // rl.Red
		Green = color.RGBA{R: 0, G: 228, B: 48, A: uint8(alpha * 255)}  // rl.Green
		Blue  = color.RGBA{R: 0, G: 121, B: 241, A: uint8(alpha * 255)} // rl.Blue
	)
	rl.DrawCubeV(pos, rl.NewVector3(maxSize, thick, thick), Red)
	rl.DrawCubeV(pos, rl.NewVector3(thick, maxSize, thick), Green)
	rl.DrawCubeV(pos, rl.NewVector3(thick, thick, maxSize), Blue)
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
	Vector3One       = rl.Vector3One()
	Vector3Zero      = rl.Vector3Zero()
	Vector2One       = rl.Vector2One()
	Vector2Zero      = rl.Vector2Zero()
	Vector3OneLength = rl.Vector3Length(Vector3One)
	Vector2OneLength = rl.Vector2Length(Vector2One)
)

func IsUnitVec3(v rl.Vector3) bool { return rl.Vector3Length(v) <= Vector3OneLength }
func IsUnitVec2(v rl.Vector2) bool { return rl.Vector2Length(v) <= Vector2OneLength }

const (
	Phi                = math.Phi
	InvMathPhi         = 1 / Phi
	OneMinusInvMathPhi = 1 - InvMathPhi
)
