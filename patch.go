package testmark

import (
	"bytes"
	"unicode"
)

func Patch(oldDoc *Document, hunks ...Hunk) (newDoc *Document) {
	// First pool up the hunk names we've been asked to patch.
	// We want to go over things in the order already present in the document,
	// so the order our varargs came in is not relevant nor helpful.
	//
	// Also, validate them real quick.
	// Empty names and names with whitespace are unacceptable.
	newHunks := make(map[string]Hunk, len(hunks))
	for _, hunk := range hunks {
		if hunk.Name == "" || bytes.IndexFunc([]byte(hunk.Name), unicode.IsSpace) >= 0 {
			panic("hunk name must not be empty and cannot contain whitespace")
		}
		newHunks[hunk.Name] = hunk
	}

	// Mutation is bad.
	// Immediately start making a new document.
	// Prep it with about the same amount of memory as the old one.
	newDoc = &Document{
		Lines:       make([][]byte, 0, len(oldDoc.Lines)),
		DataHunks:   make([]DocHunk, 0, len(oldDoc.DataHunks)),
		HunksByName: make(map[string]DocHunk, len(oldDoc.HunksByName)),
	}

	// Range over the document and apply patches.
	// We'll build up a whole new document as we go (byte slices and all!).
	var leftOff int
	for _, hunk := range oldDoc.DataHunks {
		// Copy any prose lines from wherever we left off, up to the start of the new hunk.
		// And advance the marker for leftOff marker to past the end of the old hunk.
		newDoc.Lines = append(newDoc.Lines, oldDoc.Lines[leftOff:hunk.LineStart]...)
		leftOff = hunk.LineEnd + 1

		var newBodyLines [][]byte
		if newHunk, exists := newHunks[hunk.Name]; exists {
			// Split our new hunk's body into lines, ready to append to the total content lines.
			// The rest... copy it into 'hunk', actually, it's a local variable and it makes the code slightly more DRY.
			newBodyLines = bytes.Split(newHunk.Body, sigilLineBreak)
			// If the last byte was a linebreak, the split will tend to exaggerate it a bit, so let's trim that back down.
			if len(newHunk.Body) > 0 && newHunk.Body[len(newHunk.Body)-1] == '\n' {
				newBodyLines = newBodyLines[0 : len(newBodyLines)-1]
			}
			hunk.BlockTag = newHunk.BlockTag

			// Yeet from newHunks, as it's now handled.
			delete(newHunks, hunk.Name)
		} else {
			// Just... keep the old lines, which we can sub-slice back out of the old document.
			newBodyLines = oldDoc.Lines[hunk.LineStart+2 : hunk.LineEnd]
		}

		// Append the hunk framing, and the body lines.
		// Watch how this changes the offsets, so we can build a new DocHunk with info that's correct.
		// (If you're just going to serialize this, it wouldn't matter, but if you want to patch multiple times, it matters.)
		newLineStart := len(newDoc.Lines)
		newDoc.Lines = appendHunkLines(newDoc.Lines, hunk.Name, hunk.BlockTag, newBodyLines)
		newLineEnd := len(newDoc.Lines)
		docHunk := DocHunk{
			LineStart: newLineStart,
			LineEnd:   newLineEnd,
			Hunk:      hunk.Hunk,
		}
		// Append the updated hunk info to newDoc.
		newDoc.DataHunks = append(newDoc.DataHunks, docHunk)
		newDoc.HunksByName[hunk.Name] = docHunk
	}

	// Copy any remaining trailing prose lines.
	newDoc.Lines = append(newDoc.Lines, oldDoc.Lines[leftOff:]...)

	// Now for any hunks we have left... We'll just stick them on the end, I guess.
	// And *now* the dang order of our original args matters.  We wouldn't want this to be randomized.
	for _, hunk := range hunks {
		// If it was already done, skip it.
		if _, stillTodo := newHunks[hunk.Name]; !stillTodo {
			continue
		}
		// If we're about to need to append something, make sure there's at least one blank line first.
		if len(newDoc.Lines[len(newDoc.Lines)-1]) > 0 {
			newDoc.Lines = append(newDoc.Lines, []byte{})
		}
		// Append it.
		newDoc.Lines = appendHunkLines(newDoc.Lines, hunk.Name, hunk.BlockTag, bytes.Split(hunk.Body, sigilLineBreak))
		// And one more trailing line, at the end.
		newDoc.Lines = append(newDoc.Lines, []byte{})
	}

	return
}

func appendHunkLines(lines [][]byte, hunkName string, hunkBlockTag string, hunkBodyLines [][]byte) [][]byte {
	lines = append(lines, bytes.Join([][]byte{sigilTestmark, []byte{'('}, []byte(hunkName), []byte{')'}}, nil))
	lines = append(lines, bytes.Join([][]byte{sigilCodeBlock, []byte(hunkBlockTag)}, nil))
	lines = append(lines, hunkBodyLines...)
	lines = append(lines, sigilCodeBlock)
	return lines
}

type PatchAccumulator struct {
	Patches []Hunk
}

func (pa *PatchAccumulator) AppendPatchIfBodyDiffers(hunk Hunk, newBody []byte) {
	if !bytes.Equal(hunk.Body, newBody) {
		hunk.Body = newBody
		pa.AppendPatch(hunk)
	}
}

func (pa *PatchAccumulator) AppendPatch(hunk Hunk) {
	if pa.Patches == nil {
		pa.Patches = make([]Hunk, 0)
	}
	pa.Patches = append(pa.Patches, hunk)
}
