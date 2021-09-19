package testmark

import (
	"io"
	"os"
	"strings"
)

func (d Document) String() string {
	var sb strings.Builder
	Write(&d, &sb)
	return sb.String()
}

func Write(doc *Document, wr io.Writer) (int, error) {
	n := 0
	for _, line := range doc.Lines {
		if n2, err := wr.Write(line); err != nil {
			return n + n2, err
		} else {
			n += n2
		}
		if n2, err := wr.Write(sigilLineBreak); err != nil {
			return n + n2, err
		} else {
			n += n2
		}
	}
	return n, nil
}

func WriteFile(doc *Document, filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = Write(doc, f)
	return err
}

func WriteWithPatches(doc *Document, wr io.Writer, patches ...Hunk) (int, error) {
	if len(patches) == 0 {
		return 0, nil
	}
	doc = Patch(doc, patches...)
	return Write(doc, wr)
}

func WriteFileWithPatches(doc *Document, filename string, patches ...Hunk) error {
	if len(patches) == 0 {
		return nil
	}
	doc = Patch(doc, patches...)
	return WriteFile(doc, filename)
}

func (pa PatchAccumulator) WriteWithPatches(doc *Document, wr io.Writer) (int, error) {
	return WriteWithPatches(doc, wr, pa.Patches...)
}

func (pa PatchAccumulator) WriteFileWithPatches(doc *Document, filename string) error {
	return WriteFileWithPatches(doc, filename, pa.Patches...)
}
