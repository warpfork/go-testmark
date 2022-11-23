package fs_test

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/warpfork/go-testmark"
	tmfs "github.com/warpfork/go-testmark/fs"
)

// Assert the implementation of various interfaces in the "fs" package
var (
	_ fs.DirEntry    = &tmfs.File{}
	_ fs.File        = &tmfs.File{}
	_ fs.ReadDirFile = &tmfs.File{}
)

// TestFS tests some basic assertions about the *Document implementation of the fs.FS interface
func TestFS(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	qt.Assert(t, err, qt.IsNil)
	dfs := tmfs.DocFs(doc)

	t.Run("open dot path", func(t *testing.T) {
		f, err := dfs.Open(".")
		pathErr := new(fs.PathError)
		if qt.Check(t, err, qt.ErrorAs, &pathErr) {
			qt.Assert(t, pathErr.Op, qt.Equals, "open")
			qt.Assert(t, pathErr.Path, qt.Equals, ".")
		}
		qt.Assert(t, err, qt.ErrorIs, fs.ErrNotExist)
		qt.Assert(t, f, qt.IsNil)
	})

	t.Run("open empty path", func(t *testing.T) {
		f, err := dfs.Open("")
		qt.Assert(t, err, qt.IsNil)
		s, err := f.Stat()
		qt.Assert(t, err, qt.IsNil)
		qt.Assert(t, s.IsDir(), qt.IsTrue)
		qt.Assert(t, s.Name(), qt.Equals, "")
	})
}

func TestFSGlob(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)
	dfs := tmfs.DocFs(doc)

	matches, err := fs.Glob(dfs, "one/t*")
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, matches, qt.DeepEquals, []string{"one/three", "one/two"})
}

func TestFSReadFile(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)
	dfs := tmfs.DocFs(doc)

	data, err := fs.ReadFile(dfs, "one")
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, string(data), qt.Equals, "baz\n")
}

func ExampleWalkDir() {
	testdata, _ := filepath.Abs("../testdata")
	doc, _ := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	dfs := tmfs.DocFs(doc)

	counter := 0
	fs.WalkDir(dfs, "", func(path string, dir fs.DirEntry, err error) error {
		fmt.Printf("%d: %q\n", counter, path)
		counter++
		return nil
	})
	// Output:
	// 0: ""
	// 1: "cannot-describe-no-linebreak"
	// 2: "more-data"
	// 3: "this-is-the-data-name"
}

func ExampleRead() {
	testdata, _ := filepath.Abs("../testdata")
	doc, _ := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	dfs := tmfs.DocFs(doc)
	f, _ := dfs.Open("more-data")

	content, _ := io.ReadAll(f)
	fmt.Print(string(content))
	// Output:
	// func OtherMarkdownParsers() (shouldHighlight bool) {
	// 	return true
	// }
}

func ExampleReadDirFile() {
	testdata, _ := filepath.Abs("../testdata")
	doc, _ := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	dfs := tmfs.DocFs(doc)

	f, _ := dfs.Open("")
	fr := f.(fs.ReadDirFile)
	dirs, _ := fr.ReadDir(-1)
	for i, d := range dirs {
		fmt.Printf("%d: %q\n", i, d.Name())
	}
	// Output:
	// 0: "cannot-describe-no-linebreak"
	// 1: "more-data"
	// 2: "this-is-the-data-name"
}

// Generally true as of this writing
// A directory will return true on IsDir
// A file with data will have a non-zero size
func ExampleIsItAFileOrADirectory() {
	testdata, _ := filepath.Abs("../testdata")
	doc, _ := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	dfs := tmfs.DocFs(doc)
	{
		f, _ := dfs.Open("")
		stat, _ := f.Stat()
		fmt.Printf("the root dir is not a file: %q,%d,%t\n", stat.Name(), stat.Size(), stat.IsDir())
	}
	{
		f, _ := dfs.Open("one")
		stat, _ := f.Stat()
		fmt.Printf("this path is a dir AND a regular file: %q,%d,%t\n", stat.Name(), stat.Size(), stat.IsDir())
	}
	{
		f, _ := dfs.Open("one/four/bang")
		stat, _ := f.Stat()
		fmt.Printf("this path is a file but NOT a dir: %q,%d,%t\n", stat.Name(), stat.Size(), stat.IsDir())
	}
	// Output:
	// the root dir is not a file: "",0,true
	// this path is a dir AND a regular file: "one",4,true
	// this path is a file but NOT a dir: "bang",4,false
}

