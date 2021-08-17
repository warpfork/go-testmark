package testmark

import (
	"io"
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
