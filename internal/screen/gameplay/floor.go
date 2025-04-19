package gameplay

import (
	"cmp"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

const IS_ENABLE_FLOOR_SHADER = false

var floorTextureTiling []float32

type Floor struct {
	Position    rl.Vector3
	Size        rl.Vector3
	BoundingBox rl.BoundingBox
}

func NewFloor(pos, size rl.Vector3) Floor {
	return Floor{
		Position:    pos,
		Size:        size,
		BoundingBox: common.GetBoundingBoxFromPositionSizeV(pos, size),
	}
}

// InitFloor creates a new floor, loads a mesh, assigns material parameters and texture to a model.
// Set floor model texture tiling and emissive color parameters on shader
// NOTE: A basic plane shape can be generated instead of being loaded from a model file
func InitFloor() {
	pos := rl.NewVector3(0., -0.5, 0.)
	size := rl.NewVector3(32., 1., 32.)
	floor = NewFloor(pos, size)

	// Load floor model mesh
	const floorScale = 5
	floorRoadPBRModel = cmp.Or(
		rl.LoadModelFromMesh(cmp.Or(
			rl.GenMeshCube(size.X/floorScale, size.Y/floorScale, size.Z/floorScale),
			rl.GenMeshPlane(size.X/floorScale, size.Z/floorScale, 10, 10))),
		rl.LoadModel(filepath.Join("res", "model", "plane.glb")),
	)

	if IS_ENABLE_FLOOR_SHADER {
		// Assign material parameters
		floorRoadPBRModel.Materials.Shader = common.Shader.PBR
		floorRoadPBRModel.Materials.GetMap(rl.MapAlbedo).Color = rl.White
		floorRoadPBRModel.Materials.GetMap(rl.MapMetalness).Value = 0.0
		floorRoadPBRModel.Materials.GetMap(rl.MapRoughness).Value = 0.0
		floorRoadPBRModel.Materials.GetMap(rl.MapOcclusion).Value = 1.0
		floorRoadPBRModel.Materials.GetMap(rl.MapEmission).Color = rl.Black
	}

	// Assign texture parameters
	floorRoadPBRModel.Materials.GetMap(rl.MapAlbedo).Texture =
		rl.LoadTexture(filepath.Join("res", "texture", "road_a.png"))
	floorRoadPBRModel.Materials.GetMap(rl.MapMetalness).Texture =
		rl.LoadTexture(filepath.Join("res", "texture", "road_mra.png"))
	floorRoadPBRModel.Materials.GetMap(rl.MapNormal).Texture =
		rl.LoadTexture(filepath.Join("res", "texture", "road_n.png"))

	textureTilingLoc = rl.GetShaderLocation(common.Shader.PBR, "tiling")
	emissiveColorLoc = rl.GetShaderLocation(common.Shader.PBR, "emissiveColor")
	floorTextureTiling = []float32{.5, .5}
}

func (fl Floor) Draw() {
	rl.SetShaderValue(common.Shader.PBR, textureTilingLoc, floorTextureTiling, rl.ShaderUniformVec2)
	fecVector4 := rl.ColorNormalize(floorRoadPBRModel.Materials.GetMap(rl.MapEmission).Color)
	floorEmissiveColor := []float32{fecVector4.X, fecVector4.Y, fecVector4.Z, fecVector4.W}
	rl.SetShaderValue(common.Shader.PBR, emissiveColorLoc, floorEmissiveColor, rl.ShaderUniformVec2)

	rl.DrawModel(floorRoadPBRModel, fl.Position, 5.0, rl.White)
}
