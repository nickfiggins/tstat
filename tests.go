package tstat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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

func (t Test) withEvent(event gotest.Event) Test {
	t.actions = append(t.actions, event.Action)

	switch event.Action { //nolint:exhaustive // no need to handle all actions
	case gotest.Start:
		t.start = event.Time
	case gotest.Out:
		if event.Elapsed != 0 {
			t.end = event.Time
		}
	}
	return t
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

// does the test name look like a sub test of the current test?
func (t *Test) looksLikeSub(subName string) bool {
	return strings.HasPrefix(subName+"/", t.FullName)
}

func (t *Test) addSubtests(sub Test) {
	trimmed := strings.TrimPrefix(sub.FullName, t.FullName+"/")
	remainingSubs := strings.Split(trimmed, "/")
	if len(remainingSubs) == 1 {
		t.Subtests = append(t.Subtests, &sub)
		return
	}

	for _, subtest := range t.Subtests {
		if t.looksLikeSub(subtest.FullName) {
			subtest.addSubtests(sub)
		}
	}
}

func newTest(pkg, name string) Test {
	// add 1 to pull the part after the slash, and conveniently
	// handle the case of no subtests as well
	subStart := strings.LastIndex(name, "/") + 1
	return Test{
		Subtests: make([]*Test, 0),
		actions:  []gotest.Action{},
		FullName: name,
		Package:  pkg,
		Name:     name[subStart:],
	}
}

func parseTestOutputs(pkgs []*gotest.PackageEvents) (TestRun, error) {
	suite := TestRun{}
	for _, pkg := range pkgs {
		run, err := parsePackageEvents(pkg)
		if err != nil {
			return TestRun{}, err
		}
		if suite.start.IsZero() || run.start.Before(suite.start) {
			suite.start = run.start
		}

		if suite.end.IsZero() || run.end.After(suite.end) {
			suite.end = run.end
		}

		suite.pkgs = append(suite.pkgs, run)
	}
	return suite, nil
}

func parsePackageEvents(events *gotest.PackageEvents) (PackageRun, error) {
	packageTests, err := getPackageTests(events.Events)
	if err != nil {
		return PackageRun{}, err
	}

	var start, end time.Time
	var failed bool
	if events.Start != nil {
		start = events.Start.Time
	}

	if events.End != nil {
		end = events.End.Time
		failed = events.End.Action == gotest.Fail
	}

	return PackageRun{
		pkgName: events.Package,
		start:   start, end: end,
		Tests:  packageTests,
		Seed:   events.Seed,
		failed: failed,
	}, nil
}

func getPackageTests(events []gotest.Event) ([]*Test, error) {
	packageTests := make(map[string]Test, 0)
	for _, event := range events {
		if event.Test == "" {
			continue
		}

		test, ok := packageTests[event.Test]
		if !ok {
			test = newTest(event.Package, event.Test)
		}

		packageTests[event.Test] = test.withEvent(event)
	}

	testsByName := maps.Values(packageTests)
	sort.Sort(byTestName(testsByName))

	rootTests, err := nestSubtests(testsByName)
	if err != nil {
		return nil, err
	}
	return rootTests, nil
}

func clean(s []string) []string {
	out := make([]string, 0)
	for _, v := range s {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func nestSubtests(tests []Test) ([]*Test, error) {
	rootTests := map[string]*Test{}
	for _, to := range tests {
		subs := clean(strings.Split(to.FullName, "/"))
		switch len(subs) {
		case 0:
		case 1:
			out := to
			rootTests[to.FullName] = &out
		default:
			test, ok := rootTests[subs[0]]
			if !ok && len(subs) > 1 {
				return nil, fmt.Errorf("subtest found without corresponding parent: %v", to.FullName)
			}
			test.addSubtests(to)
		}
	}

	return maps.Values(rootTests), nil
}

// byTestName sorts tests by their name, which ensures that all parent tests come
// before the subtests.
type byTestName []Test

func (b byTestName) Len() int      { return len(b) }
func (b byTestName) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byTestName) Less(i, j int) bool {
	return b[i].FullName < b[j].FullName
}
