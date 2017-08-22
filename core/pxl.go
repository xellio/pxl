package core

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"time"
)

// The Pxl struct definition
type Pxl struct {
	IsEncodeMode   bool
	IsDecodeMode   bool
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
		originalInfo, err := os.Stat(p.Source)
		if err != nil {
			return false, err
		}
		fmt.Println("Original size:", originalInfo.Size())

		if err := p.encodeTar(); err != nil {
			return false, err
		}

		if err := p.Encode(); err != nil {

			if err := p.removeTar(); err != nil {
				return false, err
			}
			return false, err
		}

		if err := p.removeTar(); err != nil {
			return false, err
		}

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
		//==========================================

		targetInfo, err := os.Stat(p.Target)
		if err != nil {
			return false, err
		}
		fmt.Println("PXL size:", targetInfo.Size())

		fmt.Printf("Encoding to PXL: %s\n", elapsed)
	}

	if p.IsDecodeMode {
		//******************************************
		start := time.Now()
		//==========================================
		if err := p.Decode(); err != nil {
			return false, err
		}
		//******************************************
		elapsed := time.Since(start)
		fmt.Printf("Decoding PXL: %s\n", elapsed)
		//==========================================
		if err := p.decodeTar(); err != nil {
			return false, err
		}
	}

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

func (p *Pxl) encodeTar() error {
	//******************************************
	//start := time.Now()
	//==========================================

	finfo, err := os.Stat(p.Source)
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: finfo.Name(),
		Mode: int64(finfo.Mode()),
		Size: finfo.Size(),
	}

	tarfilename := p.Source + ".tar"
	tarfile, err := os.OpenFile(tarfilename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tw := tar.NewWriter(tarfile)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(p.Source)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	//******************************************
	//elapsed := time.Since(start)
	//fmt.Printf("tar: %s\n", elapsed)
	//==========================================
	p.Source = tarfilename

	return nil
}

func (p *Pxl) decodeTar() error {
	r := bytes.NewReader(p.decodedPayload)
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		file, err := os.OpenFile(hdr.FileInfo().Name(), os.O_WRONLY|os.O_CREATE, hdr.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(file, tr); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pxl) removeTar() error {
	return os.Remove(p.Source)
}
