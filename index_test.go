package testmark_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/warpfork/go-testmark"
)

func TestIndexingDirs(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	if err != nil {
		panic(err)
	}
	doc.BuildDirIndex()

	if len(doc.DirEnt.ChildrenList) != 2 {
		t.Errorf("root dirent list should be length 2 but was %d", len(doc.DirEnt.ChildrenList))
	}
	if len(doc.DirEnt.Children) != 2 {
		t.Errorf("root dirent map should be length 2 but was %d", len(doc.DirEnt.Children))
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

func TestIndexingTree(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	if err != nil {
		panic(err)
	}
	if len(doc.DataHunks) != len(doc.HunksByName) {
		t.Errorf("doc hunk list has different length than hunks-by-name: %d != %d",
			len(doc.DataHunks), len(doc.HunksByName))
	}
	doc.BuildDirIndex()

	for _, _hunk := range doc.DataHunks {
		hunk := _hunk
		t.Run("index:"+hunk.Name, func(t *testing.T) {
			assertHunkReachable(t, doc, hunk)
		})
	}
	assertChildren(t, doc.DirEnt)
}

func assertHunkReachable(t *testing.T, doc *testmark.Document, hunk testmark.DocHunk) {
	splits := strings.Split(hunk.Name, "/")
	dir := doc.DirEnt
	for _, split := range splits {
		if len(dir.Children) != len(dir.ChildrenList) {
			t.Errorf("expected dir to have equal number of children in both data structures")
		}
		child, ok := dir.Children[split]
		if !ok {
			t.Errorf("expected dir %q to have child named %q", dir.Name, split)
		}
		dir = child
	}
	assert(t, hunk.Hunk, fmt.Sprintf("%v", *dir.Hunk))
}

func assertChildren(t *testing.T, dir *testmark.DirEnt) {
	foundChildren := make(map[string]struct{})
	for _, _child := range dir.ChildrenList {
		child := _child
		t.Run(child.Name, func(t *testing.T) {
			_, exists := foundChildren[child.Name]
			if exists {
				t.Errorf("dir %q has duplicate child: %q", dir.Name, child.Name)
			}
			foundChildren[child.Name] = struct{}{}

			mapChild, exists := dir.Children[child.Name]
			if !exists {
				t.Errorf("dir %q missing child: %q", dir.Name, child.Name)
			}
			if child != mapChild {
				t.Errorf("child %q should have equivalent pointers: %p %p", child.Name, child, mapChild)
			}

			assertChildren(t, child)
		})
	}
	// If the lengths are equal then the map and list should contain entries with the same names.
	// We don't know if the dir entries are _actually_ equivalent but the test recurses above so it should be fine.
	if len(dir.Children) != len(dir.ChildrenList) {
		t.Errorf(
			"expected dir to have equal number of children in both data structures"+
				"\n\t%s:\n\tlist: %v\n\tkeys: %v",
			dir.Name, names(dir.ChildrenList), keys(dir.Children))
	}
}

func keys(dirs map[string]*testmark.DirEnt) []string {
	names := make([]string, 0, len(dirs))
	for k := range dirs {
		names = append(names, k)
	}
	return names
}

func names(dirs []*testmark.DirEnt) []string {
	names := make([]string, 0, len(dirs))
	for _, d := range dirs {
		names = append(names, d.Name)
	}
	return names
}
