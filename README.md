## Synopsis

pxl is a small command-line tool for converting text or files to png images. In some cases this can save some bytes when transfering the data or just retuns a spacy image.

## Code Example
```
package main

import (
	"fmt"
	"github.com/xellio/pxl/core/cli"
	"os"
)

func main() {
	// Parsing and validate arguments before running
	var context, err = cli.InitFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if context.IsEncodeMode {
		fmt.Println("ENCODE")
	}

	if context.IsDecodeMode {
		fmt.Println("DECODE")
	}
}
```

## Motivation

Learning go

## Installation

go get github.com/xellio/pxl

## API Reference


## Tests


## Contributors


## License
