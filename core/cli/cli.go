package cli

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/xellio/pxl/core/common"
	"os"
)

var (
	isEncode      = pflag.BoolP("encode", "e", false, "Enable encode mode")
	isDecode      = pflag.BoolP("decode", "d", false, "Enable decode mode")
	isDebugMode   = pflag.Bool("debug", false, "Enable debug mode")
	isVersionMode = pflag.BoolP("version", "v", false, "Display version number")
	source        = pflag.StringP("input", "i", "", "File (path) to convert")
	target        = pflag.StringP("output", "o", "", "Output file")
)

func UseVersionMode() bool {
	return *isVersionMode
}

func DisplayDebug(message string) {
	if *isDebugMode {
		fmt.Println(message)
	}
}

func InitFlags() (Context, error) {

	pflag.Parse()

	if UseVersionMode() {
		fmt.Printf("%s : Build %s version %f\nAuthor : %s (see: %s)\n", common.PRODUCT_NAME, common.BUILD, common.VERSION, common.AUTHOR, common.CONTACT)
		os.Exit(0)
	}

	return generateContextFromFlags()
}

func generateContextFromFlags() (Context, error) {
	context := new(Context)

	if len(*source) <= 0 {
		return *context, fmt.Errorf("Missing argument: -i (--input) is required")
	}

	if *isDecode == *isEncode {
		return *context, fmt.Errorf("Logic error: encode and decode flags are the same")
	}

	context.IsDecodeMode = *isDecode
	context.IsEncodeMode = *isEncode
	context.IsDebugMode = *isDebugMode
	context.Source = *source
	context.Target = *target

	DisplayDebug(context.DebugString())

	return *context, nil
}
