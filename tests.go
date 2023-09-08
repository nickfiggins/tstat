package tstat

import (
	"strings"
	"time"

	"slices"

	"github.com/nickfiggins/tstat/internal/gotest"
)

// Test is a single test, which may have subtests.
type Test struct {
	Subtests []*Test         // Subtests is a list of subtests for this test.
	actions  []gotest.Action // actions is a list of actions that occurred during the test.
	FullName string          // FullName is the full name of the test, including subtests.
	Name     string          // Name is the name of the test, without the parent test name.
	Package  string          // Package is the package that the test belongs to.

	start, end time.Time
}

func (t *Test) withEvent(event gotest.Event) *Test {
	t.actions = append(t.actions, event.Action)

	switch event.Action { //nolint:exhaustive // no need to handle all actions
	case gotest.Start, gotest.Run:
		if isBefore(t.start, event.Time) {
			t.start = event.Time
		}
	case gotest.Pass, gotest.Fail, gotest.Out:
		if isAfter(t.end, event.Time) {
			t.end = event.Time
		}
	}
	return t
}

// isBefore returns true if t2 is before t1, or if t1 is zero.
func isBefore(t1, t2 time.Time) bool {
	return t1.IsZero() || (!t2.IsZero() && t2.Before(t1))
}

func isAfter(t1, t2 time.Time) bool {
	return t1.IsZero() || (!t2.IsZero() && t2.After(t1))
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
	return slices.Contains(t.actions, gotest.Fail)
}

// Skipped returns true if the test was skipped.
func (t *Test) Skipped() bool {
	return slices.Contains(t.actions, gotest.Skip)
}

// Count returns the total number of tests, including subtests.
func (t *Test) Count() int {
	count := 1
	for _, sub := range t.Subtests {
		count += sub.Count()
	}
	return count
}

// Duration returns the total duration of the test, including subtests.
func (t *Test) Duration() time.Duration {
	return t.end.Sub(t.start)
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
