package gocover

import (
	"io"
	"strings"

	"github.com/nickfiggins/tstat/internal/mathutil"
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
	ps.Percent = mathutil.Percent(ps.CoveredStmts, ps.Stmts)

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
	fs.Percent = mathutil.Percent(fs.CoveredStmts, fs.Stmts)
}

func ReadByPackage(r io.Reader) ([]*PackageStatements, error) {
	profiles, err := cover.ParseProfilesFromReader(r)
	if err != nil {
		return nil, err
	}
	return ByPackage(profiles), nil
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
		Percent:      mathutil.Percent(coveredStmts, stmts),
		Stmts:        stmts,
		CoveredStmts: coveredStmts,
	}
}
