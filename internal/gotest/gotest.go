package gotest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"
)

type Event struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Output  string    `json:"Output"`
	Test    string    `json:"Test"`
	Package string    `json:"Package"`
	Elapsed float64   `json:"Elapsed"` // Elapsed is the number of seconds that have passed.
}

// see https://pkg.go.dev/cmd/test2json#hdr-Output_Format
type Action int

const (
	Start     Action = 0
	Pass      Action = 1
	Fail      Action = 2
	Skip      Action = 3
	Out       Action = 4
	Run       Action = 5
	Undefined Action = -1
)

func ToAction(s string) Action {
	toAction := map[string]Action{
		"start": Start, "pass": Pass, "fail": Fail, "skip": Skip,
		"output": Out, "run": Run, "undefined": Undefined,
	}
	a, ok := toAction[strings.ToLower(s)]
	if !ok {
		return Undefined
	}

	return a
}

func (a Action) String() string {
	toStr := map[Action]string{
		Start: "start", Pass: "pass", Fail: "fail", Skip: "skip",
		Out: "output", Run: "run", Undefined: "undefined",
	}
	s, ok := toStr[a]
	if !ok {
		return toStr[Undefined]
	}
	return s
}

func (a Action) IsFinal() bool {
	switch a {
	case Pass, Fail, Skip:
		return true
	case Start, Run, Out, Undefined:
	default:
	}
	return false
}

func ReadJSON(r io.Reader) ([]Event, error) {
	sc := bufio.NewScanner(r)
	var lines []Event
	for sc.Scan() {
		var line Event
		s := strings.TrimFunc(sc.Text(), unicode.IsSpace)
		if len(s) == 0 {
			continue
		}
		err := json.Unmarshal([]byte(s), &line)
		if err != nil {
			return nil, fmt.Errorf("couldn't unmarshal json: %w, bytes: %v", err, s)
		}
		lines = append(lines, line)
	}
	err := sc.Err()
	if err != nil {
		return nil, fmt.Errorf("error while scanning: %w", err)
	}
	return lines, nil
}
