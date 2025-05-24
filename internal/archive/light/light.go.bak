package light

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Light type
type LightType int

const (
	DirectionalLight LightType = iota
	PointLight
	SpotLight
)

const (
	MaxLights = 4
)

var (
	Lights     [MaxLights]Light
	LightCount int32
)

type LightBool uint

const (
	True  LightBool = 1
	False LightBool = 0
)

// Light data
type Light struct {
	Type          LightType
	EnabledBinary LightBool
	Position      rl.Vector3
	Target        rl.Vector3
	Color         [4]float32
	Intensity     float32

	// Shader light parameters locations
	TypeLoc      int32
	EnabledLoc   int32
	PositionLoc  int32
	TargetLoc    int32
	ColorLoc     int32
	IntensityLoc int32
}

// Create light with provided data
// NOTE: It updated the global lightCount and it's limited to MAX_LIGHTS
func CreateLight(typ LightType, position, target rl.Vector3, color color.RGBA, intensity float32, shader rl.Shader) Light {
	var light Light

	if LightCount < MaxLights {
		light.EnabledBinary = True
		light.Type = typ
		light.Position = position
		light.Target = target
		light.Color[0] = float32(color.R) / 255.0
		light.Color[1] = float32(color.G) / 255.0
		light.Color[2] = float32(color.B) / 255.0
		light.Color[3] = float32(color.A) / 255.0
		light.Intensity = intensity

		// NOTE: Shader parameters names for lights must match the requested ones
		light.EnabledLoc = rl.GetShaderLocation(shader, fmt.Sprintf("lights[%d].enabled", LightCount))
		light.TypeLoc = rl.GetShaderLocation(shader, fmt.Sprintf("lights[%d].type", LightCount))
		light.PositionLoc = rl.GetShaderLocation(shader, fmt.Sprintf("lights[%d].position", LightCount))
		light.TargetLoc = rl.GetShaderLocation(shader, fmt.Sprintf("lights[%d].target", LightCount))
		light.ColorLoc = rl.GetShaderLocation(shader, fmt.Sprintf("lights[%d].color", LightCount))
		light.IntensityLoc = rl.GetShaderLocation(shader, fmt.Sprintf("lights[%d].intensity", LightCount))

		UpdateLight(shader, light)

		LightCount++
	}

	return light
}

// Send light properties to shader
// NOTE: Light shader locations should be available
func UpdateLight(shader rl.Shader, light Light) {
	rl.SetShaderValue(shader, light.EnabledLoc, []float32{float32(light.EnabledBinary)}, rl.ShaderUniformInt)
	rl.SetShaderValue(shader, light.TypeLoc, []float32{float32(light.Type)}, rl.ShaderUniformInt)

	// Send to shader light position values
	position := [3]float32{light.Position.X, light.Position.Y, light.Position.Z}
	rl.SetShaderValue(shader, light.PositionLoc, position[:], rl.ShaderUniformVec3)

	// Send to shader light target position values
	target := [3]float32{light.Target.X, light.Target.Y, light.Target.Z}
	rl.SetShaderValue(shader, light.TargetLoc, target[:], rl.ShaderUniformVec3)
	rl.SetShaderValue(shader, light.ColorLoc, light.Color[:], rl.ShaderUniformVec4)
	rl.SetShaderValue(shader, light.IntensityLoc, []float32{light.Intensity}, rl.ShaderUniformFloat)
}

func GetToggledEnabledBinary(index int) LightBool {
	val := Lights[index].EnabledBinary
	switch val {
	case False:
		return True
	case True:
		return False
	default:
		panic(fmt.Sprintf("unexpected gameplay.LightBool: %#v", val))
	}
}
