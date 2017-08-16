package core

import (
	"archive/tar"
	"bytes"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/xellio/pxl/core/common"
	"io"
	"os"
	"time"
)

var (
	isEncode      = pflag.BoolP("encode", "e", false, "Enable encode mode")
	isDecode      = pflag.BoolP("decode", "d", false, "Enable decode mode")
	isDebugMode   = pflag.Bool("debug", false, "Enable debug mode")
	isVersionMode = pflag.BoolP("version", "v", false, "Display version number")
	source        = pflag.StringP("input", "i", "", "File (path) to convert")
	target        = pflag.StringP("output", "o", "out.png", "Output file") // remove this as soon as the tar logic works
	useTar        = pflag.BoolP("tar", "t", false, "tar option")           // for temp disabling the tar logic
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

	if *isEncode && *useTar {
		source, err := convertToTar(*source)
		pxl.Source = source
		if err != nil {
			return *pxl, err
		}
	} else {
		pxl.Source = *source
	}

	pxl.Target = *target
	pxl.IsDecodeMode = *isDecode
	pxl.IsEncodeMode = *isEncode
	pxl.IsDebugMode = *isDebugMode

	DisplayDebug(pxl.DebugString("initialized pxl"))

	return *pxl, nil
}

func convertToTar(source string) (string, error) {
	//******************************************
	start := time.Now()
	//==========================================
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	finfo, err := os.Stat(source)
	if err != nil {
		return source, err
	}

	header := &tar.Header{
		Name: finfo.Name(),
		Mode: int64(finfo.Mode()),
		Size: finfo.Size(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return source, err
	}

	f, err := os.Open(source)
	defer f.Close()
	if err != nil {
		return source, err
	}

	//if _, err := tw.Write([]byte(file.Body)); err != nil {
	if _, err := io.Copy(tw, f); err != nil {
		return source, err
	}

	if err := tw.Close(); err != nil {
		return source, err
	}

	//******************************************
	elapsed := time.Since(start)
	fmt.Printf("tar: %s\n", elapsed)
	//==========================================
	return source, nil
}
