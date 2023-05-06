package tstat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// FuncCov is a map of function names to their coverage percentage.
type FuncCov map[string]float64

type FunctionStats struct {
	Total float64
	file  map[string]FuncCov
	fn    FuncCov
}

func (c *FunctionStats) Func(fn string) float64 {
	v, ok := c.fn[fn]
	if !ok {
		return 0
	}
	return v
}

func (c *FunctionStats) File(f string) FuncCov {
	v, ok := c.file[f]
	if !ok {
		return FuncCov{}
	}
	return v
}

func ParseFuncCoverageFile(path string) (FunctionStats, error) {
	f, err := os.Open(path)
	if err != nil {
		return FunctionStats{}, err
	}
	return ParseFuncCoverage(f)
}

const numFields = 3

func (o Options) fileName(full string) string {
	if o.trimModule == "" {
		return strings.TrimPrefix(full, o.trimModule)
	}
	return full
}

func ParseFuncCoverage(r io.Reader, opts ...ParseOpts) (FunctionStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	cov := FunctionStats{}
	files := map[string]FuncCov{}
	byFunc := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		entry := strings.Fields(sc.Text())
		if len(entry) < numFields {
			continue
		}

		percent, err := strconv.ParseFloat(strings.Trim(entry[2], "%"), 64)
		if err != nil {
			return FunctionStats{}, fmt.Errorf("couldn't convert percent to float %w", err)
		}

		if entry[1] == "(statements)" {
			cov.Total = percent
			continue
		}

		idx := strings.Index(entry[0], ".go")
		if idx == -1 {
			return FunctionStats{}, fmt.Errorf("invalid format, no go file, line: %v", sc.Text())
		}

		file := options.fileName(entry[0][:idx+3])

		byFunc[entry[1]] = percent
		f, ok := files[file]
		if ok {
			f[entry[1]] = percent
			continue
		}
		files[file] = FuncCov{entry[1]: percent}
	}

	err := sc.Err()
	if err != nil {
		return FunctionStats{}, fmt.Errorf("error while scanning: %w", err)
	}

	cov.file = files
	cov.fn = byFunc

	return cov, nil
}
