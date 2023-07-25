package tstat

import (
	"strings"
	"time"
)

// TestRun represents the results of a test run, which may contain multiple packages.
type TestRun struct {
	root       string // root is the root package of the test run
	start, end time.Time
	pkgs       []PackageRun
}

func (tr *TestRun) Packages() []PackageRun {
	return tr.pkgs
}

func (tr *TestRun) Root() (PackageRun, bool) {
	return tr.Package(tr.root)
}

// PackageRun represents the results of a package test run. If the package was run with the -shuffle flag,
// the Seed field will be populated. Otherwise, it will be 0.
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

// Count returns the total number of tests, including subtests.
func (tr *TestRun) Count() int {
	var count int
	for _, pkg := range tr.pkgs {
		count += pkg.Count()
	}
	return count
}

func (tr *TestRun) Failed() bool {
	if len(tr.pkgs) == 0 {
		return false
	}
	for _, pkg := range tr.pkgs {
		if pkg.Failed() {
			return true
		}
	}
	return false
}

func (pr *PackageRun) Test(name string) (*Test, bool) {
	return findTest(name, pr.Tests...)
}

func findTest(name string, tests ...*Test) (*Test, bool) {
	for _, t := range tests {
		if strings.EqualFold(t.Name, name) {
			return t, true
		}

		if t.looksLikeSub(name) {
			if sub, ok := findTest(name, t.Subtests...); ok {
				return sub, true
			}
		}
	}
	return nil, false
}

type PackageRun struct {
	pkgName    string
	start, end time.Time
	Tests      []*Test
	Seed       int64
	failed     bool
}

func (pr *PackageRun) Duration() time.Duration {
	return pr.end.Sub(pr.start)
}

// Count returns the total number of tests, including subtests.
func (pr *PackageRun) Count() int {
	var count int
	for _, test := range pr.Tests {
		count += test.Count()
	}
	return count
}

func (pr *PackageRun) Failed() bool {
	return pr.failed
}

func (pr *PackageRun) Failures() []*Test {
	var failures []*Test
	for _, test := range pr.Tests {
		if test.Failed() {
			failures = append(failures, test)
		}
	}
	return failures
}
