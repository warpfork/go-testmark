package testmark_test

import (
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/warpfork/go-testmark"
)

var _ fs.DirEntry = &testmark.File{}
var _ fs.File = &testmark.File{}

func TestFS(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	qt.Assert(t, err, qt.IsNil)

	t.Run("open dot path", func(t *testing.T) {
		f, err := doc.Open(".")
		pathErr := new(fs.PathError)
		if qt.Check(t, err, qt.ErrorAs, &pathErr) {
			qt.Assert(t, pathErr.Op, qt.Equals, "open")
			qt.Assert(t, pathErr.Path, qt.Equals, ".")
		}
		qt.Assert(t, err, qt.ErrorIs, fs.ErrNotExist)
		qt.Assert(t, f, qt.IsNil)
	})

	t.Run("open empty path", func(t *testing.T) {
		f, err := doc.Open("")
		qt.Assert(t, err, qt.IsNil)
		s, err := f.Stat()
		qt.Assert(t, err, qt.IsNil)
		qt.Assert(t, s.IsDir(), qt.IsTrue)
		qt.Assert(t, s.Name(), qt.Equals, "")
	})
}

func TestWalk(t *testing.T) {
	qt.Assert(t, fs.ValidPath("."), qt.IsTrue)
	testdata, err := filepath.Abs("testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)

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
	err = fs.WalkDir(doc, "", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		order := expected[orderIdx]
		f, err := doc.Open(path)
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
func TestOpenFileDir(t *testing.T) {
	testdata, err := filepath.Abs("testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)

	f, err := doc.Open("one")
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
