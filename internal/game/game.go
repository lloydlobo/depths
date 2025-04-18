// Copied and adapted from https://github.com/raysan5/raylib-game-template/blob/main/src/raylib_game.c
package game

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	audiopro "example/depths/internal/audio/processor"
	"example/depths/internal/common"
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

	common.Music.Ambient = rl.LoadMusicStream("res/music/ambient.ogg")
	common.Music.Theme = rl.LoadMusicStream(filepath.Join("res", "music", cmp.Or("mini1111.xm", "infraction-moments_passed.wav")))
	common.Music.Theme.Looping = false
	rl.SetMusicVolume(common.Music.Theme, 0.125)
	if false {
		rl.PauseMusicStream(common.Music.Theme)
	} else {
		rl.PlayMusicStream(common.Music.Theme)
	}
	common.FX.Coin = rl.LoadSound("res/fx/coin.wav")
	rl.SetSoundVolume(common.FX.Coin, 0.3)

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
	rl.UnloadMusicStream(common.Music.Theme)
	rl.UnloadMusicStream(common.Music.Ambient)
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

	rl.UpdateMusicStream(common.Music.Theme)

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
