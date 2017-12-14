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
	"runtime"
	"sync"
	"time"
)

var (
	maxBufferSize = 838860800
	bufferSize    = int64(838860800)
)

//
// The Pxl struct definition
//
type Pxl struct {
	IsEncodeMode   bool
	IsDecodeMode   bool
	Source         string
	Target         string
	encodedPayload image.Image
	decodedPayload []byte
}

//
// internal helper struct
//
type splittedResult struct {
	Index   int
	Payload []color.NRGBA
}

//
// internal scope struct
//
type scope struct {
	Start int64
	End   int64
}

// Checks the context on the Pxl struct
// Encode the Source if Pxl.IsEncodeMode
// Decode the Source if Pxl.IsDecodeMode
func (p Pxl) Process() error {

	if p.IsEncodeMode {
		originalInfo, err := os.Stat(p.Source)
		if err != nil {
			return err
		}
		fmt.Println("Original size:", originalInfo.Size())
		fmt.Println("Start encoding... This can take some time, CPU and memory. Be patient...")
		//******************************************
		start := time.Now()
		//==========================================
		if err := p.encodeTar(); err != nil {
			return err
		}

		if err := p.Encode(); err != nil {

			if err := p.removeTar(); err != nil {
				return err
			}
			return err
		}

		if err := p.removeTar(); err != nil {
			return err
		}

		f, err := os.OpenFile(p.Target, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		png.Encode(f, p.encodedPayload)
		//******************************************
		elapsed := time.Since(start)
		//==========================================

		targetInfo, err := os.Stat(p.Target)
		if err != nil {
			return err
		}
		fmt.Println("PXL size:", targetInfo.Size())

		fmt.Printf("Encoding to PXL: %s\n", elapsed)
	}

	if p.IsDecodeMode {
		//******************************************
		start := time.Now()
		//==========================================
		if err := p.Decode(); err != nil {
			return err
		}
		//******************************************
		elapsed := time.Since(start)
		fmt.Printf("Decoding PXL: %s\n", elapsed)
		//==========================================
		if err := p.decodeTar(); err != nil {
			return err
		}
	}

	return nil
}

// Encodes the Pxl.Source and stores it to Pxl.encodedPayload
func (p *Pxl) Encode() error {

	finfo, err := os.Stat(p.Source)
	if err != nil {
		return err
	}

	chunksize := finfo.Size() / int64(runtime.NumCPU())
	scopes, err := calculateScopes(chunksize)
	if err != nil {
		return err
	}

	c := make(chan splittedResult, len(scopes))

	var wg sync.WaitGroup
	wg.Add(len(scopes))

	for index, se := range scopes {
		go func(index int, se scope) {
			p.encodeChunk(index, se, c)
			wg.Done()
		}(index, se)
	}
	wg.Wait()
	close(c)
	p.setEncodedPayload(c)

	return nil
}

// Encode a Chunk for later processing
func (p *Pxl) encodeChunk(index int, se scope, c chan splittedResult) {
	res := splittedResult{Index: index}
	f, _ := os.OpenFile(p.Source, os.O_RDONLY, 0444)
	var buffer = make([]byte, bufferSize)
	tmp := make([]byte, 4)
	offset := se.Start
	for {
		num, err := f.ReadAt(buffer, offset)

		if err != errors.New("EOF") && num == 0 {
			break
		} else if num == 0 {
			break
		} else if err != nil {
			break
		}

		//loop msg bytes
		for pos := 0; pos < num; pos += 4 {
			for p := 0; p < 4; p++ {
				tmp[p] = byte(255)
				if (len(buffer) - 1) < pos+p {
					tmp[p] = byte(255)
				} else {
					tmp[p] = buffer[pos+p]
				}
			}
			res.Payload = append(res.Payload, color.NRGBA{tmp[pos%4], tmp[pos%4+1], tmp[pos%4+2], tmp[pos%4+3]})
		}

		offset := offset + bufferSize
		if offset >= se.End {
			break
		}
	}
	c <- res
}

// Append encoded data to Pxl struct
func (p *Pxl) setEncodedPayload(c <-chan splittedResult) {
	sorted := make(map[int]splittedResult)

	for sr := range c {
		sorted[sr.Index] = sr
	}

	var data []color.NRGBA
	for i := 0; i < len(sorted); i++ {
		data = append(data, sorted[i].Payload...)
	}

	dimensions := int(math.Sqrt(float64(len(data)))) + 1
	img := image.NewNRGBA((image.Rect(0, 0, dimensions, dimensions)))

	x := 0
	y := 0
	for _, rgba := range data {
		img.Set(x, y, rgba)
		x++
		if x >= dimensions {
			y++
			x = 0
		}
	}
	for posY := y; posY < dimensions; posY++ {
		for posX := x; posX < dimensions; posX++ {
			img.Set(posX, posY, color.NRGBA{0, 0, 0, 255})
		}
	}
	p.encodedPayload = img
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

// Tar data to encode
func (p *Pxl) encodeTar() error {

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

	p.Source = tarfilename

	return nil
}

// Untar decoded PXL data
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

// Remove the temporary tar file
func (p *Pxl) removeTar() error {
	return os.Remove(p.Source)
}

// Calculate the scopes depending on the CPU
func calculateScopes(chunksize int64) (map[int]scope, error) {
	if chunksize < int64(maxBufferSize) {
		bufferSize = chunksize
	}

	scopes := make(map[int]scope)
	// calculate start and end of each chunk
	for i := 0; i < runtime.NumCPU(); i++ {
		start := int64(i) * chunksize
		end := start + chunksize
		scopes[i] = scope{start, end}
	}
	return scopes, nil
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
