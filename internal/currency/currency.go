package currency

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type CurrencyType int32

const (
	Copper   CurrencyType = iota // 0
	Pearl                        // 1
	Bronze                       // 2
	Silver                       // 3
	Ruby                         // 4
	Gold                         // 5
	Diamond                      // 6
	Sapphire                     // 7

	MaxCurrencyTypes
)

var CurrencyColorMap = map[CurrencyType]color.RGBA{
	Copper:   rl.Beige,
	Pearl:    rl.ColorBrightness(rl.Black, 0.05),
	Bronze:   rl.Orange,
	Silver:   rl.ColorBrightness(rl.Black, 0.2),
	Ruby:     rl.Maroon,
	Gold:     rl.Gold,
	Diamond:  rl.SkyBlue,
	Sapphire: rl.Yellow,
}

// CurrencyConversionRateMap maps Currency in Copper units.
var CurrencyConversionRateMap = map[CurrencyType]int32{
	Copper:   1,
	Pearl:    25,
	Bronze:   25,
	Silver:   30,
	Ruby:     35,
	Gold:     40,
	Diamond:  80,
	Sapphire: 80,
}

var CurrencyStringMap = map[CurrencyType]string{
	Copper:   "Copper",
	Pearl:    "Pearl",
	Bronze:   "Bronze",
	Silver:   "Silver",
	Ruby:     "Ruby",
	Gold:     "Gold",
	Diamond:  "Diamond",
	Sapphire: "Sapphire",
}

func (ct CurrencyType) Next() CurrencyType {
	next := (ct + 1) % MaxCurrencyTypes
	fmt.Printf("next currency type: %v\n", next)
	return next
}
