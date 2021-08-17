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
		Hunk{Name: "this-one-is-new", BlockTag: "json", Body: []byte(`{"hayo": "new data!"}`)},
		Hunk{Name: "so-is-this", BlockTag: "json", Body: []byte(`{"appending": "is fun"}`)},
	)
	t.Logf("%s", doc.String())
}
