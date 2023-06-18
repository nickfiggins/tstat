package tstat

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

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

func parseProfiles(profiles []*cover.Profile, fileName func(string) string) *StatementStats {
	cov := map[string]File{}
	total := 0
	covered := 0
	for _, prof := range profiles {
		fileCov := parseProfile(prof)
		file := fileName(prof.FileName)
		cov[file] = fileCov
		total += fileCov.Stmts
		covered += fileCov.CoveredStmts
		fileCov.CoverPct = percent(fileCov.CoveredStmts, fileCov.Stmts)
	}

	return &StatementStats{fileCov: cov, CoverPct: percent(covered, total)}
}

func percent(covered, total int) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(covered) / float64(total)
}

func parseProfile(cp *cover.Profile) File {
	stmts, coveredStmts := 0, 0
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

type pkgFuncCov map[string]Function

type FunctionStats struct {
	CoverPct float64 `json:"coverPct,omitempty"`
	pkgCov   map[string]pkgFuncCov
}

func (st *FunctionStats) addFunc(fn Function) {
	file, ok := st.pkgCov[fn.File]
	if ok {
		file[fn.Name] = fn
		return
	}

	if i := strings.LastIndex(fn.File, "/"); i != -1 {
		pkg := fn.File[:i]
		if pc, pkgOK := st.pkgCov[pkg]; pkgOK {
			pc[fn.Name] = fn
			return
		}

		st.pkgCov[pkg] = pkgFuncCov{
			fn.Name: fn,
		}
	}
}

func (st *FunctionStats) Func(pkg, fn string) (float64, bool) {
	v, ok := st.pkgCov[pkg]
	if !ok {
		return 0, false
	}

	cov, ok := v[fn]
	if !ok {
		return 0, false
	}

	return cov.CoverPct, true
}

func (st *FunctionStats) File(f string) ([]Function, bool) {
	var fns []Function
	for _, pkg := range st.pkgCov {
		for _, fn := range pkg {
			if fn.File == f {
				fns = append(fns, fn)
			}
		}
	}

	sort.Slice(fns, func(i, j int) bool {
		return fns[i].Name < fns[j].Name
	})

	return fns, true
}

func (st *FunctionStats) Package(name string) ([]Function, bool) {
	fileFuncs, ok := st.pkgCov[name]
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

func parseFuncProfile(output gofunc.Output, fileName func(string) string) *FunctionStats {
	funcStats := FunctionStats{pkgCov: map[string]pkgFuncCov{}}
	for _, fn := range output.Funcs {
		file := fileName(fn.File)
		funcStats.addFunc(Function{Name: fn.Function, File: file, CoverPct: fn.Percent, line: fn.Line})
	}

	funcStats.CoverPct = output.Percent

	return &funcStats
}
