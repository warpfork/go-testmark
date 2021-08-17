package testmark

import (
	"path/filepath"
	"testing"
)

func TestPatch(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	doc, err := ReadFile(filepath.Join(testdata, "example.md"))
	doc = Patch(doc,
		Hunk{Name: "more-data", BlockTag: "text", Body: []byte("you have been...\nreplaced.\nand gotten\nrather longer.")},
		Hunk{Name: "this-is-the-data-name", BlockTag: "", Body: []byte("this one gets shorter.")},
	)
	t.Logf("%s", doc.String())
}
