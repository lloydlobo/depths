// Copied and adapted from https://github.com/raysan5/raylib-game-template/blob/main/src/raylib_game.c
package game

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	audiopro "example/depths/internal/audio/processor"
	"example/depths/internal/common"
	"example/depths/internal/light"
	"example/depths/internal/model"
	endingSC "example/depths/internal/screen/ending"
	gameplaySC "example/depths/internal/screen/gameplay"
	logoSC "example/depths/internal/screen/logo"
	optionsSC "example/depths/internal/screen/options"
	titleSC "example/depths/internal/screen/title"
)

// XM, standing for "extended module", is an audio file type introduced by
// Triton's FastTracker 2.[2] XM introduced multisampling-capable[3]
// instruments with volume and panning envelopes,[4] sample looping[5] and
// basic pattern compression. It also expanded the available effect commands
// and channels, added 16-bit sample support, and offered an alternative
// frequency table for portamentos.

// Main entry point
func Run() {
	// Initialize

	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), "tiny game ─ depths")

	rl.InitAudioDevice()

	// Disabled: distorts when no audio is playing
	if false {
		rl.AttachAudioMixedProcessor(audiopro.ProcessAudio)
		defer rl.DetachAudioMixedProcessor(audiopro.ProcessAudio)
		audiopro.InitAudioProcessor()
	}

	// Load common assets once
	common.Font.Primary = rl.GetFontDefault()
	common.Font.Secondary = rl.LoadFont("res/mecha.png")

	common.Music.UIScreen000 = rl.LoadMusicStream(filepath.Join("res", "music", "inspiring-cinematic-ambient-116199.mp3")) // Menu/Options/Credits
	common.Music.UIScreen000.Looping = true
	rl.SetMusicVolume(common.Music.UIScreen000, common.InvPhi)
	rl.PauseMusicStream(common.Music.UIScreen000)

	common.Music.OpenWorld000 = rl.LoadMusicStream(filepath.Join("res", "music", "ambient-music-329699.mp3"))
	common.Music.OpenWorld001 = rl.LoadMusicStream(filepath.Join("res", "music", "sinnesloschen-beam-117362.mp3"))
	common.Music.OpenWorld000.Looping = true
	common.Music.OpenWorld001.Looping = true
	rl.SetMusicVolume(common.Music.OpenWorld000, 1.0)
	rl.SetMusicVolume(common.Music.OpenWorld001, 1.0)
	rl.PauseMusicStream(common.Music.OpenWorld000)
	rl.PauseMusicStream(common.Music.OpenWorld001)

	common.Music.DrillRoom000 = rl.LoadMusicStream(filepath.Join("res", "music", "mandarin-dream-118311.mp3"))
	common.Music.DrillRoom001 = rl.LoadMusicStream(filepath.Join("res", "music", "serenity-329278.mp3"))
	common.Music.DrillRoom000.Looping = true
	common.Music.DrillRoom001.Looping = true
	rl.SetMusicVolume(common.Music.DrillRoom000, 1.0)
	rl.SetMusicVolume(common.Music.DrillRoom001, 1.0)
	rl.PauseMusicStream(common.Music.DrillRoom000)
	rl.PauseMusicStream(common.Music.DrillRoom001)

	common.Music.Ambient000 = rl.LoadMusicStream(filepath.Join("res", "music", "ambient.ogg"))
	common.Music.Ambient000.Looping = true

	common.FX.Coin = rl.LoadSound("res/fx/coin.wav")
	rl.SetSoundVolume(common.FX.Coin, 0.3)

	{
		var fxAudioDir = filepath.Join("res", "fx", "kenney_impact-sounds", "Audio")
		common.FXS.FootStepsConcrete = []rl.Sound{
			rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_000.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_001.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_002.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_003.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "footstep_concrete_004.ogg")),
		}
		common.FXS.ImpactsSoftHeavy = []rl.Sound{
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_000.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_001.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_002.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_003.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_heavy_004.ogg")),
		}
		common.FXS.ImpactsSoftMedium = []rl.Sound{
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_000.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_001.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_002.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_003.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactSoft_medium_004.ogg")),
		}
		common.FXS.ImpactsGenericLight = []rl.Sound{
			rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_000.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_001.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_002.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_003.ogg")),
			rl.LoadSound(filepath.Join(fxAudioDir, "impactGeneric_light_004.ogg")),
		}
	}

	common.Texture.CubicmapAtlas = rl.LoadTexture(filepath.Join("res", "texture", "cubicmap_atlas.png"))
	common.Model.OBJ = model.LoadAssetModelOBJ()

	/* Load PBR shader and setup all required locations */
	common.Shader.PBR = rl.LoadShader(
		filepath.Join("res", "shader", "glsl330_"+"pbr.vs"), // Vertex shader
		filepath.Join("res", "shader", "glsl330_"+"pbr.fs"), // Fragment shader
	)

	updateShaderLoc := func(shader rl.Shader, index int32, uniformName string) {
		shader.UpdateLocation(index, rl.GetShaderLocation(shader, uniformName))
		// common.Shader.PBR.UpdateLocation(rl.MapAlbedo, rl.GetShaderLocation(common.Shader.PBR, "albedoMap"))
		// common.Shader.PBR.UpdateLocation(rl.MapMetalness, rl.GetShaderLocation(common.Shader.PBR, "mraMap"))
	}

	if false {
		updateShaderLoc(common.Shader.PBR, rl.ShaderLocMapAlbedo, "albedoMap")
		// WARNING: Metalness, roughness, and ambient occlusion are all packed into a MRA texture
		// They are passed as to the SHADER_LOC_MAP_METALNESS location for convenience,
		// shader already takes care of it accordingly
		updateShaderLoc(common.Shader.PBR, rl.ShaderLocMapMetalness, "mraMap")
		updateShaderLoc(common.Shader.PBR, rl.ShaderLocMapNormal, "normalMap")
		// WARNING: Similar to the MRA map, the emissive map packs different information
		// into a single texture: it stores height and emission data
		// It is binded to SHADER_LOC_MAP_EMISSION location an properly processed on shader
		updateShaderLoc(common.Shader.PBR, rl.ShaderLocMapEmission, "emissiveMap")
		updateShaderLoc(common.Shader.PBR, rl.ShaderLocColorDiffuse, "albedoColor")
	}

	const MAX_LIGHTS = 4

	if false {
		// Setup additional required shader locations, including lights data
		updateShaderLoc(common.Shader.PBR, rl.ShaderLocVectorView, "viewPos")
		lightCountLoc := rl.GetShaderLocation(common.Shader.PBR, "numOfLights")
		maxLightCount := []float32{MAX_LIGHTS}
		rl.SetShaderValue(common.Shader.PBR, lightCountLoc, maxLightCount, rl.ShaderUniformInt)

		// Setup ambient color and intensity parameters
		ambientIntensity := []float32{0.02}
		ambientColor := rl.NewColor(26, 32, 135, 255)
		acnVec3 := rl.NewVector3(float32(ambientColor.R)/255.0, float32(ambientColor.G)/255.0, float32(ambientColor.B)/255.0)
		ambientColorNormalized := []float32{acnVec3.X, acnVec3.Y, acnVec3.Z}
		rl.SetShaderValue(common.Shader.PBR, rl.GetShaderLocation(common.Shader.PBR, "ambientColor"), ambientColorNormalized, rl.ShaderUniformVec3)
		rl.SetShaderValue(common.Shader.PBR, rl.GetShaderLocation(common.Shader.PBR, "ambient"), ambientIntensity, rl.ShaderUniformFloat)
	}

	if false {
		// Create some lights
		light.Lights[0] = light.CreateLight(light.PointLight, rl.NewVector3(-1.0, 1.0, -2.0), rl.NewVector3(0.0, 0.0, 0.0), rl.Yellow, 4.0, common.Shader.PBR)
		light.Lights[1] = light.CreateLight(light.PointLight, rl.NewVector3(2.0, 1.0, 1.0), rl.NewVector3(0.0, 0.0, 0.0), rl.Green, 3.3, common.Shader.PBR)
		light.Lights[2] = light.CreateLight(light.PointLight, rl.NewVector3(-2.0, 1.0, 1.0), rl.NewVector3(0.0, 0.0, 0.0), rl.Red, 8.3, common.Shader.PBR)
		light.Lights[3] = light.CreateLight(light.PointLight, rl.NewVector3(1.0, 1.0, -2.0), rl.NewVector3(0.0, 0.0, 0.0), rl.Blue, 2.0, common.Shader.PBR)
	}

	// Setup material texture maps usage in shader
	// NOTE: By default, the texture maps are always used
	usage := []float32{1}
	rl.SetShaderValue(common.Shader.PBR, rl.GetShaderLocation(common.Shader.PBR, "useTexAlbedo"), usage, rl.ShaderUniformInt)
	rl.SetShaderValue(common.Shader.PBR, rl.GetShaderLocation(common.Shader.PBR, "useTexNormal"), usage, rl.ShaderUniformInt)
	rl.SetShaderValue(common.Shader.PBR, rl.GetShaderLocation(common.Shader.PBR, "useTexMRA"), usage, rl.ShaderUniformInt)
	rl.SetShaderValue(common.Shader.PBR, rl.GetShaderLocation(common.Shader.PBR, "useTexEmissive"), usage, rl.ShaderUniformInt)

	currentScreen = logoGameScreen
	logoSC.Init()

	if _, ok := os.LookupEnv("PLATFORM_WEB"); ok {
		// emscripten_set_main_loop(UpdateDrawFrame, 60, 1)
		log.Printf("env: %v\n", "PLATFORM_WEB")
	} else {
		rl.SetTargetFPS(60)

		// Main game loop
		for !rl.WindowShouldClose() {
			UpdateDrawFrame()
		}
	}

	// De-Initialization

	// Unload current screen data before closing
	switch currentScreen {
	case logoGameScreen:
		logoSC.Unload()
	case titleGameScreen:
		titleSC.Unload()
	case optionsGameScreen:
		optionsSC.Unload()
	case gameplayGameScreen:
		gameplaySC.Unload()
	case endingGameScreen:
		endingSC.Unload()
	case unknownGameScreen:
		break
	default:
		panic(fmt.Sprintf("unexpected main.GameScreen: %#v", currentScreen))
	}

	// Unload global data loaded
	rl.UnloadFont(common.Font.Primary)
	rl.UnloadFont(common.Font.Secondary)
	rl.UnloadMusicStream(common.Music.OpenWorld001)
	rl.UnloadMusicStream(common.Music.Ambient000)
	rl.UnloadSound(common.FX.Coin)

	// Close audio context
	rl.CloseAudioDevice()

	// Close window and OpenGL context
	rl.CloseWindow()
}

