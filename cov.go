package tstat

import (
	"fmt"
	"io"

	"golang.org/x/tools/cover"
)

type CoverStats struct {
	Percent float64
	fileCov map[string]FileCov
}

type FileCov struct {
	Percent      float64 // percent
	Stmts        int     // num statments
	CoveredStmts int
}

func (c *CoverStats) File(f string) FileCov {
	v, ok := c.fileCov[f]
	if ok {
		return v
	}
	return v
}

func ParseCoverProfile(fileName string, opts ...ParseOpts) (CoverStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	profiles, err := cover.ParseProfiles(fileName)
	if err != nil {
		return CoverStats{}, fmt.Errorf("couldn't parse coverage from file: %w", err)
	}
	return parseProfiles(profiles, options), nil
}

func ParseCoverProfileFromReader(r io.Reader, opts ...ParseOpts) (CoverStats, error) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	profiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return CoverStats{}, fmt.Errorf("couldn't parse coverage from reader: %w", err)
	}
	return parseProfiles(profiles, options), nil
}

func parseProfiles(profiles []*cover.Profile, options Options) CoverStats {
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

	return CoverStats{fileCov: cov, Percent: float64(totalCovered) / float64(totalStmts)}
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
