package testmark

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	doc, err := ReadFile(filepath.Join(testdata, "example.md"))
	fmt.Printf("%v -- %v\n", doc.DataHunks, err)
}
