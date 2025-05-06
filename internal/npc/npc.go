package npc

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

const (
	MaxNPC = 32
)

type NPCSOA struct { // size=2916 (0xb64)
	BoundingBox         [MaxNPC]rl.BoundingBox
	Collisions          [MaxNPC]rl.Quaternion
	Position            [MaxNPC]rl.Vector3
	Size                [MaxNPC]rl.Vector3
	Type                [MaxNPC]NPCType
	Action              [MaxNPC]NPCActionType
	State               [MaxNPC]NPCStateFlag
	Color               [MaxNPC]color.RGBA
	Rotation            [MaxNPC]float32 // (degrees) XZ plane
	Health              [MaxNPC]float32 // [0..1]
	IsActive            [MaxNPC]bool
	IsWallCollision     [MaxNPC]bool
	IsWalk              [MaxNPC]bool
	CircularBufferIndex int32
}

// See https://book.leveldesignbook.com/process/combat/enemy#enemy-roster
type NPCType int32

const (
	TypeGrunt  NPCType = iota // Close range melee attack player straightforward, easy to pull
	TypeSquad                 // Attack from mid-range / long range, ideally take turns as a group
	TypeLeader                // High survivability, buffs nearby allies somehow
	TypeTank                  // High survivability, but slower and larger (easy to hit or avoid)
	TypeSwarm                 // Small fast attacker with low health but high close-range damage
	TypeSniper                // Slow weak long range attacker, very vulnerable without others
)

// See https://devforum.roblox.com/t/custom-enumerations/2626065/3
type NPCActionType int32

const (
	Kick NPCActionType = iota
	Punch
	WeaponMelee
	WeaponShoot
	Spell
	Special
	Custom // 100
)

// See https://devforum.roblox.com/t/custom-enumerations/2626065/5
//
//	Although this is the approach I have been using as well, it leads to
//	problems like discoverability and typing. There is also a larger use for
//	enums, for example, with flags and states
//
//	myNpc.state = NPCState.IS_ALIVE | NPCState.HAS_TARGET
//	myNPC->state = NPCState::IS_ALIVE | NPCState::HAS_TARGET;
type NPCStateFlag int32

const (
	FlagNPCDefault   NPCStateFlag = 0 << iota       // 0
	FlagNPCInCombat  NPCStateFlag = 1 << (iota - 1) // 1
	FlagNPCIsInjured NPCStateFlag = 1 << (iota - 1) // 2
	FlagNPCIsAlive   NPCStateFlag = 1 << (iota - 1) // 4
	FlagNPCHasTarget NPCStateFlag = 1 << (iota - 1) // 8
)

func (gs *NPCSOA) Reset() {
	for i := range MaxNPC {
		gs.Position[i] = rl.Vector3{}
		gs.Size[i] = rl.Vector3{}
		gs.IsActive[i] = false
	}
	gs.CircularBufferIndex = 0
}

func (gs *NPCSOA) Emit(position, size rl.Vector3, rotationDegree float32) {
	gs.Position[gs.CircularBufferIndex] = position
	gs.Size[gs.CircularBufferIndex] = size

	gs.BoundingBox[gs.CircularBufferIndex] = common.GetBoundingBoxPositionSizeV(position, size)

	gs.Collisions[gs.CircularBufferIndex] = rl.Quaternion{}

	gs.Color[gs.CircularBufferIndex] = rl.White
	gs.Rotation[gs.CircularBufferIndex] = rotationDegree
	gs.Health[gs.CircularBufferIndex] = 1. // [0..1]
	gs.IsActive[gs.CircularBufferIndex] = true
	gs.IsWallCollision[gs.CircularBufferIndex] = false
	gs.IsWalk[gs.CircularBufferIndex] = true // IDK (should add timer)

	// Increment index: (ring like data structure / circular reusable buffer)
	gs.CircularBufferIndex = (gs.CircularBufferIndex + 1) % MaxNPC
}
