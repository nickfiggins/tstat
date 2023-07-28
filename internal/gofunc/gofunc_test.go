package gofunc

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name    string
		have    io.Reader
		want    Output
		wantErr bool
	}{
		{
			name: "happy",
			have: strings.NewReader(`github.com/nickfiggins/tstat/testdata/prog/prog.go:3:	add		100.0%
				github.com/nickfiggins/tstat/testdata/prog/prog.go:7:	isOdd		0.0%
				total:							(statements)	25.0%`,
			),
			want: Output{
				Percent: 25,
				Functions: []Function{
					{
						Package:  "github.com/nickfiggins/tstat/testdata/prog",
						File:     "github.com/nickfiggins/tstat/testdata/prog/prog.go",
						Line:     3,
						Function: "add",
						Percent:  100,
					},
					{
						Package:  "github.com/nickfiggins/tstat/testdata/prog",
						File:     "github.com/nickfiggins/tstat/testdata/prog/prog.go",
						Line:     7,
						Function: "isOdd",
						Percent:  0,
					},
				},
			},
		},
		{
			name: "empty reader",
			have: strings.NewReader(""),
			want: Output{Functions: []Function{}},
		},
		{
			name: "just percent",
			have: strings.NewReader("total:							(statements)	25.0%"),
			want: Output{Functions: []Function{}, Percent: 25},
		},
		{
			name: "invalid percent",
			have: strings.NewReader(`github.com/nickfiggins/tstat/testdata/prog/prog.go:3:	add x%
			github.com/nickfiggins/tstat/testdata/prog/prog.go:7:	isOdd		0.0%
			total:							(statements)	25.0%`,
			),
			want:    Output{},
			wantErr: true,
		},
		{
			name: "missing line num",
			have: strings.NewReader(`github.com/nickfiggins/tstat/testdata/prog/prog.go:	add 0.00%
			github.com/nickfiggins/tstat/testdata/prog/prog.go:7:	isOdd		0.0%
			total:							(statements)	25.0%`,
			),
			want:    Output{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readProfile(tt.have)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestByPackage(t *testing.T) {
	type args struct {
		output Output
	}
	tests := []struct {
		name string
		args args
		want []*PackageFunctions
	}{
		{
			name: "happy",
			args: args{
				output: Output{
					Functions: []Function{newFunc("pkg", "prog.go", "foo"), newFunc("pkg", "prog.go", "foo2")},
				},
			},
			want: []*PackageFunctions{
				{
					Package: "pkg",
					Files: map[string]*FileFunctions{
						"pkg/prog.go": {
							File:      "pkg/prog.go",
							Functions: []Function{newFunc("pkg", "prog.go", "foo"), newFunc("pkg", "prog.go", "foo2")},
						},
					},
				},
			},
		},
		{
			name: "happy; multiple",
			args: args{
				output: Output{
					Functions: []Function{newFunc("pkg", "prog.go", "foo"), newFunc("pkg", "prog2.go", "foo2")},
				},
			},
			want: []*PackageFunctions{
				{
					Package: "pkg",
					Files: map[string]*FileFunctions{
						"pkg/prog.go": {
							File:      "pkg/prog.go",
							Functions: []Function{newFunc("pkg", "prog.go", "foo")},
						},
						"pkg/prog2.go": {
							File:      "pkg/prog2.go",
							Functions: []Function{newFunc("pkg", "prog2.go", "foo2")},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ByPackage(tt.args.output); !cmp.Equal(got, tt.want) {
				t.Errorf("ByPackage() diff = %v ", cmp.Diff(tt.want, got))
			}
		})
	}
}

func newFunc(pkg, file, fn string) Function {
	return Function{
		Package:  pkg,
		File:     pkg + "/" + file,
		Line:     3,
		Percent:  100,
		Function: fn,
	}
}
