// Package common provides assets and resources initialized once (if possible)
// before the game is run.
package common

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/model"
)

var (
	Font struct {
		Primary   rl.Font
		Secondary rl.Font
	}
	Music struct {
		Theme   rl.Music
		Ambient rl.Music
	}
	FX struct {
		Coin rl.Sound
	}
	Shader struct {
		// Physically-Based Rendering
		// See https://marmoset.co/posts/basic-theory-of-physically-based-rendering/
		PBR rl.Shader
	}
	Texture struct {
		CubicmapAtlas rl.Texture2D // Load cubeTexture to be applied to the cubes sides (256x256 png)
	}
	Model struct {
		// OBJ Text file format. Must include vertex position-texcoords-normals
		// information, if files references some .mtl materials file, it will be loaded (or try to).
		OBJ     model.ModelsObj

		GLB     interface{}
		FBX     interface{}
		Texture interface{}
	}
)

func GetBoundingBoxFromPositionSizeV(pos, size rl.Vector3) rl.BoundingBox {
	return rl.NewBoundingBox(
		rl.NewVector3(pos.X-size.X/2, pos.Y-size.Y/2, pos.Z-size.Z/2),
		rl.NewVector3(pos.X+size.X/2, pos.Y+size.Y/2, pos.Z+size.Z/2))
}
