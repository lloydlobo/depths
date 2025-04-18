package gameplay

import (
	"cmp"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

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
	size := rl.NewVector3(30., 1., 30.)
	floor = NewFloor(pos, size)

	// Load floor model mesh
	const floorScale = 5
	floorModel = cmp.Or(
		rl.LoadModelFromMesh(cmp.Or(
			rl.GenMeshCube(size.X/floorScale, size.Y/floorScale, size.Z/floorScale),
			rl.GenMeshPlane(size.X/floorScale, size.Z/floorScale, 10, 10))),
		rl.LoadModel(filepath.Join("res", "model", "plane.glb")))

	// Assign material parameters
	if isEnableLight := false; isEnableLight {
		floorModel.Materials.Shader = common.Shader.PBR
		floorModel.Materials.GetMap(rl.MapAlbedo).Color = rl.White
		floorModel.Materials.GetMap(rl.MapMetalness).Value = 0.0
		floorModel.Materials.GetMap(rl.MapRoughness).Value = 0.0
		floorModel.Materials.GetMap(rl.MapOcclusion).Value = 1.0
		floorModel.Materials.GetMap(rl.MapEmission).Color = rl.Black
	}

	// Assign texture parameters
	floorModel.Materials.GetMap(rl.MapAlbedo).Texture =
		rl.LoadTexture(filepath.Join("res", "texture", "road_a.png"))
	floorModel.Materials.GetMap(rl.MapMetalness).Texture =
		rl.LoadTexture(filepath.Join("res", "texture", "road_mra.png"))
	floorModel.Materials.GetMap(rl.MapNormal).Texture =
		rl.LoadTexture(filepath.Join("res", "texture", "road_n.png"))
}

func (fl Floor) Draw() {
	if isEnableLight := false; isEnableLight {
		textureTilingLoc := rl.GetShaderLocation(floorModel.Materials.Shader, "tiling")
		emissiveColorLoc := rl.GetShaderLocation(floorModel.Materials.Shader, "emissiveColor")
		floorTextureTiling := []float32{.5, .5}

		rl.SetShaderValue(floorModel.Materials.Shader, textureTilingLoc, floorTextureTiling, rl.ShaderUniformVec2)

		floorEmissiveColorVector4 := rl.ColorNormalize(floorModel.Materials.GetMap(rl.MapEmission).Color)
		floorEmissiveColor := []float32{floorEmissiveColorVector4.X, floorEmissiveColorVector4.Y, floorEmissiveColorVector4.Z, floorEmissiveColorVector4.W}
		rl.SetShaderValue(floorModel.Materials.Shader, emissiveColorLoc, floorEmissiveColor, rl.ShaderUniformVec2)
	}

	rl.DrawModel(floorModel, fl.Position, 5.0, rl.Beige)
	rl.DrawBoundingBox(fl.BoundingBox, rl.Brown)

	DrawXYZOrbitV(rl.Vector3Zero(), 2.)
	DrawWorldXYZAxis()
}
