package testexec

import (
	"testing"

	"github.com/warpfork/go-testmark"
)

func NewSuiteTester(tcfg Tester) SuiteTester {
	return SuiteTester{tcfg}
}

type SuiteTester struct {
	tcfg Tester
}

func (st SuiteTester) Name() string          { return "testexec" }
func (st SuiteTester) OwnsAllChildren() bool { return true }
func (st SuiteTester) Run(
	t *testing.T,
	filename string,
	subject *testmark.DirEnt,
	reportUse func(string),
	reportUnrecog func(string, string),
	patchAccum *testmark.PatchAccumulator,
) error {
	st.tcfg.reportUse = reportUse
	st.tcfg.reportUnrecog = reportUnrecog
	st.tcfg.Patches = patchAccum
	st.tcfg.Test(t, subject)
	return nil
}
