package testmark

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := ReadFile(filepath.Join(testdata, "example.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, hunk := range doc.DataHunks {
		t.Logf("hunk %q is on lines %d:%d, has body %q\n", hunk.Name, hunk.LineStart+1, hunk.LineEnd+1, string(hunk.Body))
	}
}

func TestParseCRLF(t *testing.T) {
	input, err := ioutil.ReadFile(filepath.Join("testdata", "example.md"))
	if err != nil {
		t.Fatal(err)
	}
	input = bytes.ReplaceAll(input, []byte("\n"), []byte("\r\n"))
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, hunk := range doc.DataHunks {
		t.Logf("hunk %q is on lines %d:%d, has body %q\n", hunk.Name, hunk.LineStart+1, hunk.LineEnd+1, string(hunk.Body))
	}
}
