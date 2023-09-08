package tstat

import (
	"github.com/nickfiggins/tstat/internal/gocover"
	"github.com/nickfiggins/tstat/internal/gofunc"
	"github.com/nickfiggins/tstat/internal/mathutil"
	"golang.org/x/exp/maps"
)

// Coverage is the coverage statistics parsed from a single test profile.
type Coverage struct {
	Percent  float64            // Percent is the total percent of statements covered.
	Packages []*PackageCoverage // Packages is the coverage of each package.
}

// Package returns the coverage of a single package in the run. It's a convenience method
// for finding a file in the list of file coverages.
func (c *Coverage) Package(name string) (*PackageCoverage, bool) {
	for _, pkg := range c.Packages {
		if pkg.Name == name {
			return pkg, true
		}
	}
	return nil, false
}

func newCoverage(coverPkgs []*gocover.PackageStatements, funcProfile []*gofunc.PackageFunctions) *Coverage {
	packages := make(map[string]*PackageCoverage)
	covered, total := int64(0), int64(0)
	for _, pkg := range coverPkgs {
		packages[pkg.Package] = newPackageCoverage(pkg)
		covered += pkg.CoveredStmts
		total += pkg.Stmts
	}

	for _, pkg := range funcProfile {
		pkgCov, ok := packages[pkg.Package]
		if !ok {
			continue
		}
		pkgCov.add(pkg)
	}
	return &Coverage{
		Percent:  mathutil.Percent(covered, total),
		Packages: maps.Values(packages),
	}
}

// PackageCoverage is the coverage of a package.
type PackageCoverage struct {
	Name    string          // Name is the name of the package.
	Percent float64         // Percent is the percentage of statements covered in the package.
	Files   []*FileCoverage // Files is the coverage of each file in the package.
}

// File returns the coverage of a file in the package. It's a convenience method
// for finding a file in the list of file coverages.
func (pc *PackageCoverage) File(name string) (*FileCoverage, bool) {
	for _, f := range pc.Files {
		if f.Name == name {
			return f, true
		}
	}
	return nil, false
}

// Functions returns all functions in the package.
func (pc *PackageCoverage) Functions() []FunctionCoverage {
	funcs := make([]FunctionCoverage, 0)
	for _, f := range pc.Files {
		funcs = append(funcs, f.Functions...)
	}
	return funcs
}

func newPackageCoverage(stmts *gocover.PackageStatements) *PackageCoverage {
	files := make([]*FileCoverage, len(stmts.Files))
	i := 0
	for name, statements := range stmts.Files {
		files[i] = &FileCoverage{
			Name:         name,
			Functions:    make([]FunctionCoverage, 0),
			Percent:      statements.Percent,
			Stmts:        int(statements.Stmts),
			CoveredStmts: int(statements.CoveredStmts),
		}
		i++
	}

	return &PackageCoverage{
		Name:    stmts.Package,
		Percent: mathutil.Percent(stmts.CoveredStmts, stmts.Stmts),
		Files:   files,
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
	Name         string             // Name is the name of the file.
	Percent      float64            // Percent is the percent of statements covered.
	Functions    []FunctionCoverage // Functions is the coverage of each function in the file.
	Stmts        int                // Stmts is the total number of statements in the file.
	CoveredStmts int                // CoveredStmts is the number of statements covered in the file.
}

// FunctionCoverage is the coverage of a function.
type FunctionCoverage struct {
	Name     string  // Name is the name of the function.
	Percent  float64 // Percent is the percent of statements covered.
	File     string  // File is the file the function is defined in.
	Line     int     // Line is the line the function is defined on.
	Internal bool    // Internal is true if the function is internal to the package.
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
