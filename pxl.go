package pxl

import (
	"archive/tar"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"time"
)

//
// The Pxl struct definition
//
type Pxl struct {
	IsEncodeMode   bool
	IsDecodeMode   bool
	Source         string
	Target         string
	encodedPayload *image.NRGBA
	decodedPayload []byte
}

// Process checks the context on the Pxl struct
// Encode the Source if Pxl.IsEncodeMode
// Decode the Source if Pxl.IsDecodeMode
func (p *Pxl) Process() error {

	if p.IsEncodeMode {
		originalInfo, err := os.Stat(p.Source)
		if err != nil {
			return err
		}
		fmt.Println("Original size:", originalInfo.Size())
		fmt.Println("Start encoding... This can take some time, CPU and memory. Be patient...")

		start := time.Now()

		if err = p.Encode(); err != nil {
			return err
		}

		f, err := os.OpenFile(p.Target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer func() {
			if derr := f.Close(); derr != nil {
				fmt.Println(derr)
			}
		}()

		if err = png.Encode(f, p.encodedPayload); err != nil {
			return err
		}

		elapsed := time.Since(start)

		targetInfo, err := os.Stat(p.Target)
		if err != nil {
			return err
		}
		fmt.Println("PXL size:", targetInfo.Size())

		fmt.Printf("Encoding to PXL: %s\n", elapsed)
	}

	if p.IsDecodeMode {
		start := time.Now()
		if err := p.Decode(); err != nil {
			return err
		}
		elapsed := time.Since(start)
		fmt.Printf("Decoding PXL: %s\n", elapsed)
		if err := p.decodeTar(); err != nil {
			return err
		}
	}

	return nil
}

// Encode reads the Source file, wraps it in tar and encodes it as pixel data
func (p *Pxl) Encode() error {
	tarData, err := tarToMemory(p.Source)
	if err != nil {
		return err
	}

	p.encodedPayload = buildImage(tarData)
	return nil
}

// Decode reads a PXL PNG and stores the raw pixel data
func (p *Pxl) Decode() error {
	img, err := loadImage(p.Source)
	if err != nil {
		return err
	}

	p.decodedPayload = img.Pix
	return nil
}

// tarToMemory wraps a file in a tar archive in memory
func tarToMemory(source string) ([]byte, error) {
	finfo, err := os.Stat(source)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Grow(int(finfo.Size()) + 1024) // pre-allocate: file size + tar overhead

	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name: finfo.Name(),
		Mode: int64(finfo.Mode()),
		Size: finfo.Size(),
	}

	if err = tw.WriteHeader(header); err != nil {
		return nil, err
	}

	file, err := os.Open(source)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err = io.Copy(tw, file); err != nil {
		return nil, err
	}

	if err = tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// buildImage creates an NRGBA image from raw byte data
func buildImage(data []byte) *image.NRGBA {
	pixelCount := (len(data) + 3) / 4
	dimensions := int(math.Sqrt(float64(pixelCount))) + 1

	img := image.NewNRGBA(image.Rect(0, 0, dimensions, dimensions))
	copy(img.Pix, data)

	// Set alpha to 255 for padding pixels
	for i := len(data); i < len(img.Pix); i++ {
		if i%4 == 3 {
			img.Pix[i] = 255
		}
	}

	return img
}

// decodeTar extracts files from the decoded pixel data
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

		if err := extractTarEntry(hdr, tr); err != nil {
			return err
		}
	}
	return nil
}

// extractTarEntry writes a single tar entry to disk
func extractTarEntry(hdr *tar.Header, tr *tar.Reader) error {
	file, err := os.OpenFile(hdr.FileInfo().Name(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, tr)
	return err
}

// loadImage loads a PNG image from filesystem
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

	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		return nil, fmt.Errorf("unexpected image format: %T", img)
	}
	return nrgba, nil
}
