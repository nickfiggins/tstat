package tstat

import (
	"testing"
	"time"

	"github.com/nickfiggins/tstat/internal/gotest"
	"github.com/stretchr/testify/assert"
)

func Test_eventConverter_convert(t *testing.T) {
	tests := []struct {
		name    string
		have    *gotest.PackageEvents
		want    PackageRun
		wantErr bool
	}{
		{
			name: "simple happy",
			have: &gotest.PackageEvents{
				Package: "pkg",
				Events: []gotest.Event{
					{Time: time.Now(), Action: gotest.Run, Package: "pkg", Test: "Test"},
					{Time: time.Now(), Action: gotest.Fail, Package: "pkg", Test: "Test"},
				},
			},
			want: PackageRun{
				pkgName: "pkg",
				Tests: []*Test{
					{Name: "Test", Package: "pkg", FullName: "Test", actions: []gotest.Action{gotest.Run, gotest.Fail}, Subtests: []*Test{}},
				},
			},
		},
		{
			name: "empty test name",
			have: &gotest.PackageEvents{
				Package: "pkg",
				Events: []gotest.Event{
					{Time: time.Now(), Action: gotest.Run, Package: "pkg", Test: "/"},
				},
			},
			want: PackageRun{
				pkgName: "pkg",
				Tests:   []*Test{},
			},
		},
		{
			name: "test contains /, not subtest",
			have: &gotest.PackageEvents{
				Package: "pkg",
				Events: []gotest.Event{
					{Time: time.Now(), Action: gotest.Pass, Package: "pkg", Test: "Test2"},
					{Time: time.Now(), Action: gotest.Pass, Package: "pkg", Test: "Test2/sub3/sub4"},
				},
			},
			want: PackageRun{
				pkgName: "pkg",
				Tests: []*Test{
					{
						Name:     "Test2",
						Package:  "pkg",
						FullName: "Test2",
						actions:  []gotest.Action{gotest.Pass},
						Subtests: []*Test{
							{
								Name:     "sub4",
								Package:  "pkg",
								Subtests: []*Test{},
								FullName: "Test2/sub3/sub4",
								actions:  []gotest.Action{gotest.Pass},
							},
						},
					},
				},
			},
		},
		{
			name: "subtests nested",
			have: &gotest.PackageEvents{
				Package: "pkg",
				Events: []gotest.Event{
					{Time: time.Now(), Action: gotest.Pass, Package: "pkg", Test: "Test2"},
					{Time: time.Now(), Action: gotest.Run, Package: "pkg", Test: "Test2/sub"},
					{Time: time.Now(), Action: gotest.Out, Package: "pkg", Test: "Test2/sub"},
					{Time: time.Now(), Action: gotest.Pass, Package: "pkg", Test: "Test2/sub"},
					{Time: time.Now(), Action: gotest.Pass, Package: "pkg", Test: "Test2/sub/sub2"},
				},
			},
			want: PackageRun{
				pkgName: "pkg",
				Tests: []*Test{
					{
						Name:     "Test2",
						Package:  "pkg",
						FullName: "Test2",
						actions:  []gotest.Action{gotest.Pass},
						Subtests: []*Test{
							{Name: "sub", Package: "pkg", FullName: "Test2/sub", actions: []gotest.Action{gotest.Run, gotest.Out, gotest.Pass},
								Subtests: []*Test{
									{Name: "sub2", Package: "pkg", Subtests: []*Test{}, FullName: "Test2/sub/sub2", actions: []gotest.Action{gotest.Pass}},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertEvents(tt.have)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventConverter.convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
