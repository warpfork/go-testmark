package testexec_test

import (
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
			test := testexec.Tester{
				Patches: &patches,
			}
			test.TestScript(t, *dir)
		})
	}
	patches.WriteFileWithPatches(doc, filename)
}
