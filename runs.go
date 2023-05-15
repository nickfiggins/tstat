package tstat

import (
	"strings"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
)

type TestRun struct {
	start, end time.Time
	pkgs       []PackageRun
}

func (tr *TestRun) Packages() []PackageRun {
	return tr.pkgs
}

func (tr *TestRun) Package(name string) (PackageRun, bool) {
	for _, pkg := range tr.pkgs {
		if strings.EqualFold(name, pkg.pkgName) {
			return pkg, true
		}
	}
	return PackageRun{}, false
}

func (tr *TestRun) Duration() time.Duration {
	return tr.end.Sub(tr.start)
}

func (tr *TestRun) Count() int {
	var count int
	for _, pkg := range tr.pkgs {
		count += pkg.Count()
	}
	return count
}

func (tr *TestRun) Passed() bool {
	for _, pkg := range tr.pkgs {
		if !pkg.Passed() {
			return false
		}
	}
	return true
}

type PackageRun struct {
	pkgName    string
	start, end time.Time
	Tests      []*Test
	cmdOut     string
}

func (tr *PackageRun) Duration() time.Duration {
	return tr.end.Sub(tr.start)
}

// Count returns the total number of tests, including subtests.
func (tr *PackageRun) Count() int {
	var count int
	for _, test := range tr.Tests {
		count += test.count()
	}
	return count
}

func (tr *PackageRun) Passed() bool {
	return len(withAction(tr.Tests, gotest.Fail)) == 0
}
