package pxl

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

func Encode(msg []byte, outputFileName string) (encodedImage image.Image, err error) {
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

	f, _ := os.OpenFile(outputFileName, os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	png.Encode(f, img)

	return img, nil
}

func Decode(path string) (msg []byte, err error) {
	img, err := loadImage(path)
	if err != nil {
		return nil, err
	}

	for _, c := range img.Pix {
		msg = append(msg, c)
	}
	return msg, nil
}

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
