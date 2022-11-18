package testmark

import (
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestIndexingDirs(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		panic(err)
	}
	doc, err := ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
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
	qt.Assert(t, err, qt.IsNil)
	doc, err := ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	if err != nil {
		panic(err)
	}
	if len(doc.DataHunks) != len(doc.HunksByName) {
		t.Errorf("doc hunk list has different length than hunks-by-name: %d != %d", len(doc.DataHunks), len(doc.HunksByName))
	}
	doc.BuildDirIndex()

	qt.Assert(t, len(doc.DataHunks), qt.Equals, len(doc.HunksByName))
	for _, hunk := range doc.DataHunks {
		t.Run("index:"+hunk.Name, func(t *testing.T) {
			assertHunkReachable(t, doc, hunk)
		})
	}
	assertChildren(t, doc.DirEnt)
}

func assertHunkReachable(t *testing.T, doc *Document, hunk DocHunk) {
	splits := strings.Split(hunk.Name, "/")
	dir := doc.DirEnt
	for _, s := range splits {
		qt.Assert(t, len(dir.Children), qt.Equals, len(dir.ChildrenList))
		d, ok := dir.Children[s]
		qt.Assert(t, ok, qt.IsTrue)
		dir = d
	}
	qt.Assert(t, &hunk.Hunk, qt.DeepEquals, dir.Hunk)
}

func assertChildren(t *testing.T, dir *DirEnt) {
	cm := make(map[string]struct{})
	for _, child := range dir.ChildrenList {
		{
			// Check that a child of this name has not been found already
			_, exists := cm[child.Name]
			qt.Check(t, exists, qt.IsFalse, qt.Commentf("duplicate child"))
			cm[child.Name] = struct{}{}
		}
		{
			// Check that the child in the list exists in the child map
			_, exists := dir.Children[child.Name]
			qt.Check(t, exists, qt.IsTrue)
		}
		assertChildren(t, child)
	}
	qt.Check(t, len(dir.Children), qt.Equals, len(dir.ChildrenList),
		// If the lengths are equal then the map and list should contain entries with the same names.
		// We don't know if the dir entries are _actually_ equivalent but the test recurses above so it should be fine.
		qt.Commentf("%s", dir.Name),
		qt.Commentf("list: %v", names(dir.ChildrenList)),
		qt.Commentf("keys: %v", keys(dir.Children)),
	)
}

func keys(dirs map[string]*DirEnt) []string {
	names := make([]string, 0, len(dirs))
	for k := range dirs {
		names = append(names, k)
	}
	return names
}

func names(dirs []*DirEnt) []string {
	names := make([]string, 0, len(dirs))
	for _, d := range dirs {
		names = append(names, d.Name)
	}
	return names
}
