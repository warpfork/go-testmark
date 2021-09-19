package testexec

import (
	"io"
	"os/exec"
	"testing"
)

func ExecFn_Exec(args []string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int, oshit error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	// FIXME: get the error code.
	return -1, err
}

func ScriptFn_ExecBash(script string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int, oshit error) {
	return ExecFn_Exec([]string{"bash", "-c", script}, stdin, stdout, stderr)
}

func defaultAssertFn(t *testing.T, actual, expect string) {
	if actual != expect {
		t.Errorf("expected: %q; actual: %q", expect, actual)
	}
}
