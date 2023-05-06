package tstat

import (
	"fmt"
	"io"

	"golang.org/x/tools/cover"
)

type StatementStats struct {
	Total   float64
	fileCov map[string]FileCov
}

type FileCov struct {
	Total       float64 // percent
	TotalStmt   int     // num statments
	StmtCovered int
}

func (c *StatementStats) File(f string) FileCov {
	v, ok := c.fileCov[f]
	if ok {
		return v
	}
	return v
}

func ParseCovProfileFromFile(p string, opts ...ParseOpts) (StatementStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	profiles, err := cover.ParseProfiles(p)
	if err != nil {
		return StatementStats{}, fmt.Errorf("couldn't parse coverage from file: %w", err)
	}
	return parseProfiles(profiles, options), nil
}

func ParseCovProfile(r io.Reader, opts ...ParseOpts) (StatementStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	profiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return StatementStats{}, fmt.Errorf("couldn't parse coverage from reader: %w", err)
	}
	return parseProfiles(profiles, options), nil
}

func parseProfiles(profiles []*cover.Profile, options Options) StatementStats {
	cov := map[string]FileCov{}
	totalStmts := 0
	totalCovered := 0
	for _, prof := range profiles {
		fileCov := parseProfile(prof)
		file := options.fileName(prof.FileName)
		cov[file] = fileCov
		totalStmts += fileCov.TotalStmt
		totalCovered += fileCov.StmtCovered
	}

	return StatementStats{fileCov: cov, Total: float64(totalCovered) / float64(totalStmts)}
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
		Total:       float64(coveredStmts) / float64(stmts),
		TotalStmt:   stmts,
		StmtCovered: coveredStmts,
	}
}
