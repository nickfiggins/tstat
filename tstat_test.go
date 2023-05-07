package tstat_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			covFile: "cover.out",
			want:    nil,
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() = %v", cmp.Diff(got, tt.want))
			}
		})
	}
}
