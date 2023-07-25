package tstat

import (
	"math"

	"github.com/nickfiggins/tstat/internal/gocover"
	"github.com/nickfiggins/tstat/internal/gofunc"
	"golang.org/x/exp/maps"
)

type Coverage struct {
	Percent  float64 // Percent is the total percent of statements covered.
	Packages []*PackageCoverage
}

func newCoverage(coverPkgs []*gocover.PackageStatements, funcProfile gofunc.Output) *Coverage {
	fnPkgs := gofunc.ByPackage(funcProfile)
	packages := make(map[string]*PackageCoverage)
	covered, total := int64(0), int64(0)
	for _, pkg := range coverPkgs {
		packages[pkg.Package] = newPackageCoverage(pkg)
		covered += pkg.CoveredStmts
		total += pkg.Stmts
	}

	for _, pkg := range fnPkgs {
		pkgCov, ok := packages[pkg.Package]
		if !ok {
			continue
		}
		pkgCov.add(pkg)
	}
	return &Coverage{
		Percent:  percent(covered, total),
		Packages: maps.Values(packages),
	}
}

func percent(num, den int64) float64 {
	if den == 0 {
		return 0
	}
	return math.Round(float64(num)/float64(den)*1000) / 10
}

type PackageCoverage struct {
	Name  string
	Files []*FileCoverage
}

func (pc *PackageCoverage) Functions() []FunctionCoverage {
	funcs := make([]FunctionCoverage, 0)
	for _, f := range pc.Files {
		funcs = append(funcs, f.Functions...)
	}
	return funcs
}

func newPackageCoverage(stmts *gocover.PackageStatements) *PackageCoverage {
	files := make(map[string]*FileCoverage)
	for name, statements := range stmts.Files {
		files[name] = &FileCoverage{
			Name:         name,
			Functions:    make([]FunctionCoverage, 0),
			Percent:      statements.Percent,
			Stmts:        int(statements.Stmts),
			CoveredStmts: int(statements.CoveredStmts),
		}
	}

	return &PackageCoverage{
		Name:  stmts.Package,
		Files: maps.Values(files),
	}
}

func (pc *PackageCoverage) add(pkgFn *gofunc.PackageFunctions) {
	for name, file := range pkgFn.Files {
		for _, f := range pc.Files {
			if f.Name == name {
				f.Functions = toFunctions(file.Functions)
				return
			}
		}
	}
}

type FileCoverage struct {
	Name         string
	Percent      float64 // percent
	Functions    []FunctionCoverage
	Stmts        int // num statments
	CoveredStmts int
}

type FunctionCoverage struct {
	Name     string
	Percent  float64
	File     string
	Line     int
	Internal bool
}

func toFunctions(fn []gofunc.Function) []FunctionCoverage {
	fns := make([]FunctionCoverage, len(fn))
	for i, f := range fn {
		fns[i] = FunctionCoverage{
			Name:     f.Function,
			Percent:  f.Percent,
			Internal: isLower(f.Function[0]),
			File:     f.File,
			Line:     f.Line,
		}
	}
	return fns
}

func isLower(c byte) bool {
	return c >= 'a' && c <= 'z'
}
