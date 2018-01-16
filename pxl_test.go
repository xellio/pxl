// +build !testing

package pxl

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestFile(text string) (string, error) {
	file := "/tmp/testfile.txt"
	err := ioutil.WriteFile(file, []byte(text), 0644)
	return file, err
}

func TestProcess(t *testing.T) {

	source, err := createTestFile("this is a test")
	if err != nil {
		t.Error("Unable to create testing file")
	}

	pxl := &Pxl{
		IsEncodeMode: true,
		Source:       source,
		Target:       source + ".pxl",
	}

	err = pxl.Process()
	if err != nil {
		t.Error(err)
	}

	stats, err := os.Stat(pxl.Target)
	if err != nil {
		t.Error(err)
	}
	require.NotEqual(t, stats.Size(), 0)

	err = os.Remove(source)
	if err != nil {
		t.Error(err)
	}

	decodePxl := &Pxl{
		IsDecodeMode: true,
		Source:       pxl.Target,
	}

	err = decodePxl.Process()
	if err != nil {
		t.Error(err)
	}

}
