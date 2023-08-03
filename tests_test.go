package tstat

import (
	"testing"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
	"github.com/stretchr/testify/assert"
)

func Test_parseTestOutputs(t *testing.T) {
	type args struct {
		outputs []*gotest.PackageEvents
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
				outputs: []*gotest.PackageEvents{
					{
						Package: "pkg",
						Start:   nil, End: nil,
						Events: []gotest.Event{},
						Seed:   1686798048639894000,
					},
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
				outputs: []*gotest.PackageEvents{
					{
						Package: "pkg",
						Start:   nil, End: nil,
						Events: []gotest.Event{
							{
								Time:    time.Now(),
								Action:  gotest.Run,
								Package: "pkg",
								Test:    "TestAdd",
							},
							{
								Time:    time.Now(),
								Action:  gotest.Out,
								Package: "pkg",
								Test:    "TestAdd",
							},
							{
								Time:    time.Now(),
								Action:  gotest.Pass,
								Package: "pkg",
								Test:    "TestAdd",
							},
						},
					},
				},
			},
			want: TestRun{
				pkgs: []PackageRun{
					{
						pkgName: "pkg",
						Tests: []*Test{
							{
								FullName: "TestAdd",
								Name:     "TestAdd",
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

func TestTest_addSubtests(t *testing.T) {
	type fields struct {
		Subtests []*Test
		actions  []gotest.Action
		FullName string
		Name     string
		Package  string
	}
	type args struct {
		sub Test
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantSub []*Test
	}{
		{
			name: "simple",
			fields: fields{
				Subtests: []*Test{},
				actions:  []gotest.Action{},
				FullName: "TestAdd",
				Name:     "TestAdd",
				Package:  "pkg",
			},
			args: args{
				sub: Test{
					Subtests: []*Test{},
					actions:  []gotest.Action{},
					FullName: "TestAdd/sub",
					Name:     "sub",
					Package:  "pkg",
				},
			},
			wantSub: []*Test{
				{
					Subtests: []*Test{},
					actions:  []gotest.Action{},
					FullName: "TestAdd/sub",
					Name:     "sub",
					Package:  "pkg",
				},
			},
		},
		{
			name: "test with /, not a subtest",
			fields: fields{
				Subtests: []*Test{},
				actions:  []gotest.Action{},
				FullName: "TestAdd",
				Name:     "TestAdd",
				Package:  "pkg",
			},
			args: args{
				sub: Test{
					Subtests: []*Test{},
					actions:  []gotest.Action{},
					FullName: "TestAdd/sub/sub2",
					Name:     "sub2",
					Package:  "pkg",
				},
			},
			wantSub: []*Test{
				{
					Subtests: []*Test{},
					actions:  []gotest.Action{},
					FullName: "TestAdd/sub/sub2",
					Name:     "sub2",
					Package:  "pkg",
				},
			},
		},
		{
			name: "adds nested subtest",
			fields: fields{
				Subtests: []*Test{
					{
						Subtests: []*Test{},
						actions:  []gotest.Action{},
						FullName: "TestAdd/sub",
						Name:     "sub",
						Package:  "pkg",
					},
				},
				actions:  []gotest.Action{},
				FullName: "TestAdd",
				Name:     "TestAdd",
				Package:  "pkg",
			},
			args: args{
				sub: Test{
					Subtests: []*Test{},
					actions:  []gotest.Action{},
					FullName: "TestAdd/sub/sub2",
					Name:     "sub2",
					Package:  "pkg",
				},
			},
			wantSub: []*Test{
				{
					Subtests: []*Test{
						{
							Subtests: []*Test{},
							actions:  []gotest.Action{},
							FullName: "TestAdd/sub/sub2",
							Name:     "sub2",
							Package:  "pkg",
						},
					},
					actions:  []gotest.Action{},
					FullName: "TestAdd/sub",
					Name:     "sub",
					Package:  "pkg",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Test{
				Subtests: tt.fields.Subtests,
				actions:  tt.fields.actions,
				FullName: tt.fields.FullName,
				Name:     tt.fields.Name,
				Package:  tt.fields.Package,
			}
			tr.addSubtests(tt.args.sub)
			assert.Equal(t, tt.wantSub, tr.Subtests)
		})
	}
}
