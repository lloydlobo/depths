package options

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

func Init() {
	framesCounter = 0
	finishScreen = 0
}

func Update() {
	// TODO: Update options screen variables here!

	// Press enter or tap to change to ENDING screen
	if rl.IsKeyDown(rl.KeyEnter) || rl.IsGestureDetected(rl.GestureDoubletap) {
		finishScreen = 1
		rl.PlaySound(common.FX.Coin)
	}
}

func Draw() {
	// TODO: Draw options screen here!
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.Black, 0.98))
	fontThatIsInGameDotGo := rl.GetFontDefault()

	fontSize := float32(fontThatIsInGameDotGo.BaseSize) * 3.0
	pos := rl.NewVector2(
		float32(rl.GetScreenWidth())/2-float32(rl.MeasureText(screenTitleText, int32(fontSize)))/2,
		float32(rl.GetScreenHeight())/2.25,
	)
	rl.DrawTextEx(fontThatIsInGameDotGo, screenTitleText, pos, fontSize, 4, rl.Orange)

	position := rl.NewVector2(float32(rl.GetScreenWidth())/2-float32(rl.MeasureText(screenSubtitleText, 20)/2), float32(rl.GetScreenHeight())/2)
	rl.DrawTextEx(common.Font.SimpleMono, screenSubtitleText, position, float32(common.Font.SimpleMono.BaseSize), 1.0, rl.Orange)
}

func Unload() {
	// TODO: Unload options screen variables here!
}

// Options screen should finish?
func Finish() int {
	return finishScreen
}

const (
	screenTitleText    = "OPTIONS" // This should be temporary during prototype
	screenSubtitleText = "return"
)

var (
	framesCounter int32 = 0
	finishScreen  int   = 0
)
