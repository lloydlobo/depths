package currency

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
)

//go:embed template_currency_items.json
var templateInventoryCurrencyJSON []byte

const (
	defaultJSONSaveDirname  = "storage"
	defaultJSONSaveFilename = "inventory_currency.json"
)

type CurrencyType int32

const (
	Copper CurrencyType = iota // Base currency
	Pearl
	Bronze
	Silver
	Ruby
	Gold
	Diamond
	Sapphire

	MaxCurrencyTypes
)

var ToColorMap = map[CurrencyType]color.RGBA{
	Copper:   rl.Beige,
	Pearl:    rl.ColorBrightness(rl.DarkGray, 0.05),
	Bronze:   rl.Orange,
	Silver:   rl.ColorBrightness(rl.DarkGray, 0.15),
	Ruby:     rl.Maroon,
	Gold:     rl.Gold,
	Diamond:  rl.SkyBlue,
	Sapphire: rl.Yellow,
}

// ToCopperUnitsMap maps any Currency type into its equivalent Copper units.
var ToCopperUnitsMap = map[CurrencyType]int32{
	Copper:   1,
	Pearl:    25,
	Bronze:   25,
	Silver:   30,
	Ruby:     35,
	Gold:     40,
	Diamond:  80,
	Sapphire: 80,
}

var ToStringMap = map[CurrencyType]string{
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
	name := filepath.Join(common.Must(os.Getwd()), defaultJSONSaveDirname, defaultJSONSaveFilename)
	data := common.Must(json.Marshal(input))
	common.Must(common.Must(os.Create(name)).Write(data))
}

func LoadCurrencyItems(output *[MaxCurrencyTypes]CurrencyItem) {
	saveDefaultFileTemplate := func() {
		var input [MaxCurrencyTypes]CurrencyItem
		common.MustNotErrOn(json.Unmarshal(templateInventoryCurrencyJSON, &input))
		fmt.Printf("input: %v\n", input)
		SaveCurrencyItems(input)
	}

	{ // Create new save file if not found
		var isFound bool
		dirs := common.Must(os.ReadDir(filepath.Join(common.Must(os.Getwd()), defaultJSONSaveDirname)))
		for i := range dirs { // Search only the first directory hierarchy
			if entry := dirs[i]; entry.Type().IsRegular() && entry.Name() == defaultJSONSaveFilename {
				isFound = true
				break
			}
		}
		if !isFound {
			slog.Warn(defaultJSONSaveFilename + " file not found. creating new...")
			saveDefaultFileTemplate()
		}
	}

	// Read and unmarshal file contents
	name := filepath.Join(defaultJSONSaveDirname, defaultJSONSaveFilename)
	data := common.Must(os.ReadFile(name))
	var temp [MaxCurrencyTypes]CurrencyItem
	common.MustNotErrOn(json.Unmarshal(data, &temp))

	if false { // Verify file contents
		var seen = make(map[CurrencyType]struct{})
		var isInvalid bool
		for i := range temp {
			if temp[i].Type != CurrencyType(i) {
				slog.Error("want sorted currency type", "index", i, "got", temp[i].Type)
				isInvalid = true
			}
			if temp[i].Type >= MaxCurrencyTypes {
				slog.Error("want currency type less than maximum", "index", i, "got", temp[i].Type)
				isInvalid = true
			}
			if _, ok := seen[temp[i].Type]; ok {
				slog.Error("want unique currency type", "index", i, "got", temp[i].Type)
				isInvalid = true
			}
			seen[temp[i].Type] = struct{}{}
		}
		if isInvalid {
			slog.Error(defaultJSONSaveFilename + " file is invalid. overwriting it with default template...")
			common.MustNotErrOn(os.Remove(filepath.Join(name)))
			saveDefaultFileTemplate()
		}
	}

	// Transfer data to receiver
	for i := range output {
		output[i] = temp[i]
	}
}

func HandleWalletToBankTransaction(currencyItems *[MaxCurrencyTypes]CurrencyItem) {
	{
		fmt.Printf("000: currencyItems: %v\n", currencyItems)
		for i := range MaxCurrencyTypes {
			fmt.Printf("i: %v\n", i)
			(&currencyItems[i]).Bank += currencyItems[i].Wallet
			(&currencyItems[i]).Wallet = 0
		}
		fmt.Printf("111: currencyItems: %v\n", currencyItems)
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
0644     file mode specifies that the file is readable and writable by the owner, and readable by everyone else.
0666     read & write
0740     owner can read, write, & execute; group can only read; others have no permissions
*/
func runExample() {
	var input [MaxCurrencyTypes]CurrencyItem
	common.MustNotErrOn(json.Unmarshal(templateInventoryCurrencyJSON, &input))
	SaveCurrencyItems(input)

	var output [MaxCurrencyTypes]CurrencyItem
	LoadCurrencyItems(&output)
	for i := range output {
		fmt.Printf("output[i]: %+v\n", output[i])
	}
}
func init() {
	if false {
		slog.Warn(strings.Repeat("=", 79))
		for range 16 {
			slog.Warn(strings.Repeat("v", 20) + "\trunExample()\t" + strings.Repeat("v", 20))
		}
		runExample()
		for range 16 {
			slog.Warn(strings.Repeat("^", 20) + "\trunExample()\t" + strings.Repeat("^", 20))
		}
		slog.Warn(strings.Repeat("=", 79))
	}
}
