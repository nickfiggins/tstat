package tstat_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nickfiggins/tstat"
)

// go:embed testdata/bigtest.json
var seed []byte

func FuzzTestsFromReaders(f *testing.F) {
	f.Fuzz(func(t *testing.T, testOut string) {
		defer func() {
			if err := recover(); err != nil {
				t.Fatalf("failed fuzz, panic occurred: %v", err)
			}
		}()
		_, err := tstat.TestsFromReader(strings.NewReader(testOut))
		var se *json.SyntaxError
		var de *json.UnmarshalTypeError
		var ie *json.InvalidUnmarshalError
		if err != nil && !errors.As(err, &se) && !errors.As(err, &de) && !errors.As(err, &ie) {
			t.Fatal(err)
		}
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
				if pkg.Test(test.Name) != test {
					t.Fatal("test should be accessible from package")
				}
			}
		}
	})
}

func FuzzCoverFromReaders(f *testing.F) {
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

		if cov.Function == nil {
			t.Fatalf("function coverage should both be defined for non-error")
		}

		if cov.Statement == nil {
			t.Fatalf("statement coverage should both be defined for non-error")
		}

		if cov.Function.CoverPct < 0 || cov.Function.CoverPct > 100 {
			t.Fatalf("function coverage should be between 0 and 100, got %v", cov.Function.CoverPct)
		}
		if cov.Statement.CoverPct < 0 || cov.Statement.CoverPct > 100 {
			t.Fatalf("statement coverage should be between 0 and 100, got %v", cov.Statement.CoverPct)
		}
	})
}
