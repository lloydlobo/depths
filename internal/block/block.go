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
	Pos      rl.Vector3
	Size     rl.Vector3
	Rotn     float32
	Health   float32 // [0..1]
	IsActive bool
	State    BlockState
}

var (
	blockModels [MaxBlockState]rl.Model
)

func NewBlock(pos, size rl.Vector3) Block {
	return Block{
		Pos:      pos,
		Size:     size,
		Rotn:     0.0,
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

	for i := range positions {
		size := rl.Vector3Multiply(
			rl.NewVector3(1, 1, 1),
			rl.NewVector3(
				float32(rl.GetRandomValue(88, 101))/100.,
				float32(rl.GetRandomValue(100, 300))/100.,
				float32(rl.GetRandomValue(88, 101))/100.))

		obj := NewBlock(positions[i], size)
		obj.Rotn = cmp.Or(float32(rl.GetRandomValue(-50, 50)/10.), 0.)

		*dst = append(*dst, obj)
	}
	mu.Unlock()
}

func SetupBlockModels() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	for i := range MaxBlockState {
		switch i {
		case DirtBlockState:
			blockModels[i] = common.Model.OBJ.Dirt
		case RockBlockState:
			blockModels[i] = common.Model.OBJ.Rocks
		case StoneBlockState:
			blockModels[i] = common.Model.OBJ.Stones
		case FloorDetailBlockState:
			blockModels[i] = common.Model.OBJ.FloorDetail
		default:
			panic(fmt.Sprintf("unexpected gameplay.BlockState: %#v", i))
		}
		rl.SetMaterialTexture(blockModels[i].Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
	}
}

func (b Block) Draw() {
	if /* b.State > 0 && */ b.IsActive {
		rl.DrawModelEx(blockModels[b.State], b.Pos, common.YAxis, b.Rotn, b.Size, rl.White)
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
			positions = append(positions, rl.NewVector3(x, y, z))
		}
	}
	return positions
}
