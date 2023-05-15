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
	testDir := "testdata/"
	tests := []struct {
		name, covFile string
		wantPercent   float64
		wantErr       bool
	}{
		{name: "happy", covFile: "prog/cover.out", wantPercent: 25},
		{name: "file not found", covFile: "cover-not-found.out", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := tstat.NewParser()
			got, err := cp.CoverageStats(filepath.Join(testDir, tt.covFile))
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
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
	testDir := "testdata/"
	covFile := "prog/cover.out"
	cp := tstat.NewParser()
	t.Setenv("GOROOT", "bad go root")
	_, err := cp.CoverageStats(filepath.Join(testDir, covFile))
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
			cp := tstat.NewParser()
			f1, _ := os.Open(filepath.Join(testDir, tt.covFile))
			f2, _ := os.Open(filepath.Join(testDir, tt.funcFile))
			got, err := cp.CoverageStatsFromReaders(f1, f2)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
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

func TestParser_TestRun(t *testing.T) {
	stats, _ := tstat.NewParser().TestRun("testdata/bigtest.json")
	if !stats.Passed() {
		t.Fatalf("Passed() returned false, wanted true")
	}

	if count := stats.Count(); count != 50 {
		t.Fatalf("wanted 50 tests, got %v", count)
	}

}
