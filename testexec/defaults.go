package testexec

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	"github.com/warpfork/go-testmark"
)

func ExecFn_Exec(args []string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int, oshit error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	// A bunch of processing is required to get typical unix error codes out of golang's exec system, unfortunately.
	// We're going to:
	// - Try to return the exit code, if we can parse it -- and then *not* return an error.
	// - Try to return a negative number with the signal if that's what killed things -- and then again *not* return an error.
	// - Or return -1000 and an error, for anything else we can't make sense of.
	// The -1000 return probably doesn't matter much -- you should check the error first (and where testmark uses this callback, it does),
	//  but a non-zero value there seems like better defense-in-depth anyway.
	if exitErr, ok := err.(*exec.ExitError); ok {
		waitStatus, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok {
			return exitErr.ExitCode(), nil
		}
		if waitStatus.Exited() {
			return waitStatus.ExitStatus(), nil
		} else if waitStatus.Signaled() {
			return -int(waitStatus.Signal()), nil
		}
	}
	if err != nil {
		return -1000, err
	}
	return 0, nil
}

func ScriptFn_ExecBash(script string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int, oshit error) {
	return ExecFn_Exec([]string{"bash", "-c", script}, stdin, stdout, stderr)
}

func defaultAssertFn(t *testing.T, actual, expect string) {
	if actual != expect {
		t.Errorf("expected: %q; actual: %q", expect, actual)
	}
}

// Hash Table of all the "special" nodes used by testmark.
var DefaultLeafNodes = map[string]struct{}{
	"exitcode": {},
	"stderr":   {},
	"stdout":   {},
	"output":   {},
	"input":    {},
	"sequence": {},
	"script":   {},
	"fs":       {},
}

func RecursionFn_Then(t *testing.T, dir testmark.DirEnt) error {
	t.Helper()
	if _, exists := DefaultLeafNodes[dir.Name]; exists {
		t.Logf("will not recurse into special leaf node %q", dir.Name)
		return IgnoreRecursion
	}
	if len(dir.Name) > 5 && strings.HasPrefix(dir.Name, "then-") {
		return nil
	}
	t.Logf("skipping dir recursion: %q does not begin with %q ", dir.Name, "then-")
	return fmt.Errorf("%q does not begin with %q: %w", dir.Name, "then-", SkipRecursion)
}
