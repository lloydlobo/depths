// *   NOTE: raylib supports multiple models file formats:
// *
// *     - OBJ  > Text file format. Must include vertex position-texcoords-normals information,
// *              if files references some .mtl materials file, it will be loaded (or try to).
// *     - GLTF > Text/binary file format. Includes lot of information and it could
// *              also reference external files, raylib will try loading mesh and materials data.
// *     - IQM  > Binary file format. Includes mesh vertex data but also animation data,
// *              raylib can load .iqm animations.
// *     - VOX  > Binary file format. MagikaVoxel mesh format:
// *              https://github.com/ephtracy/voxel-model/blob/master/MagicaVoxel-file-format-vox.txt
// *     - M3D  > Binary file format. Model 3D format:
// *              https://bztsrc.gitlab.io/model3d
package model

import (
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AssetType uint8

const (
	Banner AssetType = iota
	Barrel
	CharacterHuman
	CharacterOrc
	Chest
	Coin
	Column
	Dirt
	Floor
	FloorDetail
	Gate
	Rocks
	Stairs
	Stones
	Trap
	Wall
	WallHalf
	WallNarrow
	WallOpening
	WoodStructure
	WoodSupport

	MaxModelTypeCount // 21
)

// OBJ Text file format. Must include vertex position-texcoords-normals
// information, if files references some .mtl materials file, it will be loaded (or try to).
type ModelsObj struct {
	Colormap rl.Texture2D

	Banner, Barrel, CharacterHuman, CharacterOrc, Chest, Coin, Column, Dirt,
	Floor, FloorDetail, Gate, Rocks, Stairs, Stones, Trap, Wall, WallHalf,
	WallNarrow, WallOpening, WoodStructure, WoodSupport rl.Model
}

func LoadAssetModelOBJ() ModelsObj {
	dir := filepath.Join("res", "kenney_mini-dungeon", "Models", "OBJ format")
	return ModelsObj{
		Colormap: rl.LoadTexture(filepath.Join(dir, "Textures", "colormap.png")),

		Banner:         rl.LoadModel(filepath.Join(dir, "banner.obj")),
		Barrel:         rl.LoadModel(filepath.Join(dir, "barrel.obj")),
		CharacterHuman: rl.LoadModel(filepath.Join(dir, "character-human.obj")),
		CharacterOrc:   rl.LoadModel(filepath.Join(dir, "character-orc.obj")),
		Chest:          rl.LoadModel(filepath.Join(dir, "chest.obj")),
		Coin:           rl.LoadModel(filepath.Join(dir, "coin.obj")),
		Column:         rl.LoadModel(filepath.Join(dir, "column.obj")),
		Dirt:           rl.LoadModel(filepath.Join(dir, "dirt.obj")),
		Floor:          rl.LoadModel(filepath.Join(dir, "floor.obj")),
		FloorDetail:    rl.LoadModel(filepath.Join(dir, "floor-detail.obj")),
		Gate:           rl.LoadModel(filepath.Join(dir, "gate.obj")),
		Rocks:          rl.LoadModel(filepath.Join(dir, "rocks.obj")),
		Stairs:         rl.LoadModel(filepath.Join(dir, "stairs.obj")),
		Stones:         rl.LoadModel(filepath.Join(dir, "stones.obj")),
		Trap:           rl.LoadModel(filepath.Join(dir, "trap.obj")),
		Wall:           rl.LoadModel(filepath.Join(dir, "wall.obj")),
		WallHalf:       rl.LoadModel(filepath.Join(dir, "wall-half.obj")),
		WallNarrow:     rl.LoadModel(filepath.Join(dir, "wall-narrow.obj")),
		WallOpening:    rl.LoadModel(filepath.Join(dir, "wall-opening.obj")),
		WoodStructure:  rl.LoadModel(filepath.Join(dir, "wood-structure.obj")),
		WoodSupport:    rl.LoadModel(filepath.Join(dir, "wood-support.obj")),
	}
}

type ModelsGLB struct { // size=2520 (0x9d8)
	Banner, Barrel,
	CharacterHuman, CharacterOrc, Chest, Coin, Column,
	Dirt,
	Floor, FloorDetail,
	Gate,
	Rocks,
	Stairs, Stones,
	Trap,
	Wall, WallHalf, WallNarrow, WallOpening, WoodStructure, WoodSupport rl.Model
}

func LoadAssetModelGLB() ModelsGLB {
	dir := filepath.Join("res", "kenney_mini-dungeon", "Models", "OBJ format")
	return ModelsGLB{
		Banner:         rl.LoadModel(filepath.Join(dir, "banner.obj")),
		Barrel:         rl.LoadModel(filepath.Join(dir, "barrel.obj")),
		CharacterHuman: rl.LoadModel(filepath.Join(dir, "character-human.obj")),
		CharacterOrc:   rl.LoadModel(filepath.Join(dir, "character-orc.obj")),
		Chest:          rl.LoadModel(filepath.Join(dir, "chest.obj")),
		Coin:           rl.LoadModel(filepath.Join(dir, "coin.obj")),
		Column:         rl.LoadModel(filepath.Join(dir, "column.obj")),
		Dirt:           rl.LoadModel(filepath.Join(dir, "dirt.obj")),
		Floor:          rl.LoadModel(filepath.Join(dir, "floor.obj")),
		FloorDetail:    rl.LoadModel(filepath.Join(dir, "floor-detail.obj")),
		Gate:           rl.LoadModel(filepath.Join(dir, "gate.obj")),
		Rocks:          rl.LoadModel(filepath.Join(dir, "rocks.obj")),
		Stairs:         rl.LoadModel(filepath.Join(dir, "stairs.obj")),
		Stones:         rl.LoadModel(filepath.Join(dir, "stones.obj")),
		Trap:           rl.LoadModel(filepath.Join(dir, "trap.obj")),
		Wall:           rl.LoadModel(filepath.Join(dir, "wall.obj")),
		WallHalf:       rl.LoadModel(filepath.Join(dir, "wall-half.obj")),
		WallNarrow:     rl.LoadModel(filepath.Join(dir, "wall-narrow.obj")),
		WallOpening:    rl.LoadModel(filepath.Join(dir, "wall-opening.obj")),
		WoodStructure:  rl.LoadModel(filepath.Join(dir, "wood-structure.obj")),
		WoodSupport:    rl.LoadModel(filepath.Join(dir, "wood-support.obj")),
	}
}
