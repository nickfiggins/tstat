package tstat

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/tools/cover"
)

func Test_Internal_TestRunFromReader(t *testing.T) {
	tests := []struct {
		name    string
		parser  Parser
		want    TestRun
		wantErr bool
	}{
		{
			name: "happy",
			parser: Parser{
				opts: []ParseOpts{},
				testParser: func(r io.Reader) ([]gotest.Output, error) {
					return []gotest.Output{
						{
							Time:    time.Now(),
							Action:  "pass",
							Test:    "Test1",
							Package: "pkg",
						},
						{
							Time:    time.Now(),
							Action:  "pass",
							Test:    "Test2",
							Package: "pkg",
						},
						{
							Time:    time.Now(),
							Action:  "pass",
							Test:    "Test2/sub",
							Package: "pkg",
						},
					}, nil
				},
			},
			want: TestRun{
				Tests: []*Test{
					{
						Subtests: []Test{},
						Action:   "pass",
						Name:     "Test1",
						SubName:  "Test1",
						Package:  "pkg",
					},
					{
						Subtests: []Test{
							{
								Subtests: []Test{},
								Action:   "pass",
								Name:     "Test2/sub",
								SubName:  "sub",
								Package:  "pkg",
							},
						},
						Action:  "pass",
						Name:    "Test2",
						SubName: "Test2",
						Package: "pkg",
					},
				},
			},
		},
		{
			name: "error parsing tests",
			parser: Parser{
				opts: []ParseOpts{},
				testParser: func(r io.Reader) ([]gotest.Output, error) {
					return []gotest.Output{}, errors.New("error parsing tests")
				},
			},
			want:    TestRun{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := tt.parser
			got, err := cp.TestRunFromReader(strings.NewReader(""))
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got.Tests, tt.want.Tests) {
				t.Fatalf(cmp.Diff(got.Tests, tt.want.Tests))
			}
		})
	}
}

type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("err reading")
}
func Test_Internal_CoverageStatsFromReaders(t *testing.T) {
	tests := []struct {
		name                 string
		profile, funcProfile io.Reader
		parser               Parser
		want                 Coverage
		wantErr              bool
	}{
		{
			name: "happy cov",
			parser: Parser{
				opts: []ParseOpts{},
				coverParser: func(r io.Reader) ([]*cover.Profile, error) {
					return []*cover.Profile{
						{
							FileName: "prog.go",
							Mode:     "atomic",
							Blocks: []cover.ProfileBlock{
								{StartLine: 1, EndLine: 10, StartCol: 1, EndCol: 2, NumStmt: 5, Count: 2},
							},
						},
					}, nil
				},
			},
			profile:     strings.NewReader(""),
			funcProfile: strings.NewReader(""),
			want: Coverage{
				Function: &FunctionStats{fileCov: map[string]fileFuncCov{}},
				Statement: &StatementStats{
					CoverPct: 20,
					fileCov:  map[string]File{"prog.go": {CoverPct: 0.2, Stmts: 5, CoveredStmts: 1}},
				},
			},
		},
		{
			name: "happy cov, trim module",
			parser: Parser{
				opts: []ParseOpts{TrimModule("github.com/mod/")},
				coverParser: func(r io.Reader) ([]*cover.Profile, error) {
					return []*cover.Profile{
						{
							FileName: "github.com/mod/prog.go",
							Mode:     "atomic",
							Blocks: []cover.ProfileBlock{
								{StartLine: 1, EndLine: 10, StartCol: 1, EndCol: 2, NumStmt: 5, Count: 2},
							},
						},
					}, nil
				},
			},
			profile:     strings.NewReader(""),
			funcProfile: strings.NewReader(""),
			want: Coverage{
				Function: &FunctionStats{fileCov: map[string]fileFuncCov{}},
				Statement: &StatementStats{
					CoverPct: 20,
					fileCov:  map[string]File{"prog.go": {CoverPct: 0.2, Stmts: 5, CoveredStmts: 1}},
				},
			},
		},
		{
			name: "error reading coverage",
			parser: Parser{
				opts: []ParseOpts{},
				coverParser: func(r io.Reader) ([]*cover.Profile, error) {
					return nil, errors.New("error parsing")
				},
			},
			profile:     strings.NewReader(""),
			funcProfile: strings.NewReader(""),
			want:        Coverage{},
			wantErr:     true,
		},
		{
			name: "error reading func profile",
			parser: Parser{
				opts: []ParseOpts{},
				coverParser: func(r io.Reader) ([]*cover.Profile, error) {
					return nil, errors.New("error parsing")
				},
			},
			profile:     strings.NewReader(""),
			funcProfile: &errReader{},
			want:        Coverage{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parser.CoverageStatsFromReaders(tt.profile, tt.funcProfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v wanted %v", got, tt.want)
			}
		})
	}
}
