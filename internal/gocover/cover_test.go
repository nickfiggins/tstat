package gocover_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nickfiggins/tstat/internal/gocover"
	"golang.org/x/tools/cover"
)

func TestByPackage(t *testing.T) {
	profs, err := cover.ParseProfiles("../../testdata/go-cmp/cover.out")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		profiles []*cover.Profile
	}
	tests := []struct {
		name         string
		args         args
		wantPercents map[string]float64
	}{
		{
			name: "happy",
			args: args{profiles: profs},
			wantPercents: map[string]float64{
				"github.com/google/go-cmp/cmp":                   93.5,
				"github.com/google/go-cmp/cmp/cmpopts":           96.6,
				"github.com/google/go-cmp/cmp/internal/diff":     93.2,
				"github.com/google/go-cmp/cmp/internal/function": 51.5,
				"github.com/google/go-cmp/cmp/internal/value":    90.1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gocover.ByPackage(tt.args.profiles)
			gotPercents := make(map[string]float64, 0)
			for _, pkg := range got {
				gotPercents[pkg.Package] = pkg.Percent
			}
			if !cmp.Equal(gotPercents, tt.wantPercents) {
				t.Errorf("ByPackage() mismatch, diff (-want, +got):\n %v", cmp.Diff(tt.wantPercents, gotPercents))
			}
		})
	}
}
