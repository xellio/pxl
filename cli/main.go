package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/pflag"
	"github.com/xellio/pxl"
)

var (
	encodeFlag    = pflag.StringP("encode", "e", "", "Enable encode mode")
	decodeFlag    = pflag.StringP("decode", "d", "", "Enable decode mode")
	isVersionMode = pflag.BoolP("version", "v", false, "Display version number")
	maxProcs      = pflag.IntP("procs", "p", runtime.NumCPU(), "Number of threads to use")
)

func main() {
	// Parsing and validate arguments before running
	pxl, err := initFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = pxl.Process()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if pxl.IsEncodeMode {
		fmt.Println("PXL-File:", pxl.Target)
	} else {
		fmt.Println("Success")
	}

}

// Returns the isVersionMode flag
func useVersionMode() bool {
	return *isVersionMode
}

// InitFlags initializes and validates the given flags.
// Returns a new Pxl struct if valid
// If isVersionMode flag is set, some information is shown and the programm is exited
func initFlags() (pxl.Pxl, error) {
	pflag.Parse()

	if useVersionMode() {
		fmt.Printf("%s : Version %f\nAuthor : %s (see: %s)\n", pxl.ProductName, pxl.Version, pxl.Author, pxl.Contact)
		os.Exit(0)
	}
	runtime.GOMAXPROCS(*maxProcs)
	pxl, err := generatePxlFromFlags()
	if err != nil {
		usage()
	}
	return pxl, err

}

// Creates and returns a new Pxl struct
// Returns an error if invalid values were passed by pflags
func generatePxlFromFlags() (pxl.Pxl, error) {
	pxl := new(pxl.Pxl)

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

// show usage information
func usage() {
	fmt.Println(`
Usage: 
    pxl [option] [file]

The options are:
    -v, --version
        Display version information
    -e, --encode
        Encode the given file
    -d, --decode
        Decode the given (pxl) file
    -p, --procs
    	Specify the number of threads to use (default = NumCPU)
		`)
}
