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

	t.Run("open empty path", func(t *testing.T) {
		f, err := doc.Open("")
		pathErr := new(fs.PathError)
		if qt.Check(t, err, qt.ErrorAs, &pathErr) {
			qt.Assert(t, pathErr.Op, qt.Equals, "open")
			qt.Assert(t, pathErr.Path, qt.Equals, "")
		}
		qt.Assert(t, err, qt.ErrorIs, fs.ErrInvalid)
		qt.Assert(t, f, qt.IsNil)
	})

	t.Run("open dot path", func(t *testing.T) {
		f, err := doc.Open(".")
		qt.Assert(t, err, qt.IsNil)
		s, err := f.Stat()
		qt.Assert(t, err, qt.IsNil)
		qt.Assert(t, s.IsDir(), qt.IsTrue)
	})
}

func TestWalk(t *testing.T) {
	qt.Assert(t, fs.ValidPath("."), qt.IsTrue)
	testdata, err := filepath.Abs("testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)

	type expectT struct {
		path string
		dir  bool
	}
	expected := []expectT{
		{".", true},
		{"one", true}, // one is both dir and file
		{"one/four", true},
		{"one/four/bang", false},
		{"one/three", false},
		{"one/two", false},
		{"really", true},
		{"really/deep", true},
		{"really/deep/dirs", true},
		{"really/deep/dirs/wow", false},
	}
	orderIdx := 0

	mark := func(b bool) string {
		if b {
			return "✔"
		}
		t.Fail()
		return "✖"
	}
	err = fs.WalkDir(doc, ".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		order := expected[orderIdx]
		t.Logf("%d %s-%s %q: %t", orderIdx, mark(order.path == path), mark(dir.IsDir() == order.dir), path, dir.IsDir())
		orderIdx++
		return nil
	})
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, orderIdx, qt.Equals, len(expected))
}

// This is a weird edge of object filesystems where pseudo-dirs and files can overlap.
func TestOpenFileDir(t *testing.T) {
	qt.Assert(t, fs.ValidPath("."), qt.IsTrue)
	testdata, err := filepath.Abs("testdata")
	qt.Assert(t, err, qt.IsNil)

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	qt.Assert(t, err, qt.IsNil)

	f, err := doc.Open("one")
	qt.Assert(t, err, qt.IsNil)

	data, err := io.ReadAll(f)
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, string(data), qt.Equals, "baz\n")

	stat, err := f.Stat()
	qt.Assert(t, err, qt.IsNil)
	qt.Assert(t, stat.IsDir(), qt.IsTrue)
	qt.Assert(t, stat.Name(), qt.Equals, "one")
	qt.Assert(t, stat.Size(), qt.Equals, int64(len(data)))

	rf := f.(fs.ReadDirFile)
	dirs, err := rf.ReadDir(-1)
	qt.Assert(t, err, qt.IsNil)
	dirNames := make([]string, 0, len(dirs))
	for _, d := range dirs {
		dirNames = append(dirNames, d.Name())
	}
	qt.Assert(t, dirNames, qt.DeepEquals, []string{"four", "three", "two"})
}
