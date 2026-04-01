package pxl

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestProcessRoundTrip(t *testing.T) {
	testContent := "this is a test"
	source := createTestFile(t, testContent)
	target := source + ".pxl"

	p := &Pxl{
		IsEncodeMode: true,
		Source:        source,
		Target:        target,
	}

	require.NoError(t, p.Process())

	stats, err := os.Stat(target)
	require.NoError(t, err)
	require.NotZero(t, stats.Size())

	// Decode into temp dir
	decodeDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(decodeDir))
	defer os.Chdir(origDir)

	decodePxl := &Pxl{
		IsDecodeMode: true,
		Source:        target,
	}
	require.NoError(t, decodePxl.Process())

	decoded, err := os.ReadFile(filepath.Join(decodeDir, "testfile.txt"))
	require.NoError(t, err)
	require.Equal(t, testContent, string(decoded))
}

func TestProcessRoundTripLargeContent(t *testing.T) {
	// Content larger than 4 bytes to exercise multi-pixel encoding
	// and not evenly divisible by 4 to test padding
	testContent := "Hello, this is a longer test string that spans multiple pixels!"
	source := createTestFile(t, testContent)
	target := source + ".pxl"

	p := &Pxl{
		IsEncodeMode: true,
		Source:        source,
		Target:        target,
	}
	require.NoError(t, p.Process())

	decodeDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(decodeDir))
	defer os.Chdir(origDir)

	decodePxl := &Pxl{
		IsDecodeMode: true,
		Source:        target,
	}
	require.NoError(t, decodePxl.Process())

	decoded, err := os.ReadFile(filepath.Join(decodeDir, "testfile.txt"))
	require.NoError(t, err)
	require.Equal(t, testContent, string(decoded))
}

func TestProcessRoundTripBinaryData(t *testing.T) {
	// Binary content with all byte values 0-255
	var buf []byte
	for i := 0; i < 256; i++ {
		buf = append(buf, byte(i))
	}
	source := filepath.Join(t.TempDir(), "binary.bin")
	require.NoError(t, os.WriteFile(source, buf, 0644))
	target := source + ".pxl"

	p := &Pxl{
		IsEncodeMode: true,
		Source:        source,
		Target:        target,
	}
	require.NoError(t, p.Process())

	decodeDir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(decodeDir))
	defer os.Chdir(origDir)

	decodePxl := &Pxl{
		IsDecodeMode: true,
		Source:        target,
	}
	require.NoError(t, decodePxl.Process())

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
	err := p.Process()
	require.Error(t, err)
}

func TestProcessDecodeNonexistentFile(t *testing.T) {
	p := &Pxl{
		IsDecodeMode: true,
		Source:        "/tmp/does-not-exist-at-all.pxl",
	}
	err := p.Process()
	require.Error(t, err)
}

func TestProcessNoMode(t *testing.T) {
	p := &Pxl{
		Source: "/tmp/whatever",
	}
	// Neither encode nor decode — should be a no-op without error
	require.NoError(t, p.Process())
}

func TestEncodeNonexistentSource(t *testing.T) {
	p := &Pxl{Source: "/tmp/nonexistent-file-xyz"}
	err := p.Encode()
	require.Error(t, err)
}

func TestDecodeInvalidPNG(t *testing.T) {
	// Write a non-PNG file and try to decode it
	bad := filepath.Join(t.TempDir(), "bad.pxl")
	require.NoError(t, os.WriteFile(bad, []byte("not a png"), 0644))

	p := &Pxl{Source: bad}
	err := p.Decode()
	require.Error(t, err)
}

func TestDecodeNonNRGBAImage(t *testing.T) {
	// Create a paletted PNG that won't decode to *image.NRGBA
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

func TestCalculateScopes(t *testing.T) {
	numCPU := runtime.NumCPU()

	t.Run("normal file", func(t *testing.T) {
		fileSize := int64(1000)
		scopes, bufSize := calculateScopes(fileSize)

		require.Len(t, scopes, numCPU)
		require.True(t, bufSize > 0, "bufSize should be > 0")

		// Last scope must end at fileSize
		lastScope := scopes[numCPU-1]
		require.Equal(t, fileSize, lastScope.End)

		// First scope must start at 0
		require.Equal(t, int64(0), scopes[0].Start)
	})

	t.Run("small file", func(t *testing.T) {
		scopes, bufSize := calculateScopes(int64(1))
		require.Len(t, scopes, numCPU)
		require.Equal(t, int64(1), bufSize) // clamped to minimum 1

		// Last scope must cover the entire file
		lastScope := scopes[numCPU-1]
		require.Equal(t, int64(1), lastScope.End)
	})

	t.Run("large file caps buffer", func(t *testing.T) {
		fileSize := int64(1024 * 1024 * 1024) // 1GB
		_, bufSize := calculateScopes(fileSize)
		require.True(t, bufSize <= int64(maxBufferSize), "bufSize should be capped at maxBufferSize")
	})
}

func TestEncodeTarCreatesAndCleans(t *testing.T) {
	source := createTestFile(t, "tar test content")

	p := &Pxl{Source: source}
	require.NoError(t, p.encodeTar())

	// Source should now point to .tar file
	require.True(t, strings.HasSuffix(p.Source, ".tar"))
	_, err := os.Stat(p.Source)
	require.NoError(t, err)

	// removeTar should delete it
	require.NoError(t, p.removeTar())
	_, err = os.Stat(p.Source)
	require.True(t, os.IsNotExist(err))
}

func TestEncodeTarNonexistentSource(t *testing.T) {
	p := &Pxl{Source: "/tmp/this-does-not-exist.txt"}
	err := p.encodeTar()
	require.Error(t, err)
}

func TestEncodeDecodeDirectAPI(t *testing.T) {
	// Test Encode/Decode methods directly without going through Process
	source := createTestFile(t, "direct API test")
	tarSource := source

	p := &Pxl{Source: tarSource}
	require.NoError(t, p.encodeTar())

	require.NoError(t, p.Encode())
	require.NotNil(t, p.encodedPayload)

	// Write PNG to disk
	target := source + ".pxl"
	f, err := os.Create(target)
	require.NoError(t, err)
	require.NoError(t, png.Encode(f, p.encodedPayload))
	require.NoError(t, f.Close())

	require.NoError(t, p.removeTar())

	// Decode
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
	err := p.Process()
	require.Error(t, err)
}
