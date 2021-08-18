package testmark

import (
	"strings"
)

// BuildDirIndex mutates the Document to set the DirEnt field.
//
// The order of ChildrenList in the DirEnt
// is determined by the order in which things are first seen in the Document's Hunk list.
// The "directories" can be implied,
// e.g. a Hunk with name="foo/bar" will cause the creation of a DirEnt with name "foo".
//
// No concept of path "cleaning" is applied.  Paths like "." and ".." are not treated specially.
// A path containing repeated slashes is a fairly deranged thing to do, but also won't be rejected.
func (doc *Document) BuildDirIndex() {
	doc.DirEnt = &DirEnt{}
	for _, hunk := range doc.DataHunks {
		doc.DirEnt.fill(strings.Split(hunk.Name, "/"), hunk.Hunk)
	}
}

func (dirent *DirEnt) fill(pathSegs []string, hunk Hunk) {
	if len(pathSegs) == 0 {
		dirent.Hunk = &hunk
		return
	}
	if dirent.Children == nil {
		dirent.Children = make(map[string]*DirEnt)
	}
	if next, exists := dirent.Children[pathSegs[0]]; exists {
		next.fill(pathSegs[1:], hunk)
	} else {
		l := len(dirent.ChildrenList)
		dirent.ChildrenList = append(dirent.ChildrenList, DirEnt{
			Name: pathSegs[0],
		})
		dirent.Children[pathSegs[0]] = &dirent.ChildrenList[l]
		dirent.ChildrenList[l].fill(pathSegs[1:], hunk)
	}
}
