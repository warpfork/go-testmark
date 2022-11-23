package testmark

import (
	"bytes"
	"io/fs"
	"sort"
	"time"
)

// We don't implement writing to hunks through this interface so everything is read-only
const defaultFileMode fs.FileMode = 0444

// MultiDoc allows combining multiple documents into a single entity for ease of use.
// This can be used to spread hunks across multiple real files or other sources.
type MultiDoc struct {
	// DocByHunkName allows you to get back to the original document for a hunk.
	// This would be necessary for the patching features and any other time you need
	// features of the original document
	DocByHunkName map[string]*Document
	// Document that stores the combined set of hunks
	// Original and Lines will not be relevant in this case
	Document
}

// Add will combine the given Document with this MultiDoc.
// If a duplicate hunk is found, an fs.PathError will be returned.
// Add will check for duplicates before mutating the MultiDoc; however, this is NOT threadsafe.
// Add will not mutate the given Document.
func (d *MultiDoc) Add(doc *Document) error {
	if d.DocByHunkName == nil {
		d.DocByHunkName = make(map[string]*Document)
	}
	if d.Document.DataHunks == nil {
		d.Document.HunksByName = make(map[string]DocHunk)
	}
	for _, hunk := range doc.DataHunks {
		// Guard against mutating a DocFS that already has hunks of a particular name
		// TODO: add merge strategies to allow overwrite
		if _, exists := d.HunksByName[hunk.Name]; exists {
			return &fs.PathError{Op: "add", Path: hunk.Name, Err: fs.ErrExist}
		}
	}
	for _, hunk := range doc.DataHunks {
		d.DocByHunkName[hunk.Name] = doc
		d.HunksByName[hunk.Name] = hunk
		d.DataHunks = append(d.DataHunks, hunk)
	}
	return nil
}

// File is a representation of DirEnt used when working with Documents as an fs.FS
// Generally this is done by calling Document.Open or using a function from the fs package.
type File struct {
	// buffer contains the hunk data after opening
	// a nil buffer implies that the file is closed
	buffer *bytes.Buffer
	stat   fileStat
	// childIdx tracks index of children for readdir
	childIdx int
	// children sorted by name
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
	if f.buffer == nil {
		return 0, fs.ErrClosed
	}
	return f.buffer.Read(b)
}

// ReadDir will return a []*File as an []fs.DirEntry
// Returned entries will be sorted by filename as required by fs.ReadDir
// If n < 0, ReadDir will return all remaining entries.
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if len(f.children) == 0 {
		return []fs.DirEntry{}, nil
	}
	start := f.childIdx
	end := f.childIdx + n
	if end >= len(f.children) || n < 0 {
		end = len(f.children)
	}
	defer func() { f.childIdx = end }()
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
	if f.buffer == nil {
		return fs.ErrClosed
	}
	f.buffer = nil
	return nil
}

// fileStat stores data as an implementation of fs.FileInfo
type fileStat struct {
	name string
	size int64
	mode fs.FileMode
	sys  interface{}
}

// IsDir will be true if a DirEnt has children
func (s fileStat) IsDir() bool {
	return s.mode.IsDir()
}

// ModTime is meaningless. Testmark hunks don't have a modtime
// ModTime will always return a time.Time zero value
func (fileStat) ModTime() time.Time {
	return time.Time{}
}

// Mode returns the FileMode for the file.
// The filemode will have the fs.ModeDir bit enabled if the File contains references to children
func (s fileStat) Mode() fs.FileMode {
	return s.mode
}

// Name returns the base name of the file
func (s fileStat) Name() string {
	return s.name
}

// Size returns the size of the underlying hunk when the file was opened
func (s fileStat) Size() int64 {
	return s.size
}

// Sys returns the _thing_ from whence this file came.
// Generally that's the *DirEnt
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
// Open does NOT follow conventions for fs.ValidPath.
// Opening an empty path will return the root directory for the document.
// This is different than the fs.ValidPath special case of using "." as the root path.
// The testmark document treats "." and ".." the same as any other character.
func (doc *Document) Open(name string) (fs.File, error) {
	if doc.DirEnt == nil {
		doc.BuildDirIndex()
	}
	if name == "" {
		return doc.DirEnt.file(), nil
	}
	ent := findDir(doc.DirEnt, splitpath(name)...)
	if ent == nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return ent.file(), nil
}

// file returns the File representation of a DirEnt.
func (d *DirEnt) file() *File {
	size := int64(0)
	buf := bytes.NewBuffer([]byte{})
	if d.Hunk != nil {
		buf = bytes.NewBuffer(d.Hunk.Body)
		size = int64(len(d.Hunk.Body))
	}
	childrenSorted := make([]string, 0, len(d.Children))
	for name := range d.Children {
		childrenSorted = append(childrenSorted, name)
	}
	sort.Strings(childrenSorted)
	mode := defaultFileMode
	if len(d.Children) > 0 {
		mode = mode | fs.ModeDir
	}
	return &File{
		buffer:         buf,
		childrenSorted: childrenSorted,
		children:       d.Children,
		stat: fileStat{
			name: d.Name,
			mode: mode,
			sys:  d,
			// size is generally the number of directory entires or the length of the file in bytes.
			// We have to choose one or the other because files and directories can overlap
			// In my opinion it's best to go with file length, so directories will always have a size of zero.
			size: size,
		},
	}
}

// findDir will traverse down a DirEnt by each path split and return a DirEnt for that path if it exists.
// Otherwise, findDir will return nil.
func findDir(dir *DirEnt, pathsplits ...string) *DirEnt {
	if len(pathsplits) == 0 {
		return dir
	}
	if child, exists := dir.Children[pathsplits[0]]; exists {
		return findDir(child, pathsplits[1:]...)
	}
	return nil
}
