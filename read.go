package testmark

import (
	"bytes"
	"fmt"
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
	doc := Document{
		Original:    data,
		HunksByName: make(map[string]DocHunk),
	}

	// Markdown can be effectively parsed line by line.
	doc.Lines = bytes.Split(data, sigilLineBreak)

	// The first layer of our parse is "is this in a code block".
	// Code blocks are the only feature of markdown that meaningfully changes what mode you're in at the start of a line.
	// After that, we look for our magic prefix (on any lines that *aren't* already in a code block).
	// Then, the rest is... pretty straightforward.
	var offset int
	var inCodeBlock bool
	var expectCodeBlock bool
	var codeBlockOffset int
	hunkInProgress := DocHunk{LineStart: -1}
	for i, line := range doc.Lines {
		// Check for transition in or out of codeblock.
		if bytes.HasPrefix(line, sigilCodeBlock) {
			switch inCodeBlock {
			case false: // starting a block
				if expectCodeBlock {
					hunkInProgress.BlockTag = string(line[len(sigilCodeBlock):])
					codeBlockOffset = offset + len(line) + 1
				}
				expectCodeBlock = false
			case true: // ending a block
				if hunkInProgress.LineStart > -1 {
					hunkInProgress.LineEnd = i
					hunkInProgress.Body = doc.Original[codeBlockOffset:offset]
					doc.DataHunks = append(doc.DataHunks, hunkInProgress)
					doc.HunksByName[hunkInProgress.Name] = hunkInProgress
					hunkInProgress = DocHunk{LineStart: -1}
				}
			}
			inCodeBlock = !inCodeBlock
			goto next
		}
		// If we're in a code block, just fly by.
		if inCodeBlock {
			goto next
		}
		// If we were expecting a code block just now, we didn't get it.
		if expectCodeBlock {
			// ... Log a complaint?  I don't think halting with a parse error would be helpful.
			// But definitely don't wait around for arbitrarily distant code blocks.
			expectCodeBlock = false
		}
		// Look for testmark block indicators.
		if bytes.HasPrefix(line, sigilTestmark) {
			// If this line, after the sigil prefix, doesn't begin with "(" and end with ")", it's not a well-formed markdown comment, and you should probably be told about that.
			remainder := line[len(sigilTestmark):]
			if len(remainder) < 2 || remainder[0] != '(' || remainder[len(remainder)-1] != ')' {
				return &doc, fmt.Errorf("invalid markdown comment on line %d (should look like %q, mind the parens)", i+1, "[testmark]:# (data-name-here)")
			}
			remainder = remainder[1 : len(remainder)-1]

			// Parse the name.  If there's whitespace inside here, we'll just quietly stop at that.  Maybe we'll do extensions with that info space in the future.
			nameEnd := bytes.IndexFunc(remainder, unicode.IsSpace)
			if nameEnd < 0 {
				nameEnd = len(remainder)
			}
			name := string(remainder[0:nameEnd])
			if len(name) == 0 {
				return &doc, fmt.Errorf("invalid markdown comment on line %d, hunk name is empty", i+1)
			}

			// Error if the hunk name is repeated.
			if already, exists := doc.HunksByName[name]; exists {
				// You can actually ignore this error, and things will even still mostly work.  HunksByName will only look up the first occurence, and Patch will change only the first occurence, and that is weird, but perhaps fine.
				return &doc, fmt.Errorf("repeated testmark hunk name %q, first seen on line %d, and again on line %d", name, already.LineStart+1, i+1)
			}

			// Okay: hunk started (probably), name is parsed out, we expect a code block to start on the next line, cool.
			expectCodeBlock = true
			hunkInProgress.LineStart = i
			hunkInProgress.Name = name
		}
		// Any other text?  It's prose.  No action.
	next:
		// Track total offset, so we can use it to subslice out document hunks.
		// Mind: this is going to be off by one at the very end of the file... but that turns out never to matter to us.
		offset += len(line) + 1
	}
	return &doc, nil
}
