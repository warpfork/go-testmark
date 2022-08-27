package testexec_test

import (
	"strings"
	"testing"

	"github.com/warpfork/go-testmark"
	"github.com/warpfork/go-testmark/testexec"
)

func Test(t *testing.T) {
	filename := "selfexercise.md"
	doc, err := testmark.ReadFile(filename)
	if err != nil {
		t.Fatalf("spec file parse failed?!: %s", err)
	}

	doc.BuildDirIndex()
	patches := testmark.PatchAccumulator{}
	for _, dir := range doc.DirEnt.ChildrenList {
		t.Run(dir.Name, func(t *testing.T) {
			t.Logf("testmark describe\n%s", strings.Join(doc.Describe(&dir), "\n"))
			test := testexec.Tester{
				Patches: &patches,
			}
			test.TestScript(t, dir)
			t.Fatalf("forced failure")
		})
	}
	patches.WriteFileWithPatches(doc, filename)
}
