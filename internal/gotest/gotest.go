package gotest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Action int

const (
	Start Action = iota
	Pass
	Fail
	Skip
	Out
)

func (a Action) String() string {
	switch a {
	case Pass:
		return "pass"
	case Fail:
		return "fail"
	case Skip:
		return "skip"
	case Start:
		return "start"
	case Out:
		return "output"
	}
	return "invalid status"
}

func IsFinal(status string) bool {
	return status == Pass.String() || status == Fail.String() || status == Skip.String()
}

type Output struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Output  string    `json:"Output"`
	Test    string    `json:"Test"`
	Package string    `json:"Package"`
}

func ReadJSON(r io.Reader) ([]Output, error) {
	sc := bufio.NewScanner(r)
	var lines []Output
	for sc.Scan() {
		var line Output
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
