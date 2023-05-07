package tstat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type TestOutput struct {
	Action  string `json:"Action"`
	Output  string `json:"Output"`
	Test    string `json:"Test"`
	Package string `json:"Package"`
}

type TestStats struct {
	Tests  map[string]TestOutput
	passed bool
	cmdOut string
}

func statusFilter(statuses ...string) func(out TestOutput) bool {
	return func(out TestOutput) bool {
		for _, s := range statuses {
			if out.Action == s {
				return true
			}
		}
		return false
	}
}

func ParseTestOutput(jsonOut io.Reader) (*TestStats, error) {
	outputs, err := readByLine(jsonOut)
	if err != nil {
		return nil, err
	}

	tests := byTestName(outputs, statusFilter("pass", "fail"))
	failed := byTestName(outputs, statusFilter("fail"))

	return &TestStats{Tests: tests, passed: len(failed) == 0, cmdOut: consoleOutputs(outputs)}, nil
}

func readByLine(r io.Reader) ([]TestOutput, error) {
	sc := bufio.NewScanner(r)
	var lines []TestOutput
	for sc.Scan() {
		var line TestOutput
		b := sc.Bytes()
		err := json.Unmarshal(b, &line)
		if err != nil {
			return nil, fmt.Errorf("couldn't unmarshal json: %w, bytes: %v", err, string(b))
		}
		lines = append(lines, line)
	}
	err := sc.Err()
	if err != nil {
		return nil, fmt.Errorf("error while scanning: %w", err)
	}
	return lines, nil
}

func consoleOutputs(outputs []TestOutput) string {
	sb := strings.Builder{}
	for _, o := range outputs {
		if o.Action == "output" {
			sb.WriteString(o.Output)
		}
	}

	return sb.String()
}

// byTestName returns a map of TestOutput structs by the test name, only
// including those which the filter func returns true.
func byTestName(outputs []TestOutput, filter func(out TestOutput) bool) map[string]TestOutput {
	tests := map[string]TestOutput{}
	for _, o := range outputs {
		if o.Test != "" && filter(o) {
			tests[o.Test] = o
		}
	}
	return tests
}
