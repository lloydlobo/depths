// Package common provides assets and resources initialized once (if possible)
// before the game is run.
//
// Physically-Based Rendering
//
//	See https://marmoset.co/posts/basic-theory-of-physically-based-rendering/
//
// OBJ Text file format.
//
//	Must include vertex position-texcoords-normals information, if files
//	references some .mtl materials file, it will be loaded (or try to).
package common

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/model"
)

var (
	// User data

	SavedgameSlotData SavedgameSlotDataType

	// Text Resource

	Font struct{ Primary, Secondary rl.Font }

	// Audio Resource

	Music struct {
		OpenWorld000,
		OpenWorld001,
		DrillRoom000,
		DrillRoom001,
		UIScreen000,
		UIScreen001,
		Ambient000 rl.Music
	}

	FX  struct{ Coin rl.Sound }
	FXS struct {
		ImpactsSoftHeavy, ImpactsSoftMedium, ImpactsGenericLight,
		FootStepsConcrete []rl.Sound
	}

	// Models Resource

	Shader struct{ PBR rl.Shader }

	Texture struct{ CubicmapAtlas rl.Texture2D }

	ModelDungeonKit struct {
		OBJ model.ModelsOBJ
		GLB model.ModelsGLB
	}

	ModelPrototypeKit struct{}
)
