package gotest

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	Time    time.Time `json:"Time"`
	Action  Action    `json:"Action"`
	Output  string    `json:"Output"`
	Test    string    `json:"Test"`
	Package string    `json:"Package"`
	Elapsed float64   `json:"Elapsed"` // Elapsed is the number of seconds that have passed.
}

func (e *Event) Seed() (int64, bool) {
	flag := "-test.shuffle"
	idx := strings.Index(e.Output, flag)
	if idx == -1 {
		return 0, false
	}
	num := strings.Trim(e.Output[idx+len(flag):], " \n")
	n, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// PackageEvent returns true if the event is a package event (no test referenced in event).
func (e *Event) PackageEvent() bool {
	return e.Test == "" && e.Package != ""
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

func (a *Action) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	act := ToAction(s)
	*a = act
	return nil
}

func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
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
