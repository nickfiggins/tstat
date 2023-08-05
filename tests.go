package tstat

import (
	"strings"

	"github.com/nickfiggins/tstat/internal/gotest"
)

// Test is a single test, which may have subtests.
type Test struct {
	Subtests []*Test         // Subtests is a list of subtests for this test.
	actions  []gotest.Action // actions is a list of actions that occurred during the test.
	FullName string          // FullName is the full name of the test, including subtests.
	Name     string          // Name is the name of the test, without the parent test name.
	Package  string          // Package is the package that the test belongs to.
}

// Test returns the test with the given name. If the test name matches the current test, it will be returned.
func (t *Test) Test(name string) (*Test, bool) {
	if name == t.FullName {
		return t, true
	}

	return findTest(name, t.Subtests...)
}

// Failed returns true if the test failed.
func (t *Test) Failed() bool {
	for _, act := range t.actions {
		if act == gotest.Fail {
			return true
		}
	}
	return false
}

// Skipped returns true if the test was skipped.
func (t *Test) Skipped() bool {
	for _, act := range t.actions {
		if act == gotest.Skip {
			return true
		}
	}
	return false
}

// Count returns the total number of tests, including subtests.
func (t *Test) Count() int {
	count := 1
	for _, sub := range t.Subtests {
		count += sub.Count()
	}
	return count
}

// does the subName look like a sub test of the current test?
func (t *Test) looksLikeSub(subName string) bool {
	return strings.HasPrefix(subName, t.FullName+"/")
}

func (t *Test) addSubtests(sub *Test) {
	trimmed := strings.TrimPrefix(sub.FullName, t.FullName+"/")
	remainingSubs := strings.Split(trimmed, "/")
	if len(remainingSubs) == 1 {
		t.Subtests = append(t.Subtests, sub)
		return
	}

	for _, subtest := range t.Subtests {
		if subtest.looksLikeSub(sub.FullName) {
			subtest.addSubtests(sub)
			return
		}
	}

	t.Subtests = append(t.Subtests, sub)
}
