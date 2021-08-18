package testmark

import (
	"path/filepath"
	"testing"
)

func TestIndexingDirs(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	doc, err := ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	doc.BuildDirIndex()
	if len(doc.DirEnt.ChildrenList) != 2 {
		t.Errorf("root dirent list should be length 2")
	}
	if len(doc.DirEnt.Children) != 2 {
		t.Errorf("root dirent map should be length 2")
	}
	if doc.DirEnt.ChildrenList[0].Name != "one" {
		t.Errorf("first child of root dirent should've been 'one'")
	}
	if doc.DirEnt.ChildrenList[1].Name != "really" {
		t.Errorf("second child of root dirent should've been 'really'")
	}
	if doc.DirEnt.Children["one"].Hunk.Name != "one" {
		t.Errorf("hunk 'one' looked up through dir maps should still have the right name")
	}
	if string(doc.DirEnt.Children["one"].Hunk.Body) != "baz\n" {
		t.Errorf("hunk 'one' looked up through dir maps should have the right content")
	}
	if doc.DirEnt.Children["really"].Children["deep"].Children["dirs"].Children["wow"].Hunk.Name != "really/deep/dirs/wow" {
		t.Errorf("hunk 'really/deep/dirs/wow' looked up through dir maps should still have the right name")
	}
	if string(doc.DirEnt.Children["really"].Children["deep"].Children["dirs"].Children["wow"].Hunk.Body) != "zot\n" {
		t.Errorf("hunk 'really/deep/dirs/wow' looked up through dir maps should have the right content")
	}
}
