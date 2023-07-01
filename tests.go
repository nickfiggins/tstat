package tstat

import (
	"fmt"
	"sort"
	"strconv"
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

func (t *Test) Test(name string) *Test {
	if name == t.Name {
		return t
	}

	sub, ok := findTest(name, t.Subtests...)
	if !ok {
		return &Test{}
	}
	return sub
}

func (t *Test) Failed() bool {
	for _, act := range t.actions {
		if act == gotest.Fail {
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
		actions:  []gotest.Action{gotest.ToAction(to.Action)},
		Name:     to.Test,
		Package:  to.Package,
		SubName:  to.Test[subStart:],
	}
}

func byPackage(outputs []gotest.Event) map[string][]gotest.Event {
	pkgs := make(map[string][]gotest.Event)
	for _, out := range outputs {
		if out.Package == "" {
			continue
		}
		pkg, ok := pkgs[out.Package]
		if !ok {
			pkgs[out.Package] = append(pkgs[out.Package], out)
			continue
		}
		pkg = append(pkg, out)
		pkgs[out.Package] = pkg
	}

	return pkgs
}

func parseTestOutputs(outputs []gotest.Event) (TestRun, error) {
	pkgTests := byPackage(outputs)
	suite := TestRun{}
	for name, tests := range pkgTests {
		run, err := parsePackageTests(tests)
		if err != nil {
			return TestRun{}, err
		}
		if suite.start.IsZero() || run.start.Before(suite.start) {
			suite.start = run.start
		}

		if suite.end.IsZero() || run.end.After(suite.end) {
			suite.end = run.end
		}
		run.pkgName = name
		suite.pkgs = append(suite.pkgs, run)
	}
	return suite, nil
}

func getSeed(out gotest.Event) (int64, bool) {
	flag := "-test.shuffle"
	idx := strings.Index(out.Output, flag)
	if idx == -1 {
		return 0, false
	}
	num := strings.Trim(out.Output[idx+len(flag):], " \n")
	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func parsePackageTests(outputs []gotest.Event) (PackageRun, error) {
	tmap := make(map[string]Test, 0)
	var start, end time.Time
	var seed int64
	for _, out := range outputs {
		if isPackageEvent(out) {
			switch {
			case isStart(out):
				start = out.Time
			case isEnd(out):
				end = out.Time
			case gotest.ToAction(out.Action) == gotest.Out:
				if s, ok := getSeed(out); ok {
					seed = s
				}
			}
			continue
		}

		if out.Test == "" {
			continue
		}

		test, ok := tmap[out.Test]
		if !ok {
			t := toTest(out)
			tmap[out.Test] = t
			continue
		}

		test.actions = append(test.actions, gotest.ToAction(out.Action))
		tmap[out.Test] = test
	}

	tests := maps.Values(tmap)
	sort.Sort(byTestName(tests))

	rootTests, err := nestSubtests(tests)
	if err != nil {
		return PackageRun{}, err
	}

	return PackageRun{
		start: start, end: end,
		Tests: rootTests,
		Seed:  seed,
	}, nil
}

// isPackageEvent returns true if the event is a package event (no test referenced in event).
func isPackageEvent(out gotest.Event) bool {
	return out.Test == "" && out.Package != ""
}

func isEnd(out gotest.Event) bool {
	return out.Test == "" && out.Elapsed != 0
}

func isStart(out gotest.Event) bool {
	return gotest.ToAction(out.Action) == gotest.Start
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

func withAction(tests []*Test, action gotest.Action) []*Test {
	sl := make([]*Test, 0)
	for _, test := range tests {
		for _, act := range test.actions {
			if act == action {
				sl = append(sl, test)
				break
			}
		}
	}
	return sl
}

// byTestName sorts tests by their name, which ensures that all parent tests come
// before the subtests.
type byTestName []Test

func (b byTestName) Len() int      { return len(b) }
func (b byTestName) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byTestName) Less(i, j int) bool {
	return b[i].Name < b[j].Name
}
