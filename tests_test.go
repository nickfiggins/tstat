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
					{Action: gotest.Start, Package: "pkg"},
					{Action: gotest.Out, Package: "pkg", Output: "-test.shuffle 1686798048639894000\n"},
				},
			},
			want: TestRun{
				root: "pkg",
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
					{Action: gotest.Run, Package: "pkg", Test: "TestAdd"},
					{Action: gotest.Out, Package: "pkg", Test: "TestAdd"},
					{Action: gotest.Pass, Package: "pkg", Test: "TestAdd"},
				},
			},
			want: TestRun{
				root: "pkg",
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
