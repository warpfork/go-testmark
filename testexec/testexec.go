/*
	testexec is a package that offers some helper functions for testing
	execution of commands (either with actual os.Exec, or a callback of your defining),
	and tests them using certain conventions of testmark dir+filename names for the fixture data.

	This package isn't part of testmark's core, nor is it particularly special.
	It's just some conventions that may be handy.
*/
package testexec

import (
	"io"
	"testing"

	"github.com/warpfork/go-testmark"
)

// ExecFn is the outline for a callback that can run individual commands in a sequence of commands.
// An ExecFn can be placed in the Tester config struct to control the behavior of TestSequence.
//
// Semantically, an ExecFn is supposed to treat its args slice roughly as `os.StartProcess` treats `argv`:
// the zeroth element of the slice is the command name (or path), and subsequent slice elements are positional arguments to that command.
// However, it's a freeform callback: you can do as you wish.
type ExecFn func(args []string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int, oshit error)

// ScriptFn is the outline for a callback that can run a script.
// A ScriptFn can be placed in the Tester config struct to control the behavior of TestScript.
//
// Semantically, ScriptFn is... quite free to do whatever it wishes.
// A typical implementation (and the default behavior of this package, if you don't supply a ScriptFn)
// is to handle the script as if it was a bash shell script (so, piping, env vars, etc, are all possible).
// However, one could place any kind of script parser and interpreter within this callback.
type ScriptFn func(script string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int, oshit error)

// FilterFn is the outline for a callback that can be used to normalize some parts of output strings.
// A common example of this is to strip timestamps back out of log messages for a program that emits logs with timestamps.
type FilterFn func(line string) (replacement string)

// AssertFn can be used in Tester to specify a better test assertion function.
// For example, if you use the quicktest package, this is an excellent snippet which will result in good diffs:
//
//		func(...) { quicktest.Assert(t, actual, quicktest.CmpEquals(), expect) }
//
// The default behavior, if a Tester object doesn't get an AssertFn, is to fall back to a very basic call to `t.Errorf`.
type AssertFn func(t *testing.T, actual, expect string)

// Tester is a configuration-gathering structure.
// Each of the `Test*` methods upon it will use these callbacks to define their behavior.
//
// All of the fields can be nil, which will result in default behaviors.
// (A nil ExecFn will result in an OS exec being used (specifically: `ExecFn_Exec`);
// a nil FilterFn means no filtering will occur;
// a nil AssertFn means a very basic check using t.Errorf will be used.)
type Tester struct {
	ExecFn
	ScriptFn
	FilterFn
	AssertFn
}

// TestSequence runs a test based on a "sequence" instruction -- a hunk called "sequence" should have a series of lines,
// and each line is a command to be executed.
//
// The model of execution and model of environment used is very simple:
// it's just consecutive lines which will be fed one after another to an ExecFn.
// There is *not* a scripting engine involved (you can't "pipe", use "env vars", use "subcommands",
// "$?" has no special meaning, etc; there's no shell interpreting those things).
// (The ExecFn can be as simple or as complex as you want, of course -- but none of these features are provided for you.)
//
// The parser that splits up each line of the sequences into the args string slice is primitive.  It's simply whitespace splitting.
// If you need to test the handling of arguments that involve whitespace, there are two options:
// One, you may want to use TestScript instead (this will let you parse it yourself, and or just hand off to a shell of some kind which means you can use the shell's quoting rules);
// or, Two, you can change your sequence hunk name to `sequences.jsonl`, and put a json list on each line, which will be parsed and that list becomes the args.
//
// Each DirEnt can also contain several other named entries which will be treated specially.
// "output" -- if present, will cause the commands to be given a unified stdout and stderr buffer, and it will be checked against this data when done.
// "stdout" -- if present, works like "output", but only collects stdout.  Cannot be combined with "output".
// "stderr" -- ditto "stdout", but for (you guessed it) stderr.
// "exitcode" -- if present, should contains a base-10 number for the expected exit code.  If not present, an exitcode of 0 will be expected.
//
// Not every data DirEnt has to contain any of "output", "stdout", "stderr", or "exitcode".
// Containing none of them means the commands in the sequence will all be run,
// and the exitcode is expected to be zero from each,
// and nothing about their stdout nor stderr will be checked.
// (If you want to mandate that their output is empty, you must create a blank data block to say so explicitly.)
//
// If you wish to run some commands, and gather and test (or ignore!) their output as one block,
// then run additional commands in a subtest, you can use another special path name
// in the testmark.DirEnt for that: `then-(.*)`.
// Anything starting with "then-" will cause a new nested `t.Run` block,
// which will have a name that is the rest of the path name once the "then-" prefix has been stripped.
// Inside that DirEnt, all the same rules apply again (we'll look for a "sequence" hunk, etc).
//
// Note that there are no faculties to expect a specific exitcode from a specific line of the sequence,
// nor are their faculties to extract outputs from one specific line of the sequence.
// If you are wanting to do these things, you can use the "then-" feature.
// (This is probably a nudge towards writing better documentation anyway:
// if you have a whole series of commands and it's a very specific one that should stand out,
// you should probably give it separate data blocks for sheer human readability anyway.)
//
// There is no faculty for ignoring an exitcode.
// Surely the systems you're so rigorously ensuring the quality of don't have nondeterministic exit codes;
// so why would you need to do this?  If it's allowed to be anything but zero, it should be expected to be that; and if it's expected, say so.
func (tcfg Tester) TestSequence(t *testing.T, data testmark.DirEnt) {
	panic("not yet implemented")
}

func (tcfg Tester) TestScript(t *testing.T, data testmark.DirEnt) {
	panic("not yet implemented")
}

func (tcfg Tester) Test(t *testing.T, data testmark.DirEnt) {
	panic("not yet implemented")
}

// Not yet defined how these will nest.  ISTM if you specify one or the other, it shouldn't become willing to switch when it goes deeper into then-trees.

// Not yet defined if these should complain loudly if they _don't_ find something that matches.  I think being able to shrug is useful; otherwise that check will often get offloaded to callers.
