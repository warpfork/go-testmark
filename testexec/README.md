the testexec extension to testmark
==================================

See the README in the parent directory for a broader introduction to testmark as a format.

"testexec" is a convention used within the testmark format to describe
tests for executing processes.

Use the [well-known hunk names](well-known-hunk-names) to write your fixtures,
then connect it to `go test` with the connecting the code you can see in the [example code](#examples).
It supports [autopatching and fixture generation](#autopatching),
and the [exec hooks are configurable](#configurable-exec-hooks).


Well-known hunk names
---------------------

The "testexec" extension is as simple as a series of well-known hunk names:

Firstly, either of these two forms can be used to specify the commands to test:

- "`script`" -- will feed the entire hunk of text as a script to a shell (by default, bash).
- "`sequence`" -- will run each line as a command as its own executable (parsing args by simply whitespace split).

("sequence" mode is meant to be a little less complicated to implement, and more portable, but is also a bit less flexible.
"script" mode can do whatever the shell does -- so, handle quoting, redirections, etc -- the only challenge is then you have to know what your shell does!)
i
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

- "`fs/somedir/thefile.ext`" -- (and names of that general pattern: anything starting with "fs/")

And last of all, sequences of causally related tests can be created:
given a series of hunks with names like the following:

```
test-one/script
test-one/then-test-two/script
test-one/then-test-two/then-test-three/script
test-one/then-test-two-b/script
```

In a scenario like the above, the nested "then-*" tests will inherit the working directory from their parent --
or more specifically, a *copy* of it.
So: the script for `then-test-two` receives whatever the filesystem state left behind by `test-one` was;
the `then-test-three` script receives whatever filesystem state left behind by `then-test-two` was;
and `then-test-two-b` receives the filesystem as it was *after `test-one`* (!  _not_ what was left by `then-test-three`).
(This lets you do one-time setup work, or build narratives -- while also still keeping tests isolated.)


Examples
--------

See the `selfexercise.md` file in this directory for fully worked examples.

See the `testexec_test.go` file for the complete wiring to execute the examples --
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


Configurable Exec Hooks
-----------------------

The default behaviors for this testexec implementation are:

- the exec mode does literally the system exec calls (e.g. for a single process).
- the script mode will invoke bash and feed the script to it.

This is customizable in the Go code.
See the `testexec.Tester` struct, and its fields `ExecFn` and `ScriptFn`
for the two callbacks which can replace the default execution behaviors.
