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

	success, err := pxl.Process()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if success {
		fmt.Println("Success")
		fmt.Println("Output:", pxl.Target)
	}

}
