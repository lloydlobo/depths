package block

import (
	"cmp"
	"fmt"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	"example/depths/internal/floor"
)

// The game world is composed of rough 3D objects—mainly cubes, referred to as
// blocks—representing various materials, such as dirt, stone, ores, tree
// trunks, water, and lava. The core gameplay revolves around picking up and
// placing these objects. These blocks are arranged in a 3D grid, while players
// can move freely around the world. Players can break, or mine, blocks and
// then place them elsewhere, enabling them to build
// things.[6](https://en.wikipedia.org/wiki/Minecraft#cite_note-10)
type BlockState uint8

const (
	DirtBlockState BlockState = iota
	RockBlockState
	StoneBlockState
	FloorDetailBlockState // decorated floor tile

	MaxBlockState
)

type Block struct {
	Position rl.Vector3
	Size     rl.Vector3
	Rotation float32
	Health   float32 // [0..1]
	IsActive bool
	State    BlockState
}

var (
	blockModels [MaxBlockState]rl.Model
)

// WARN: The asset uses a 1:1:1 block but places the bottom at the unit cubes bottom
func (b *Block) GetBlockBoundingBox() rl.BoundingBox {
	modelCenterPositionY := (b.Position.Y + b.Size.Y/2)
	return rl.BoundingBox{
		Min: rl.Vector3{X: b.Position.X - b.Size.X/2, Y: modelCenterPositionY - b.Size.Y/2, Z: b.Position.Z - b.Size.Z/2},
		Max: rl.Vector3{X: b.Position.X + b.Size.X/2, Y: modelCenterPositionY + b.Size.Y/2, Z: b.Position.Z + b.Size.Z/2},
	}
}

const (
	blockModelYSize = 1.0
)

func NewBlock(pos, size rl.Vector3) Block {
	return Block{
		Position: pos,
		Size:     size,
		Rotation: 0.0,
		State:    DirtBlockState,
		IsActive: true,
	}
}

func (o *Block) NextState() {
	o.State++
	if o.State >= MaxBlockState {
		o.State = MaxBlockState - 1
		o.IsActive = false
	}
}

func InitBlocks(dst *[]Block, positions []rl.Vector3) {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	for i := range positions {
		size := rl.Vector3Multiply(
			rl.NewVector3(1, 1, 1),
			rl.NewVector3(
				float32(rl.GetRandomValue(92, 98))/100.,
				float32(rl.GetRandomValue(100/1.25, 100*1.5))/100.,
				float32(rl.GetRandomValue(92, 98))/100.,
			),
		)
		obj := NewBlock(positions[i], size)
		obj.Rotation = cmp.Or(float32(rl.GetRandomValue(-80, 80)/10.), 0.)
		*dst = append(*dst, obj)
	}
}

func SetupBlockModels() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	for i := range MaxBlockState {
		switch i {
		case DirtBlockState:
			blockModels[i] = common.ModelDungeonKit.OBJ.Dirt
		case RockBlockState:
			blockModels[i] = common.ModelDungeonKit.OBJ.Rocks
		case StoneBlockState:
			blockModels[i] = common.ModelDungeonKit.OBJ.Stones
		case FloorDetailBlockState:
			blockModels[i] = common.ModelDungeonKit.OBJ.FloorDetail
		default:
			panic(fmt.Sprintf("unexpected gameplay.BlockState: %#v", i))
		}
		rl.SetMaterialTexture(blockModels[i].Materials, rl.MapDiffuse, common.ModelDungeonKit.OBJ.Colormap)
	}
}

func (b Block) Draw() {
	if b.IsActive {
		rotationAxis := common.YAxis
		rotationAxis = rl.Vector3Normalize(b.Position)
		rotationAxis = rl.Vector3Lerp(rotationAxis, common.YAxis, .5)
		rl.DrawModelEx(blockModels[b.State], b.Position, rotationAxis, b.Rotation, b.Size, rl.White)
	}
}

// - Avoid spawning where player is standing
// - Randomly skip a position
// - A noise map or simplex/perlin noise "can" serve better
func GenerateRandomBlockPositions(gameFloor floor.Floor) []rl.Vector3 {
	var positions []rl.Vector3 // 61% of maxPositions

	var (
		y    = (gameFloor.BoundingBox.Min.Y + gameFloor.BoundingBox.Max.Y) / 2.0
		bb   = gameFloor.BoundingBox
		offX = float32(3)
		offZ = float32(3)
	)
	y = gameFloor.Position.Y // gameFloor is a plane

	var (
		maxGridCells            = gameFloor.Size.X * gameFloor.Size.Z // just-in-case
		maxSkipLoopPositionOdds = int32(2)                            // if 2 -> 0,1,2 -> 1/3 odds
	)

NextCol:
	for x := bb.Min.X + 1; x < bb.Max.X; x++ {
	NextRow:
		for z := bb.Min.Z + 1; z < bb.Max.Z; z++ {
			if len(positions) >= int(maxGridCells) {
				break NextCol
			}
			// Reserve space for area in offset from origin
			for i := -offX; i <= offX; i++ {
				for k := -offZ; k <= offZ; k++ {
					if i == x && k == z {
						continue NextRow
					}
					if rl.Vector3Distance(rl.NewVector3(i, y, k), rl.NewVector3(x, y, z)) < (offX+offZ)/2 {
						continue NextRow
					}
				}
			}
			if rl.GetRandomValue(0, maxSkipLoopPositionOdds) == 0 {
				continue NextRow
			}
			pos := rl.NewVector3(x, y, z)
			pos = rl.NewVector3(pos.X, pos.Y+blockModelYSize/2., pos.Z)
			pos = rl.NewVector3(pos.X, pos.Y-blockModelYSize/2., pos.Z)
			positions = append(positions, pos)
		}
	}
	return positions
}
