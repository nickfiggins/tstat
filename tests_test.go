package tstat

import (
	"testing"

	"github.com/nickfiggins/tstat/internal/gotest"
	"github.com/stretchr/testify/assert"
)

func Test_parseTestOutputs(t *testing.T) {
	type args struct {
		outputs []gotest.Event
	}
	tests := []struct {
		name    string
		args    args
		want    TestRun
		wantErr bool
	}{
		{
			name: "with seed",
			args: args{
				outputs: []gotest.Event{
					{Action: "start", Package: "pkg"},
					{Action: "output", Package: "pkg", Output: "-test.shuffle 1686798048639894000\n"},
				},
			},
			want: TestRun{
				pkgs: []PackageRun{
					{
						pkgName: "pkg",
						Seed:    1686798048639894000,
						Tests:   []*Test{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with seed",
			args: args{
				outputs: []gotest.Event{
					{Action: "run", Package: "pkg", Test: "TestAdd"},
					{Action: "output", Package: "pkg", Test: "TestAdd"},
					{Action: "pass", Package: "pkg", Test: "TestAdd"},
				},
			},
			want: TestRun{
				pkgs: []PackageRun{
					{
						pkgName: "pkg",
						Tests: []*Test{
							{
								Name:     "TestAdd",
								SubName:  "TestAdd",
								Package:  "pkg",
								Subtests: []*Test{},
								actions:  []gotest.Action{gotest.Run, gotest.Out, gotest.Pass},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTestOutputs(tt.args.outputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTestOutputs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
