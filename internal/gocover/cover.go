package gocover

import (
	"math"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/tools/cover"
)

type PackageStatements struct {
	Package      string
	Files        map[string]*FileStatements
	Percent      float64
	Stmts        int64
	CoveredStmts int64
}

func (ps *PackageStatements) add(prof *cover.Profile) {
	fileStatements := parseProfile(prof.Blocks)

	ps.Stmts += fileStatements.Stmts
	ps.CoveredStmts += fileStatements.CoveredStmts
	ps.Percent = percent(ps.Stmts, ps.CoveredStmts)

	file, ok := ps.Files[prof.FileName]
	if !ok {
		ps.Files[prof.FileName] = fileStatements
		return
	}

	file.join(fileStatements)
}

type FileStatements struct {
	Percent      float64
	Stmts        int64
	CoveredStmts int64
}

func (fs *FileStatements) join(other *FileStatements) {
	fs.Stmts += other.Stmts
	fs.CoveredStmts += other.CoveredStmts
	fs.Percent = percent(fs.Stmts, fs.CoveredStmts)
}

func ByPackage(profiles []*cover.Profile) []*PackageStatements {
	packages := make(map[string]*PackageStatements)
	for _, prof := range profiles {
		pkg := packageFromFileName(prof.FileName)
		_, ok := packages[pkg]
		if !ok {
			fs := parseProfile(prof.Blocks)
			packages[pkg] = &PackageStatements{
				Package:      pkg,
				Files:        map[string]*FileStatements{prof.FileName: fs},
				Percent:      fs.Percent,
				Stmts:        fs.Stmts,
				CoveredStmts: fs.CoveredStmts,
			}
			continue
		}
		packages[pkg].add(prof)
	}

	return maps.Values(packages)
}

func packageFromFileName(fileName string) string {
	if i := strings.LastIndex(fileName, "/"); i != -1 {
		return fileName[:i]
	}
	return ""
}

func parseProfile(blocks []cover.ProfileBlock) *FileStatements {
	stmts, coveredStmts := int64(0), int64(0)
	for _, bk := range blocks {
		if bk.NumStmt == 0 {
			continue
		}
		stmts += int64(bk.NumStmt)
		if bk.Count > 0 {
			coveredStmts += int64(bk.NumStmt)
		}
	}

	return &FileStatements{
		Percent:      percent(stmts, coveredStmts),
		Stmts:        stmts,
		CoveredStmts: coveredStmts,
	}
}

func round(f float64) float64 {
	return math.Round(f*10) / 10
}

func percent(den, num int64) float64 {
	if den == 0 {
		return 0
	}
	return round((float64(num) / float64(den)) * 100)
}
