package gotest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func readJSON(r io.Reader) ([]Event, error) {
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

func ReadByPackage(r io.Reader) ([]*PackageEvents, error) {
	events, err := readJSON(r)
	if err != nil {
		return nil, err
	}
	return ByPackage(events), nil
}

func ByPackage(events []Event) []*PackageEvents {
	packages := make(map[string]*PackageEvents)
	for _, e := range events {
		if e.Package == "" {
			continue
		}
		pkg, ok := packages[e.Package]
		if !ok {
			packages[e.Package] = newFromEvent(e)
			continue
		}
		pkg.Add(e)
	}

	vals := make([]*PackageEvents, 0, len(packages))
	for _, v := range packages {
		vals = append(vals, v)
	}

	return vals
}

type PackageEvents struct {
	Package    string
	Start, End *Event
	Seed       int64
	Events     []Event
}

func (pe *PackageEvents) Add(e Event) {
	if pe.Start == nil && e.Action == Start {
		pe.Start = &e
	}

	if pe.End == nil && e.Elapsed != 0 && e.Test == "" {
		pe.End = &e
	}

	if pe.Seed == 0 {
		s, ok := e.Seed()
		if ok {
			pe.Seed = s
		}
	}

	pe.Events = append(pe.Events, e)
}

func newFromEvent(e Event) *PackageEvents {
	pe := &PackageEvents{
		Package: e.Package,
		Start:   nil, End: nil,
		Events: []Event{},
	}
	pe.Add(e)
	return pe
}
