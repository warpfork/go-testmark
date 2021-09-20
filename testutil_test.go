package testmark_test

import (
	"fmt"
	"testing"
)

/*
	This file is for the quickest, dumbest, most essential test helpers.

	I'm refraining from using a full-blown testing library,
	because (at the time of writing) the golang module system
	does not differentiate test deps from runtime deps,
	and if this library is going to be easy to use widely,
	I'd like for its transitive dependency tree to not foist
	my personal preference of testing library's onto other people's module graphs.
*/

// assert is a quick and dirty test helper.
// It stringifies anything given and uses string equality.
// You can probably give it strings or bytes and it'll probably "DTRT";
// anything else relies on "%v".
// It'll emit both the expected and actual values as strings if there's a mismatch.
func assert(t *testing.T, actual interface{}, expect string) {
	var actualStr string
	if s, ok := actual.(string); ok {
		actualStr = s
	} else if bs, ok := actual.([]byte); ok {
		actualStr = string(bs)
	} else {
		actualStr = fmt.Sprintf("%v", actual)
	}
	if actualStr != expect {
		t.Errorf("expected: %q;\nactual: %q", expect, actualStr)
	}
}
