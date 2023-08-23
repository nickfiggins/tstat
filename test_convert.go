package tstat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/exp/maps"
)

const testDelim = "/"

// convertTests converts a gotest.PackageEvents into a PackageRun.
func convertEvents(pkg *gotest.PackageEvents) (PackageRun, error) {
	tests := getPackageTests(pkg.Events)
	nested, err := nestSubtests(tests)
	if err != nil {
		return PackageRun{}, err
	}

	var start, end time.Time
	var failed bool
	if pkg.Start != nil {
		start = pkg.Start.Time
	}

	if pkg.End != nil {
		end = pkg.End.Time
		failed = pkg.End.Action == gotest.Fail
	}

	return PackageRun{
		pkgName: pkg.Package,
		start:   start, end: end,
		Tests:  nested,
		Seed:   pkg.Seed,
		failed: failed,
	}, nil
}

func getPackageTests(events []gotest.Event) []*Test {
	packageTests := make(map[string]*Test, 0)
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
	return testsByName
}

func newTest(pkg, name string) *Test {
	// add 1 to pull the part after the slash, and conveniently
	// handle the case of no subtests as well
	subStart := strings.LastIndex(name, "/") + 1
	return &Test{
		Subtests: make([]*Test, 0),
		actions:  []gotest.Action{},
		FullName: name,
		Package:  pkg,
		Name:     name[subStart:],
	}
}

// nestSubtests takes a list of tests and nests subtests under their parent.
// It returns a list of root tests.
func nestSubtests(tests []*Test) ([]*Test, error) {
	rootTests := map[string]*Test{}
	for _, to := range tests {
		subs := removeEmpty(strings.Split(to.FullName, testDelim))
		subDepth := len(subs)
		if subDepth == 1 { // root test; no subtests
			out := to
			rootTests[to.FullName] = out
		} else if subDepth > 1 { // at least one subtest
			test, ok := rootTests[subs[0]]
			if !ok {
				return nil, fmt.Errorf("subtest found without corresponding parent: %v", to.FullName)
			}
			test.addSubtests(to)
		}
	}

	return maps.Values(rootTests), nil
}

// removeEmpty removes empty strings from a slice of strings.
func removeEmpty(s []string) []string {
	out := make([]string, 0)
	for _, v := range s {
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

// byTestName sorts tests by their name, which ensures that all parent tests come
// before the subtests.
type byTestName []*Test

func (b byTestName) Len() int      { return len(b) }
func (b byTestName) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byTestName) Less(i, j int) bool {
	return b[i].FullName < b[j].FullName
}
