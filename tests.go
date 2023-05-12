package tstat

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
)

type TestRun struct {
	start, end time.Time
	Tests      []*Test
	Passed     bool
	cmdOut     string
}

func (ts *TestRun) Duration() time.Duration {
	return ts.end.Sub(ts.start)
}

func (ts *TestRun) Count() int {
	count := 0
	for _, test := range ts.Tests {
		count += test.count()
	}
	return count
}

type Test struct {
	Subtests []Test
	Action   string `json:"Action"`
	Name     string
	SubName  string
	Package  string `json:"Package"`
}

func (t *Test) count() int {
	count := 0
	for _, sub := range t.Subtests {
		count += sub.count()
	}
	return count
}

func (t *Test) addSubtests(sub Test) {
	subs := strings.Split(sub.Name, "/")
	nestedCount := len(subs)
	if nestedCount == 2 {
		t.Subtests = append(t.Subtests, sub)
		return
	}

	for _, subtest := range t.Subtests {
		if strings.Contains(sub.Name, subtest.Name) {
			subtest.addSubtests(sub)
		}
	}
}

func toTest(to gotest.Output) Test {
	nameParts := strings.Split(to.Test, "/")
	return Test{
		Subtests: make([]Test, 0),
		Action:   to.Action,
		Name:     to.Test,
		Package:  to.Package,
		SubName:  nameParts[len(nameParts)-1],
	}
}

func parseTestOutput(outputs []gotest.Output) (TestRun, error) {
	output := byTestName(outputs)
	start, end := start(output), output[len(output)-1]
	sort.Sort(byTestName(outputs))

	var tests []Test
	for _, out := range outputs {
		if out.Test != "" && gotest.IsFinal(out.Action) {
			tests = append(tests, toTest(out))
		}
	}

	rootTests, err := nestSubtests(tests)
	if err != nil {
		return TestRun{}, err
	}

	return TestRun{
		start: start.Time, end: end.Time,
		Tests:  rootTests,
		Passed: len(withStatus(rootTests, gotest.Fail)) == 0,
		cmdOut: consoleOutputs(outputs),
	}, nil
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

	var testSlice []*Test
	for _, t := range rootTests {
		testSlice = append(testSlice, t)
	}

	return testSlice, nil
}

func withStatus(tests []*Test, s gotest.Action) []*Test {
	sl := make([]*Test, 0)
	strStatus := s.String()
	for _, test := range tests {
		if test.Action == strStatus {
			sl = append(sl, test)
		}
	}
	return sl
}

func consoleOutputs(outputs []gotest.Output) string {
	sb := strings.Builder{}
	for _, o := range outputs {
		if o.Action == "output" {
			sb.WriteString(o.Output)
		}
	}

	return sb.String()
}

// byTestName sorts tests by their name, which ensures that all parent tests come
// before the subtests.
type byTestName []gotest.Output

func start(b []gotest.Output) gotest.Output {
	for _, out := range b {
		if out.Action == "start" {
			return out
		}
	}
	return gotest.Output{}
}

func (b byTestName) Len() int      { return len(b) }
func (b byTestName) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byTestName) Less(i, j int) bool {
	return b[i].Test < b[j].Test
}
