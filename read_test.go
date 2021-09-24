package testmark_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/warpfork/go-testmark"
)

func TestRead(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	if err != nil {
		t.Fatal(err)
	}

	readFixturesExample(t, doc)
}

func readFixturesExample(t *testing.T, doc *testmark.Document) {
	assert(t, doc.DataHunks[0].Name, "this-is-the-data-name")
	assert(t, doc.DataHunks[0].LineStart+1, "13")
	assert(t, doc.DataHunks[0].LineEnd+1, "17")
	assert(t, doc.DataHunks[0].Body, "the content of this code block is data which can be read,\nand *replaced*, by testmark.\n")

	assert(t, doc.DataHunks[1].Name, "more-data")
	assert(t, doc.DataHunks[1].LineStart+1, "36")
	assert(t, doc.DataHunks[1].LineEnd+1, "41")
	assert(t, doc.DataHunks[1].Body, "func OtherMarkdownParsers() (shouldHighlight bool) {\n\treturn true\n}\n")

	assert(t, doc.DataHunks[2].Name, "cannot-describe-no-linebreak")
	assert(t, doc.DataHunks[2].LineStart+1, "70")
	assert(t, doc.DataHunks[2].LineEnd+1, "73")
	assert(t, doc.DataHunks[2].Body, "A markdown codeblock always has a trailing linebreak before its close indicator, you see.\n")
}

func TestParseCRLF(t *testing.T) {
	input, err := ioutil.ReadFile(filepath.Join("testdata", "example.md"))
	if err != nil {
		t.Fatal(err)
	}
	input = bytes.ReplaceAll(input, []byte("\n"), []byte("\r\n"))
	doc, err := testmark.Parse(input)
	if err != nil {
		t.Fatal(err)
	}

	readFixturesExample(t, doc)
}
