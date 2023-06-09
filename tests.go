package tstat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/exp/maps"
)

type Test struct {
	Subtests []*Test
	actions  []gotest.Action
	Name     string
	SubName  string
	Package  string
}

func (t *Test) Test(name string) (*Test, bool) {
	if name == t.Name {
		return t, true
	}

	return findTest(name, t.Subtests...)
}

func (t *Test) Failed() bool {
	for _, act := range t.actions {
		if act == gotest.Fail {
			return true
		}
	}
	return false
}

func (t *Test) Skipped() bool {
	for _, act := range t.actions {
		if act == gotest.Skip {
			return true
		}
	}
	return false
}

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

func parseTestOutputs(events []gotest.Event) (TestRun, error) {
	pkgs := gotest.ByPackage(events)
	suite := TestRun{root: root(events)}
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

func root(events []gotest.Event) string {
	if len(events) == 0 {
		return ""
	}
	for _, e := range events {
		if e.Package != "" {
			return e.Package
		}
	}
	return ""
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

func nestSubtests(tests []Test) ([]*Test, error) {
	rootTests := map[string]*Test{}
	for _, to := range tests {
		subs := strings.Split(to.Name, "/")
		if len(subs) == 1 {
			out := to
			rootTests[to.Name] = &out
			continue
		}
		test, ok := rootTests[subs[0]]
		if !ok && len(subs) > 1 {
			return nil, fmt.Errorf("subtest found without corresponding parent: %v", to.Name)
		}
		test.addSubtests(to)
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
