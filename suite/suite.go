package suite

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"testing"

	"github.com/warpfork/go-fsx"
	"github.com/warpfork/go-testmark"
)

// NewSuiteManager constructs a suite manager with which you can use WorkWith to register
// patterns of filenames and hunknames within them with a TestingPattern,
// and then run the whole suite together.
//
// Note that if you want automatic fixture regen features to work,
// the filesystem you hand in to this constructor must support `fsx.FSSupportingWrite`.
// (In practice, this often means you want to use `github.com/warpfork/go-fsx/osfs.DirFS`
// where you otherwise might've used stdlib `os.DirFS` to construct the filesystem reference.)
func NewSuiteManager(fs fsx.FS) *SuiteManager {
	return &SuiteManager{
		fs:      fs,
		workset: make(map[string]fileContentExpectations),
	}
}

// Using SuiteManager to run tests is not strictly necessary,
// but saves a lot of boilerplate, offers a lot of features and guardrails,
// and makes your test setup more declarative and easy to read:
//
//   - SuiteManager lets you specify groups of files to treat as test data.
//   - SuiteManager lets you associate TestingPattern callbacks with hunk names by globbing.
//   - SuiteManager automatically names your tests based on the filename and hunk names.
//   - SuiteManager automatically rigs up fixture regeneration for you when `-testmark.regen=true`.
//   - SuiteManager will warn you about any hunks that go unused in a file (helps detect typos!).
//   - SuiteManager will warn you about any hunk globs that go unmatched in a file.
//
// In general, using SuiteManager will help make sure your fixture files and test cases
// stay aligned as the code evolves: if you remove tests that used some hunks, you'll get notified;
// if you add tests that expect certain patterns of hunks in some files, but didn't write them yet,
// you'll get notified; etc.
//
// Approximately the only thing SuiteManager can't help you with is if you use whole files for organization,
// and you remove some files (or add some features or tests that would expect whole new files),
// and you're using globbing to register filenames, then SuiteManager can't help you infer what's missing.
type SuiteManager struct {
	fs fsx.FS

	workset map[string]fileContentExpectations
}

type fileContentExpectations struct {
	filename string

	handlers map[HunkGlob]TestingPattern
}

// TestingPattern is the interface that test code should implement to become
// usable by SuiteManager as a handler.
//
// Examples of TestingPattern that are well-known and easy to use immediately
// include `testexec.TestingPattern`
//
// When implementing this interface:
// The Run method will do the majority of the work.
// Other methods on the interface let the SuiteManager know how recursion should work,
// and provide some information that may be used in logging.
//
// Note that it is not necessary to import this package in order to implement this interface.
// Only symbols from testmark's core package and the go standard library are used.
type TestingPattern interface {
	Run(
		t *testing.T, // The standard testing object, for obvious purposes.
		filename string, // The filename the subject data was loaded from.  Typically not needed (data comes in via `subject` and, if appropriate, can flow out through `patchAccum`).
		subject *testmark.DirEnt, // The subject hunk (and enclosing dirent, in case you want to navigate to child hunks).
		reportUse func(string), // Should be called with the full path of any hunk that's consumed by this test.  Used to detect orphaned hunks that went unused by the whole suite.
		reportUnrecog func(string), // If this test code owns all child hunks, it may call this to report one that it doesn't recognize.
		patchAccum *testmark.PatchAccumulator, // If non-nil, means regenerating golden master data is requested instead of testing.
	) error // Run may return errors or call t.Fatal itself.

	Name() string // Purely for diagnostic purposes.

	OwnsAllChildren() bool // If true, the suite manager will not look at any hunks beneath this subject or attempt to match other globs and testing patterns to them.
}

type FilenameGlob string
type HunkGlob string

