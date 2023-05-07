package tstat_test

import (
	"path/filepath"
	"testing"

	"github.com/nickfiggins/tstat"
)

func TestRead(t *testing.T) {
	testDir := "testdata/"
	tests := []struct {
		name    string
		covFile string
		want    *tstat.Coverage
		wantErr bool
	}{
		{
			name:    "happy",
			covFile: "prog/cover.out",
			want: &tstat.Coverage{
				Function: &tstat.FunctionStats{
					Percent: 25,
				},
				Statement: &tstat.StatementStats{
					Percent: 25,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := tstat.NewParser()
			got, err := cp.CoverageStats(filepath.Join(testDir, tt.covFile))
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Statement.Percent != tt.want.Statement.Percent {
				t.Errorf("Read() = got statement pct %v, wanted %v", got.Statement.Percent, tt.want.Statement.Percent)
			}

			if got.Function.Percent != tt.want.Function.Percent {
				t.Errorf("Read() = got function pct %v, wanted %v", got.Function.Percent, tt.want.Function.Percent)
			}
		})
	}
}
