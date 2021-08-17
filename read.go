package testmark

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"unicode"
)

func Read(r io.Reader) (*Document, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func ReadFile(name string) (*Document, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

var (
	sigilLineBreak = []byte{'\n'}
	sigilCodeBlock = []byte("```")
	sigilTestmark  = []byte("[testmark]:# ")
)

func Parse(data []byte) (*Document, error) {
	doc := Document{Original: data}

	// Markdown can be effectively parsed line by line.
	doc.OriginalLines = bytes.Split(data, sigilLineBreak)

	// The first layer of our parse is "is this in a code block".
	// Code blocks are the only feature of markdown that meaningfully changes what mode you're in at the start of a line.
	// After that, we look for our magic prefix (on any lines that *aren't* already in a code block).
	// Then, the rest is... pretty straightforward.
	var inCodeBlock bool
	var expectCodeBlock bool
	hunkInProgress := Hunk{Line: -1}
	for i, line := range doc.OriginalLines {
		// Check for transition in or out of codeblock.
		if bytes.HasPrefix(line, sigilCodeBlock) {
			switch inCodeBlock {
			case false: // starting a block
				if expectCodeBlock {
					hunkInProgress.BlockTag = string(line[len(sigilCodeBlock):])
				}
				expectCodeBlock = false
			case true: // ending a block
				if hunkInProgress.Line > -1 {
					hunkInProgress.EndLine = i
					doc.DataHunks = append(doc.DataHunks, hunkInProgress)
					hunkInProgress = Hunk{Line: -1}
				}
			}
			inCodeBlock = !inCodeBlock
			continue
		}
		// If we're in a code block, just fly by.
		if inCodeBlock {
			continue
		}
		// If we were expecting a code block just now, we didn't get it.
		if expectCodeBlock {
			// ... Log a complaint?  I don't think halting with a parse error would be helpful.
			// But definitely don't wait around for arbitrarily distant code blocks.
			expectCodeBlock = false
		}
		// Look for testmark block indicators.
		if bytes.HasPrefix(line, sigilTestmark) {
			remainder := line[len(sigilTestmark):]
			nameEnd := bytes.IndexFunc(remainder, unicode.IsSpace)
			if nameEnd < 0 {
				nameEnd = len(remainder)
			}
			name := remainder[0:nameEnd]
			if len(name) == 0 {
				// ... Log a complaint?  Empty block names are very silly.
			}
			expectCodeBlock = true
			hunkInProgress.Line = i
			hunkInProgress.Name = string(name)
		}
		// Any other text?  It's prose.  No action.
	}
	return &doc, nil
}