// ChangeToScreen changes to next screen, no transition.
func ChangeToScreen(screen GameScreen) {

	// Unload current screen
	switch currentScreen {
	case logoGameScreen:
		logoSC.Unload()
	case titleGameScreen:
		titleSC.Unload()
	case optionsGameScreen:
		optionsSC.Unload()
	case gameplayGameScreen:
		gameplaySC.Unload()
	case endingGameScreen:
		endingSC.Unload()
	case unknownGameScreen:
		break
	default:
		panic(fmt.Sprintf("unexpected main.GameScreen: %#v", currentScreen))
	}

	// Init next screen
	switch screen {
	case logoGameScreen:
		logoSC.Init()
	case titleGameScreen:
		titleSC.Init()
	case optionsGameScreen:
		optionsSC.Init()
	case gameplayGameScreen:
		gameplaySC.Init()
	case endingGameScreen:
		endingSC.Init()
	case unknownGameScreen:
		break
	default:
		panic(fmt.Sprintf("unexpected main.GameScreen: %#v", currentScreen))
	}

	currentScreen = screen
}

// TransitionToScreen requests transition to next screen.
func TransitionToScreen(screen GameScreen) {
	onTransition = true
	transFadeout = false
	transFromScreen = int(currentScreen)
	transToScreen = screen
	transAlpha = float32(0.0)
}

