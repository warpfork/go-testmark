go-testmark
===========

Do you need test fixtures and example data for your project, in a language agnostic way?

Do you want it to be easy to combine with documentation, and easy for others to read?

Do you want those fixtures to be easy to maintain, because they're programmatically parsable, and programmatically patchable?

Do you want to be able to display those fixtures and examples in the middle of the docs you're already writing?
(Are those docs already in markdown, and all you need is some glue code to make those code blocks sing?)

You're looking for `testmark`.  And you found it.

This is `go-testmark`, a library implementing a parser (and patcher!) for the `testmark` format,
which is itself a subset of markdown that you can use anywhere you're already using markdown.

It'll make your code blocks sing.

Read on:

---

- [The Testmark Format](#what-is-the-testmark-format)
	- [Examples](#testmark-format-by-example)
	- [Purpose](#the-purpose-of-testmark)
- [Implementations](#implementations)
	- [go-testmark](#go-testmark)
	- [js-testmark](#js-testmark)
	- [rust-testmark](#rust-testmark)
	- [python-testmark](#python-testmark)
	- [your-language-here-testmark](#your-language-here-testmark)
- [The go-testmark Library](#this-is-go-testmark)
	- [Features](#features)
		- [Parsing](#parsing)
		- [Walking](#walking-and-indexing)
		- [Patching](#patching)
		- [Writing](#writing)
		- [Autopatching](#autopatching)
	- [Examples](#examples)
		- [Examples in the Wild](#examples-in-the-wild)
	- [Extensions](#extensions)
		- [The `testexec` Convention](#the-testexec-convention)
- [License](#license)
- [Acknowledgements](#acknowledgements)

---


What is the testmark format?
----------------------------

Testmark is a very simple, and language-agnostic, format.
It's also a subset of markdown.
Your markdown documents can also be testmark documents!

Testmark data is contained in a markdown codeblock -- you know, the things that start and end with triple-backticks.

Testmark data is labled by a markdown comment.

This means you can easily embed large blocks of data in a markdown file,
and give them whatever label is appropriate,
and do this seamlessly in a document you're also writing for a human audience.

This makes for great development of test fixtures, specs, and documentation, all in one.

### testmark format by example

You see markdown code blocks all the time.  They render like this:

[testmark]:# (this-is-testmark-btw)
```json
{"these things": "you know?",
 "syntax highlighed, typically": "etc, etc"}
```

That right there?  That was also testmark fixture data.

Check out the "raw" form of this page, if you're looking a rendered form.
There's a specially structured "comment" in markdown which is tagging that markdown code block,
marking it as testmark data, and giving it a name.

The comment looks like this:

```
[testmark]:# (the-data-name-goes-here)
```

... and then the triple-backticks go right after that.
Data follows until the next line starting with triple-backticks, as is usual in markdown.

So, in total, it looks like this:

```
	[testmark]:# (the-data-name-goes-here)
	```text
	this is your data block
	as big as you like
	```
```

That's it.

Check out the [testdata/example.md](testdata/example.md) file (and other files in that directory) for more examples.
Be sure to mind the [raw form of the file](https://raw.githubusercontent.com/warpfork/go-testmark/master/testdata/example.md) too.

### the purpose of testmark

Formats for test fixtures and example data are extremely useful.
Some kind of language-agnostic format is critically important any time you're working on a project that involves codebases in more than one language, or any kind of networked interoperability.

Yet, picking a format (and getting people to agree on it) is hard.  And then getting it in your documentation is hard.

And then *maintaining* the fixtures and your documentation is hard, because you're typically stuck between two choices that are both bad:
either you can put some very ugly fixture formats in your documentation (and eventually realize that most users won't read them anyway, because of the eyeburn);
or you can maintain the fixtures and documentation separately,
while manually putting very pretty examples in the middle of your documentation (but then fail to make them load-bearing, so eventually coderot strikes,
and now your documentation misleads users, and adoption drops and frustration rises, and oh dear).

**Testmark is meant to solve all of these things.**

- Because testmark is "just markdown", you can easily use it together with other things that are already markdown.  That means including in documentation, websites, etc.
- Because you can intersperse markdown and prose with the code blocks, you can make good, readable, living and verifiable documentation, directly intertwined with your test data.
  It's great for commenting; it's great for docs.  People will actually read this!  And you have full control over the formatting and presentation.  Annotate things however you like.
- Because testmark is "just markdown", you can probably conclusively leap past a lot of bikeshedding conversations about test fixture data formats.
  Markdown isn't great, but good heavens is it useful, and ubiquitous.  Your colleagues will probably agree.
- Because testmark is "just markdown", you get all the other tools that work on markdown, for free.  For example, that tasty, tasty syntax highlighting.
  There's no smoother way to getting pretty example data on a website and getting it directly used in tests at the same time.
  (No fancy website or publishing tool pipeline needed, either.  You can just use github readme files -- just like this one!)
- Because testmark is "just markdown", you can probably hand people links that jump them directly to headings in your fixtures files.
  Users who need those references will appreciate this; you who authors the fixtures and specs will probably take joy from being able to point directly at your latest work.
- Because more than one code block can be in a file, and you can tag them with names, you can cram many fixtures in one file.  (Or not.  Up to you.)
- Because it's machine-parsable, we can have tools and libraries that programmatically update the data blocks, too.
  And because testmark interacts with the markdown format in a very deterministic and well-bounded way, your markdown prose stays exactly where you put it, too.
  Easy fixture maintenance and automation, _and_ good human readability?  Yes, we _can_ have both.

tl;dr: deduplicate the work of spec fixtures and docs, both saving time, and getting more confident in the results, simultaneously.

### parsing testmark is easy

See [README_parsing](README_parsing.md) if you would like to write a testmark parser in another language.
It's _extremely_ straightforward.


Implementations
---------------

Testmark has implementations in a variety of languages already!

### go-testmark

It's right here :) https://github.com/warpfork/go-testmark/

### js-testmark

Check out https://github.com/rvagg/js-testmark/ !

### rust-testmark

Check out https://github.com/bsundsrud/rust-testmark !

### python-testmark

Check out https://github.com/iameru/python-testmark !

### your-language-here-testmark

It is _extremely easy_ to write new testmark parsers.

The average time taken to produce a testmark library in a new language empirically appears to be about 5 hours.
It's significantly less than a day's work.

See [README_parsing](README_parsing.md) for a guide.

Send us a note (or a PR to this list!) when you publish it!


This is go-testmark
-------------------

This is a golang library that implements a parser, a patcher, and a writer for the testmark format.

You should be able to look at the [godoc](https://pkg.go.dev/github.com/warpfork/go-testmark) for about five seconds and figure it out.  There's not much to it.

### Features

#### parsing

`go-testmark` can parse any markdown file and look for testmark data hunks.

When you've parsed a testmark file, you can iterate over all the data hunks in it, and see their names, or look up them up by name.

Parsing works in the simplest way possible.
It only looks at the code blocks tagged as testmark.
(It actually ignores the actual markdown content as completely as possible.
Simple is good.  And it turns out it's possible to parse testmark data out,
and even later support patching the testmark data blocks, without a complete markdown parser.)

#### walking and indexing

You can range linearly over the slice of parsed hunks in a `Document` once you've parsed it.

Each hunk has a name (from the testmark comment),
a body (the blob from inside the code block),
and optionally may have the code block's tag (if any; usually this is already used by other people, for syntax highlighting indicators).

If you use hunk names that look like filesystem paths (e.g. "foo/bar/baz", with slashes),
you can also get an indexed view that lets you easily walk it as if it was directories.
Just call [Document.BuildDirIndex](https://pkg.go.dev/github.com/warpfork/go-testmark#Document.BuildDirIndex).
("Directories" for names with many segments will be created implicitly; it's very low friction.)

Once you've built a directory index, you can range over `DirEnt` either as an ordered list of its contents,
or look things up by path segment like a map.

#### patching

When using the patch operation, the markdown you wrote will be maintained by the operation; only the testmark data blocks change.
(No markdown gets reformated; nothing tries to normalize anything.  Whatever you write is safe.
Use whatever other markdown extensions you like; we're not gonna error if there's something fancy we didn't expect.  It's chill.)

Patching is _really_ simple.  It looks like this:

```go
doc, err := testmark.ReadFile("example.md")
doc = testmark.Patch(doc,
	testmark.Hunk{Name: "more-data", BlockTag: "text", Body: []byte("you have been...\nreplaced.\nand gotten\nrather longer.")},
	testmark.Hunk{Name: "this-is-the-data-name", BlockTag: "", Body: []byte("this one gets shorter.")},
	testmark.Hunk{Name: "this-one-is-new", BlockTag: "json", Body: []byte(`{"hayo": "new data!"}`)},
	testmark.Hunk{Name: "so-is-this", BlockTag: "json", Body: []byte(`{"appending": "is fun"}`)},
)
fmt.Printf("%s", doc.String())
```

(That's real code from our tests, and it applies on the `example.md` file in the [testdata](testdata) directory.)

#### writing

`go-testmark` can write back out a document that it's holding in memory.

You'll produce these by parsing, and by patching.

It's not really encouraged to try to create a new document purely via the `go-testmark` APIs.
We don't offer any APIs for writing and formatting markdown _outside_ of the testmark data blocks;
it's better to just write that yourself, in an editor or with other tools fit for the purpose.

(You probably can start with an empty document and just patch hunk into it, and it'll be fine.
It's just dubious if you'll really want to do that in practice.)

#### autopatching

`go-testmark` supports automatically regenerating fixtures.

Use the `-testmark.regen` flag during testing.  E.g., `go test ./... -testmark.regen`.

If you're using [extensions](#extensions) like [testexec](#the-testexec-convention), it'll automatically do the right thing for you.
If you're writing your own tests, you will need to wire up patching and check the [Regen flag](https://pkg.go.dev/github.com/warpfork/go-testmark#Regen)
in order to take advantage of this.  (Testmark can't read your mind on data sources and what should be patched unfortunately.)

Use this kind of regeneration when you make a change to code and would rather review the result via
something like `git diff` instead of editing the fixture content manually.
It's a huge timesaver.

It's also great for just making blank hunks and filling them in the first time, quickly and easily.

### Examples

Check out the [`patch_test.go`](patch_test.go) file for an example of what updating a testmark file looks like with this library.

#### Examples in the Wild

Check out how the IPLD project uses testmark:

- This document is both prose documentation for humans, and full of testmark data (using directory naming conventions for the hunk names, too):
  [ipld/specs/selectors/selector-fixtures-1](https://github.com/ipld/ipld/blob/17fd0efb695eb4933a68ca55d48c8e1dd765734b/specs/selectors/fixtures/selector-fixtures-1.md)
- This code reads that document, and in a handful of lines, iterates over the "directories" of hunks, and then plucks data out of them:
  [go-ipld/selector/spec_test.go](https://github.com/ipld/go-ipld-prime/blob/5c39e6803594f599a85d2545ad72faf584bf6f19/traversal/selector/spec_test.go#L29-L39)

### Extensions

Testmark can be "extended" by giving behaviors to certain conventions of data hunk names.
(These aren't changing the serial format at all -- just decorating it with behaviors, by writing reusable test code against a naming convention.)

It's easy to build these kinds of extensions and conventions on top of a testmark library.
They don't need to be integrated!  You can develop your reusable test code wherever and however you want.

As a general rule, an "extension" tends to take the form of a function,
and it simply takes a parsed testmark document as a parameter,
and optionally a patch accumulator object (or just a filename).
You have to call the extension yourself in your test code.
(This means there's no plugin arcitecture to worry about.
You also never need to fear an extension giving magic special meaning to a hunk name that you didn't expect.
You just... call the code you want, and that's it.)

Some testmark libraries may also bundle some extensions (for convenience),
and try to form relatively well-known conventions.
We have at least one of these in go-testmark; see below.

#### The `testexec` Convention

The "testexec" convention is for testing CLI programs quickly and conveniently.

Long story short: you make a hunk called `mytest/script` and put commands in it,
and a hunk called `mytest/output` and put the expected output in it, and you're already good to go.

(No, this is not enabled by default.  You'll have to call a function in your test code explicitly to use this behavior.)

There's a bunch more features too, like ways to check exit codes,
ways to check stdout and stderr streams separately,
ways to populate fixture filesystems (using hunks like `mytest/fs/dirname/filename.ext`), etc.
See that package for more details!



License
-------

SPDX-License-Identifier: Apache-2.0 OR MIT


Acknowledgements
----------------

This is probably inspired by a lot of things.  (Mostly my own failings.  But hey.  Those *do* teach you something.)

- It's probably inspired heavily by rvagg's work on making IPLD Schema DSL parsers, which could parse content out of markdown codeblocks, and did this for similar "being able to embed the real data in the docs is cool" reasons.
  (That work differs slightly, in that that system just ran with the code block syntax tag hint, and also, has no patching capabilities, and also, aggregated all the data rather than making it accessible as named blocks.  But the goals and reasons are very similar!)
- It's probably inspired a bit by campoy's [`embedmd`](https://github.com/campoy/embedmd) tool.
  (That work uses markdown comments in a similar way.  Testmark differs in that it's meant for programmatic patching rather than use as a file-wangling tool, and also programmating reading; and that it treats the markdown file as the source of truth, rather than a terminal output.)
- It's influenced by "[taffy](https://github.com/warpfork/go-taffy)", another test fixture format I wrote not long before this one.
  (Taffy didn't get very far.  It was special for no reason.  Technically, it's "more correct" than testmark, because you can put *any* data in it.  But: describing things attractively within the taffy format was basically impossible.  That turns out to kill.  This lesson informed the idea for testmark.)
- It's also influenced by an even older attempt at test fixture format called [wishfix](https://github.com/warpfork/go-wish/blob/master/wishfix/format.md) ([example](https://github.com/polydawn/repeatr/blob/d581713218bad916aee5b67a55e93806bb8873f2/examples/hello-cached.tcase)).
  (You can see the "it should be attractive" rule applied more strongly in wishfix than in taffy; and yet, still, a lack of flexibility about formatting.  The lesson to learn was again: don't be special; just use a format that's already capable of being decorative.)
- Probably other things as well.  A lot of test fixture formats have passed, however briefly, through my brain over the years.  My apologies for any acknowledgements forgotten.
