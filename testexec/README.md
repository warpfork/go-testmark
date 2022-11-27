the testexec extension to testmark
==================================

See the [README](../README.md) in the parent directory for a broader introduction to testmark as a format.

"testexec" is a convention used within the testmark format to describe
tests for executing processes.

Use the [well-known hunk names](#well-known-hunk-names) to write your fixtures,
then connect it to `go test` with the connecting the code you can see in the [example code](#examples).
It supports [autopatching and fixture generation](#autopatching),
and the [exec hooks are configurable](#configurable-exec-hooks).


Well-known hunk names
---------------------

The "testexec" extension is as simple as a series of well-known hunk names:

Firstly, either of these two forms can be used to specify the commands to test:

- "`script`" -- will feed the entire hunk of text as a script to a shell (by default, bash).
- "`sequence`" -- will run each line as a command as its own executable (parsing args by simply whitespace split).

(See [Script vs Sequence](#script-vs-sequence) for more on why there's two options.)

Second, there's the hunks that contain output assertions:

- "`output`" -- if present, will cause the commands to be given a unified stdout and stderr buffer, and it will be checked against this data when done.
- "`stdout`" -- if present, works like "output", but only collects stdout.  Cannot be combined with "output".
- "`stderr`" -- ditto "stdout", but for (you guessed it) stderr.
- "`exitcode`" -- if present, should contains a base-10 number for the expected exit code.  If not present, an exitcode of 0 will be expected.

For hunks not present, nothing will be checked.
(In other words, if you have a `stdout` hunk, but not a `stderr` hunk, whatever the stderr output is, it will be ignored by the test.)

Note that not every testexec dataset has to contain _any_ of "output", "stdout", "stderr", or "exitcode":
containing none of them means the commands in the sequence will all be run,
and the exitcode is expected to be zero from each,
and nothing about their stdout nor stderr will be checked.
(If you want to mandate that their output is empty, you must create a blank data hunk to say so explicitly.)

Input streams (a.k.a. "stdin") can also be specified:

- "`input`" -- if present, will be fed to the stdin stream of the execution.

Beyond this, you can also specify fixture filesystems:

- "`fs/*`" -- everything under here will be placed in a (temporary!) working directory during the run.
- "`fs/somedir/thefile.ext`" -- for example, causes "somedir" to be created, and places "thefile.ext" inside it.

And last of all, sequences of causally related tests can be created.
Any time a testexec script or sequence has siblings named "`then-*`",
that creates sub-tests.

For example, given a series of hunks with names like the following:

```
test-one/script
test-one/then-test-two/script
test-one/then-test-two/then-test-three/script
test-one/then-test-two-b/script
```

In the scenario above, a total of four tests will occur.
Each of the tests that are related by the nested "`then-*`" hunk naming will inherit the working directory from their parent --
or more specifically, a *copy* of it.
Tests that *aren't* related don't receive copies of the working directories!
(This means you can make diverging tests of various filesystem state evolution stories easily.)

Walking through the scenario above step by step:

- the script for `test-one` executes first;
- the script for `then-test-two` receives whatever filesystem state was left behind by the evalution of `test-one`;
- the script for `then-test-three` receives whatever filesystem state was left behind by `then-test-two`;
- and the script for `then-test-two-b` receives the filesystem as it was *after `test-one`*.
	- Note!  _not_ what was left by `then-test-three`!  The relationships are only explicit :) and the order of the testmark hunks doesn't matter.

The "`then-*`" system (and its helpful support for filesystem forking),
lets you do one-time setup work, or build narratives -- while still keeping tests cleanly isolated and very explicit.


Examples
--------

See the [`selfexercise.md`](./selfexercise.md) file in this directory for fully worked examples.
(Note: if you're reading this on Github: You may also want to look at the raw mode of it!
You won't see the testmark annotations in the rendered markdown on Github!)

See the [`testexec_test.go`](./testexec_test.go) file for an example of the complete wiring to execute the examples --
including the autopatcher support.


Autopatching
------------

**Yes** -- this testexec implementation supports automatic fixture regeneration.

There's a tiny amount of setup required:
you do have to create a PatchAccumulator,
assign it to the `Tester{}.Patches` field,
and then tell the PatchAccumulator where to write any updates to.
(This is the norm for use of the patching system.)
(PRs welcome for further automating this.)
Once that setup is done:

Just write your 'script' hunk,
and whichever of the 'output', 'stdout', 'stderr', or 'exitcode' hunks that you want to track,
and run your tests like `go test ./... -testmark.regen`.
The document will be patched with the actual output inserted into those hunks.

Note that regen mode will only update hunks that already exist; it won't add them.
(E.g., if you only have a `stdout` hunk, regen won't _add_ a `stderr` hunk.)


Script vs Sequence
------------------

The testexec convention covers two different kinds of instruction specification hunk:
"script" and "sequence".

Each mode is similar, and the content of each looks roughly like a shell script, but they differ slightly.

"sequence" mode is defined as just splitting words based on whitespace, and processing args that way.
It's meant to be a little less complicated to implement, and is a bit more portable (it's not invoking a shell!),
but of course it's also a bit less flexible.

"script" mode is defined as doing whatever a shell does -- so, it should handle quoting, support redirections, etc.
In exchange, the challenge is then you have to know what your shell does!
(And internally, the default implementation is literally executing the host shell -- so there's some portability considerations, there.)

In this library, each of these modes can be set up to do something different, by use of different callbacks when you're setting up the code --
see the next section about [configurable exec hooks](#configurable-exec-hooks).

In practice: we often find that the "sequence" mode's callback is easy to implement
and connect to the program's "main" method...
which makes it very easy to write tests that exercise a CLI-style application!
(The alternative would likely be doing a separate "go install" phase and invoking the application again as a subprocess,
but this is considerably more complicated to implement and maintain.)
But we still also find people usually reaching for "script" mode when testing how one program interacts with others,
because "sequence" mode is incapable of supporting piping programs together, etc.
Ultimately, both approaches have some attractions :)


Configuration Hooks
-------------------

### Configurable Exec Hooks

The default behaviors for this testexec implementation are:

- the sequence mode does literally the system exec calls (e.g. for a single process).
- the script mode will invoke bash and feed the script to it.

This is customizable in the Go code.
See the `testexec.Tester` struct, and its fields `ExecFn` and `ScriptFn`
for the two callbacks which can replace the default execution behaviors.

### Bring your own assertion library

... if you want.

You can configure your choice of test equality checking function in the `testexec.Tester` struct by setting the `AssertFn` field.

(The authors of this package happen to like `frankban/quicktest`, for example -- but we didn't want to force that choice on you!)

The default function does some very simple string comparisons,
and has zero dependencies,
but does not provide rich diff output or any other nice-to-have features.

### Filtrations

Sometimes you want to test an application that is mostly predictable, but perhaps includes some unpredictable outputs, like timestamps for example.

You can set the `FilterFn` field in the `testexec.Tester` struct to apply some normalizing transforms to the output streams before doing the comparisons.
