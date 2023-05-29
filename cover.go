package tstat

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/nickfiggins/tstat/internal/gofunc"
	"golang.org/x/tools/cover"
)

type Coverage struct {
	Function  *FunctionStats  `json:"function,omitempty"`
	Statement *StatementStats `json:"statement,omitempty"`
}

func (c *Coverage) writeTo(w io.Writer) error {
	b, err := json.Marshal(c) // TODO: write to the writer in a better format
	if err != nil {
		return fmt.Errorf("couldn't marshal json: %w", err)
	}
	_, err = w.Write(b)
	return err
}

type StatementStats struct {
	CoverPct float64 `json:"coverPct,omitempty"`
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

func parseProfiles(profiles []*cover.Profile, opts Options) *StatementStats {
	cov := map[string]File{}
	total := 0
	covered := 0
	for _, prof := range profiles {
		fileCov := parseProfile(prof)
		file := opts.fileName(prof.FileName)
		cov[file] = fileCov
		total += fileCov.Stmts
		covered += fileCov.CoveredStmts
		fileCov.CoverPct = percent(fileCov.CoveredStmts, fileCov.Stmts)
	}

	return &StatementStats{fileCov: cov, CoverPct: percent(covered, total)}
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
	CoverPct float64 `json:"coverPct,omitempty"`
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

	sort.Slice(vals, func(i, j int) bool {
		return vals[i].Name < vals[j].Name
	})

	return vals, true
}

func parseFuncProfile(output gofunc.Output, opts Options) *FunctionStats {
	funcStats := FunctionStats{fileCov: map[string]fileFuncCov{}}
	for _, fn := range output.Funcs {
		file := opts.fileName(fn.File)
		funcStats.addFunc(Function{Name: fn.Function, File: file, CoverPct: fn.Percent, line: fn.Line})
	}

	funcStats.CoverPct = output.Percent

	return &funcStats
}
