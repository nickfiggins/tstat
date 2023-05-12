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
	CoverPct float64
	fileCov  map[string]File
}

type File struct {
	CoverPct     float64 // percent
	Stmts        int     // num statments
	CoveredStmts int
}

func (st *StatementStats) File(f string) (File, bool) {
	v, ok := st.fileCov[f]
	return v, ok
}
func parseProfiles(profiles []*cover.Profile, opts ...ParseOpts) StatementStats {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	cov := map[string]File{}
	total := 0
	covered := 0
	for _, prof := range profiles {
		fileCov := parseProfile(prof)
		file := options.fileName(prof.FileName)
		cov[file] = fileCov
		total += fileCov.Stmts
		covered += fileCov.CoveredStmts
		fileCov.CoverPct = percent(fileCov.CoveredStmts, fileCov.Stmts)
	}

	return StatementStats{fileCov: cov, CoverPct: percent(covered, total)}
}

func percent(covered, total int) float64 {
	if total == 0 {
		return -1
	}
	return 100 * float64(covered) / float64(total)
}

func parseProfile(cp *cover.Profile) File {
	stmts := 0
	coveredStmts := 0
	for _, bk := range cp.Blocks {
		stmts += bk.NumStmt
		if bk.Count > 0 {
			coveredStmts++
		}
	}

	return File{
		CoverPct:     float64(coveredStmts) / float64(stmts),
		Stmts:        stmts,
		CoveredStmts: coveredStmts,
	}
}

type Function struct {
	Name     string
	File     string
	CoverPct float64
	line     int
}

type fileFuncCov map[string]Function

type FunctionStats struct {
	CoverPct float64
	fileCov  map[string]fileFuncCov
}

func (st *FunctionStats) addFunc(fn Function) {
	file, ok := st.fileCov[fn.File]
	if ok {
		file[fn.Name] = fn
		return
	}
	st.fileCov[fn.File] = fileFuncCov{
		fn.Name: fn,
	}
}

func (st *FunctionStats) Func(file, fn string) float64 {
	v, ok := st.fileCov[file]
	if !ok {
		return -1
	}

	cov, ok := v[fn]
	if !ok {
		return -1
	}

	return cov.CoverPct
}

func (st *FunctionStats) File(f string) ([]Function, bool) {
	fileFuncs, ok := st.fileCov[f]
	if !ok {
		return nil, false
	}

	vals := make([]Function, 0, len(fileFuncs))
	for _, v := range fileFuncs {
		vals = append(vals, v)
	}

	return vals, true
}

const numFields = 3

func ParseFuncProfile(fileName string) (FunctionStats, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return FunctionStats{}, err
	}
	return ParseFuncProfileFromReader(f)
}

func ParseFuncProfileFromReader(r io.Reader, opts ...ParseOpts) (FunctionStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	funcStats := FunctionStats{fileCov: map[string]fileFuncCov{}}
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
			funcStats.CoverPct = percent
			continue
		}

		s := strings.Split(entry[0], ":")
		if len(s) < 2 {
			return FunctionStats{}, fmt.Errorf("unexpected format for filename: %v", entry[0])
		}

		file, line := s[0], s[1]

		file = options.fileName(file)

		lineInt, err := strconv.Atoi(line)
		if err != nil {
			return FunctionStats{}, fmt.Errorf("invalid line number in row %v, num '%v'", sc.Text(), line)
		}

		funcStats.addFunc(Function{Name: entry[1], File: file, CoverPct: percent, line: lineInt})
	}

	err := sc.Err()
	if err != nil {
		return FunctionStats{}, fmt.Errorf("error while scanning: %w", err)
	}

	return funcStats, nil
}
