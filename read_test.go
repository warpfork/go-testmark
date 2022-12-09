package testmark_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
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
	if len(doc.DataHunks) != len(doc.HunksByName) {
		t.Errorf("document hunk list has different length than hunks-by-name: %d != %d", len(doc.DataHunks), len(doc.HunksByName))
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

func TestParseSimple(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fmt.Fprintln(buf, "[testmark]:# (simple)")
	fmt.Fprintln(buf, "```")
	fmt.Fprintln(buf, "Hello, World!")
	fmt.Fprintln(buf, "```")
	doc, err := testmark.Parse(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	hunk := doc.HunksByName["simple"]
	assert(t, hunk.Body, "Hello, World!\n")
}

func TestParseTrailingWhitespace(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	fmt.Fprintln(buf, "[testmark]:# (trailing/whitespace) ")
	fmt.Fprintln(buf, "```")
	fmt.Fprintln(buf, "foo")
	fmt.Fprintln(buf, "```")
	doc, err := testmark.Parse(buf.Bytes())
	assert(t, err.Error(), `invalid markdown comment on line 1 (should look like "[testmark]:# (data-name-here)"; remove trailing whitespace)`)
	_, exists := doc.HunksByName["trailing/whitespace"]
	if exists {
		t.Errorf("hunk should not exist")
	}
}

func TestParseTrailingExtraLineBreak(t *testing.T) {
	buf := bytes.NewBuffer([]byte("[testmark]:# (extra/newline)\n\n"))
	fmt.Fprintln(buf, "```")
	fmt.Fprintln(buf, "foo")
	fmt.Fprintln(buf, "```")
	doc, err := testmark.Parse(buf.Bytes())
	assert(t, err.Error(), `invalid markdown comment on line 2. Missing code block for hunk extra/newline`)
	hunk, exists := doc.HunksByName["extra/newline"]
	if exists {
		t.Errorf("testmark should ignore these")
	}
	assert(t, hunk.Body, "")
}

func TestDuplicateHunk(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	mdPath := filepath.Join(testdata, "exampleWithDuplicateHunks.md")
	_, err = testmark.ReadFile(mdPath)
	assertRegex(t, err.Error(), "repeated testmark hunk name \"one/two/three\".*")
}

func assertRegex(t testing.TB, actual string, pattern string) {
	matched, err := regexp.MatchString(pattern, actual)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("expected %q to match pattern %q", actual, pattern)
	}
}
