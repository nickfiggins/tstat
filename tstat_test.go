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
			cp, err := tstat.Cover(tstat.WithCoverProfile("testdata/" + tt.covFile))
			if (err != nil) != tt.wantInitErr {
				t.Errorf("Cover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantInitErr {
				return
			}

			got, err := cp.Stats()
			if (err != nil) != tt.wantErr {
				t.Errorf("Stats() error = %v, wantErr %v", err, tt.wantErr)
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
	_, err := tstat.Cover(tstat.WithCoverProfile(filepath.Join(testDir, covFile)))
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
			cp, err := tstat.Cover(tstat.WithCoverProfile(testDir+"/"+tt.covFile), func(cp *tstat.CoverageParser) error {
				f, _ := os.Open(testDir + "/" + tt.funcFile)
				cp.FuncProfile = f
				return nil
			})
			if err != nil {
				t.Fatalf("failed to create coverage parser: %v", err)
			}

			got, err := cp.Stats()
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
	tp, err := tstat.TestsFromFile("testdata/bigtest.json")
	if err != nil {
		t.Fatalf("failed to create test parser: %v", err)
	}
	stats, err := tp.Stats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if !stats.Passed() {
		t.Fatalf("Passed() returned false, wanted true")
	}

	if count := stats.Count(); count != 50 {
		t.Fatalf("wanted 50 tests, got %v", count)
	}

}
