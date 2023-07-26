package tstat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/exp/maps"
)

// Test is a single test, which may have subtests.
type Test struct {
	Subtests []*Test         // Subtests is a list of subtests for this test.
	actions  []gotest.Action // actions is a list of actions that occurred during the test.
	Name     string          // Name is the full name of the test, including subtests.
	SubName  string          // SubName is the name of the test, without the parent test name.
	Package  string          // Package is the package that the test belongs to.
}

// Test returns the test with the given name. If the test name matches the current test, it will be returned.
func (t *Test) Test(name string) (*Test, bool) {
	if name == t.Name {
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

// does the test name look like a sub test of the current test?
func (t *Test) looksLikeSub(subName string) bool {
	return strings.HasPrefix(subName+"/", t.Name)
}

func (t *Test) addSubtests(sub Test) {
	trimmed := strings.TrimPrefix(sub.Name, t.Name+"/")
	remainingSubs := strings.Split(trimmed, "/")
	if len(remainingSubs) == 1 {
		t.Subtests = append(t.Subtests, &sub)
		return
	}

	for _, subtest := range t.Subtests {
		if t.looksLikeSub(subtest.Name) {
			subtest.addSubtests(sub)
		}
	}
}

func toTest(to gotest.Event) Test {
	// add 1 to pull the part after the slash, and conveniently
	// handle the case of no subtests as well
	subStart := strings.LastIndex(to.Test, "/") + 1
	return Test{
		Subtests: make([]*Test, 0),
		actions:  []gotest.Action{to.Action},
		Name:     to.Test,
		Package:  to.Package,
		SubName:  to.Test[subStart:],
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
	testsByName := getPackageTests(events.Events)
	rootTests, err := nestSubtests(testsByName)
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
		Tests:  rootTests,
		Seed:   events.Seed,
		failed: failed,
	}, nil
}

func getPackageTests(events []gotest.Event) []Test {
	packageTests := make(map[string]Test, 0)
	for _, out := range events {
		if out.Test == "" {
			continue
		}

		test, ok := packageTests[out.Test]
		if !ok {
			t := toTest(out)
			packageTests[out.Test] = t
			continue
		}

		test.actions = append(test.actions, out.Action)
		packageTests[out.Test] = test
	}

	testsByName := maps.Values(packageTests)
	sort.Sort(byTestName(testsByName))
	return testsByName
}

func clean(s []string) []string {
	out := make([]string, 0)
	for _, v := range s {
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func nestSubtests(tests []Test) ([]*Test, error) {
	rootTests := map[string]*Test{}
	for _, to := range tests {
		subs := clean(strings.Split(to.Name, "/"))
		switch len(subs) {
		case 0:
		case 1:
			out := to
			rootTests[to.Name] = &out
		default:
			test, ok := rootTests[subs[0]]
			if !ok && len(subs) > 1 {
				return nil, fmt.Errorf("subtest found without corresponding parent: %v", to.Name)
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
	return b[i].Name < b[j].Name
}
