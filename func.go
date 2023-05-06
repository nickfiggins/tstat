package tstat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type FuncFileCov map[string]float64

type FuncCov struct {
	Total float64
	file  map[string]FuncFileCov
	fn    map[string]float64
}

func (c *FuncCov) Coverage(fn string) float64 {
	v, ok := c.fn[fn]
	if !ok {
		return 0
	}
	return v
}

func FromFuncCovFile(path string) (FuncCov, error) {
	f, err := os.Open(path)
	if err != nil {
		return FuncCov{}, err
	}
	return ReadFunc(f)
}

const numFields = 3

func ReadFunc(r io.Reader) (FuncCov, error) {
	cov := FuncCov{}
	files := map[string]FuncFileCov{}
	byFunc := map[string]float64{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		entry := strings.Fields(sc.Text())
		if len(entry) < numFields {
			continue
		}

		percent, err := strconv.ParseFloat(strings.Trim(entry[2], "%"), 64)
		if err != nil {
			return FuncCov{}, fmt.Errorf("couldn't convert percent to float %w", err)
		}

		if entry[1] == "(statements)" {
			cov.Total = percent
			continue
		}

		idx := strings.Index(entry[0], ".go")
		if idx == -1 {
			return FuncCov{}, fmt.Errorf("invalid format, no go file, line: %v", sc.Text())
		}
		file := entry[0][:idx+3]

		byFunc[entry[1]] = percent
		f, ok := files[file]
		if ok {
			f[entry[1]] = percent
			continue
		}
		files[file] = FuncFileCov{entry[1]: percent}
	}

	err := sc.Err()
	if err != nil {
		return FuncCov{}, fmt.Errorf("error while scanning: %w", err)
	}

	cov.file = files
	cov.fn = byFunc

	return cov, nil
}
