package tstat

import (
	"testing"

	"github.com/nickfiggins/tstat/internal/gotest"
	"github.com/stretchr/testify/assert"
)

func TestTest_addSubtests(t *testing.T) {
	type fields struct {
		Subtests []*Test
		actions  []gotest.Action
		FullName string
		Name     string
		Package  string
	}
	type args struct {
		sub *Test
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
				sub: &Test{
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
				sub: &Test{
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
				sub: &Test{
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

func TestTest_looksLikeSub(t *testing.T) {
	tests := []struct {
		name       string
		parentName string
		subName    string
		want       bool
	}{
		{
			name:       "simple",
			parentName: "TestAdd",
			subName:    "TestAdd/sub",
			want:       true,
		},
		{
			name:       "nested",
			parentName: "TestAdd",
			subName:    "TestAdd/sub/sub3",
			want:       true,
		},
		{
			name:       "not sub",
			parentName: "TestAdd",
			subName:    "TestAdd2",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Test{FullName: tt.parentName}
			if got := tr.looksLikeSub(tt.subName); got != tt.want {
				t.Errorf("Test.looksLikeSub() = %v, want %v", got, tt.want)
			}
		})
	}
}
