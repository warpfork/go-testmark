/*
	testexec is a package that offers some helper functions for testing
	execution of commands (either with actual os.Exec, or a callback of your defining),
	and tests them using certain conventions of testmark dir+filename names for the fixture data.

	This package isn't part of testmark's core, nor is it particularly special.
	It's just some conventions that may be handy.
*/
package testexec

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
//
// The 'Patches' accumulator will be used to gather new fixture data if `testmark.Regen` is true.
// (If the pointer is nil, a warning will be logged.)
// If the NewSuiteTester constructor is used, `Patches` will be wired and handled automatically.
// Otherwise, if `Patches` is handled manually, it becomes the user's responsibility
// to actually apply the patches and save the updated document.
type Tester struct {
	ExecFn
	ScriptFn
	FilterFn
	AssertFn

	Patches *testmark.PatchAccumulator

	reportUse     func(string)         // Used to wire with suite, if you use NewSuiteTester.
	reportUnrecog func(string, string) // Used to wire with suite, if you use NewSuiteTester.
}

func (tcfg *Tester) init() {
	if tcfg.ExecFn == nil {
		tcfg.ExecFn = ExecFn_Exec
	}
	if tcfg.ScriptFn == nil {
		tcfg.ScriptFn = ScriptFn_ExecBash
	}
	if tcfg.AssertFn == nil {
		tcfg.AssertFn = defaultAssertFn
	}
	if tcfg.reportUse == nil {
		tcfg.reportUse = func(string) {}
	}
	if tcfg.reportUnrecog == nil {
		tcfg.reportUnrecog = func(string, string) {}
	}
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
// Additionally, the commands can be run within a temp directory, with some files pre-populated,
// by use of data hunks under the "fs/" name.  So, "fs/foo.bar" will result in a file named "foo.bar" in the temp directory.
// The temp directory is applied by using `os.Chdir`, and so is not safe for use with concurrent tests.
// There are no special faculties for making empty directories, symlinks, setting file properties, etc;
// if you need to do anything fancy, a setup script may be a good direction to pursue.
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
//
// This test system comes complete with fixture regeneration support.
// If `testmark.Regen` is true (e.g. you have invoked "go test" with the argument "-testmark.regen"),
// then instead of making any assertions, this function will accumulate patches
// in the `Tester.Patches` slice.
// Regen mode will only update hunks that already exist; it won't add them.
// As an edge case, note that if that an exitcode hunk is absent, but a nonzero exitcode is encountered,
// the test will still be failed, even though in patch regen mode most assertions are usually skipped.
func (tcfg Tester) TestSequence(t *testing.T, data *testmark.DirEnt) {
	t.Helper()
	tcfg.test(t, data, true, false, "")
}

func (tcfg Tester) TestScript(t *testing.T, data *testmark.DirEnt) {
	t.Helper()
	tcfg.test(t, data, false, true, "")
}

func (tcfg Tester) Test(t *testing.T, data *testmark.DirEnt) {
	t.Helper()
	tcfg.test(t, data, true, true, "")
}

func (tcfg Tester) test(t *testing.T, data *testmark.DirEnt, allowExec, allowScript bool, parentTmpdir string) {
	t.Helper()
	tcfg.init()

	sequenceHunk, sequenceMode := data.Children["sequence"]
	scriptHunk, scriptMode := data.Children["script"]
	if !sequenceMode && !scriptMode {
		t.Fatalf("dir %q does not contain a 'script' or 'sequence' hunk", data.Path)
	}
	if sequenceMode && scriptMode {
		t.Fatalf("dir %q contained both a 'script' and a 'sequence' hunk, which is nonsensical", data.Path)
	}
	if sequenceMode && !allowExec {
		t.Fatalf("found sequence hunk but the test framework was invoked without permission to run those")
	}
	if scriptMode && !allowScript {
		t.Fatalf("found script hunk but the test framework was invoked without permission to run those")
	}
	if *testmark.Regen && tcfg.Patches == nil {
		t.Fatalf("%s\n%s",
			"testmark.regen mode engaged, but there is no patch accumulator available here",
			"nothing to do if requested to regenerate test fixtures but have nowhere to put data",
		)
	}

	// Create a tempdir, and fill it with any files.
	// (This used to be conditional on if this test, or any parents, had use of a 'fs/*' hunk...
	//  but in v0.9.0, we started making tmpdirs unconditionally, because it's fairly common for
	//   a test script to start of with commands that initialize a filesystem, and
	//    forcing such tests to start with a dummy 'fs/this-is-not-relevant' hunk seemed unfortunate.
	//  I've left the flow as written earlier, in case we want to make a flag to *dis*able this again.)
	var dir string
	useTmpdirs := true // This was previously conditional: `fsEnt, exists := data.Children["fs"]; exists || parentTmpdir != ""`
	if useTmpdirs {
		// Create and get into a tempdir.
		var err error
		dir, err = ioutil.TempDir("", "testmarkexec")
		if err != nil {
			t.Fatalf("test aborted: could not create tempdir: %s", err)
		}
		defer os.RemoveAll(dir)
		retreat, err := os.Getwd()
		if err != nil {
			t.Fatalf("test aborted: could not find cwd: %s", err)
		}
		defer os.Chdir(retreat)
		if err := os.Chdir(dir); err != nil {
			t.Fatalf("test aborted: could not chdir to tempdir: %s", err)
		}

		// If there was a parent tempdir: copy those files first.
		if parentTmpdir != "" {
			if err := copyFiles(parentTmpdir, dir); err != nil {
				t.Fatalf("test aborted: could not populate files to tempdir: %s", err)
			}
		}

		// Create any new files.
		fsEnt := data.Children["fs"]
		if fsEnt != nil {
			if err := tcfg.createFiles(fsEnt, "."); err != nil {
				t.Fatalf("test aborted: could not populate files to tempdir: %s", err)
			}
		}
	}

	// Prepare output buffers.
	var stdout, stderr io.Writer
	if _, exists := data.Children["output"]; exists {
		tcfg.reportUse(data.Children["output"].Path)
		stdout = &bytes.Buffer{}
		stderr = stdout
		if _, exists := data.Children["stdout"]; exists {
			t.Errorf("testexec entry %q shouldn't contain 'stdout' hunk if it also specifies a unified 'output' hunk", data.Name)
		}
		if _, exists := data.Children["stderr"]; exists {
			t.Errorf("testexec entry %q shouldn't contain 'stderr' hunk if it also specifies a unified 'output' hunk", data.Name)
		}
	}
	if _, exists := data.Children["stdout"]; exists {
		tcfg.reportUse(data.Children["stdout"].Path)
		stdout = &bytes.Buffer{}
	}
	if _, exists := data.Children["stderr"]; exists {
		tcfg.reportUse(data.Children["stderr"].Path)
		stderr = &bytes.Buffer{}
	}
	var exitcode int

	// Prepare an input buffer, if applicable.
	var stdin io.Reader
	if ent := data.Children["input"]; ent != nil {
		tcfg.reportUse(data.Children["input"].Path)
		stdin = bytes.NewReader(ent.Hunk.Body)
	} else {
		stdin = bytes.NewReader(nil)
	}

	// Do the thing.
	switch {
	case sequenceMode:
		tcfg.reportUse(sequenceHunk.Path)
		exitcode = tcfg.doSequence(t, sequenceHunk.Hunk, stdin, stdout, stderr)
	case scriptMode:
		tcfg.reportUse(scriptHunk.Path)
		exitcode = tcfg.doScript(t, scriptHunk.Hunk, stdin, stdout, stderr)
	}

	// Okay, comparisons time.
	// Or, regen time!
	if ent, exists := data.Children["output"]; exists {
		// stdout buffer should be prepared to be both stdout and stderr earlier before execution.
		bs := stdout.(*bytes.Buffer).Bytes()
		if *testmark.Regen {
			tcfg.Patches.AppendPatchIfBodyDiffers(*ent.Hunk, bs)
		} else {
			t.Run("check-combined-output", func(t *testing.T) {
				tcfg.AssertFn(t, string(bs), string(ent.Hunk.Body))
			})
		}
	}
	if ent, exists := data.Children["stdout"]; exists {
		bs := stdout.(*bytes.Buffer).Bytes()
		if *testmark.Regen {
			tcfg.Patches.AppendPatchIfBodyDiffers(*ent.Hunk, bs)
		} else {
			t.Run("check-stdout", func(t *testing.T) {
				tcfg.AssertFn(t, string(bs), string(ent.Hunk.Body))
			})
		}
	}
	if ent, exists := data.Children["stderr"]; exists {
		bs := stderr.(*bytes.Buffer).Bytes()
		if *testmark.Regen {
			tcfg.Patches.AppendPatchIfBodyDiffers(*ent.Hunk, bs)
		} else {
			t.Run("check-stderr", func(t *testing.T) {
				tcfg.AssertFn(t, string(bs), string(ent.Hunk.Body))
			})
		}
	}
	t.Run("check-exitcode", func(t *testing.T) {
		if ent, exists := data.Children["exitcode"]; exists {
			tcfg.reportUse(data.Children["exitcode"].Path)
			if *testmark.Regen {
				tcfg.Patches.AppendPatchIfBodyDiffers(*ent.Hunk, []byte(strconv.Itoa(exitcode)))
			} else {
				tcfg.AssertFn(t, strconv.Itoa(exitcode), strings.TrimSpace(string(ent.Hunk.Body)))
			}
		} else {
			tcfg.AssertFn(t, strconv.Itoa(exitcode), "0")
		}
	})

	tcfg.recurse(t, data, allowExec, allowScript, dir)
}

func (tcfg Tester) recurse(t *testing.T, data *testmark.DirEnt, allowExec bool, allowScript bool, parentTmpDir string) {
	alreadyFailed := t.Failed()
	for _, child := range data.ChildrenList {
		if _, exists := leafNodeTable[child.Name]; exists {
			//t.Logf("%s will not recurse into special leaf node %q", t.Name(), child.Name)
			continue
		}
		t.Run(child.Name, func(t *testing.T) {
			if len(child.Name) <= 5 || !strings.HasPrefix(child.Name, "then-") {
				tcfg.reportUnrecog(child.Path, "a child test in testmark must be named beginning with \"then-\"")
			}
			if alreadyFailed {
				// This comes after the recursion test because a file structure error should still fail
				t.Skipf("parent commands failed, so while more commands are specified, testing them is not meaningful")
			}
			tcfg.test(t, child, allowExec, allowScript, parentTmpDir)
		})
	}
}

// Hash Table of all the "special" nodes used by testexec.
var leafNodeTable = map[string]struct{}{
	"exitcode": {},
	"stderr":   {},
	"stdout":   {},
	"output":   {},
	"input":    {},
	"sequence": {},
	"script":   {},
	"fs":       {},
}

func (tcfg Tester) doSequence(t *testing.T, hunk *testmark.Hunk, stdin io.Reader, stdout, stderr io.Writer) (exitcode int) {
	t.Helper()
	// Loop over the lines in the sequence.
	lines := bytes.Split(hunk.Body, []byte{'\n'})
	for _, line := range lines {
		args := strings.Fields(string(line))
		if len(args) < 1 {
			continue
		}

		var err error
		exitcode, err = tcfg.ExecFn(args, stdin, stdout, stderr)
		if err != nil {
			t.Fatalf("execution failed: error from ExecFn is %q", err)
		}
		if exitcode != 0 {
			break // TODO: it's probably still an error if that happens before the end?
		}
	}
	return
}

func (tcfg Tester) doScript(t *testing.T, hunk *testmark.Hunk, stdin io.Reader, stdout, stderr io.Writer) (exitcode int) {
	t.Helper()
	var err error
	exitcode, err = tcfg.ScriptFn(string(hunk.Body), stdin, stdout, stderr)
	if err != nil {
		t.Fatalf("execution failed: error from script is %q", err)
	}
	return
}

// createFiles makes files and directories matching testmark hunks.
// It creates them relative to the os cwd plus prefix -- use with care.
func (tcfg Tester) createFiles(dir *testmark.DirEnt, prefix string) error {
	tcfg.reportUse(dir.Path)
	if dir.Hunk != nil {
		return ioutil.WriteFile(prefix, dir.Hunk.Body, 0644)
	} else {
		if err := os.MkdirAll(prefix, 0755); err != nil {
			return err
		}
	}
	for _, ent := range dir.Children {
		if err := tcfg.createFiles(ent, filepath.Join(prefix, ent.Name)); err != nil {
			return err
		}
	}
	return nil
}

// copyFiles does what it says on the tin, recursively,
// and handles dirs, files, and symlinks, with basic posix perms (the 0777 bits).
// It doesn't pay attention to dev nodes, uid/gid, etc etc.
func copyFiles(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, fileInfo os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			return os.Mkdir(filepath.Join(destination, relPath), fileInfo.Mode())
		case os.ModeSymlink:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, filepath.Join(destination, relPath))
		default:
			fnew, err := os.OpenFile(filepath.Join(destination, relPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileInfo.Mode())
			if err != nil {
				return err
			}
			defer fnew.Close()
			fold, err := os.OpenFile(path, os.O_RDONLY, 0)
			if err != nil {
				return err
			}
			defer fold.Close()
			_, err = io.Copy(fnew, fold)
			return err
		}
	})
	return err
}
