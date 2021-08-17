package testmark

type Document struct {
	// The whole thing, complete, but split into lines.
	// We always save this, because if we are going to write this document back out,
	// it's going to be by patching this.  (We don't try to understand, much less normalize, a full markdown AST!)
	Original []byte

	// Original, sliced into lines.  Shares backing memory with Original.
	// Useful because we made it during parse anyway, and it can save us a lot of work during edits.
	OriginalLines [][]byte

	// Each data hunk.
	// Contains just offset information, and the parsed name header.
	DataHunks []Hunk

	// Like it says on the tin.
	HunksByName map[string]Hunk
}

type Hunk struct {
	// Index into Document.OriginalLines where the comment block is found.
	// The code block indicator is necessarily is the following line,
	// and the code block body one line after that.
	// N.B. zero-indexed.  You probably want to +1 before printing to a human.
	Line int

	// Index into Document.OriginalLines that contains the closing code block indicator.
	EndLine int

	// The hunk name (e.g. whatever comes after `[testmark]:# ` and before any more whitespace).
	// Cannot be empty.
	Name string

	// The code block syntax hint (or more literally: anything that comes after the triple-tick that starts the code block).
	// Usually we don't encourage use of this much in testmark, but it's here.  Can be empty.
	BlockTag string
}
