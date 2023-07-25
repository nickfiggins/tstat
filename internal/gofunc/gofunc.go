package gofunc

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

type Output struct {
	Funcs   []Function
	Percent float64
}

type Function struct {
	Package  string
	File     string
	Line     int
	Function string
	Percent  float64
}

const numFields = 3

func Read(r io.Reader) (Output, error) {
	funcs := make([]Function, 0)
	totalPercent := float64(0)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		entry := strings.Fields(sc.Text())
		if len(entry) < numFields {
			continue
		}

		percent, err := strconv.ParseFloat(strings.Trim(entry[2], "%"), 64)
		if err != nil {
			return Output{}, fmt.Errorf("couldn't convert percent to float %w", err)
		}

		if entry[1] == "(statements)" {
			totalPercent = percent
			break
		}

		s := strings.Split(entry[0], ":")
		if len(s) < 2 {
			return Output{}, fmt.Errorf("unexpected format for filename: %v", entry[0])
		}

		file, line := s[0], s[1]

		lineInt, err := strconv.Atoi(line)
		if err != nil {
			return Output{}, fmt.Errorf("invalid line number in row %v, num '%v'", sc.Text(), line)
		}

		var pkg string
		if i := strings.LastIndex(file, "/"); i != -1 {
			pkg = file[:i]
		}

		funcs = append(funcs, Function{
			Package:  pkg,
			File:     file,
			Line:     lineInt,
			Function: entry[1],
			Percent:  percent,
		})
	}

	err := sc.Err()
	if err != nil {
		return Output{}, fmt.Errorf("error while scanning: %w", err)
	}

	return Output{Funcs: funcs, Percent: totalPercent}, nil
}

type PackageFunctions struct {
	Package string
	Files   map[string]*FileFunctions
}

func (pf *PackageFunctions) add(fn Function) {
	file, ok := pf.Files[fn.File]
	if !ok {
		file = &FileFunctions{
			File:      fn.File,
			Functions: []Function{fn},
		}
		pf.Files[fn.File] = file
		return
	}

	file.Functions = append(file.Functions, fn)
}

type FileFunctions struct {
	File      string
	Functions []Function
}

func ByPackage(output Output) []*PackageFunctions {
	packages := make(map[string]*PackageFunctions)
	for _, function := range output.Funcs {
		if function.Package == "" {
			continue
		}
		pkg, ok := packages[function.Package]
		if !ok {
			packages[function.Package] = newPkgFunctions(function)
			continue
		}
		pkg.add(function)
	}

	return maps.Values(packages)
}

func newPkgFunctions(fn Function) *PackageFunctions {
	return &PackageFunctions{
		Package: fn.Package,
		Files: map[string]*FileFunctions{
			fn.File: {
				File:      fn.File,
				Functions: []Function{fn},
			},
		},
	}
}
