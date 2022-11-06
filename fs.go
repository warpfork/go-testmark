package testmark

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

const defaultFileMode = 0777

// File implements both fs.File and fs.DirEntry
type File struct {
	*bytes.Buffer
	stat           fileStat
	childrenIdx    int // track index for readdir
	childrenSorted []string
	children       map[string]*DirEnt
}

// Name returns the name of the file (or subdirectory) described by the entry.
// This name is only the final element of the path (the base name), not the entire path.
// For example, Name would return "hello.go" not "home/gopher/hello.go".
func (f *File) Name() string {
	return f.stat.name
}

// Info is equivalent to Stat
func (f *File) Info() (fs.FileInfo, error) {
	return f.stat, nil
}

// Stat returns a fs.FileInfo for the file
// If the file is both a file AND a directory then the length will be the length of the file body
// If the fil
func (f *File) Stat() (fs.FileInfo, error) {
	return f.stat, nil
}

// IsDir reports whether the entry describes a directory.
func (f *File) IsDir() bool {
	return f.stat.IsDir()
}

// Type returns the type bits for the entry.
// The type bits are a subset of the usual FileMode bits, those returned by the FileMode.Type method.
func (f *File) Type() fs.FileMode {
	return f.stat.mode.Type()
}

// Read reads up to len(b) bytes from the File and stores them in b. It returns
// the number of bytes read and any error encountered. At end of file, Read
// returns 0, io.EOF.
func (f *File) Read(b []byte) (int, error) {
	if f.Buffer == nil {
		return 0, fs.ErrClosed
	}
	return f.Buffer.Read(b)
}

//ReadDir will return a []*File as an []fs.DirEntry
// Returned entries will be sorted by filename as required by fs.ReadDir
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if len(f.children) == 0 {
		return []fs.DirEntry{}, nil
	}
	start := f.childrenIdx
	end := f.childrenIdx + n
	if end >= len(f.children) || n < 0 {
		end = len(f.children)
	}
	defer func() { f.childrenIdx = end }()
	fmt.Printf("%v[%d:%d]\n", f.childrenSorted, start, end)
	names := f.childrenSorted[start:end]
	return f.readDir(names...)
}

// readDir takes the list of children names and returns an []fs.DirEntry in the same order.
// Will return an fs.PathError if a child name does not exist.
func (f *File) readDir(children ...string) ([]fs.DirEntry, error) {
	result := make([]fs.DirEntry, 0, len(children))
	for _, name := range children {
		child, exists := f.children[name]
		if !exists {
			return []fs.DirEntry{}, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
		}
		result = append(result, child.file())
	}
	return result, nil
}

// Close will set the underlying buffer to nil
// If the buffer is already nil then Close will return an fs.ErrClosed
func (f *File) Close() error {
	if f.Buffer == nil {
		return fs.ErrClosed
	}
	f.Buffer = nil
	return nil
}

// There's basically nothing meaningful in the fileStat structure
type fileStat struct {
	name string
	size int64
	mode fs.FileMode
	sys  interface{}
}

//IsDir will be true if a DirEnt has children
func (s fileStat) IsDir() bool {
	return s.mode.IsDir()
}

// ModTime is meaningless. Testmark hunks don't have a modtime
func (fileStat) ModTime() time.Time {
	return time.Time{}
}

// Mode is meaningless for testmark "files"
func (s fileStat) Mode() fs.FileMode {
	return s.mode
}

// base name of the file
func (s fileStat) Name() string {
	return s.name
}

func (s fileStat) Size() int64 {
	return s.size
}

func (s fileStat) Sys() interface{} {
	return s.sys
}

// Open opens the named hunk or a directory path.
// Open does not follow relative paths such as .. or .
//
// When Open returns an error, it should be of type *PathError
// with the Op field set to "open", the Path field set to name,
// and the Err field describing the problem.
//
// Open should reject attempts to open names that do not satisfy
// ValidPath(name), returning a *PathError with Err set to
// ErrInvalid or ErrNotExist.
func (doc *Document) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	// if hunk, exists := doc.HunksByName[name]; exists {
	// 	return hunk.file(), nil
	// }

	if strings.HasSuffix(name, "/") {
		name = strings.TrimRight(name, "/")
	}
	if doc.DirEnt == nil {
		err := doc.BuildDirIndex()
		if err != nil {
			return nil, err
		}
	}
	if name == "." {
		return doc.DirEnt.file(), nil
	}
	dir := findDir(doc.DirEnt, splitpath(name)...)
	if dir != nil {
		// Create a directory type file
		return dir.file(), nil
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

func (h *DocHunk) file() *File {
	buf := bytes.NewBuffer(h.Body)
	return &File{
		Buffer:         buf,
		stat:           h.stat(),
		children:       make(map[string]*DirEnt),
		childrenSorted: make([]string, 0),
	}
}

func (h *DocHunk) stat() fileStat {
	basename := path.Base(h.Name)
	return fileStat{
		name: basename,
		mode: defaultFileMode,
		sys:  h,
		size: int64(len(h.Body)),
	}
}

func (d *DirEnt) stat() fileStat {
	mode := fs.FileMode(defaultFileMode)
	length := len(d.Children)
	if len(d.Children) > 0 {
		mode = mode | fs.ModeDir
	}
	if d.Hunk != nil {
		length = len(d.Hunk.Body)
	}
	return fileStat{
		name: d.Name,
		mode: mode,
		sys:  d,
		size: int64(length),
	}
}

func (d *DirEnt) file() *File {
	buf := bytes.NewBuffer([]byte{})
	if d.Hunk != nil {
		//writing to the buffer won't modify hunk body
		buf = bytes.NewBuffer(d.Hunk.Body)
	}
	childrenSorted := d.childrenNames()
	sort.Strings(childrenSorted)
	return &File{
		Buffer:         buf,
		stat:           d.stat(),
		childrenSorted: childrenSorted,
		children:       d.Children,
	}
}
func (d *DirEnt) childrenNames() []string {
	result := make([]string, 0, len(d.Children))
	for name := range d.Children {
		result = append(result, name)
	}
	return result
}

func (d *DirEnt) IsFile() bool {
	return d.Hunk != nil
}

func (d *DirEnt) IsDir() bool {
	return len(d.Children) > 0
}

func findDir(dir *DirEnt, pathsplits ...string) *DirEnt {
	if len(pathsplits) == 0 {
		return dir
	}
	if child, exists := dir.Children[pathsplits[0]]; exists {
		return findDir(child, pathsplits[1:]...)
	}
	return nil
}