// WorkWith registers one or more files to be handled with some TestingPattern.
// A globbing pattern can be used for both the filenames, and for what hunk labels
// within the file should be handled.
//
// WorkWith can be called more than one time with the name filename, or with filename
// globs that cause the same files to be matched more than once.
// In this case, the mapping of filenames and hunk globs to TestingPattern simply
// continue to accumulate.
//
// Both filename and hunk label globbing are per `path.Match`.
// Note that using a "glob" that's a literal, with no actual pattern matching markers, is acceptable.
//
// The filename glob is matched against the SuiteManager's filesystem immediately.
// If it matches no files, an error is returned.
// (Use MustWorkWith to panic instead.)
//
// If the hunk name glob doesn't compile as a globbing pattern, an error is returned.
// (Use MustWorkWith to panic instead.)
//
// If many testing patterns are registered, and the exact same hunk glob is associated with
// the same filename, the last call of WorkWith overrides the earlier one for that pattern.
// However, if distinct hunk globs later simply happen to match the same hunk label,
// then both will be called, and the order in which they are called is unspecified.
func (sm *SuiteManager) WorkWith(files FilenameGlob, hunks HunkGlob, action TestingPattern) error {
	_, err := path.Match(string(hunks), "")
	if err != nil {
		return fmt.Errorf("hunk label glob does not compile: %w", err)
	}
	matches, err := fs.Glob(sm.fs, string(files))
	if err != nil {
		return fmt.Errorf("filename glob does not compile: %w", err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("filename glob %q matched no files", files)
	}
	for _, match := range matches {
		if _, exists := sm.workset[match]; !exists {
			sm.workset[match] = fileContentExpectations{
				filename: match,
				handlers: make(map[HunkGlob]TestingPattern),
			}
		}
		sm.workset[match].handlers[hunks] = action
	}
	return nil
}

// MustWorkWith is exactly as per WorkWith, but panics in case of errors.
func (sm *SuiteManager) MustWorkWith(files FilenameGlob, hunks HunkGlob, action TestingPattern) {
	if err := sm.WorkWith(files, hunks, action); err != nil {
		panic(err)
	}
}

// IgnoreUnrecognized tells the suite manager that if a hunk is reported as unrecognized or unused,
// then it should be ignored instead of becoming an error.
//
// A typical value to see this called with might be `"*/comment"`, for example.
//
// Such a setup would allow you to put comment nodes even deep inside hunk trees that are used
// by other test patterns (such as e.g. `testexec.TestPattern`) which would otherwise report them as an error.
func (sm *SuiteManager) IgnoreUnrecognized(files FilenameGlob, pattern HunkGlob) {
	panic("not yet implemented") // TODO
}

// Run launches the test suite.
// WorkWith should have been called to populate the suite before this.
//
// Subtests will be created with `t.Run` for each file,
// and another nested subtest for each hunk that matches and causes
// a TestingPattern to be invoked.
// (The TestingPattern gets full control of `t` thereafter, and may
// create even further additional subtests.)
//
// By default, each file that the suite works with will be handled in parallel,
// and within each file, the hunks are handled in the order they appear.
//
// Calling Run more than one time is nonsensical.
func (sm *SuiteManager) Run(t *testing.T) {
	for filename, fileContentExpectations := range sm.workset {
		filename := filename
		fileContentExpectations := fileContentExpectations
		t.Run(filename, func(t *testing.T) {
			t.Parallel()

			// Open the file.
			f, err := sm.fs.Open(filename)
			if err != nil {
				t.Fatalf("could not load testmark file: %s", err)
			}

			// Parse the file.
			tmDoc, err := testmark.Read(f)
			f.Close()
			if err != nil {
				t.Fatalf("could not parse testmark file %q: %s", filename, err)
			}
			tmDoc.BuildDirIndex()

			// Prepare to write back patches, if appropriate.
			var patchAccum *testmark.PatchAccumulator
			if *testmark.Regen {
				patchAccum = &testmark.PatchAccumulator{}
				defer func() {
					f, err := fsx.OpenFile(sm.fs, filename, fsx.O_TRUNC|fsx.O_WRONLY, 0777)
					if err != nil {
						t.Fatalf("could not open file to write regenerated fixture: %s", err)
					}
					patchAccum.WriteWithPatches(tmDoc, f.(io.Writer))
					f.Close()
				}()
			}

			// Before we begin walking, prepare to remember which things are usedHunks... or explicitly flagged as unknown.
			usedHunks := map[string]struct{}{}
			unrecognizedHunks := map[string]struct{}{}
			usedGlobs := map[HunkGlob]struct{}{}
			reportUse := func(hunkName string) { usedHunks[hunkName] = struct{}{} }
			reportUnrecog := func(hunkName string) { unrecognizedHunks[hunkName] = struct{}{} }

			// Range over the hunks, treating labels as if they're filesystem paths (e.g., "/" groups them).
			// For every match, create a subtest with the hunk path as a name, and call the test pattern.
			for _, ent := range tmDoc.DirEnt.ChildrenList {
				// Check every handler associated with this filename if it matches this hunk.
				for hunkGlob, action := range fileContentExpectations.handlers {
					if match, _ := path.Match(string(hunkGlob), ent.Path); match {
						usedGlobs[hunkGlob] = struct{}{}
						t.Run(ent.Path, func(t *testing.T) {
							err := action.Run(t, filename, ent, reportUse, reportUnrecog, patchAccum)
							if err != nil {
								t.Fatalf("error while running the %s testing pattern on hunk %q in file %q: %s", action.Name(), ent.Path, filename, err)
							}
						})
					}
				}
			}

			// Make another subtest to report any unused or explicitly unrecognized hunks.
			// Only do this if there are any of them; otherwise, don't bother cluttering up the reports.
			if len(tmDoc.HunksByName) == 0 ||
				len(usedHunks) < len(tmDoc.HunksByName) ||
				len(unrecognizedHunks) > 0 ||
				len(usedGlobs) < len(fileContentExpectations.handlers) {
				t.Run("dangling references", func(t *testing.T) {
					if len(tmDoc.HunksByName) == 0 {
						t.Errorf("file %q contained no testmark hunks at all and caused no tests to be exercised in this suite", filename)
					}
					for hunkName, _ := range tmDoc.HunksByName {
						if _, exists := usedHunks[string(hunkName)]; !exists {
							t.Errorf("hunk label %q in file %q was not used by any tests in this suite", hunkName, filename)
						}
					}
					for _, hunkName := range unrecognizedHunks {
						t.Errorf("hunk label %q in file %q was flagged as unrecognized by one of the tests in this suite", hunkName, filename)
					}
					for hunkGlob, _ := range fileContentExpectations.handlers {
						if _, exists := usedGlobs[hunkGlob]; !exists {
							t.Errorf("the glob %q matched zero hunk labels in file %q and caused no tests to be exercised in this suite", hunkGlob, filename)
						}
					}
				})
			}
		})

	}
}
