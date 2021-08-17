package testmark

type Document struct {
	// The whole thing, complete, but split into lines.
	// We always save this, because if we are going to write this document back out,
	// it's going to be by patching this.  (We don't try to understand, much less normalize, a full markdown AST!)
	// May be nil if this document is the result of Patch operations rather than Parse.
	Original []byte

	// The document, sliced into lines.  Shares backing memory with Original, if Original is non-nil.
	// Useful because we made it during parse anyway, and it can save us a lot of work during edits.
	Lines [][]byte

	// Each data hunk.
	// Contains just offset information, and the parsed name header.
	// Is in order of hunk appearance.
	DataHunks []DocHunk

	// Like it says on the tin.
	HunksByName map[string]DocHunk
}

// DocHunk is the Document's internal idea of where hunks are.
type DocHunk struct {
	// Index into Document.OriginalLines where the comment block is found.
	// The code block indicator is necessarily is the following line,
	// and the code block body one line after that.
	// N.B. zero-indexed.  You probably want to +1 before printing to a human.
	LineStart int

	// Index into Document.OriginalLines that contains the closing code block indicator.
	LineEnd int

	Hunk
}

// Hunk is a simple tuple of hunk name string and body bytes.
// Optionally, it may also have a BlockTag (which is whatever markdown has in the code block; usually, in practice, this is used to state a syntax for highlighting, which does not have much to do with testmark.)
type Hunk struct {
	// The hunk name (e.g. whatever comes after `[testmark]:# ` and before any more whitespace).
	// Cannot be empty.
	Name string

	// The code block syntax hint (or more literally: anything that comes after the triple-tick that starts the code block).
	// Usually we don't encourage use of this much in testmark, but it's here.  Can be empty.
	BlockTag string

	// The full body of the hunk, as bytes.
	// (This is *still* a subslice of Document.Original, if this hunk was created by Parse, but probably a unique slice otherwise.)
	Body []byte
}
