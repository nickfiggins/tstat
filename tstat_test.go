package tstat_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/nickfiggins/tstat"
)

func Test_Cover(t *testing.T) {
	tests := []struct {
		name, covFile string
		wantPercent   float64
		wantErr       bool
		wantInitErr   bool
	}{
		{name: "happy", covFile: "prog/cover.out", wantPercent: 25},
		{name: "file not found", covFile: "cover-not-found.out", wantInitErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tstat.Cover("testdata/" + tt.covFile)
			if (err != nil) != tt.wantInitErr {
				t.Errorf("Cover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Percent != tt.wantPercent {
				t.Errorf("Cover() = got statement pct %v, wanted %v", got.Percent, tt.wantPercent)
			}
		})
	}
}

func Test_Cover_CmdError(t *testing.T) {
	t.Setenv("GOROOT", "bad go root")
	testDir := "testdata/"
	covFile := "prog/cover.out"
	_, err := tstat.Cover(filepath.Join(testDir, covFile))
	wantErr := &exec.ExitError{}
	if !errors.As(err, &wantErr) {
		t.Errorf("wanted exec exit error, got err = %v", err)
		return
	}
}

func Test_CoverFromReaders(t *testing.T) {
	testDir := "testdata/"
	tests := []struct {
		name, covFile, funcFile string
		wantPercent             float64
		wantErr                 bool
	}{
		{name: "happy", covFile: "prog/cover.out", funcFile: "prog/func.out", wantPercent: 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.Open(testDir + "/" + tt.covFile)
			f2, _ := os.Open(testDir + "/" + tt.funcFile)
			got, err := tstat.CoverFromReaders(f, f2)
			if err != nil {
				t.Fatalf("failed to create coverage parser: %v", err)
			}

			if got.Percent != tt.wantPercent {
				t.Errorf("CoverFromReaders() = got statement pct %v, wanted %v", got.Percent, tt.wantPercent)
			}

		})
	}
}

func Test_Tests(t *testing.T) {
	type wantPackage struct {
		name  string
		seed  int64
		tests int
	}

	type want struct {
		count    int
		failed   bool
		duration time.Duration
		pkgs     []wantPackage
	}
	compare := func(t *testing.T, want want, got tstat.TestRun) {
		t.Helper()
		if got.Count() != want.count {
			t.Errorf("got count %v, want %v", got.Count(), want.count)
		}
		if got.Failed() != want.failed {
			t.Errorf("got failed %v, want %v", got.Failed(), want.failed)
		}
		gotRounded := got.Duration().Round(time.Millisecond)
		if gotRounded != want.duration {
			t.Errorf("got duration %v, want %v", gotRounded, want.duration)
		}

		for _, pkg := range want.pkgs {
			gotPkg, ok := got.Package(pkg.name)
			if !ok {
				t.Errorf("package %v not found", pkg.name)
				continue
			}
			if gotPkg.Seed != pkg.seed {
				t.Errorf("got pkg %v seed %v, want %v", pkg.name, gotPkg.Seed, pkg.seed)
			}
			if gotPkg.Count() != pkg.tests {
				t.Errorf("got pkg %v, %v tests, want %v", pkg.name, gotPkg.Count(), pkg.tests)
			}
		}
	}

	tests := []struct {
		testFile string
		want     want
		wantErr  bool
	}{
		{
			testFile: "testdata/bigtest.json",
			want:     want{50, false, 473 * time.Millisecond, []wantPackage{}},
			wantErr:  false,
		},
		{
			testFile: "testdata/go-cmp/go-cmp.json",
			want: want{709, false, 1082 * time.Millisecond, []wantPackage{
				{
					name:  "github.com/google/go-cmp/cmp",
					seed:  1688261989310323000,
					tests: 301,
				},
				{
					name:  "github.com/google/go-cmp/cmp/cmpopts",
					seed:  1688261989453195000,
					tests: 151,
				},
				{
					name:  "github.com/google/go-cmp/cmp/internal/diff",
					seed:  1688261989662787000,
					tests: 238,
				},
			}},
			wantErr: false,
		},
		{
			testFile: "not-found.json",
			want:     want{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testFile, func(t *testing.T) {
			stats, err := tstat.Tests(tt.testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			compare(t, tt.want, stats)
		})
	}
}
