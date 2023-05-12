package tstat_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/nickfiggins/tstat"
	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/tools/cover"
)

func mockTestParser(r io.Reader) ([]gotest.Output, error) {
	return nil, nil
}

func mockCoverParser(r io.Reader) ([]*cover.Profile, error) {
	return nil, nil
}

func Test_CoverageStats(t *testing.T) {
	testDir := "testdata/"
	tests := []struct {
		name, covFile string
		wantPercent   float64
		wantErr       bool
	}{
		{name: "happy", covFile: "prog/cover.out", wantPercent: 25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := tstat.NewParser()
			got, err := cp.CoverageStats(filepath.Join(testDir, tt.covFile))
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

func Test_TestStats(t *testing.T) {

}