// UpdateTransition updates transition effect (fade-in, fade-out).
func UpdateTransition() {
	if !transFadeout {
		transAlpha += 0.05

		if transAlpha > 1.01 {
			transAlpha = 1.0

			// Unload current screen
			switch GameScreen(transFromScreen) {
			case logoGameScreen:
				logoSC.Unload()
			case titleGameScreen:
				titleSC.Unload()
			case optionsGameScreen:
				optionsSC.Unload()
			case gameplayGameScreen:
				gameplaySC.Unload()
			case endingGameScreen:
				endingSC.Unload()
			case unknownGameScreen:
				break
			default:
				panic(fmt.Sprintf("unexpected main.GameScreen: %#v", currentScreen))
			}

			// Load next screen
			switch transToScreen {
			case logoGameScreen:
				logoSC.Init()
			case titleGameScreen:
				titleSC.Init()
			case optionsGameScreen:
				optionsSC.Init()
			case gameplayGameScreen:
				gameplaySC.Init()
			case endingGameScreen:
				endingSC.Init()
			case unknownGameScreen:
				break
			default:
				panic(fmt.Sprintf("unexpected main.GameScreen: %#v", currentScreen))
			}

			currentScreen = transToScreen

			// Activate fade out effect to next loaded screen
			transFadeout = true
		}
	} else { // Transition fade out logic
		transAlpha -= 0.02
		if transAlpha < -0.01 {
			onTransition = false
			transFadeout = false
			transFromScreen = int(GameScreen(-1))
			transToScreen = unknownGameScreen
			transAlpha = 0.0
		}
	}
}

