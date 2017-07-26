package pxl

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"time"
)

// The Pxl struct definition
type Pxl struct {
	IsEncodeMode   bool
	IsDecodeMode   bool
	IsDebugMode    bool
	Source         string
	Target         string
	encodedPayload image.Image
	decodedPayload []byte
}

// Checks the context on the Pxl struct
// Encode the Source if Pxl.IsEncodeMode
// Decode the Source if Pxl.IsDecodeMode
func (p Pxl) Process() (bool, error) {

	if p.IsEncodeMode {
		if err := p.Encode(); err != nil {
			return false, err
		}
		p.DebugString()

		f, _ := os.OpenFile(p.Target, os.O_WRONLY|os.O_CREATE, 0600)
		defer f.Close()
		png.Encode(f, p.encodedPayload)
	}

	if p.IsDecodeMode {
		if err := p.Decode(); err != nil {
			return false, err
		}

		err := ioutil.WriteFile(p.Target, p.decodedPayload, 0644)
		if err != nil {
			return false, err
		}
	}

	DisplayDebug(p.DebugString("DONE"))

	return true, nil
}

// Encodes the Pxl.Source and stores it to Pxl.encodedPayload
func (p *Pxl) Encode() error {

	msg, err := ioutil.ReadFile(p.Source)
	if err != nil {
		return err
	}

	dimensions := int(math.Sqrt(float64(len(msg)/4))) + 1

	var pxl [][][]byte

	var line [][]byte
	var pixel []byte

	//loop msg bytes
	for i, c := range msg {
		if i > 0 && (i%4) == 0 {
			line = append(line, pixel)
			pixel = nil
		}
		pixel = append(pixel, c)

		if len(line) > 0 && (len(line)%dimensions) == 0 {
			pxl = append(pxl, line)
			line = nil
		}
	}

	//calculate n missing values in incomplete color
	if len(pixel) < 4 {
		missing := 4 - len(pixel)
		for i := 0; i < missing; i++ {
			pixel = append(pixel, 255)
		}
	}

	//append incomplete color to line
	line = append(line, pixel)

	//calculate n missing values in line
	if len(line) < dimensions {
		missing := (dimensions - len(line))
		missingColor := []byte{0, 0, 0, 255}
		//add n missing colors to line
		for i := 0; i < missing; i++ {
			line = append(line, missingColor)
		}
	}
	pxl = append(pxl, line)

	//create image with dimensions x dimensions
	img := image.NewNRGBA((image.Rect(0, 0, dimensions, dimensions)))

	x := 0
	y := 0

	for i, line := range pxl {
		x = 0
		if i > 0 {
			y++
		}

		//each pixel in line
		for _, pixel := range line {
			img.Set(x, y, color.NRGBA{pixel[0], pixel[1], pixel[2], pixel[3]})
			x++
		}
	}

	p.encodedPayload = img

	return nil

}

// Decoded the Pxl.Source and stores it to Pxl.decodedPayload
func (p *Pxl) Decode() error {
	img, err := loadImage(p.Source)
	if err != nil {
		return err
	}

	for _, i := range img.Pix {
		p.decodedPayload = append(p.decodedPayload, i)
	}
	return nil
}

// Load image from filesystem
func loadImage(path string) (*image.NRGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	return img.(*image.NRGBA), nil
}

// Outputs a petty presentation of the Pxl struct
func (p Pxl) DebugString(msg ...string) string {
	x, y := 0, 0
	if p.encodedPayload != nil {
		x = p.encodedPayload.Bounds().Max.X
		y = p.encodedPayload.Bounds().Max.Y
	}

	if len(msg) <= 0 {
		msg = append(msg, time.Now().String())
	}

	return fmt.Sprintf(`%s
isDebugMode     : %t
isEncode        : %t
isDecode        : %t
source          : %s
target          : %s
encodedPayload  : %d x %d
decodedPayload  : %d`,
		msg,
		p.IsEncodeMode,
		p.IsDecodeMode,
		p.IsDebugMode,
		p.Source,
		p.Target,
		x, y,
		len(p.decodedPayload),
	)
}
