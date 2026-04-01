# PXL

pxl encodes arbitrary files into PNG images and decodes them back. It wraps the input in a tar archive, maps the raw bytes to pixel data and produces a lossless PNG. Decoding reverses the process exactly.

[![go report card](https://goreportcard.com/badge/github.com/xellio/pxl "go report card")](https://goreportcard.com/report/github.com/xellio/pxl)
[![CI](https://github.com/xellio/pxl/actions/workflows/ci.yml/badge.svg)](https://github.com/xellio/pxl/actions/workflows/ci.yml)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/xellio/pxl.svg)](https://pkg.go.dev/github.com/xellio/pxl)

A ~20 MB log file with repetitive content compresses down to ~100 KB as a PNG:

```
./pxl -e logfile.log
```

## Installation

Build from source:

```
make build
```

The binary will be at `./bin/pxl`.

## Usage

### Encoding

```
./pxl -e example/xellio.jpg
```

<img src="./example/xellio.jpg" width="163" alt="original file"> ➜ ![pxl image](./example/xellio.jpg.pxl?raw=true "pxl image")

### Decoding

```
./pxl -d example/xellio.jpg.pxl
```

![pxl image](./example/xellio.jpg.pxl?raw=true "pxl image") ➜ <img src="./example/xellio.jpg" width="163" alt="original file">

### Options

```
-e, --encode       Encode the given file
-d, --decode       Decode the given (pxl) file
-v, --version      Display version information
    --cpuprofile   Write CPU profile to file
    --memprofile   Write memory profile to file
```

## Library usage

```go
import "github.com/xellio/pxl"

p := &pxl.Pxl{
    IsEncodeMode: true,
    Source:        "input.txt",
    Target:        "input.txt.pxl",
}
if err := p.Process(); err != nil {
    log.Fatal(err)
}
```

## Is this better than other compression?

No, probably not. But it looks good.
