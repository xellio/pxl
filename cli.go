package pxl

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"

	"github.com/xellio/pxl/common"
)

var (
	isEncode      = pflag.BoolP("encode", "e", false, "Enable encode mode")
	isDecode      = pflag.BoolP("decode", "d", false, "Enable decode mode")
	isDebugMode   = pflag.Bool("debug", false, "Enable debug mode")
	isVersionMode = pflag.BoolP("version", "v", false, "Display version number")
	source        = pflag.StringP("input", "i", "", "File (path) to convert")
	target        = pflag.StringP("output", "o", "out.png", "Output file")
)

// Returns the isVersionMode flag
func UseVersionMode() bool {
	return *isVersionMode
}

// Shows debug information if this isDebugMode is set
func DisplayDebug(message string) {
	if *isDebugMode {
		fmt.Println(message)
	}
}

// Initializes and validates the given flags.
// Returns a new Pxl struct if valid
// If isVersionMode flag is set, some information is shown and the programm is exited
func InitFlags() (Pxl, error) {
	pflag.Parse()

	if UseVersionMode() {
		fmt.Printf("%s : Build %s version %f\nAuthor : %s (see: %s)\n", common.PRODUCT_NAME, common.BUILD, common.VERSION, common.AUTHOR, common.CONTACT)
		os.Exit(0)
	}

	return generatePxlFromFlags()
}

// Creates and returns a new Pxl struct
// Returns an error if invalid values were passed by pflags
func generatePxlFromFlags() (Pxl, error) {
	pxl := new(Pxl)

	if len(*source) <= 0 {
		return *pxl, fmt.Errorf("Missing argument: -i (--input) is required")
	}

	if *isDecode == *isEncode {
		return *pxl, fmt.Errorf("Logic error: encode and decode flags are the same")
	}

	pxl.IsDecodeMode = *isDecode
	pxl.IsEncodeMode = *isEncode
	pxl.IsDebugMode = *isDebugMode
	pxl.Source = *source
	pxl.Target = *target

	DisplayDebug(pxl.DebugString("initialized pxl"))

	return *pxl, nil
}
