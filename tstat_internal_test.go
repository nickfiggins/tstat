package tstat

import (
	"errors"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/nickfiggins/tstat/internal/gofunc"
	"github.com/nickfiggins/tstat/internal/gotest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/cover"
)

func Test_Internal_TestRunFromReader(t *testing.T) {
	fileName := t.TempDir() + "/" + "test.json"
	if _, err := os.Create(fileName); err != nil {
		t.Fatal(err)
	}
	testFile, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		parser  TestParser
		want    []PackageRun
		wantErr bool
	}{
		{
			name: "happy, nested subtests",
			parser: TestParser{
				testParser: func(r io.Reader) ([]gotest.Event, error) {
					return []gotest.Event{
						{Time: time.Now(), Action: gotest.Pass, Test: "Test1", Package: "pkg"},
						{Time: time.Now(), Action: gotest.Pass, Test: "Test2", Package: "pkg"},
						{Time: time.Now(), Action: gotest.Pass, Test: "Test2/sub", Package: "pkg"},
						{Time: time.Now(), Action: gotest.Pass, Test: "Test2/sub/sub2", Package: "pkg"},
						{Time: time.Now(), Action: gotest.Pass, Test: "Test2", Package: "pkg2"},
					}, nil
				},
			},
			want: []PackageRun{
				{
					Tests: []*Test{
						{Name: "Test1", SubName: "Test1", Package: "pkg", Subtests: []*Test{}, actions: []gotest.Action{gotest.Pass}},
						{Name: "Test2", SubName: "Test2", Package: "pkg", actions: []gotest.Action{gotest.Pass},
							Subtests: []*Test{{Subtests: []*Test{
								{Name: "Test2/sub/sub2", SubName: "sub2", Package: "pkg", Subtests: []*Test{}, actions: []gotest.Action{gotest.Pass}},
							}, Name: "Test2/sub", SubName: "sub", Package: "pkg", actions: []gotest.Action{gotest.Pass}}},
						},
					},
				},
				{
					pkgName: "pkg2",
					Tests: []*Test{
						{Name: "Test2", SubName: "Test2", Package: "pkg2", Subtests: []*Test{}, actions: []gotest.Action{gotest.Pass}},
					},
				},
			},
		},
		{
			name: "error parsing tests",
			parser: TestParser{
				testParser: func(r io.Reader) ([]gotest.Event, error) {
					return []gotest.Event{}, errors.New("error parsing tests")
				},
			},
			want:    []PackageRun{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := tt.parser
			got, err := cp.Stats(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].pkgName > tt.want[j].pkgName
			})
			sort.Slice(got.pkgs, func(i, j int) bool {
				return got.pkgs[i].pkgName > got.pkgs[j].pkgName
			})
			for i, p := range tt.want {
				assert.ElementsMatch(t, p.Tests, got.pkgs[i].Tests)
			}
		})
	}
}

func Test_Internal_CoverageStatsFromReaders(t *testing.T) {
	fileName := t.TempDir() + "/" + "test.json"
	if _, err := os.Create(fileName); err != nil {
		t.Fatal(err)
	}
	testFile, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		funcProfile io.Reader
		parser      CoverageParser
		want        Coverage
		wantErr     bool
	}{
		{
			name: "happy cov",
			parser: CoverageParser{
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
				funcParser: func(r io.Reader) (gofunc.Output, error) {
					return gofunc.Output{}, nil
				},
			},
			funcProfile: strings.NewReader(""),
			want: Coverage{
				Function: &FunctionStats{pkgCov: map[string]pkgFuncCov{}},
				Statement: &StatementStats{
					CoverPct: 20,
					fileCov:  map[string]File{"prog.go": {CoverPct: 0.2, Stmts: 5, CoveredStmts: 1}},
				},
			},
		},
		{
			name: "happy cov, trim module",
			parser: CoverageParser{
				trimModule: "github.com/mod",
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
				funcParser: func(r io.Reader) (gofunc.Output, error) {
					return gofunc.Output{}, nil
				},
			},
			funcProfile: strings.NewReader(""),
			want: Coverage{
				Function: &FunctionStats{pkgCov: map[string]pkgFuncCov{}},
				Statement: &StatementStats{
					CoverPct: 20,
					fileCov:  map[string]File{"prog.go": {CoverPct: 0.2, Stmts: 5, CoveredStmts: 1}},
				},
			},
		},
		{
			name: "error reading coverage",
			parser: CoverageParser{
				coverParser: func(r io.Reader) ([]*cover.Profile, error) {
					return nil, errors.New("error parsing")
				},
				funcParser: func(r io.Reader) (gofunc.Output, error) {
					return gofunc.Output{}, nil
				},
			},
			funcProfile: strings.NewReader(""),
			want:        Coverage{},
			wantErr:     true,
		},
		{
			name: "error reading func profile",
			parser: CoverageParser{
				coverParser: func(r io.Reader) ([]*cover.Profile, error) {
					return nil, errors.New("error parsing")
				},
				funcParser: func(r io.Reader) (gofunc.Output, error) {
					return gofunc.Output{}, nil
				},
			},
			funcProfile: &errReader{},
			want:        Coverage{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parser.Stats(testFile, tt.funcProfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("err reading")
}
