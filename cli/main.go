package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/spf13/pflag"
	"github.com/xellio/pxl"
)

var (
	encodeFlag    = pflag.StringP("encode", "e", "", "Enable encode mode")
	decodeFlag    = pflag.StringP("decode", "d", "", "Enable decode mode")
	isVersionMode = pflag.BoolP("version", "v", false, "Display version number")
	maxProcs      = pflag.IntP("procs", "p", runtime.NumCPU(), "Number of threads to use")
	cpuprofile    = pflag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile    = pflag.String("memprofile", "", "write memory profile to `file`")
)

func main() {
	// Parsing and validate arguments before running
	p, err := initFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	err = p.Process()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	if p.IsEncodeMode {
		fmt.Println("PXL-File:", p.Target)
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
		fmt.Printf("%s : Version %s\nAuthor : %s (see: %s)\n", pxl.ProductName, pxl.Version, pxl.Author, pxl.Contact)
		os.Exit(0)
	}

	runtime.GOMAXPROCS(*maxProcs)
	p, err := generatePxlFromFlags()
	if err != nil {
		usage()
	}
	return p, err

}

// Creates and returns a new Pxl struct
// Returns an error if invalid values were passed by pflags
func generatePxlFromFlags() (pxl.Pxl, error) {
	p := new(pxl.Pxl)

	if len(*encodeFlag) <= 0 && len(*decodeFlag) <= 0 {
		return *p, fmt.Errorf("Missing argument: input is required")

	}

	if len(*encodeFlag) > 0 && len(*decodeFlag) > 0 {
		return *p, fmt.Errorf("Logic error: only encode or decode flag is allowed")
	}

	if len(*encodeFlag) > 0 {
		p.IsEncodeMode = true
		p.Source = *encodeFlag
		p.Target = *encodeFlag + ".pxl"
	}

	if len(*decodeFlag) > 0 {
		p.Source = *decodeFlag
		p.IsDecodeMode = true
	}

	return *p, nil
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
