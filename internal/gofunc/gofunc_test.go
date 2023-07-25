package gofunc

import (
	"io"
	"strings"
	"testing"

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
				Funcs: []Function{
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
			want: Output{Funcs: []Function{}},
		},
		{
			name: "just percent",
			have: strings.NewReader("total:							(statements)	25.0%"),
			want: Output{Funcs: []Function{}, Percent: 25},
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
			got, err := Read(tt.have)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
