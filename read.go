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
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Read(f)
}

var (
	sigilLineBreak      = []byte{'\n'}
	sigilCarriageReturn = []byte{'\r'}
	sigilCrLf           = []byte{'\r', '\n'}
	sigilCodeBlock      = []byte("```")
	sigilTestmark       = []byte("[testmark]:# ")
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
	for i, origLine := range doc.Lines {
		// Support CRLF line endings, for Windows.
		line := bytes.TrimSuffix(origLine, sigilCarriageReturn)

		// Check for transition in or out of codeblock.
		if bytes.HasPrefix(line, sigilCodeBlock) {
			switch inCodeBlock {
			case false: // starting a block
				if expectCodeBlock {
					hunkInProgress.InfoString = string(line[len(sigilCodeBlock):])
					codeBlockOffset = offset + len(origLine) + 1
				}
				expectCodeBlock = false
			case true: // ending a block
				if hunkInProgress.LineStart > -1 {
					hunkInProgress.LineEnd = i
					hunkInProgress.Body = normalizeEndings(doc.Original[codeBlockOffset:offset])
					doc.DataHunks = append(doc.DataHunks, hunkInProgress)
					doc.HunksByName[hunkInProgress.Name] = hunkInProgress
					hunkInProgress = DocHunk{LineStart: -1}
				}
			}
			inCodeBlock = !inCodeBlock
			goto next
		}
		if inCodeBlock {
			// If we're in a code block, just fly by.
			goto next
		}
		if expectCodeBlock {
			// If we were expecting a code block just now, we didn't get it.
			return &doc, fmt.Errorf("invalid markdown comment on line %d. Missing code block for hunk %s", i+1, hunkInProgress.Name)
		}
		// Look for testmark block indicators.
		if bytes.HasPrefix(line, sigilTestmark) {
			// If this line, after the sigil prefix, doesn't begin with "(" and end with ")", it's not a well-formed markdown comment, and you should probably be told about that.
			remainder := line[len(sigilTestmark):]
			if len(remainder) < 2 || remainder[0] != '(' || remainder[len(remainder)-1] != ')' {
				if unicode.IsSpace(rune(remainder[len(remainder)-1])) {
					return &doc, fmt.Errorf("invalid markdown comment on line %d (should look like %q; remove trailing whitespace)", i+1, "[testmark]:# (data-name-here)")
				}
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
		offset += len(origLine) + 1
	}
	return &doc, nil
}

// normalizeEndings looks for instances of "\r\n" and flattens them to "\n".
// If it finds no instances of "\r\n", the original byte slice is returned unchanged.
//
// This function does not bring joy; however,
// see https://github.com/warpfork/go-testmark/pull/4#issuecomment-922760414
// and see https://github.com/warpfork/go-testmark/pull/4#issuecomment-922782549
// for discussion.  Performing this kind of normalization to data hunk boundaries
// seems to be a "least bad" behavior in a practical sense.
func normalizeEndings(in []byte) []byte {
	if bytes.Count(in, sigilCrLf) == 0 {
		return in
	}
	return bytes.Replace(in, sigilCrLf, sigilLineBreak, -1)
}
