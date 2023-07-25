package tstat_test

import (
	"bytes"
	_ "embed"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/nickfiggins/tstat"
)

//go:embed testdata/bigtest.json
var seed []byte

//go:embed testdata/go-cmp/cover.out
var goCmpCover []byte

//go:embed testdata/go-cmp/func.out
var goCmpFunc []byte

func FuzzTestsFromReaders(f *testing.F) {
	f.Add(string(seed))
	f.Fuzz(func(t *testing.T, testOut string) {
		defer func() {
			if err := recover(); err != nil {
				t.Fatalf("failed fuzz, panic occurred: %v", err)
			}
		}()
		_, _ = tstat.TestsFromReader(strings.NewReader(testOut))
	})
}

func FuzzPackagesTestsFromReaders(f *testing.F) {
	f.Add(seed)
	f.Fuzz(func(t *testing.T, b []byte) {
		defer func() {
			if err := recover(); err != nil {
				t.Fatalf("failed fuzz, panic occurred: %v", err)
			}
		}()
		tr, err := tstat.TestsFromReader(bytes.NewBuffer(b))
		if err != nil {
			if !reflect.DeepEqual(tr, tstat.TestRun{}) {
				t.Fatalf("failed fuzz: %v, run wasn't zero value with err != nil", err)
			}
			return
		}

		for _, pkg := range tr.Packages() {
			for _, test := range pkg.Tests {
				if test.Name == "" {
					t.Fatal("test name should be defined for non-error")
				}
				if _, ok := pkg.Test(test.Name); !ok {
					t.Fatal("test should be accessible from package")
				}
			}
		}
	})
}

func FuzzCoverFromReaders(f *testing.F) {
	if len(goCmpCover) == 0 || len(goCmpFunc) == 0 {
		f.Fatal("go-cmp coverage data not embedded")
	}
	f.Add(string(goCmpCover), string(goCmpFunc))
	f.Fuzz(func(t *testing.T, cover, fn string) {
		defer func() {
			if err := recover(); err != nil {
				t.Fatalf("failed fuzz, panic occurred: %v", err)
			}
		}()
		cov, err := tstat.CoverFromReaders(strings.NewReader(cover), strings.NewReader(fn))
		if err != nil {
			if !reflect.DeepEqual(cov, tstat.Coverage{}) {
				t.Fatalf("failed fuzz: %v, %v, cover wasn't zero value with err != nil", err, cover)
			}
			return
		}

		if len(cov.Packages) > 0 {
			for _, pkg := range cov.Packages {
				for _, file := range pkg.Files {
					if file.CoveredStmts > file.Stmts {
						t.Fatalf("covered statements should be less than total statements")
					}
					if file.Percent < 0 || file.Percent > 100 || math.IsNaN(file.Percent) {
						t.Fatalf("statement coverage should both be defined for non-error")
					}

					if file.Stmts > 0 && file.CoveredStmts == file.Stmts && file.Percent != 100 {
						t.Fatal("statement coverage should be 100%% if all statements are covered", file)
					}

					for _, fn := range file.Functions {
						if fn.Percent < 0 || fn.Percent > 100 || math.IsNaN(fn.Percent) {
							t.Fatalf("function coverage should both be defined for non-error")
						}
						if fn.File != file.Name {
							t.Fatalf("function should have file name")
						}
					}
				}
			}
		}

		if cov.Packages == nil {
			t.Fatalf("function coverage should both be defined for non-error")
		}

		if cov.Percent < 0 || cov.Percent > 100 || math.IsNaN(cov.Percent) {
			t.Fatalf("statement coverage should both be defined for non-error")
		}
	})
}
