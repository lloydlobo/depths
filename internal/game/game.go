// Copied and adapted from https://github.com/raysan5/raylib-game-template/blob/main/src/raylib_game.c
package game

import (
	"fmt"
	"log"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	endingSC "example/depths/internal/screen/ending"
	gameplaySC "example/depths/internal/screen/gameplay"
	logoSC "example/depths/internal/screen/logo"
	optionsSC "example/depths/internal/screen/options"
	titleSC "example/depths/internal/screen/title"
)

// Main entry point
func Run() {
	// Initialize

	rl.InitWindow(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), "tiny game ─ depths")

	rl.InitAudioDevice()

	// Load common assets once
	common.Font.Primary = rl.GetFontDefault()
	common.Font.Secondary = rl.LoadFont("res/mecha.png")
	common.Music.Ambient = rl.LoadMusicStream("res/music/ambient.ogg")
	common.Music.Theme = rl.LoadMusicStream("res/music/mini1111.xm")
	common.Music.Theme.Looping = false
	common.FX.Coin = rl.LoadSound("res/music/coin.wav")

	rl.SetMusicVolume(common.Music.Theme, 1.0)
	rl.PauseMusicStream(common.Music.Theme)

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
