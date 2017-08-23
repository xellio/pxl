package main

import (
	"fmt"
	pxl "github.com/xellio/pxl/core"
	"os"
)

func main() {
	// Parsing and validate arguments before running
	pxl, err := pxl.InitFlags()
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
