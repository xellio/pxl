package pxl

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "testfile.txt")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

func roundTrip(t *testing.T, source, target string) string {
	t.Helper()

	p := &Pxl{
		IsEncodeMode: true,
		Source:        source,
		Target:        target,
	}
	require.NoError(t, p.Process())

	stats, err := os.Stat(target)
	require.NoError(t, err)
	require.NotZero(t, stats.Size())

	decodeDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(decodeDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	decodePxl := &Pxl{
		IsDecodeMode: true,
		Source:        target,
	}
	require.NoError(t, decodePxl.Process())

	return decodeDir
}

func TestProcessRoundTrip(t *testing.T) {
	testContent := "this is a test"
	source := createTestFile(t, testContent)

	decodeDir := roundTrip(t, source, source+".pxl")

	decoded, err := os.ReadFile(filepath.Join(decodeDir, "testfile.txt"))
	require.NoError(t, err)
	require.Equal(t, testContent, string(decoded))
}

func TestProcessRoundTripLargeContent(t *testing.T) {
	testContent := "Hello, this is a longer test string that spans multiple pixels!"
	source := createTestFile(t, testContent)

	decodeDir := roundTrip(t, source, source+".pxl")

	decoded, err := os.ReadFile(filepath.Join(decodeDir, "testfile.txt"))
	require.NoError(t, err)
	require.Equal(t, testContent, string(decoded))
}

func TestProcessRoundTripBinaryData(t *testing.T) {
	var buf []byte
	for i := 0; i < 256; i++ {
		buf = append(buf, byte(i))
	}
	source := filepath.Join(t.TempDir(), "binary.bin")
	require.NoError(t, os.WriteFile(source, buf, 0644))

	decodeDir := roundTrip(t, source, source+".pxl")

	decoded, err := os.ReadFile(filepath.Join(decodeDir, "binary.bin"))
	require.NoError(t, err)
	require.Equal(t, buf, decoded)
}

func TestProcessEncodeNonexistentFile(t *testing.T) {
	p := &Pxl{
		IsEncodeMode: true,
		Source:        "/tmp/does-not-exist-at-all.txt",
		Target:        "/tmp/out.pxl",
	}
	require.Error(t, p.Process())
}

func TestProcessDecodeNonexistentFile(t *testing.T) {
	p := &Pxl{
		IsDecodeMode: true,
		Source:        "/tmp/does-not-exist-at-all.pxl",
	}
	require.Error(t, p.Process())
}

func TestProcessNoMode(t *testing.T) {
	p := &Pxl{Source: "/tmp/whatever"}
	require.NoError(t, p.Process())
}

func TestEncodeNonexistentSource(t *testing.T) {
	p := &Pxl{Source: "/tmp/nonexistent-file-xyz"}
	require.Error(t, p.Encode())
}

func TestDecodeInvalidPNG(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "bad.pxl")
	require.NoError(t, os.WriteFile(bad, []byte("not a png"), 0644))

	p := &Pxl{Source: bad}
	require.Error(t, p.Decode())
}

func TestDecodeNonNRGBAImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "paletted.png")
	palette := color.Palette{color.Black, color.White}
	img := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)

	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(f, img))
	require.NoError(t, f.Close())

	p := &Pxl{Source: path}
	err = p.Decode()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected image format")
}

func TestLoadImageNonexistent(t *testing.T) {
	_, err := loadImage("/tmp/does-not-exist.png")
	require.Error(t, err)
}

func TestLoadImageInvalidFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "garbage.png")
	require.NoError(t, os.WriteFile(path, []byte("garbage"), 0644))
	_, err := loadImage(path)
	require.Error(t, err)
}

func TestTarToMemory(t *testing.T) {
	source := createTestFile(t, "tar test content")

	data, err := tarToMemory(source)
	require.NoError(t, err)
	require.NotEmpty(t, data)
}

func TestTarToMemoryNonexistent(t *testing.T) {
	_, err := tarToMemory("/tmp/this-does-not-exist.txt")
	require.Error(t, err)
}

func TestBuildImage(t *testing.T) {
	t.Run("data aligned to 4 bytes", func(t *testing.T) {
		data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		img := buildImage(data)

		require.NotNil(t, img)
		// First 8 bytes should match data
		require.Equal(t, data, img.Pix[:len(data)])
	})

	t.Run("data not aligned to 4 bytes", func(t *testing.T) {
		data := []byte{1, 2, 3, 4, 5}
		img := buildImage(data)

		require.NotNil(t, img)
		require.Equal(t, data, img.Pix[:len(data)])
	})

	t.Run("empty data", func(t *testing.T) {
		img := buildImage([]byte{})
		require.NotNil(t, img)
	})

	t.Run("padding has alpha 255", func(t *testing.T) {
		data := []byte{1, 2, 3, 4}
		img := buildImage(data)

		// All padding pixels should have alpha=255
		for i := len(data) + 3; i < len(img.Pix); i += 4 {
			require.Equal(t, byte(255), img.Pix[i], "padding pixel alpha at index %d", i)
		}
	})
}

func TestEncodeDecodeDirectAPI(t *testing.T) {
	source := createTestFile(t, "direct API test")

	p := &Pxl{Source: source}
	require.NoError(t, p.Encode())
	require.NotNil(t, p.encodedPayload)

	// Write PNG to disk
	target := source + ".pxl"
	f, err := os.Create(target)
	require.NoError(t, err)
	require.NoError(t, png.Encode(f, p.encodedPayload))
	require.NoError(t, f.Close())

	dp := &Pxl{Source: target}
	require.NoError(t, dp.Decode())
	require.NotEmpty(t, dp.decodedPayload)
}

func TestProcessEncodeUnwritableTarget(t *testing.T) {
	source := createTestFile(t, "test")

	p := &Pxl{
		IsEncodeMode: true,
		Source:        source,
		Target:        "/nonexistent-dir/out.pxl",
	}
	require.Error(t, p.Process())
}
