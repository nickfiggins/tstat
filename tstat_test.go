package tstat_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/nickfiggins/tstat"
)

func Test_CoverageStats(t *testing.T) {
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
			if got.Statement.CoverPct != tt.wantPercent {
				t.Errorf("Read() = got statement pct %v, wanted %v", got.Statement.CoverPct, tt.wantPercent)
			}

			if got.Function.CoverPct != tt.wantPercent {
				t.Errorf("Read() = got function pct %v, wanted %v", got.Function.CoverPct, tt.wantPercent)
			}
		})
	}
}

func Test_CoverageStats_CmdError(t *testing.T) {
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

func Test_CoverageStatsFromReaders(t *testing.T) {
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

			if got.Statement.CoverPct != tt.wantPercent {
				t.Errorf("Read() = got statement pct %v, wanted %v", got.Statement.CoverPct, tt.wantPercent)
			}

			if got.Function.CoverPct != tt.wantPercent {
				t.Errorf("Read() = got function pct %v, wanted %v", got.Function.CoverPct, tt.wantPercent)
			}
		})
	}
}

func TestParser_TestRun(t *testing.T) {
	stats, err := tstat.Tests("testdata/bigtest.json")
	if err != nil {
		t.Fatalf("failed to create test parser: %v", err)
	}
	if stats.Failed() {
		t.Fatalf("Failed() returned true, wanted false")
	}

	if count := stats.Count(); count != 50 {
		t.Fatalf("wanted 50 tests, got %v", count)
	}

}
