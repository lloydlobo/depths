package currency

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

type CurrencyType int32

const (
	Copper   CurrencyType = iota // 0 Base currency
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
	Pearl:    rl.ColorBrightness(rl.DarkGray, 0.05),
	Bronze:   rl.Orange,
	Silver:   rl.ColorBrightness(rl.DarkGray, 0.15),
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
	return (ct + 1) % MaxCurrencyTypes
}

type CurrencyItem struct {
	Type   CurrencyType `json:"type"`
	Wallet int32        `json:"wallet"`
	Bank   int32        `json:"bank"`
}

// NOTE: If the file already exists, it is truncated.
// NOTE: If the file does not exist, it is created with mode 0o666 (before umask).
func SaveCurrencyItems(input [MaxCurrencyTypes]CurrencyItem) {
	name := filepath.Join("storage", "inventory_currency.json")
	data := common.Must(json.Marshal(input))
	common.Must(common.Must(os.Create(name)).Write(data))
}

func LoadCurrencyItems(output *[MaxCurrencyTypes]CurrencyItem) {
	name := filepath.Join("storage", "inventory_currency.json")
	data := common.Must(os.ReadFile(name))
	err := json.Unmarshal(data, output)
	if err != nil {
		panic(err)
	}
}

/*
0000     no permissions
0700     read, write, & execute only for the owner
0770     read, write, & execute for owner and group
0777     read, write, & execute for owner, group and others
0111     execute
0222     write
0333     write & execute
0444     read
0555     read & execute
0666     read & write
0740     owner can read, write, & execute; group can only read; others have no permissions

0644     file mode specifies that the file is readable and writable by the owner, and readable by everyone else.

[

	    {"Type":0,"AmountInWallet":0,"AmountInBank":0}, {"Type":1,"AmountInWallet":0,"AmountInBank":0},
		{"Type":2,"AmountInWallet":0,"AmountInBank":0}, {"Type":3,"AmountInWallet":0,"AmountInBank":0},
		{"Type":4,"AmountInWallet":0,"AmountInBank":0}, {"Type":5,"AmountInWallet":0,"AmountInBank":0},
		{"Type":6,"AmountInWallet":0,"AmountInBank":0}, {"Type":7,"AmountInWallet":0,"AmountInBank":0}

]
*/
func example() {
	var input [MaxCurrencyTypes]CurrencyItem
	SaveCurrencyItems(input)

	var output [MaxCurrencyTypes]CurrencyItem
	LoadCurrencyItems(&output)
	for i := range output {
		fmt.Printf("output[i]: %+v\n", output[i])
	}
}
