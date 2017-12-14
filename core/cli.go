package core

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"runtime"
)

var (
	encodeFlag    = pflag.StringP("encode", "e", "", "Enable encode mode")
	decodeFlag    = pflag.StringP("decode", "d", "", "Enable decode mode")
	isVersionMode = pflag.BoolP("version", "v", false, "Display version number")
	maxProcs      = pflag.IntP("procs", "p", runtime.NumCPU(), "Number of threads to use")
)

// Returns the isVersionMode flag
func UseVersionMode() bool {
	return *isVersionMode
}

// Initializes and validates the given flags.
// Returns a new Pxl struct if valid
// If isVersionMode flag is set, some information is shown and the programm is exited
func InitFlags() (Pxl, error) {
	pflag.Parse()

	if UseVersionMode() {
		fmt.Printf("%s : Version %f\nAuthor : %s (see: %s)\n", PRODUCT_NAME, VERSION, AUTHOR, CONTACT)
		os.Exit(0)
	}
	runtime.GOMAXPROCS(*maxProcs)
	return generatePxlFromFlags()
}

// Creates and returns a new Pxl struct
// Returns an error if invalid values were passed by pflags
func generatePxlFromFlags() (Pxl, error) {
	pxl := new(Pxl)

	if len(*encodeFlag) <= 0 && len(*decodeFlag) <= 0 {
		return *pxl, fmt.Errorf("Missing argument: input is required")

	}

	if len(*encodeFlag) > 0 && len(*decodeFlag) > 0 {
		return *pxl, fmt.Errorf("Logic error: only encode or decode flag is allowed")
	}

	if len(*encodeFlag) > 0 {
		pxl.IsEncodeMode = true
		pxl.Source = *encodeFlag
		pxl.Target = *encodeFlag + ".pxl"
	}

	if len(*decodeFlag) > 0 {
		pxl.Source = *decodeFlag
		pxl.IsDecodeMode = true
	}

	return *pxl, nil
}
