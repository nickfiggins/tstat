package tstat

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFuncProfileFromReader(t *testing.T) {
	tests := []struct {
		name       string
		funcReader io.Reader
		want       FunctionStats
		wantErr    bool
	}{
		{
			name: "happy",
			funcReader: strings.NewReader(
				`github.com/nickfiggins/tstat/testdata/prog/prog.go:3:	add		100.0%
				 github.com/nickfiggins/tstat/testdata/prog/prog.go:7:	isOdd		0.0%
				 total:							(statements)	25.0%`),
			want: FunctionStats{
				CoverPct: 25,
				fileCov: map[string]fileFuncCov{
					"github.com/nickfiggins/tstat/testdata/prog/prog.go": {
						"add": Function{
							Name: "add", File: "github.com/nickfiggins/tstat/testdata/prog/prog.go", CoverPct: 100, line: 3,
						},
						"isOdd": Function{
							Name: "isOdd", File: "github.com/nickfiggins/tstat/testdata/prog/prog.go", CoverPct: 0, line: 7,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFuncProfileFromReader(tt.funcReader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFuncProfileFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
