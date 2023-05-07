package tstat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/cover"
)

type Coverage struct {
	Function  *FunctionStats
	Statement *StatementStats
}

type StatementStats struct {
	Percent float64
	fileCov map[string]FileCov
}

type FileCov struct {
	Percent      float64 // percent
	Stmts        int     // num statments
	CoveredStmts int
}

func (c *StatementStats) File(f string) FileCov {
	v, ok := c.fileCov[f]
	if ok {
		return v
	}
	return v
}

func ParseCoverProfile(fileName string, opts ...ParseOpts) (*StatementStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	profiles, err := cover.ParseProfiles(fileName)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse coverage from file: %w", err)
	}
	return parseProfiles(profiles, options), nil
}

func ParseCoverProfileFromReader(r io.Reader, opts ...ParseOpts) (*StatementStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	profiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse coverage from reader: %w", err)
	}
	return parseProfiles(profiles, options), nil
}

func parseProfiles(profiles []*cover.Profile, options Options) *StatementStats {
	cov := map[string]FileCov{}
	totalStmts := 0
	totalCovered := 0
	for _, prof := range profiles {
		fileCov := parseProfile(prof)
		file := options.fileName(prof.FileName)
		cov[file] = fileCov
		totalStmts += fileCov.Stmts
		totalCovered += fileCov.CoveredStmts
	}

	return &StatementStats{fileCov: cov, Percent: float64(totalCovered) / float64(totalStmts)}
}

func parseProfile(cp *cover.Profile) FileCov {
	stmts := 0
	coveredStmts := 0
	for _, bk := range cp.Blocks {
		stmts += bk.NumStmt
		if bk.Count > 0 {
			coveredStmts++
		}
	}

	return FileCov{
		Percent:      float64(coveredStmts) / float64(stmts),
		Stmts:        stmts,
		CoveredStmts: coveredStmts,
	}
}

// FuncCov is a map of function names to their coverage percentage.
type FuncCov map[string]float64

type FunctionStats struct {
	Percent float64
	file    map[string]FuncCov
	fn      FuncCov
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

const numFields = 3

func ParseFuncProfile(fileName string) (*FunctionStats, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return ParseFuncProfileFromReader(f)
}

func ParseFuncProfileFromReader(r io.Reader, opts ...ParseOpts) (*FunctionStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	cov := &FunctionStats{}
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
			return nil, fmt.Errorf("couldn't convert percent to float %w", err)
		}

		if entry[1] == "(statements)" {
			cov.Percent = percent
			continue
		}

		idx := strings.Index(entry[0], ".go")
		if idx == -1 {
			return nil, fmt.Errorf("invalid format, no go file, line: %v", sc.Text())
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
		return nil, fmt.Errorf("error while scanning: %w", err)
	}

	cov.file = files
	cov.fn = byFunc

	return cov, nil
}
