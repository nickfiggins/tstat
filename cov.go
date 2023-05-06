package tstat

import (
	"fmt"
	"io"

	"golang.org/x/tools/cover"
)

type FileCov struct {
	Total       float64 // percent
	TotalStmt   int     // num statments
	StmtCovered int
}

type Coverage struct {
	Total    float64
	coverage map[string]FileCov
}

func FromCovFile(p string) (Coverage, error) {
	profiles, err := cover.ParseProfiles(p)
	if err != nil {
		return Coverage{}, fmt.Errorf("couldn't parse coverage from file: %w", err)
	}
	return parseProfiles(profiles), nil
}

func ReadCover(r io.Reader) (Coverage, error) {
	profiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return Coverage{}, fmt.Errorf("couldn't parse coverage from reader: %w", err)
	}
	return parseProfiles(profiles), nil
}

func parseProfiles(profiles []*cover.Profile) Coverage {
	cov := map[string]FileCov{}
	totalStmts := 0
	totalCovered := 0
	for _, prof := range profiles {
		fileCov := parseProfile(prof)
		cov[prof.FileName] = fileCov
		totalStmts += fileCov.TotalStmt
		totalCovered += fileCov.StmtCovered
	}

	return Coverage{coverage: cov, Total: float64(totalCovered) / float64(totalStmts)}
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
