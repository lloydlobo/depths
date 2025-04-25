package title

import (
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

func Init() {
	// TODO: Initialize title game screen variables here!
	framesCounter = 0
	finishScreen = 0
	if !rl.IsMusicStreamPlaying(common.Music.UIScreen000) {
		rl.PlayMusicStream(common.Music.UIScreen000)
	}
}

func Update() {
	rl.UpdateMusicStream(common.Music.UIScreen000)

	// Press enter or tap to change to GAMEPLAY screen
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsGestureDetected(rl.GestureDoubletap) {
		// finishScreen=1// optionsGameScreen
		finishScreen = 2 // gameplayGameScreen
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "rollover3.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_ui-audio", "Audio", "switch33.ogg")))
		rl.PlaySound(rl.LoadSound(filepath.Join("res", "fx", "kenney_interface-sounds", "Audio", "confirmation_001.ogg")))
	}
}

func Draw() {
	// TODO: Draw title game screen here!
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.Black, 0.98))
	fontThatIsInGameDotGo := rl.GetFontDefault()

	fontSize := float32(fontThatIsInGameDotGo.BaseSize) * 3.0
	pos := rl.NewVector2(
		float32(rl.GetScreenWidth())/2-float32(rl.MeasureText(screenTitleText, int32(fontSize)))/2,
		float32(rl.GetScreenHeight())/2.25,
	)
	rl.DrawTextEx(fontThatIsInGameDotGo, screenTitleText, pos, fontSize, 4, rl.White)

	posX := int32(rl.GetScreenWidth())/2 - rl.MeasureText(screenSubtitleText, 20)/2
	posY := int32(rl.GetScreenHeight()) / 2
	rl.DrawText(screenSubtitleText, posX, posY, 20, rl.White)
}

func Unload() {
	// Unload LOGO screen variables here!
}

// Title game screen should finish?
func Finish() int {
	return finishScreen
}

const (
	screenTitleText    = "DEPTHS" // "TITLE SCREEN"
	screenSubtitleText = "enter"  //"press enter or tap to jump to gameplay screen"
)

// Module Variables Definition (local)
var (
	framesCounter int32 = 0
	finishScreen  int   = 0
)
