package testexec_test

import (
	"flag"
	"testing"

	"github.com/warpfork/go-testmark"
	"github.com/warpfork/go-testmark/testexec"
)

var RunFailTest = flag.Bool("run-fail-test", false, "Executes the tests which are expected to fail")

func TestSelfExercise(t *testing.T) {
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

func TestInvalid(t *testing.T) {
	if !(*RunFailTest) {
		t.Skipf("%s requires %q flag to execute", t.Name(), "run-fail-test")
	}
	filename := "invalidexercise.md"
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

func TestStrict(t *testing.T) {
	if !(*RunFailTest) {
		t.Skipf("%s requires %q flag to execute", t.Name(), "run-fail-test")
	}
	filename := "strictexercise.md"
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

func TestStrictDisabled(t *testing.T) {
	filename := "strictexercise.md"
	doc, err := testmark.ReadFile(filename)
	if err != nil {
		t.Fatalf("spec file parse failed?!: %s", err)
	}

	doc.BuildDirIndex()
	patches := testmark.PatchAccumulator{}
	for _, dir := range doc.DirEnt.ChildrenList {
		t.Run(dir.Name, func(t *testing.T) {
			test := testexec.Tester{
				Patches:           &patches,
				DisableStrictMode: true,
			}
			test.TestScript(t, dir)
		})
	}
	patches.WriteFileWithPatches(doc, filename)
}
