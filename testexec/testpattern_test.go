package testexec_test

import (
	"testing"

	"github.com/warpfork/go-fsx/osfs"
	"github.com/warpfork/go-testmark/suite"
	"github.com/warpfork/go-testmark/testexec"
)

// Assert that our testingPattern type matches the suite.TestingPattern interface.
// By doing this in a "_test.go" file, we avoid importing the suite package in the testexec package.
var _ suite.TestingPattern = testexec.SuiteTester{}

func TestSuiteMode(t *testing.T) {
	t.Run("selfexericse file", func(t *testing.T) {
		sm := suite.NewSuiteManager(osfs.DirFS("."))
		sm.MustWorkWith("selfexercise.md", "*", testexec.NewSuiteTester(testexec.Tester{}))
		sm.Run(t)
	})
	t.Run("strictexericse file", func(t *testing.T) {
		if !(*RunFailTest) {
			t.Skipf("%s requires %q flag to execute", t.Name(), "run-fail-test")
		}
		sm := suite.NewSuiteManager(osfs.DirFS("."))
		sm.MustWorkWith("strictexercise.md", "*", testexec.NewSuiteTester(testexec.Tester{}))
		sm.Run(t)
	})
}
