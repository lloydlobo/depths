package ending

import (
	"example/depths/internal/common"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Init() {
	framesCounter = 0
	finishScreen = 0
}

func Update() {
	// TODO: Update ending screen variables here!

	// Press enter or tap to change to ENDING screen
	if rl.IsKeyDown(rl.KeyEnter) || rl.IsGestureDetected(rl.GestureDoubletap) {
		finishScreen = 1
		rl.PlaySound(common.FX.Coin)
	}
}

func Draw() {
	// TODO: Draw ending screen here!
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
	// TODO: Unload ending screen variables here!
}

// Ending screen should finish?
func Finish() int {
	return finishScreen
}

const (
	screenTitleText    = "GAMEOVER" // This should be temporary during prototype
	screenSubtitleText = "continue" // "press enter or tap to jump to title screen"
)

var (
	framesCounter int32 = 0
	finishScreen  int   = 0
)
