package fs_test

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

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
	if err != nil {
		panic(err)
	}

	doc, err := testmark.ReadFile(filepath.Join(testdata, "example.md"))
	if err != nil {
		panic(err)
	}
	dfs := tmfs.DocFs(doc)

	t.Run("open dot path", func(t *testing.T) {
		f, err := dfs.Open(".")
		pathErr := new(fs.PathError)
		if !errors.As(err, &pathErr) {
			t.Errorf("expected open to return an fs.PathError: %v", err)
		}
		assertString(t, pathErr.Op, "open")
		assertString(t, pathErr.Path, ".")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected an fs.ErrNotExist: %v", err)
		}
		if f != nil {
			t.Errorf("expected opened file to be nil")
		}
	})

	t.Run("open empty path", func(t *testing.T) {
		f, err := dfs.Open("")
		if err != nil {
			panic(err)
		}
		s, err := f.Stat()
		if err != nil {
			panic(err)
		}
		if !s.IsDir() {
			t.Errorf("expected file to be a directory")
		}
		assertString(t, s.Name(), "")
	})
}

func TestFSGlob(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	if err != nil {
		panic(err)
	}

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	if err != nil {
		panic(err)
	}
	dfs := tmfs.DocFs(doc)

	matches, err := fs.Glob(dfs, "one/t*")
	if err != nil {
		panic(err)
	}
	assertStrings(t, matches, "one/three", "one/two")
}

func TestFSReadFile(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	if err != nil {
		panic(err)
	}

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	if err != nil {
		panic(err)
	}
	dfs := tmfs.DocFs(doc)

	data, err := fs.ReadFile(dfs, "one")
	if err != nil {
		panic(err)
	}

	assertString(t, string(data), "baz\n")
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
// A directory that is not a file will have a size equal to the number of children
// A path that is a file will have a size equal to it's buffer length regardless of directory status
func ExampleIsItAFileOrADirectory() {
	testdata, _ := filepath.Abs("../testdata")
	doc, _ := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	dfs := tmfs.DocFs(doc)
	{
		f, _ := dfs.Open("")
		stat, _ := f.Stat()
		fmt.Printf("the root dir with %d children is not a file: %q,%d,%t\n", stat.Size(), stat.Name(), stat.Size(), stat.IsDir())
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
	// the root dir with 2 children is not a file: "",2,true
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
	testdata, err := filepath.Abs("../testdata")
	if err != nil {
		panic(err)
	}

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	if err != nil {
		panic(err)
	}
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
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
		content, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}

		fdir := f.(fs.ReadDirFile)
		children, err := fdir.ReadDir(-1)
		if err != nil {
			panic(err)
		}

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
	if err != nil {
		panic(err)
	}
	if orderIdx != len(expected) {
		t.Errorf("didn't iterate through all entries")
	}
}

// This is a weird edge of object filesystems where pseudo-dirs and files can overlap.
// We don't really have a great way of handling this. Just, some files are also directories. The end.
func TestFSOpenFileDir(t *testing.T) {
	testdata, err := filepath.Abs("../testdata")
	if err != nil {
		panic(err)
	}

	doc, err := testmark.ReadFile(filepath.Join(testdata, "exampleWithDirs.md"))
	if err != nil {
		panic(err)
	}
	dfs := tmfs.DocFs(doc)

	f, err := dfs.Open("one")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	assertString(t, string(data), "baz\n")

	rf := f.(fs.ReadDirFile)
	dirs, err := rf.ReadDir(-1)
	if err != nil {
		panic(err)
	}
	dirNames := make([]string, 0, len(dirs))
	for _, d := range dirs {
		dirNames = append(dirNames, d.Name())
	}

	assertStrings(t, dirNames, "four", "three", "two")

	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}
	if !stat.IsDir() {
		t.Errorf("file should be a directory")
	}
	if stat.Name() != "one" {
		t.Errorf("file name expected to be %q but got %q", "one", stat.Name())
	}
	if stat.Size() != int64(len(data)) {
		t.Errorf("file size of %d expected to be equal to data length of %d", stat.Size(), len(data))
	}
}

func assertStrings(t testing.TB, actual []string, expected ...string) {
	t.Helper()
	sharedLen := len(actual)
	if sharedLen > len(expected) {
		sharedLen = len(expected)
	}
	for i := 0; i < sharedLen; i++ {
		if actual[i] != expected[i] {
			t.Errorf("Expected actual[%d] to be %q but got %q", i, expected[i], actual[i])
		}
	}
	if len(actual) != len(expected) {
		t.Errorf("Expected len(actual)=%d to equal len(expected)=%d", len(actual), len(expected))
	}
}

func assertString(t testing.TB, actual string, expected string) {
	t.Helper()
	if actual != expected {
		t.Errorf("expected %q to equal %q", actual, expected)
	}
}
