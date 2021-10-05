Parsing Testmark
================

Parsing Testmark is extremely easy.


The Testmark Format
-------------------

The testmark format is a subset of markdown, specifically selected to be easy to parse out of a larger markdown document.

The format looks generally like this:

```
	[testmark]:# (some/hunk-name-goes-here)
	```text
	this is your data block
	as big as you like
	```
```

More specifically:

1. There must be a markdown "comment" of the form seen above (e.g. a line that starts with `[testmark]:# (` and ends in `)`).
2. It must be followed by a data block that starts with triple-backticks, which may not be indented.
3. The data block ends when there are again triple-backticks, which may again not be indented.

(Note that there are quite a few other ways of declarating code blocks in markdown -- testmark does *not* support all of them;
it focuses on exactly one, for simplicity and clarity.
(See discussion in https://github.com/warpfork/go-testmark/issues/1 .))


Implementing a Testmark Parser
------------------------------

There are two major roads you can take:

1. Use a full-on markdown parser -- then only pluck out the data you need.
2. Use a very (very!) simple parser that is sufficient to recognize testmark data within markdown.

We actually recommend [Option 2](#option-2-write-a-testmark-focused-parser).
Simple is better.


### Option 1: use a full-on markdown parser

Of course, since testmark is "just markdown",
if you are working in a language that already has a suitable markdown parsing library,
you can use that.

You need to look for two things:

1. markdown comment lines
	- (which, if the library you're using doesn't talk about that explicitly, might also be referred to as "link labels")
	- which have the label of "testmark"
2. ...which are immediately followed by a code block.

We won't spent any more time talking about how do actually do that in this document, though,
because most of what you'll need to figure out is specific to the markdown library you use.

We also don't really actually recommend this road at all.
You can probably build something that's:

- faster,
- has fewer dependencies,
- is equally correct,
- easier to debug,
- and has an easier, more reliable way to support "patching"

... by going with [Option 2](#option-2-write-a-testmark-focused-parser): writing a simple parser focused on testmark alone.


### Option 2: write a testmark-focused parser

It's extremely simple to build a direct testmark-focused parser --
the only features we need to parse are line-break delimited,
so the parser is both easy and extremely fool-proof to implement.

(We _don't_ need to parse most of markdown in order to find the testmark elements!)

Because every piece of data which is important to parse starts exactly on the beginning of a line,
we can split the document into lines first, and then parse each line in a very direct way.

- Split document into lines.
- For each line: look for lines starting with triple-backticks.  These begin or end codeblocks.
	- Store a state toggle for whether you're in a codeblock.  Do not parse lines within a codeblock.
	- ... except to look for the end of the codeblock, which is another triple-backtick.
- For each line not in a codeblock: if it starts with `[testmark]:# ` -- it may be a testmark block starting.
- If the next line is the start of a codeblock, it's definitely a testmark block.
- That's it.

See how easy that was?

There's a couple more details you should also mind:

- The testmark comment line should continue with `(`, contain some text, and end in `)`.
	- If it does not, you should notify the user of a funny thing here, and ignore it.
	- The contained text, up to the first element of whitespace, is the testmark block label.
	- Any content after the whitespace is currently ignored (but considered reserved for potential future use).
- The testmark comment line must *immediately* precede a codeblock.
	- If there is no codeblock after a testmark comment line, you should notify the user of a funny thing here, and ignore it.
- Codeblocks start lines can contain some test after the triple-backtick.  This block tag is not part of testmark, but you may wish to know it's there.
	- You can ignore it completely.  It usually contains a syntax rendering hint (this is a feature of GFM, a common markdown extension).
	- Some testmark libraries detect this field, and allow it to be set during patch operations, but this is not considered a required feature of testmark.

You can see a golang implementation of this in [read.go](read.go) if the example is helpful.
The complete parser is less than 100 lines, and many of them are comments.

### patching

If you wish to support patching operations on a testmark document, this is very straightfoward by extending the above:
when splitting the document into lines, just... remember that, and remember the line number where all the testmark hunks start and end.
This information will be sufficient to enable patching.

Now, when replacing a testmark data hunk, simply:

- Emit all the lines of the original document, up until the testmark data hunk started;
- Now serialize and emit the new testmark data hunk...
	- One comment line
	- One codeblock opener line
	- All the user-specified content body lines
	- One codeblock closer line
- Emit all the lines of the original document, starting at the line number after the old testmark data ended.

This is a very nice approach because we have (again) completely avoided parsing markdown,
and thus the operations have been very simple, very fast, and very foolproof.
This algorithm works exactly correctly no matter what other markdown extensions people may be using in their documents,
and will never rewrite anything except the testmark data blocks, and thus follows "the principle of least surprise" very well.

You can see a golang implementation of this in [patch.go](patch.go) if the example is helpful.
It is again less than 100 lines, and nearly 50% of them comments.
It also includes efficient batching application many patches at once.


About Linebreaks
----------------

This testmark implementation always reports data hunks as having LF-only line endings (e.g. unix style).
If it sees CRLF linebreaks (e.g. windows style) in a document, it will normalize them to LF-only in the byte slices yielded to the user.

We consider this a tactically useful choice, because there are other systems in the world that convert linebreaks to CRLF
(namely, git, with default configuration, on windows environments -- which is often relevant, because testmark files are often stored in git, and checked out on windows),
and yet rarely if ever do developers want to have to shield their tests against this variation.

(See https://github.com/warpfork/go-testmark/pull/4 and https://github.com/warpfork/go-testmark/pull/3 for discussion.)