// DrawTransition draws transition effect (full-screen rectangle).
func DrawTransition() {
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()),
		int32(rl.GetScreenHeight()), rl.Fade(rl.Black, transAlpha))
}

// UpdateDrawFrame  updates and draws game frame.
func UpdateDrawFrame() {
	// =============================================================================
	// Update

	if false {
		rl.UpdateMusicStream(common.Music.OpenWorld001)

		// Modify processing variables
		if rl.IsKeyPressed(rl.KeyLeft) {
			audiopro.AudioExponent -= 0.05
		}
		if rl.IsKeyPressed(rl.KeyRight) {
			audiopro.AudioExponent += 0.05
		}
		if rl.IsKeyPressed(rl.KeyDown) {
			audiopro.AudioExponent -= 0.25
		}
		if rl.IsKeyPressed(rl.KeyUp) {
			audiopro.AudioExponent += 0.25
		}
	}

	if !onTransition {
		switch currentScreen {
		case logoGameScreen:
			logoSC.Update()

			if logoSC.Finish() == 1 {
				TransitionToScreen(titleGameScreen)
			}
		case titleGameScreen:
			titleSC.Update()

			if titleSC.Finish() == 1 {
				TransitionToScreen(optionsGameScreen)
			} else if titleSC.Finish() == 2 {
				TransitionToScreen(gameplayGameScreen)
			}
		case optionsGameScreen:
			optionsSC.Unload()

			if optionsSC.Finish() == 1 {
				TransitionToScreen(titleGameScreen)
			}
		case gameplayGameScreen:
			gameplaySC.Update()

			if gameplaySC.Finish() == 1 {
				TransitionToScreen(endingGameScreen)
			} else if gameplaySC.Finish() == 2 {
				TransitionToScreen(titleGameScreen)
			}
		case endingGameScreen:
			endingSC.Update()

			if endingSC.Finish() == 1 {
				TransitionToScreen(titleGameScreen)
			}
		case unknownGameScreen:
			break
		default:
			panic(fmt.Sprintf("unexpected main.GameScreen: %#v", currentScreen))
		}
	} else {
		UpdateTransition() // Update transition (fade-in, fade-out)
	}
	// -----------------------------------------------------------------------------

	// =============================================================================
	// Draw

	rl.BeginDrawing()

	rl.ClearBackground(rl.RayWhite)

	switch currentScreen {
	case logoGameScreen:
		logoSC.Draw()
	case titleGameScreen:
		titleSC.Draw()
	case optionsGameScreen:
		optionsSC.Draw()
	case gameplayGameScreen:
		gameplaySC.Draw()
	case endingGameScreen:
		endingSC.Draw()
	case unknownGameScreen:
		break
	default:
		panic(fmt.Sprintf("unexpected main.GameScreen: %#v", currentScreen))
	}

	// Draw full screen rectangle in front of everything
	if onTransition {
		DrawTransition()
	}

	rl.DrawFPS(10, 10)

	rl.EndDrawing()
	// -----------------------------------------------------------------------------
}

type GameScreen int

const (
	unknownGameScreen  GameScreen = iota - 1 // -1
	logoGameScreen                           // 0
	titleGameScreen                          // 1
	optionsGameScreen                        // 2
	gameplayGameScreen                       // 3
	endingGameScreen                         // 4
)

// =====================================================================================
// Shared Variables Definition (global)

// NOTE: Those variables are shared between modules through C equivalent of screens.h
var (
	currentScreen GameScreen
)

// =====================================================================================
// Local Variables Definition (local to this module)

// Required variables to manage screen transitions (fade-in, fade-out)
var (
	transAlpha      float32    = float32(0.0)
	onTransition    bool       = false
	transFadeout    bool       = false
	transFromScreen int        = -1
	transToScreen   GameScreen = unknownGameScreen
)
