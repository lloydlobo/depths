package logo

import (
	textutil "example/depths/internal/util/textutil"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Init() {
	framesCounter = 0
	state3FramesCounter = 0
	finishScreen = 0

	logoPositionX = int32(rl.GetScreenWidth())/2 - 128
	logoPositionY = int32(rl.GetScreenHeight())/2 - 128

	lettersCount = 0

	topSideRecWidth = 16
	leftSideRecHeight = 16

	bottomSideRecWidth = 16
	rightSideRecHeight = 16

	state = 0
	alpha = float32(1.0)
}

func Update() {
	// Press enter or tap to change to title game screen (quick jump)
	if rl.IsKeyDown(rl.KeyEnter) || rl.IsGestureDetected(rl.GestureDoubletap) {
		finishScreen = 1
		// rl.PlaySound(fxCoin)
	}

	if state == 0 {
		framesCounter++

		if framesCounter == 80*transSpeedMultiplier {
			state = 1
			// Reset counter... will be used later...
			framesCounter = 0
		}
	} else if state == 1 {
		// State 1: Bars animation logic: top and left
		topSideRecWidth += 8
		leftSideRecHeight += 8

		if topSideRecWidth == 256 {
			state = 2
		}
	} else if state == 2 {
		// State 2: Bars animation logic: bottom and right
		bottomSideRecWidth += 8
		rightSideRecHeight += 8

		if bottomSideRecWidth == 256 {
			state = 3
		}
	} else if state == 3 {
		// State 3: "raylib" text-write animation logic
		framesCounter++
		// Tracks total frames elapsed since state==3 (avoid flicker when framesCounter = 0)
		state3FramesCounter++

		if lettersCount < 10 {
			const n = int32(len(logoText) * 2) // => 12
			if framesCounter/n > 0 {
				// Every 12 frames, one more letter
				lettersCount++
				framesCounter = 0
			}
		} else {
			// When all letters have appeared, just fade out everything
			if framesCounter > 200*transSpeedMultiplier {
				alpha -= 0.02

				if alpha <= 0.0 {
					alpha = 0.0
					// Jump to next screen
					finishScreen = 1
				}
			}
		}
	}
}

func Draw() {
	screenH := int32(rl.GetScreenHeight())
	screenW := int32(rl.GetScreenWidth())
	rl.DrawRectangle(0, 0, screenW, screenH, rl.Black)
	if state == 0 {
		// Draw blinking top-left square corner
		if (framesCounter/10)%2 > 0 {
			rl.DrawRectangle(logoPositionX, logoPositionY, 16, 16, rl.White)
		}
	} else if state == 1 {
		// Draw bars animation: top and left
		rl.DrawRectangle(logoPositionX, logoPositionY, topSideRecWidth, 16, rl.White)
		rl.DrawRectangle(logoPositionX, logoPositionY, 16, leftSideRecHeight, rl.White)
	} else if state == 2 {
		// Draw bars animation: bottom and right
		rl.DrawRectangle(logoPositionX, logoPositionY, topSideRecWidth, 16, rl.White)
		rl.DrawRectangle(logoPositionX, logoPositionY, 16, leftSideRecHeight, rl.White)

		rl.DrawRectangle(logoPositionX+240, logoPositionY, 16, rightSideRecHeight, rl.White)
		rl.DrawRectangle(logoPositionX, logoPositionY+240, bottomSideRecWidth, 16, rl.White)
	} else if state == 3 {
		// Draw "raylib" text-write animation + "powered by"
		rl.DrawRectangle(logoPositionX, logoPositionY, topSideRecWidth, 16,
			rl.Fade(rl.White, alpha))
		rl.DrawRectangle(logoPositionX, logoPositionY+16, 16,
			leftSideRecHeight-32, rl.Fade(rl.White, alpha))

		rl.DrawRectangle(logoPositionX+240, logoPositionY+16, 16,
			rightSideRecHeight-32, rl.Fade(rl.White, alpha))
		rl.DrawRectangle(logoPositionX, logoPositionY+240, bottomSideRecWidth,
			16, rl.Fade(rl.White, alpha))

		rl.DrawRectangle(screenW/2-112, screenH/2-112, 224, 224, rl.Fade(rl.Black, alpha))
		rl.DrawText(textutil.Subtext(logoText, 0, lettersCount), screenW/2-44,
			screenH/2+48, 50, rl.Fade(rl.White, alpha))
		if state3FramesCounter > 20*(transSpeedMultiplier*2) {
			rl.DrawText("powered by", logoPositionX, logoPositionY-27, 20,
				rl.Fade(rl.LightGray, alpha))
		}
	}
}

func Unload() {
	// Unload LOGO screen variables here!
}

// Logo Screen should finish?
func Finish() int {
	return finishScreen
}

const logoText = "raylib" // "raylib"
const transSpeedMultiplier = 0.125 / 5

// Module Variables Definition (local)
var (
	framesCounter       int32 = 0
	state3FramesCounter int32 = 0
	finishScreen        int   = 0

	logoPositionX int32 = 0
	logoPositionY int32 = 0

	lettersCount = 0

	topSideRecWidth   int32 = 0
	leftSideRecHeight int32 = 0

	bottomSideRecWidth int32 = 0
	rightSideRecHeight int32 = 0

	state = 0            // Logo animation states
	alpha = float32(1.0) // Useful for fading
)
