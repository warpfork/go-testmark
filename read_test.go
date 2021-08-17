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
	for _, hunk := range doc.DataHunks {
		fmt.Printf("hunk %q is on lines %d:%d, has body %q\n", hunk.Name, hunk.Line+1, hunk.EndLine+1, string(hunk.Body))
	}
}
