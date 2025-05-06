package hud

import (
	"cmp"
	"fmt"
	"math"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	"example/depths/internal/currency"
	"example/depths/internal/player"
	"example/depths/internal/util/mathutil"
)

// // Hard-coded slice
// //   - If player enters drillroom:
// //   - increment CollectedCount with OnHandCount
// //   - reset OnHandCount to 0
// var currencyItems /* = */ [currency.MaxCurrencyTypes]currency.CurrencyItem /* {
// 	currency.Copper:   {Type: currency.Copper, Wallet: 0, Bank: 0},
// 	currency.Pearl:    {Type: currency.Pearl, Wallet: 0, Bank: 0},
// 	currency.Bronze:   {Type: currency.Bronze, Wallet: 0, Bank: 0},
// 	currency.Silver:   {Type: currency.Silver, Wallet: 0, Bank: 0},
// 	currency.Ruby:     {Type: currency.Ruby, Wallet: 0, Bank: 0},
// 	currency.Gold:     {Type: currency.Gold, Wallet: 0, Bank: 0},
// 	currency.Diamond:  {Type: currency.Diamond, Wallet: 0, Bank: 0},
// 	currency.Sapphire: {Type: currency.Sapphire, Wallet: 0, Bank: 0},
// } */

// DrawHUD draws the Heads-Up-Display on 2D screen.
func DrawHUD(
	xPlayer player.Player,
	currencyItems [currency.MaxCurrencyTypes]currency.CurrencyItem,
) {
	screenW := int32(rl.GetScreenWidth())
	screenH := int32(rl.GetScreenHeight())

	//
	// Player stats: health / money / experience
	//

	const (
		marginX    = 20
		marginY    = 20
		radius     = float32(20)
		marginLeft = float32(marginX * 2. / 3.)
	)

	var (
		cargoRatio = float32(xPlayer.CargoCapacity) / float32(xPlayer.MaxCargoCapacity)
		circlePos  = rl.NewVector2(radius, 20*1)
		font       = common.Font.SourGummy
		fontSize   = float32(common.Font.SourGummy.BaseSize) * float32(cmp.Or(1.5, 3.0))
		radius0    = float32(15.)
	)

	//
	// Set the transform matrix to where the HUD Stats are
	//

	rl.PushMatrix()
	rl.Translatef(marginLeft, marginY, 0)
	{
		healthPartsCount := int32(rl.Clamp((xPlayer.Health*10.)/2., 0, 5))

		// Draw Health : Close eyes ////// Go to sleep
		//   - 1.0 == 5 hearts
		//   - 0.0 == 0 hearts
		// 		FIXME: Use a better transition effect.. circle zoom out
		// 		TODO: If player dead or health==0.. allow retry option or make it new game for that level id OR delete the level file
		if healthPartsCount <= 1 {
			var (
				f               = rl.Clamp(xPlayer.Health, 0.00025, 1.0)
				healthCirclePos = rl.Vector2Subtract(circlePos, rl.NewVector2(0, radius0/4))
				innerCol        = rl.Fade(rl.Red, 2*xPlayer.Health)
				outerCol        = rl.Fade(rl.Maroon, 3*xPlayer.Health)
				radius1         = radius0 * common.InvPhi
			)

			if f <= 0.00025 {
				radius1 /= f
				f = max(0.1, mathutil.SqrtF(f)) // Black+Red splattered screen
				outerCol = rl.Fade(rl.ColorLerp(rl.Black, outerCol, f), max(0.1, 1000*f))
				innerCol = rl.Fade(rl.ColorLerp(rl.Black, innerCol, f), max(0.1, 1000*f))
				healthCirclePos = rl.Vector2Lerp(rl.NewVector2(float32(screenW)/2, float32(screenH)/2), healthCirclePos, f*f)
			}

			rl.DrawCircleV(healthCirclePos, radius1, outerCol)
			rl.DrawCircleSector(healthCirclePos, radius1, -90, -90+xPlayer.Health*360, cmp.Or(16, (healthPartsCount-1)*2), innerCol)
		}

		for i := range healthPartsCount {
			DrawHeart(rl.Vector2Add(circlePos, rl.NewVector2(2*radius0*float32(i), 0)), radius0)
		}
	}
	rl.PopMatrix()

	rl.PushMatrix()
	rl.Translatef(marginLeft, marginY+20*3-radius/2, 0)
	{
		// Draw Cargo Capacity - [1] circle sector meter
		circlePos = rl.NewVector2(radius, radius)
		circleCutoutRec := rl.NewRectangle(radius/2., radius/2., radius, radius)

		if cargoRatio >= 1. {
			rl.DrawCircleGradient(int32(circlePos.X), int32(circlePos.Y), radius+3, rl.White, rl.Fade(rl.White, .1))
		}

		rl.DrawRectangleRoundedLinesEx(circleCutoutRec, 1., 16, 0.5+radius/2., rl.DarkGray)
		rl.DrawCircleSector(circlePos, radius, -90, -90+360*cargoRatio, 16, cmp.Or(rl.White, rl.Gold))
		rl.DrawCircleV(circlePos, radius/2, rl.Fade(rl.Gold, cargoRatio))
		rl.DrawCircleV(circlePos, radius*max(.75, (1-cargoRatio)), rl.Fade(rl.Gold, 1.-cargoRatio)) // Glass Half-Empty
		rl.DrawCircleV(circlePos, radius*max(.75, (1-cargoRatio)), rl.DarkGray)
		rl.DrawCircleV(circlePos, 8+8, rl.Gold)
		rl.DrawCircleV(circlePos, 8+4, rl.DarkGray)
		rl.DrawCircleV(circlePos, 8, rl.Gold)

		if false && cargoRatio >= 0.5 {
			rl.DrawCircleV(circlePos, radius*cargoRatio, rl.Fade(rl.Gold, 1.0)) // Glass Half-Full
		}
	}
	rl.PopMatrix()

	// Draw Cargo Capacity - [2] meter text
	rl.PushMatrix()
	rl.Translatef(marginLeft+radius*2.25, marginY+20*3+radius, 0)
	{
		fontSize := fontSize * common.InvPhi
		capText := fmt.Sprintf("%.2d", xPlayer.CargoCapacity)
		capStrLen := rl.MeasureTextEx(font, capText, fontSize, 1.0)
		divideText := fmt.Sprintf("%s", strings.Repeat("-", len(capText)))
		divideStrLen := rl.MeasureTextEx(font, divideText, fontSize, 1.0)
		maxCapText := fmt.Sprintf("%.2d", xPlayer.MaxCargoCapacity)
		maxCapStrLen := rl.MeasureTextEx(font, maxCapText, fontSize, 1.0)

		rl.DrawTextEx(font, capText, rl.NewVector2(capStrLen.X/2, -20-10/2), fontSize, 1, rl.Black)
		rl.DrawTextEx(font, capText, rl.NewVector2(capStrLen.X/2, -20-10/2), fontSize, 1, rl.Fade(rl.Yellow, 0.8))
		rl.DrawTextEx(font, divideText, rl.NewVector2(divideStrLen.X/2, -(2*10)/1.5), fontSize, 0.0625, rl.Fade(rl.LightGray, 0.3))
		rl.DrawTextEx(font, maxCapText, rl.NewVector2(maxCapStrLen.X/2, -10/2), fontSize, 1, rl.Gray)
	}
	rl.PopMatrix()

	rl.PushMatrix()
	rl.Translatef(marginLeft*.5, marginY+20*4+radius+20*.25, 0)
	{
		// currencyItems[currency.Copper].Wallet = hitScore

		for i := range currency.MaxCurrencyTypes {
			item := currencyItems[i]

			const offset = (radius * 3)
			gapY := offset * float32(i)
			// fontSize := (fontSize * 2. / 3.) - 2
			fontSize := float32(font.BaseSize)

			position := rl.NewVector2(circlePos.X, circlePos.Y+gapY)
			rl.DrawCircleV(position, min(fontSize/2, (radius*common.OneMinusInvPhi)), currency.ToColorMap[item.Type])
			{
				text := fmt.Sprintf("%d", item.Bank)
				textStringSize := rl.MeasureTextEx(font, text, fontSize, 1)
				_ = textStringSize
				pos := position
				pos.X -= textStringSize.X / 2
				pos.Y += textStringSize.Y / 4
				// pos := rl.Vector2Add(position, rl.NewVector2(-textStringSize.X/2, textStringSize.Y*.8))
				rl.DrawTextEx(font, text, pos, fontSize, 1., rl.LightGray)
			}
			if item.Wallet > 0 {
				text := fmt.Sprintf("+%d", item.Wallet)
				textStringSize := rl.MeasureTextEx(font, text, fontSize, 1)
				fontSize := fontSize - 2
				pos := rl.Vector2Add(position, rl.NewVector2(-textStringSize.X/2, textStringSize.Y*.8))
				pos = rl.Vector2Add(pos, rl.NewVector2(fontSize*1.5, -fontSize/1.5))
				rl.DrawTextEx(font, text, pos, fontSize, 1., rl.LightGray)
			}
		}
	}
	rl.PopMatrix()
}

func DrawHeart(position rl.Vector2, radius float32) {
	if isDrawBackdropCircle := false; isDrawBackdropCircle {
		rl.DrawCircleV(position, radius, rl.Fade(rl.Red, .1))
	}

	var (
		offsetX = radius / math.Pi
		l       = rl.NewVector2(position.X-offsetX, position.Y-radius/2.)
		r       = rl.NewVector2(position.X+offsetX, position.Y-radius/2.)
		ll      = rl.NewVector2(l.X-offsetX*math.Pi/2, l.Y+radius/(math.Pi*2))
		rr      = rl.NewVector2(r.X+offsetX*math.Pi/2, r.Y+radius/(math.Pi*2))
		bot     = rl.NewVector2(position.X, position.Y+radius/(math.Pi/2))
	)

	if isShowLines := false; isShowLines {
		rl.DrawTriangleLines(bot, rr, ll, rl.Red)
		rl.DrawCircleLinesV(l, radius/2., rl.Red)
		rl.DrawCircleLinesV(r, radius/2., rl.Red)
	} else {
		rl.DrawTriangle(bot, rr, ll, rl.Red)
		rl.DrawCircleV(l, radius/2., rl.Red)
		rl.DrawCircleV(r, radius/2., rl.Red)
	}
}
