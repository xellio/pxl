package pxl

import (
	"archive/tar"
	"bytes"
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

const (
	maxBufferSize = 32 * 1024 * 1024 // 32MB
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
type splitResult struct {
	Index   int
	Payload []color.NRGBA
	Error   error
}

//
// internal scope struct
//
type scope struct {
	Start int64
	End   int64
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

		if err = p.encodeTar(); err != nil {
			return err
		}

		if err = p.Encode(); err != nil {
			_ = p.removeTar()
			return err
		}

		if err = p.removeTar(); err != nil {
			return err
		}

		f, err := os.OpenFile(p.Target, os.O_WRONLY|os.O_CREATE, 0600)
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

// Encode the Pxl.Source and stores it to Pxl.encodedPayload
func (p *Pxl) Encode() error {

	finfo, err := os.Stat(p.Source)
	if err != nil {
		return err
	}

	fileSize := finfo.Size()
	scopes, bufSize := calculateScopes(fileSize)

	c := make(chan splitResult, len(scopes))

	var wg sync.WaitGroup
	wg.Add(len(scopes))

	for index, se := range scopes {
		go func(index int, se scope) {
			p.encodeChunk(index, se, bufSize, c)
			wg.Done()
		}(index, se)
	}
	wg.Wait()
	close(c)

	return p.setEncodedPayload(c)
}

// Encode a Chunk for later processing
func (p *Pxl) encodeChunk(index int, se scope, bufSize int64, c chan splitResult) {
	res := splitResult{Index: index}

	f, err := os.Open(p.Source)
	if err != nil {
		res.Error = err
		c <- res
		return
	}
	defer f.Close()

	buffer := make([]byte, bufSize)
	offset := se.Start

	for offset < se.End {
		num, err := f.ReadAt(buffer, offset)

		// Clamp to scope boundary
		if offset+int64(num) > se.End {
			num = int(se.End - offset)
		}

		// Convert bytes to NRGBA colors
		for pos := 0; pos < num; pos += 4 {
			r, g, b, a := byte(255), byte(255), byte(255), byte(255)
			if pos < num {
				r = buffer[pos]
			}
			if pos+1 < num {
				g = buffer[pos+1]
			}
			if pos+2 < num {
				b = buffer[pos+2]
			}
			if pos+3 < num {
				a = buffer[pos+3]
			}
			res.Payload = append(res.Payload, color.NRGBA{r, g, b, a})
		}

		if err == io.EOF || num == 0 {
			break
		}
		if err != nil {
			res.Error = err
			c <- res
			return
		}

		offset += int64(num)
	}

	c <- res
}

// Append encoded data to Pxl struct
func (p *Pxl) setEncodedPayload(c <-chan splitResult) error {
	sorted := make(map[int]splitResult)

	for sr := range c {
		if sr.Error != nil {
			return sr.Error
		}
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
		x = 0
	}
	p.encodedPayload = img
	return nil
}

// Decode the Pxl.Source and stores it to Pxl.decodedPayload
func (p *Pxl) Decode() error {
	img, err := loadImage(p.Source)
	if err != nil {
		return err
	}

	p.decodedPayload = append(p.decodedPayload, img.Pix...)
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

	defer func() {
		if derr := tarfile.Close(); derr != nil {
			fmt.Println(derr)
		}
	}()

	tw := tar.NewWriter(tarfile)

	if err = tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(p.Source)
	if err != nil {
		return err
	}
	defer func() {
		if derr := file.Close(); derr != nil {
			fmt.Println(derr)
		}
	}()

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

		if err := extractTarEntry(hdr, tr); err != nil {
			return err
		}
	}
	return nil
}

// extractTarEntry writes a single tar entry to disk
func extractTarEntry(hdr *tar.Header, tr *tar.Reader) error {
	file, err := os.OpenFile(hdr.FileInfo().Name(), os.O_WRONLY|os.O_CREATE, hdr.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, tr)
	return err
}

// Remove the temporary tar file
func (p *Pxl) removeTar() error {
	return os.Remove(p.Source)
}

// Calculate the scopes depending on the CPU
func calculateScopes(fileSize int64) (map[int]scope, int64) {
	numCPU := runtime.NumCPU()
	chunkSize := fileSize / int64(numCPU)

	bufSize := chunkSize
	if bufSize > int64(maxBufferSize) {
		bufSize = int64(maxBufferSize)
	}
	if bufSize <= 0 {
		bufSize = 1
	}

	scopes := make(map[int]scope)
	for i := 0; i < numCPU; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize
		if i == numCPU-1 {
			end = fileSize
		}
		scopes[i] = scope{start, end}
	}
	return scopes, bufSize
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

	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		return nil, fmt.Errorf("unexpected image format: %T", img)
	}
	return nrgba, nil
}
