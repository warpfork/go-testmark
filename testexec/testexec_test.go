package testexec_test

import (
	"sort"
	"testing"

	"github.com/warpfork/go-testmark"
	"github.com/warpfork/go-testmark/testexec"

	qt "github.com/frankban/quicktest"
)

func TestSelf(t *testing.T) {
	filename := "selfexercise.md"
	doc, err := testmark.ReadFile(filename)
	if err != nil {
		t.Fatalf("spec file parse failed?!: %s", err)
	}

	doc.BuildDirIndex()
	patches := testmark.PatchAccumulator{}
	for _, dir := range doc.DirEnt.ChildrenList {
		t.Run(dir.Name, func(t *testing.T) {
			test := testexec.Tester{
				Patches: &patches,
			}
			test.TestScript(t, dir)
		})
	}
	patches.WriteFileWithPatches(doc, filename)
}

// TestRecursion tests edge cases around which nodes are traversed via recursion
func TestRecursion(t *testing.T) {
	rtested := map[string]int{}
	rcalled := map[string]int{}
	sum := func(m map[string]int) int {
		result := 0
		for _, v := range m {
			result += v
		}
		return result
	}

	rfn := func(t *testing.T, dir testmark.DirEnt) error {
		rtested[dir.Path]++
		result := testexec.RecursionFn_Then(t, dir)
		if result == nil {
			rcalled[dir.Path]++
		}
		return result
	}

	filename := "selfexercise.md"
	doc, err := testmark.ReadFile(filename)
	if err != nil {
		t.Fatalf("spec file parse failed?!: %s", err)
	}

	doc.BuildDirIndex()

	patches := testmark.PatchAccumulator{}
	for _, dir := range doc.DirEnt.ChildrenList {
		rtested[dir.Path]++
		t.Run(dir.Name, func(t *testing.T) {
			rcalled[dir.Path]++
			test := testexec.Tester{
				Patches:     &patches,
				RecursionFn: rfn,
			}
			test.Test(t, dir)
		})
	}
	patches.WriteFileWithPatches(doc, filename)

	pathList := collectPaths(*doc.DirEnt)
	sort.Strings(pathList)
	t.Log("[idx]: [test count]-[recursion count]: [path]")
	for idx, path := range pathList {
		check := rtested[path]
		recurse := rcalled[path]
		t.Logf("%02d: %d-%d: %s", idx, check, recurse, path)
	}

	qt.Assert(t, rtested[""], qt.Equals, 0)
	qt.Assert(t, rcalled[""], qt.Equals, 0)
	qt.Assert(t, rtested["bad"], qt.Equals, 1)
	qt.Assert(t, rcalled["bad"], qt.Equals, 1)
	qt.Assert(t, rtested["bad/script"], qt.Equals, 1)
	qt.Assert(t, rcalled["bad/script"], qt.Equals, 0)
	qt.Assert(t, rtested["bad/then-missing-script"], qt.Equals, 1)
	qt.Assert(t, rcalled["bad/then-missing-script"], qt.Equals, 1)
	qt.Assert(t, rtested["bad/then-missing-script/then-another-thing"], qt.Equals, 0)
	qt.Assert(t, rcalled["bad/then-missing-script/then-another-thing"], qt.Equals, 0)
	qt.Assert(t, rtested["bad/then-missing-script/then-another-thing/script"], qt.Equals, 0)
	qt.Assert(t, rcalled["bad/then-missing-script/then-another-thing/script"], qt.Equals, 0)
	qt.Assert(t, rtested["bad/not-a-then-statement"], qt.Equals, 1)
	qt.Assert(t, rcalled["bad/not-a-then-statement"], qt.Equals, 0)
	qt.Assert(t, rtested["bad/not-a-then-statement/script"], qt.Equals, 0)
	qt.Assert(t, rcalled["bad/not-a-then-statement/script"], qt.Equals, 0)

	qt.Assert(t, rtested["using-stdin"], qt.Equals, 1)
	qt.Assert(t, rcalled["using-stdin"], qt.Equals, 1)
	qt.Assert(t, rtested["using-stdin/input"], qt.Equals, 1)
	qt.Assert(t, rcalled["using-stdin/input"], qt.Equals, 0)
	qt.Assert(t, rtested["using-stdin/output"], qt.Equals, 1)
	qt.Assert(t, rcalled["using-stdin/output"], qt.Equals, 0)
	qt.Assert(t, rtested["using-stdin/script"], qt.Equals, 1)
	qt.Assert(t, rcalled["using-stdin/script"], qt.Equals, 0)

	qt.Assert(t, rtested["whee"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee"], qt.Equals, 1)
	qt.Assert(t, rtested["whee/fs"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/fs"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/fs/a"], qt.Equals, 0)
	qt.Assert(t, rcalled["whee/fs/a"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/output"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/output"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/script"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/script"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-more-files"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-more-files"], qt.Equals, 1)
	qt.Assert(t, rtested["whee/then-more-files/fs"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-more-files/fs"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-more-files/fs/b"], qt.Equals, 0)
	qt.Assert(t, rcalled["whee/then-more-files/fs/b"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-more-files/output"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-more-files/output"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-more-files/script"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-more-files/script"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-touching-files"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-touching-files"], qt.Equals, 1)
	qt.Assert(t, rtested["whee/then-touching-files/output"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-touching-files/output"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-touching-files/script"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-touching-files/script"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-touching-files/then-subtesting-again"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-touching-files/then-subtesting-again"], qt.Equals, 1)
	qt.Assert(t, rtested["whee/then-touching-files/then-subtesting-again/output"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-touching-files/then-subtesting-again/output"], qt.Equals, 0)
	qt.Assert(t, rtested["whee/then-touching-files/then-subtesting-again/script"], qt.Equals, 1)
	qt.Assert(t, rcalled["whee/then-touching-files/then-subtesting-again/script"], qt.Equals, 0)

	qt.Assert(t, sum(rtested), qt.Equals, 22)
	qt.Assert(t, sum(rcalled), qt.Equals, 7)
}

func collectPaths(d testmark.DirEnt) []string {
	result := []string{d.Path}
	for _, c := range d.ChildrenList {
		result = append(result, collectPaths(c)...)
	}
	return result
}
