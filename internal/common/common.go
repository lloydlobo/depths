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

	FX struct{ Coin rl.Sound }

	FXS struct {
		ImpactsSoftHeavy, ImpactsSoftMedium, ImpactsGenericLight, ImpactFootStepsConcrete []rl.Sound

		RPGDrawKnife, RPGCloth []rl.Sound

		SciFiLaserLarge, SciFiLaserSmall []rl.Sound

		InterfaceClick []rl.Sound
	}

	// Models Resource

	Shader struct {
		PBR,
		Grayscale rl.Shader
	}

	Model struct {
		Dwarf rl.Model
	}

	Texture struct {
		DwarfDiffuse,
		CubicmapAtlas rl.Texture2D
	}

	ModelDungeonKit struct {
		OBJ model.ModelsOBJ
		GLB model.ModelsGLB
	}

	ModelPrototypeKit struct{}
)
