package main

import (
	"fmt"
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
	cpuprofile    = pflag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile    = pflag.String("memprofile", "", "write memory profile to `file`")
)

func main() {
	p, err := initFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create CPU profile:", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, "could not start CPU profile:", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	err = p.Process()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not create memory profile:", err)
			os.Exit(1)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, "could not write memory profile:", err)
			os.Exit(1)
		}
	}

	if p.IsEncodeMode {
		fmt.Println("PXL-File:", p.Target)
	} else {
		fmt.Println("Success")
	}
}

// initFlags parses and validates CLI flags.
// Returns a new Pxl struct if valid.
func initFlags() (pxl.Pxl, error) {
	pflag.Parse()

	if *isVersionMode {
		fmt.Printf("%s : Version %s\nAuthor : %s (see: %s)\n", pxl.ProductName, pxl.Version, pxl.Author, pxl.Contact)
		os.Exit(0)
	}

	p, err := generatePxlFromFlags()
	if err != nil {
		usage()
	}
	return p, err
}

// generatePxlFromFlags creates a Pxl struct from the parsed flags.
func generatePxlFromFlags() (pxl.Pxl, error) {
	var p pxl.Pxl

	if *encodeFlag == "" && *decodeFlag == "" {
		return p, fmt.Errorf("Missing argument: input is required")
	}

	if *encodeFlag != "" && *decodeFlag != "" {
		return p, fmt.Errorf("Logic error: only encode or decode flag is allowed")
	}

	if *encodeFlag != "" {
		p.IsEncodeMode = true
		p.Source = *encodeFlag
		p.Target = *encodeFlag + ".pxl"
	}

	if *decodeFlag != "" {
		p.Source = *decodeFlag
		p.IsDecodeMode = true
	}

	return p, nil
}

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
        Decode the given (pxl) file`)
}
