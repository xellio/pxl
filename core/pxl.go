package core

import (
	"errors"
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

		f, err := os.OpenFile(p.Target, os.O_WRONLY|os.O_CREATE, 0600)

		if err != nil {
			panic(err)
		}

		defer f.Close()
		//******************************************
		start := time.Now()
		//==========================================
		png.Encode(f, p.encodedPayload)
		//******************************************
		elapsed := time.Since(start)
		fmt.Printf("encoding to png: %s\n", elapsed)
		//==========================================
	}

	if p.IsDecodeMode {
		if err := p.Decode(); err != nil {
			return false, err
		}
		// @todo: handle tar
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

	f, err := os.OpenFile(p.Source, os.O_RDONLY, 0444)
	defer f.Close()
	if err != nil {
		return err
	}

	finfo, err := os.Stat(p.Source)
	if err != nil {
		return err
	}

	dimensions := int(math.Sqrt(float64(finfo.Size()/4))) + 1

	x := 0
	y := 0

	//create image with dimensions x dimensions
	img := image.NewNRGBA((image.Rect(0, 0, dimensions, dimensions)))

	//fillPx := color.NRGBA{0, 0, 0, 255}
	//draw.Draw(img, img.Bounds(), &image.Uniform{fillPx}, image.ZP, draw.Src)

	var buffer = make([]byte, 838860800)
	tmp := make([]byte, 4)

	for {
		num, err := f.Read(buffer)

		if err != errors.New("EOF") && num == 0 {
			break
		} else if num == 0 {
			break
		} else if err != nil {
			return err
		}

		//loop msg bytes
		for pos := 0; pos < num; pos += 4 {
			for p := 0; p < 4; p++ {
				if len(buffer) < pos+p {
					tmp[p] = byte(255)
				} else {
					tmp[p] = buffer[pos+p]
				}
			}
			img.Set(x, y, color.NRGBA{tmp[pos%4], tmp[pos%4+1], tmp[pos%4+2], tmp[pos%4+3]})
			x++
			if x >= dimensions {
				y++
				x = 0
			}
		}
	}

	for posY := y; posY < dimensions; posY++ {
		for posX := x; posX < dimensions; posX++ {
			img.Set(posX, posY, color.NRGBA{0, 0, 0, 255})
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