func ExampleConvertFileToDirEnt() {
	testdata, _ := filepath.Abs("../testdata")
	doc, _ := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	dfs := tmfs.DocFs(doc)

	f, _ := dfs.Open("more-data")
	stat, _ := f.Stat()
	ent := stat.Sys().(*testmark.DirEnt)
	fmt.Print(string(ent.Hunk.Body))
	// Output:
	// func OtherMarkdownParsers() (shouldHighlight bool) {
	// 	return true
	// }
}

// TestWalkDocument tests the implementation of fs.WalkDir against a Document
func TestFSWalkDocument(t *testing.T) {
	qt.Assert(t, fs.ValidPath("."), qt.IsTrue)
	testdata, err := filepath.Abs("../testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)
	dfs := tmfs.DocFs(doc)

	type expectT struct {
		path        string
		contents    string
		numChildren int
	}
	expected := []expectT{
		//path, isDir
		{"", "", 2},
		{"one", "baz\n", 3}, // both dir and file
		{"one/four", "", 1},
		{"one/four/bang", "mop\n", 0},
		{"one/three", "bar\n", 0},
		{"one/two", "foo\n", 0},
		{"really", "", 1},
		{"really/deep", "", 1},
		{"really/deep/dirs", "", 1},
		{"really/deep/dirs/wow", "zot\n", 0},
	}
	maxPathLen := 0
	for _, v := range expected {
		if maxPathLen < len(v.path) {
			maxPathLen = len(v.path)
		}
	}
	orderIdx := 0

	mark := func(b bool) string {
		if b {
			return "✔"
		}
		t.Fail()
		return "✖"
	}
	err = fs.WalkDir(dfs, "", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		order := expected[orderIdx]
		f, err := dfs.Open(path)
		qt.Assert(t, err, qt.IsNil)
		content, err := io.ReadAll(f)
		qt.Assert(t, err, qt.IsNil)

		fdir := f.(fs.ReadDirFile)
		children, err := fdir.ReadDir(-1)
		qt.Assert(t, err, qt.IsNil)

		t.Logf("%2d:  path: %s %-*q  children: %s%s %5t,%d  contents: %s %q",
			orderIdx,
			mark(order.path == path),
			maxPathLen+2, // path + quotes. Any path that exceeds the expected length will screw up formatting but that's fine.
			path,
			mark(order.numChildren == len(children)),
			mark(order.numChildren > 0 == dir.IsDir()),
			dir.IsDir(),
			len(children),
			mark(order.contents == string(content)),
			content,
		)
		orderIdx++
		return nil
	})
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, orderIdx, qt.Equals, len(expected))
}

// This is a weird edge of object filesystems where pseudo-dirs and files can overlap.
// We don't really have a great way of handling this. Just, some files are also directories. The end.
func TestFSOpenFileDir(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)
	dfs := tmfs.DocFs(doc)

	f, err := dfs.Open("one")
	qt.Assert(t, err, qt.IsNil)

	data, err := io.ReadAll(f)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, string(data), qt.Equals, "baz\n")

	rf := f.(fs.ReadDirFile)
	dirs, err := rf.ReadDir(-1)
	qt.Assert(t, err, qt.IsNil)
	dirNames := make([]string, 0, len(dirs))
	for _, d := range dirs {
		dirNames = append(dirNames, d.Name())
	}
	qt.Assert(t, dirNames, qt.DeepEquals, []string{"four", "three", "two"})

	stat, err := f.Stat()
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, stat.IsDir(), qt.IsTrue)
	qt.Assert(t, stat.Name(), qt.Equals, "one")
	qt.Assert(t, stat.Size(), qt.Equals, int64(len(data)))
}
